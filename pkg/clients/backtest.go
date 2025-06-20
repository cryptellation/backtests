package clients

import (
	"context"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/runtime"
	"github.com/google/uuid"
	"go.temporal.io/sdk/worker"
)

// Backtest is a local representation of a backtest running on the Cryptellation API.
type Backtest struct {
	ID     uuid.UUID
	client client
}

// CreateParams contains parameters for creating a backtest.
type CreateParams struct {
	BacktestParameters backtest.Parameters
	Runner             runtime.Runnable
	Worker             worker.Worker
	TaskQueue          string
}

// Create creates a new backtest with registered workflows.
func (bt *Backtest) Create(ctx context.Context, params CreateParams) error {
	// Register workflows and get callbacks
	callbacks := runtime.RegisterRunnable(params.Worker, params.TaskQueue, params.Runner)

	// Create backtest with callbacks
	res, err := bt.client.raw.CreateBacktest(ctx, api.CreateBacktestWorkflowParams{
		BacktestParameters: params.BacktestParameters,
		Callbacks:          callbacks,
	})
	if err != nil {
		return err
	}

	// Update the backtest ID
	bt.ID = res.ID
	return nil
}

// Run starts the backtest on Cryptellation API.
func (bt *Backtest) Run(ctx context.Context) error {
	_, err := bt.client.raw.RunBacktest(ctx, api.RunBacktestWorkflowParams{
		BacktestID: bt.ID,
	})
	return err
}
