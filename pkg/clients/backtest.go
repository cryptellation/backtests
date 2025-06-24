package clients

import (
	"context"

	"github.com/cryptellation/backtests/api"
	"github.com/google/uuid"
)

// Backtest is a local representation of a backtest running on the Cryptellation API.
type Backtest struct {
	ID     uuid.UUID
	client client
}

// Run starts the backtest on Cryptellation API.
func (bt *Backtest) Run(ctx context.Context) error {
	_, err := bt.client.raw.RunBacktest(ctx, api.RunBacktestWorkflowParams{
		BacktestID: bt.ID,
	})
	return err
}
