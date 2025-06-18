package clients

import (
	"context"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/runtime/bot"
	"github.com/google/uuid"
	"go.temporal.io/sdk/worker"
)

// Backtest is a local representation of a backtest running on the Cryptellation API.
type Backtest struct {
	ID     uuid.UUID
	client client
}

// RunParams contains parameters for running a backtest.
type RunParams struct {
	Bot       bot.Bot
	Worker    worker.Worker
	TaskQueue string
}

// Run starts the backtest on Cryptellation API.
func (bt *Backtest) Run(ctx context.Context, params RunParams) error {
	_, err := bt.client.raw.RunBacktest(ctx, api.RunBacktestWorkflowParams{
		BacktestID: bt.ID,
		Callbacks:  bot.RegisterWorkflows(params.Worker, params.TaskQueue, bt.ID, params.Bot),
	})
	return err
}
