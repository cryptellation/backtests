package clients

import (
	"context"

	"github.com/cryptellation/backtests/api"
	temporalclient "go.temporal.io/sdk/client"
)

// RawClient is a client for the cryptellation backtests service with just the
// calls to the temporal workflows.
type RawClient interface {
	CreateBacktest(
		ctx context.Context,
		params api.CreateBacktestWorkflowParams,
	) (api.CreateBacktestWorkflowResults, error)
	RunBacktest(
		ctx context.Context,
		params api.RunBacktestWorkflowParams,
	) (api.RunBacktestWorkflowResults, error)
	GetBacktest(
		ctx context.Context,
		params api.GetBacktestWorkflowParams,
	) (api.GetBacktestWorkflowResults, error)
	ListBacktests(
		ctx context.Context,
		params api.ListBacktestsWorkflowParams,
	) (api.ListBacktestsWorkflowResults, error)
	SubscribeToPrice(
		ctx context.Context,
		params api.SubscribeToPriceWorkflowParams,
	) (api.SubscribeToPriceWorkflowResults, error)
}

var _ RawClient = raw{}

type raw struct {
	temporal temporalclient.Client
}

// NewRaw creates a new raw client to execute temporal workflows.
func NewRaw(cl temporalclient.Client) RawClient {
	return &raw{temporal: cl}
}

// CreateBacktest creates a new backtest workflow.
func (c raw) CreateBacktest(
	ctx context.Context,
	params api.CreateBacktestWorkflowParams,
) (api.CreateBacktestWorkflowResults, error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.CreateBacktestWorkflowName, params)
	if err != nil {
		return api.CreateBacktestWorkflowResults{}, err
	}

	// Get result and return
	var res api.CreateBacktestWorkflowResults
	err = exec.Get(ctx, &res)

	return res, err
}

// RunBacktest runs a backtest workflow.
func (c raw) RunBacktest(
	ctx context.Context,
	params api.RunBacktestWorkflowParams,
) (api.RunBacktestWorkflowResults, error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.RunBacktestWorkflowName, params)
	if err != nil {
		return api.RunBacktestWorkflowResults{}, err
	}

	// Get result and return
	var res api.RunBacktestWorkflowResults
	err = exec.Get(ctx, &res)

	return res, err
}

// SubscribeToPrice subscribes to the backtest price workflow.
func (c raw) SubscribeToPrice(
	ctx context.Context,
	params api.SubscribeToPriceWorkflowParams,
) (api.SubscribeToPriceWorkflowResults, error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.SubscribeToPriceWorkflowName, params)
	if err != nil {
		return api.SubscribeToPriceWorkflowResults{}, err
	}

	// Get result and return
	var res api.SubscribeToPriceWorkflowResults
	err = exec.Get(ctx, &res)

	return res, err
}

// ListBacktests lists backtest workflows.
func (c raw) ListBacktests(
	ctx context.Context,
	params api.ListBacktestsWorkflowParams,
) (api.ListBacktestsWorkflowResults, error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.ListBacktestsWorkflowName, params)
	if err != nil {
		return api.ListBacktestsWorkflowResults{}, err
	}

	// Get result and return
	var res api.ListBacktestsWorkflowResults
	err = exec.Get(ctx, &res)

	return res, err
}

// GetBacktest retrieves a backtest workflow.
func (c raw) GetBacktest(
	ctx context.Context,
	params api.GetBacktestWorkflowParams,
) (api.GetBacktestWorkflowResults, error) {
	workflowOptions := temporalclient.StartWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}

	// Execute workflow
	exec, err := c.temporal.ExecuteWorkflow(ctx, workflowOptions, api.GetBacktestWorkflowName, params)
	if err != nil {
		return api.GetBacktestWorkflowResults{}, err
	}

	// Get result and return
	var res api.GetBacktestWorkflowResults
	err = exec.Get(ctx, &res)

	return res, err
}
