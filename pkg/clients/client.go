package clients

import (
	"context"

	"github.com/cryptellation/backtests/api"
	temporalclient "go.temporal.io/sdk/client"
)

// Client is a client for the cryptellation backtests service.
type Client interface {
	// NewBacktest creates a new backtest.
	NewBacktest(
		ctx context.Context,
		params api.CreateBacktestWorkflowParams,
	) (Backtest, error)
	// GetBacktest gets a backtest.
	GetBacktest(
		ctx context.Context,
		params api.GetBacktestWorkflowParams,
	) (Backtest, error)
	// ListBacktests lists backtests.
	ListBacktests(
		ctx context.Context,
		params api.ListBacktestsWorkflowParams,
	) ([]Backtest, error)
	// Info calls the service info.
	Info(ctx context.Context) (api.ServiceInfoResults, error)
}

type client struct {
	temporal temporalclient.Client
	raw      RawClient
}

// New creates a new client to execute temporal workflows.
func New(cl temporalclient.Client) Client {
	return &client{
		temporal: cl,
		raw:      NewRaw(cl),
	}
}

// NewBacktest creates a new backtest.
func (c client) NewBacktest(
	ctx context.Context,
	params api.CreateBacktestWorkflowParams,
) (Backtest, error) {
	res, err := c.raw.CreateBacktest(ctx, params)
	return Backtest{
		ID:     res.ID,
		client: c,
	}, err
}

func (c client) GetBacktest(
	ctx context.Context,
	params api.GetBacktestWorkflowParams,
) (Backtest, error) {
	res, err := c.raw.GetBacktest(ctx, params)
	if err != nil {
		return Backtest{}, err
	}

	return Backtest{
		ID:     res.Backtest.ID,
		client: c,
	}, nil
}

func (c client) ListBacktests(
	ctx context.Context,
	params api.ListBacktestsWorkflowParams,
) ([]Backtest, error) {
	res, err := c.raw.ListBacktests(ctx, params)
	if err != nil {
		return nil, err
	}

	backtests := make([]Backtest, len(res.Backtests))
	for i, bt := range res.Backtests {
		backtests[i] = Backtest{
			ID:     bt.ID,
			client: c,
		}
	}

	return backtests, nil
}

// Info calls the service info.
func (c client) Info(ctx context.Context) (res api.ServiceInfoResults, err error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.ServiceInfoWorkflowName)
	if err != nil {
		return api.ServiceInfoResults{}, err
	}

	// Get result and return
	err = exec.Get(ctx, &res)
	return res, err
}
