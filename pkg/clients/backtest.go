package clients

import (
	"context"
	"fmt"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/runtime/bot"
	"github.com/google/uuid"
	"go.temporal.io/sdk/worker"
)

// Backtest is a local representation of a backtest running on the Cryptellation API.
type Backtest struct {
	ID     uuid.UUID
	client client
}

// Run starts the backtest on Cryptellation API.
func (bt *Backtest) Run(ctx context.Context, b bot.Bot) error {
	// TODO(#4): get worker from parameters instead of creating a new one

	// Create temporary worker
	tq := fmt.Sprintf("%s-%s", runtime.ModeBacktest.String(), bt.ID.String())
	w := worker.New(bt.client.temporal, tq, worker.Options{})

	// Register workflows
	cbs := bot.RegisterWorkflows(w, tq, bt.ID, b)

	// Start worker
	go func() {
		if err := w.Run(nil); err != nil {
			panic(err)
		}
	}()
	defer w.Stop()

	_, err := bt.client.raw.RunBacktest(ctx, api.RunBacktestWorkflowParams{
		BacktestID: bt.ID,
		Callbacks:  cbs,
	})
	return err
}
