package svc

import (
	"fmt"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/svc/db"
	"go.temporal.io/sdk/workflow"
)

func (wf *workflows) SubscribeToPriceWorkflow(
	ctx workflow.Context,
	params api.SubscribeToPriceWorkflowParams,
) (api.SubscribeToPriceWorkflowResults, error) {
	logger := workflow.GetLogger(ctx)

	// Read backtest
	bt, err := wf.readBacktestFromDB(ctx, params.BacktestID)
	if err != nil {
		return api.SubscribeToPriceWorkflowResults{}, fmt.Errorf("read backtest from db: %w", err)
	}

	// Add subscription
	if _, err := bt.CreateTickSubscription(params.Exchange, params.Pair); err != nil {
		return api.SubscribeToPriceWorkflowResults{}, fmt.Errorf("cannot create subscription: %w", err)
	}
	logger.Debug("Subscribed to price",
		"exchange", params.Exchange,
		"pair", params.Pair,
		"backtest_id", bt.ID.String())

	// Save backtest
	var writeRes db.UpdateBacktestActivityResults
	err = workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		wf.db.UpdateBacktestActivity, db.UpdateBacktestActivityParams{
			Backtest: bt,
		}).Get(ctx, &writeRes)
	if err != nil {
		return api.SubscribeToPriceWorkflowResults{}, fmt.Errorf("save backtest to db: %w", err)
	}

	return api.SubscribeToPriceWorkflowResults{}, nil
}
