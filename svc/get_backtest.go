package svc

import (
	"fmt"

	"github.com/cryptellation/backtests/api"
	"go.temporal.io/sdk/workflow"
)

func (wf *workflows) GetBacktestWorkflow(
	ctx workflow.Context,
	params api.GetBacktestWorkflowParams,
) (api.GetBacktestWorkflowResults, error) {
	// Read backtest
	bt, err := wf.readBacktestFromDB(ctx, params.BacktestID)
	if err != nil {
		return api.GetBacktestWorkflowResults{}, fmt.Errorf("read backtest from db: %w", err)
	}

	return api.GetBacktestWorkflowResults{
		Backtest: bt,
	}, nil
}
