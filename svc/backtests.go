package svc

import (
	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/svc/db"
	"github.com/cryptellation/candlesticks/pkg/clients"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// Backtests is the interface for the backtests domain.
type Backtests interface {
	Register(w worker.Worker)

	// Backtests

	CreateBacktestWorkflow(
		ctx workflow.Context,
		params api.CreateBacktestWorkflowParams,
	) (api.CreateBacktestWorkflowResults, error)
	GetBacktestWorkflow(
		ctx workflow.Context,
		params api.GetBacktestWorkflowParams,
	) (api.GetBacktestWorkflowResults, error)
	ListBacktestsWorkflow(
		ctx workflow.Context,
		params api.ListBacktestsWorkflowParams,
	) (api.ListBacktestsWorkflowResults, error)
	RunBacktestWorkflow(
		ctx workflow.Context,
		params api.RunBacktestWorkflowParams,
	) (api.RunBacktestWorkflowResults, error)
	SubscribeToPriceWorkflow(
		ctx workflow.Context,
		params api.SubscribeToPriceWorkflowParams,
	) (api.SubscribeToPriceWorkflowResults, error)

	// Backtests Accounts

	GetBacktestAccountsWorkflow(
		ctx workflow.Context,
		params api.GetBacktestAccountsWorkflowParams,
	) (api.GetBacktestAccountsWorkflowResults, error)

	// Backtests Orders

	CreateBacktestOrderWorkflow(
		ctx workflow.Context,
		params api.CreateBacktestOrderWorkflowParams,
	) (api.CreateBacktestOrderWorkflowResults, error)

	GetBacktestOrdersWorkflow(
		ctx workflow.Context,
		params api.GetBacktestOrdersWorkflowParams,
	) (api.GetBacktestOrdersWorkflowResults, error)
}

// Check that the workflows implements the Backtests interface.
var _ Backtests = &workflows{}

type workflows struct {
	db            db.DB
	cryptellation clients.WfClient
}

// New creates a new backtests workflows.
func New(db db.DB) Backtests {
	return &workflows{
		cryptellation: clients.NewWfClient(),
		db:            db,
	}
}

// Register registers the candlesticks workflows to the worker.
func (wf *workflows) Register(w worker.Worker) {
	w.RegisterWorkflowWithOptions(wf.CreateBacktestOrderWorkflow, workflow.RegisterOptions{
		Name: api.CreateBacktestOrderWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.CreateBacktestWorkflow, workflow.RegisterOptions{
		Name: api.CreateBacktestWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.GetBacktestAccountsWorkflow, workflow.RegisterOptions{
		Name: api.GetBacktestAccountsWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.GetBacktestOrdersWorkflow, workflow.RegisterOptions{
		Name: api.GetBacktestOrdersWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.GetBacktestWorkflow, workflow.RegisterOptions{
		Name: api.GetBacktestWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.ListBacktestsWorkflow, workflow.RegisterOptions{
		Name: api.ListBacktestsWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.RunBacktestWorkflow, workflow.RegisterOptions{
		Name: api.RunBacktestWorkflowName,
	})
	w.RegisterWorkflowWithOptions(wf.SubscribeToPriceWorkflow, workflow.RegisterOptions{
		Name: api.SubscribeToPriceWorkflowName,
	})

	w.RegisterWorkflowWithOptions(ServiceInfoWorkflow, workflow.RegisterOptions{
		Name: api.ServiceInfoWorkflowName,
	})
}
