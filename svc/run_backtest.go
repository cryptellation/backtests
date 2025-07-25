package svc

import (
	"fmt"
	"time"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/backtests/svc/db"
	candlesticksapi "github.com/cryptellation/candlesticks/api"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/ticks/pkg/tick"
	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

func (wf *workflows) RunBacktestWorkflow(
	ctx workflow.Context,
	params api.RunBacktestWorkflowParams,
) (api.RunBacktestWorkflowResults, error) {
	// Load backtest from database to get callbacks
	bt, err := wf.readBacktestFromDB(ctx, params.BacktestID)
	if err != nil {
		return api.RunBacktestWorkflowResults{}, fmt.Errorf("loading backtest from database: %w", err)
	}

	// Init the backtest from client side
	bt, err = wf.execOnInitBacktestCallback(ctx, bt.Callbacks.OnInitCallback, params.BacktestID)
	if err != nil {
		return api.RunBacktestWorkflowResults{}, fmt.Errorf("initializing backtest from client side: %w", err)
	}

	// Loop on backtest events
	if err := wf.loopThroughBacktestEvents(ctx, bt, bt.Callbacks); err != nil {
		return api.RunBacktestWorkflowResults{}, fmt.Errorf("looping through backtest events: %w", err)
	}

	// Exit the backtest from client side
	if err := wf.execOnExitBacktestCallback(ctx, bt.Callbacks.OnExitCallback, bt.ID); err != nil {
		return api.RunBacktestWorkflowResults{}, fmt.Errorf("exit backtest from client side: %w", err)
	}

	return api.RunBacktestWorkflowResults{}, nil
}

func (wf *workflows) loopThroughBacktestEvents(
	ctx workflow.Context,
	bt backtest.Backtest,
	callbacks runtime.Callbacks,
) error {
	logger := workflow.GetLogger(ctx)

	for finished := false; !finished; {
		logger.Debug("Looping over prices",
			"backtest_id", bt.ID.String(),
			"current_time", bt.CurrentTime())

		// Get prices
		prices, err := wf.readActualPrices(ctx, bt)
		if err != nil {
			return fmt.Errorf("cannot read actual prices: %w", err)
		}
		if len(prices) == 0 {
			logger.Warn("No price detected",
				"time", bt.CurrentCandlestick.Time)
			bt.SetCurrentTime(bt.EndTime)
			break
		} else if !prices[0].Time.Equal(bt.CurrentCandlestick.Time) {
			logger.Warn("No price between current time and first event retrieved",
				"current_time", bt.CurrentCandlestick.Time,
				"first_event_time", prices[0].Time)
			bt.SetCurrentTime(prices[0].Time)
		}

		// Execute backtest with these prices
		if err := execOnPriceBacktest(ctx, callbacks.OnNewPricesCallback, prices, bt.ID); err != nil {
			return fmt.Errorf("cannot execute backtest: %w", err)
		}

		// Advance backtest
		finished, bt, err = wf.advanceBacktest(ctx, bt.ID)
		if err != nil {
			return fmt.Errorf("cannot advance backtest: %w", err)
		}
	}

	return nil
}

func (wf *workflows) execOnInitBacktestCallback(
	ctx workflow.Context,
	onInitCallback runtime.CallbackWorkflow,
	backtestID uuid.UUID,
) (backtest.Backtest, error) {
	// Load backtest
	bt, err := wf.readBacktestFromDB(ctx, backtestID)
	if err != nil {
		return backtest.Backtest{}, fmt.Errorf("load backtest from db: %w", err)
	}

	// Options
	opts := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("backtest-%s-on-init", backtestID.String()),
		TaskQueue:                onInitCallback.TaskQueueName, // Execute in the client queue
		WorkflowExecutionTimeout: time.Second * 30,             // Timeout if the child workflow does not complete
	}

	// Check if the timeout is set
	if onInitCallback.ExecutionTimeout > 0 {
		opts.WorkflowExecutionTimeout = onInitCallback.ExecutionTimeout
	}

	// Run a new child workflow
	ctx = workflow.WithChildOptions(ctx, opts)
	if err := workflow.ExecuteChildWorkflow(ctx, onInitCallback.Name, runtime.OnInitCallbackWorkflowParams{
		Context: runtime.Context{
			ID:              backtestID,
			Mode:            runtime.ModeBacktest,
			Now:             bt.StartTime,
			ParentTaskQueue: workflow.GetInfo(ctx).TaskQueueName,
		},
	}).Get(ctx, nil); err != nil {
		return backtest.Backtest{}, fmt.Errorf("starting new onInitCallback child workflow: %w", err)
	}

	// Reload backtest in case of modifications
	bt, err = wf.readBacktestFromDB(ctx, backtestID)
	if err != nil {
		return backtest.Backtest{}, fmt.Errorf("reload backtest from db: %w", err)
	}

	return bt, nil
}

func (wf *workflows) advanceBacktest(ctx workflow.Context, id uuid.UUID) (bool, backtest.Backtest, error) {
	logger := workflow.GetLogger(ctx)

	// Read backtest
	bt, err := wf.readBacktestFromDB(ctx, id)
	if err != nil {
		return false, backtest.Backtest{}, fmt.Errorf("load backtest from db: %w", err)
	}

	// Advance backtest
	finished, err := bt.Advance()
	if err != nil {
		return false, backtest.Backtest{}, fmt.Errorf("cannot advance backtest: %w", err)
	}
	logger.Info("Advancing backtest",
		"id", bt.ID.String(),
		"current_time", bt.CurrentTime())

	// Save backtest
	var writeRes db.UpdateBacktestActivityResults
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		wf.db.UpdateBacktestActivity, db.UpdateBacktestActivityParams{
			Backtest: bt,
		}).Get(ctx, &writeRes)
	if err != nil {
		return false, backtest.Backtest{}, fmt.Errorf("save backtest to db: %w", err)
	}

	return finished, bt, nil
}

func (wf *workflows) readActualPrices(ctx workflow.Context, bt backtest.Backtest) ([]tick.Tick, error) {
	logger := workflow.GetLogger(ctx)
	logger.Debug("Reading actual prices",
		"backtest_id", bt.ID.String())

	// Run for all prices subscriptions
	// TODO(#5): parallelize the read for each subscription
	prices := make([]tick.Tick, 0, len(bt.PricesSubscriptions))
	for _, sub := range bt.PricesSubscriptions {
		logger.Debug("Reading actual prices for subscription",
			"exchange", sub.Exchange,
			"pair", sub.Pair)

		// Get exchange info
		result, err := wf.cryptellation.ListCandlesticks(ctx, candlesticksapi.ListCandlesticksWorkflowParams{
			Exchange: sub.Exchange,
			Pair:     sub.Pair,
			Period:   bt.PricePeriod,
			Start:    &bt.CurrentCandlestick.Time,
			End:      &bt.EndTime,
			Limit:    1,
		}, &workflow.ChildWorkflowOptions{
			TaskQueue: candlesticksapi.WorkerTaskQueueName,
		})
		if err != nil {
			return nil, fmt.Errorf("could not get candlesticks from service: %w", err)
		}

		// Get the first candlestick if possible
		if len(result.List) == 0 {
			continue
		}
		cs := result.List[0]
		t := cs.Time

		// Create tick from candlesticks
		p := tick.FromCandlestick(sub.Exchange, sub.Pair, bt.CurrentCandlestick.Price, t, cs)
		prices = append(prices, p)
	}

	// Only keep the earliest same time ticks for time consistency
	_, prices = tick.OnlyKeepEarliestSameTime(prices, bt.EndTime)
	logger.Info("Gotten ticks on backtest",
		"quantity", len(prices),
		"backtest_id", bt.ID.String())
	return prices, nil
}

func execOnPriceBacktest(
	ctx workflow.Context,
	callback runtime.CallbackWorkflow,
	prices []tick.Tick,
	backtestID uuid.UUID,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Debug("Executing backtest callback for new prices",
		"callback", callback.Name,
		"prices", prices)

	// Options
	opts := workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("backtest-%s-on-new-prices-%s",
			backtestID.String(), prices[0].Time.Format(time.RFC3339)),
		TaskQueue:                callback.TaskQueueName, // Execute in the client queue
		WorkflowExecutionTimeout: time.Second * 30,       // Timeout if the child workflow does not complete
	}

	// Check if the timeout is set
	if callback.ExecutionTimeout > 0 {
		opts.WorkflowExecutionTimeout = callback.ExecutionTimeout
	}

	// Execute backtest
	err := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, opts),
		callback.Name, runtime.OnNewPricesCallbackWorkflowParams{
			Context: runtime.Context{
				ID:              backtestID,
				Mode:            runtime.ModeBacktest,
				Now:             prices[0].Time,
				ParentTaskQueue: workflow.GetInfo(ctx).TaskQueueName,
			},
			Ticks: prices,
		}).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (wf *workflows) execOnExitBacktestCallback(
	ctx workflow.Context,
	onExitCallback runtime.CallbackWorkflow,
	backtestID uuid.UUID,
) error {
	// Load backtest
	bt, err := wf.readBacktestFromDB(ctx, backtestID)
	if err != nil {
		return fmt.Errorf("load backtest from db: %w", err)
	}

	// Options
	opts := workflow.ChildWorkflowOptions{
		WorkflowID:               fmt.Sprintf("backtest-%s-on-exit", backtestID.String()),
		TaskQueue:                onExitCallback.TaskQueueName, // Execute in the client queue
		WorkflowExecutionTimeout: time.Second * 30,             // Timeout if the child workflow does not complete
	}

	// Check if the timeout is set
	if onExitCallback.ExecutionTimeout > 0 {
		opts.WorkflowExecutionTimeout = onExitCallback.ExecutionTimeout
	}

	// Run a new child workflow
	ctx = workflow.WithChildOptions(ctx, opts)
	if err := workflow.ExecuteChildWorkflow(
		ctx, onExitCallback.Name, runtime.OnExitCallbackWorkflowParams{
			Context: runtime.Context{
				ID:              backtestID,
				Mode:            runtime.ModeBacktest,
				Now:             bt.EndTime,
				ParentTaskQueue: workflow.GetInfo(ctx).TaskQueueName,
			},
		},
	).Get(ctx, nil); err != nil {
		return fmt.Errorf("starting new onExitCallback child workflow: %w", err)
	}

	return nil
}
