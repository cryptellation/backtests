package clients

import (
	"github.com/cryptellation/backtests/api"
	"go.temporal.io/sdk/workflow"
)

// WfClient is a client for the cryptellation backtests service from a workflow perspective.
type WfClient interface {
	// SubscribeToPrice subscribes to specific price updates.
	SubscribeToPrice(
		ctx workflow.Context,
		params api.SubscribeToPriceWorkflowParams,
	) (api.SubscribeToPriceWorkflowResults, error)
}

type wfClient struct{}

// NewWfClient creates a new workflow client.
// This client is used to call workflows from within other workflows.
// It is not used to call workflows from outside the workflow environment.
func NewWfClient() WfClient {
	return wfClient{}
}

// SubscribeToPrice subscribes to the backtest price.
func (c wfClient) SubscribeToPrice(
	ctx workflow.Context,
	params api.SubscribeToPriceWorkflowParams,
) (api.SubscribeToPriceWorkflowResults, error) {
	// Set options
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		TaskQueue: api.WorkerTaskQueueName,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOptions)

	// Execute child workflow
	var res api.SubscribeToPriceWorkflowResults
	err := workflow.ExecuteChildWorkflow(ctx, api.SubscribeToPriceWorkflowName, params).Get(ctx, &res)
	if err != nil {
		return api.SubscribeToPriceWorkflowResults{}, err
	}

	return res, nil
}
