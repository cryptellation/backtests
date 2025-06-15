package svc

import (
	"fmt"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/backtests/svc/db"
	candlesticksapi "github.com/cryptellation/candlesticks/api"
	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// CreateBacktestOrderWorkflow creates an order on a backtest.
func (wf *workflows) CreateBacktestOrderWorkflow(
	ctx workflow.Context,
	params api.CreateBacktestOrderWorkflowParams,
) (api.CreateBacktestOrderWorkflowResults, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Creating order on backtest",
		"backtest_id", params.BacktestID.String(),
		"order", params.Order)

	// Create a new ID if not provided
	if params.Order.ID == uuid.Nil {
		params.Order.ID = uuid.New()
	}

	// Read backtest and candlesticks
	bt, cs, err := wf.getBacktestAndCandlestick(ctx, params)
	if err != nil {
		return api.CreateBacktestOrderWorkflowResults{}, fmt.Errorf("could not read backtest and candlesticks: %w", err)
	}

	// Add order to backtest
	logger.Info("Adding order to backtest",
		"order", params.Order,
		"backtest_id", params.BacktestID.String())
	if err := bt.AddOrder(params.Order, cs); err != nil {
		return api.CreateBacktestOrderWorkflowResults{}, err
	}

	// Save backtest
	if err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		db.UpdateBacktestActivityName, db.UpdateBacktestActivityParams{
			Backtest: bt,
		}).Get(ctx, nil); err != nil {
		return api.CreateBacktestOrderWorkflowResults{}, fmt.Errorf("could not save backtest to service: %w", err)
	}

	return api.CreateBacktestOrderWorkflowResults{}, nil
}

func (wf *workflows) getBacktestAndCandlestick(
	ctx workflow.Context,
	params api.CreateBacktestOrderWorkflowParams,
) (backtest.Backtest, candlestick.Candlestick, error) {
	// Get backtest
	var dbBtRes db.ReadBacktestActivityResults
	if err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		db.ReadBacktestActivityName, db.ReadBacktestActivityParams{
			ID: params.BacktestID,
		}).Get(ctx, &dbBtRes); err != nil {
		return backtest.Backtest{}, candlestick.Candlestick{}, fmt.Errorf("could not get backtest from service: %w", err)
	}

	// Get candlestick for the time
	csRes, err := wf.cryptellation.ListCandlesticks(ctx, candlesticksapi.ListCandlesticksWorkflowParams{
		Exchange: params.Order.Exchange,
		Pair:     params.Order.Pair,
		Period:   dbBtRes.Backtest.Parameters.PricePeriod,
		Start:    &dbBtRes.Backtest.CurrentCandlestick.Time,
		End:      &dbBtRes.Backtest.CurrentCandlestick.Time,
		Limit:    0,
	}, nil)
	if err != nil {
		return backtest.Backtest{}, candlestick.Candlestick{},
			fmt.Errorf("could not get candlesticks from service: %w", err)
	}

	// Check if we have a candlestick
	if len(csRes.List) == 0 {
		return backtest.Backtest{}, candlestick.Candlestick{},
			fmt.Errorf("%w: %d candlesticks retrieved", backtest.ErrNoDataForOrderValidation, len(csRes.List))
	}
	cs := csRes.List[0]

	return dbBtRes.Backtest, cs, nil
}
