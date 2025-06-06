package svc

import (
	"fmt"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/svc/db"
	"go.temporal.io/sdk/workflow"
)

func (wf *workflows) ListBacktestsWorkflow(
	ctx workflow.Context,
	_ api.ListBacktestsWorkflowParams,
) (api.ListBacktestsWorkflowResults, error) {
	// Execute activity for listing backtests
	var dbRes db.ListBacktestsActivityResults
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		wf.db.ListBacktestsActivity, db.ListBacktestsActivityParams{}).Get(ctx, &dbRes)
	if err != nil {
		return api.ListBacktestsWorkflowResults{}, fmt.Errorf("adding backtest to db: %w", err)
	}

	return api.ListBacktestsWorkflowResults{
		Backtests: dbRes.Backtests,
	}, nil
}
