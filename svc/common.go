package svc

import (
	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/backtests/svc/db"
	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

func (wf *workflows) readBacktestFromDB(ctx workflow.Context, id uuid.UUID) (backtest.Backtest, error) {
	var readRes db.ReadBacktestActivityResults
	err := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, db.DefaultActivityOptions()),
		wf.db.ReadBacktestActivity, db.ReadBacktestActivityParams{
			ID: id,
		}).Get(ctx, &readRes)
	if err != nil {
		return backtest.Backtest{}, err
	}

	return readRes.Backtest, nil
}
