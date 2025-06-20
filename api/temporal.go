package api

import (
	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/runtime/account"
	"github.com/cryptellation/runtime/order"
	"github.com/google/uuid"
)

const (
	// WorkerTaskQueueName is the name of the task queue for the cryptellation worker.
	WorkerTaskQueueName = "CryptellationbacktestsTaskQueue"
)

// CreateBacktestWorkflowName is the name of the workflow to create a backtest.
const CreateBacktestWorkflowName = "CreateBacktestWorkflow"

type (
	// CreateBacktestWorkflowParams is the parameters of the CreateBacktestWorkflow workflow.
	CreateBacktestWorkflowParams struct {
		BacktestParameters backtest.Parameters
		Callbacks          runtime.Callbacks
	}

	// CreateBacktestWorkflowResults is the results of the CreateBacktestWorkflow workflow.
	CreateBacktestWorkflowResults struct {
		ID uuid.UUID
	}
)

// RunBacktestWorkflowName is the name of the workflow to run a backtest.
const RunBacktestWorkflowName = "RunBacktestWorkflow"

type (
	// RunBacktestWorkflowParams is the parameters of the RunBacktestWorkflow workflow.
	RunBacktestWorkflowParams struct {
		BacktestID uuid.UUID
	}

	// RunBacktestWorkflowResults is the results of the RunBacktestWorkflow workflow.
	RunBacktestWorkflowResults struct{}
)

// GetBacktestWorkflowName is the name of the workflow to get a backtest.
const GetBacktestWorkflowName = "GetBacktestWorkflow"

type (
	// GetBacktestWorkflowParams is the parameters of the GetBacktestWorkflow workflow.
	GetBacktestWorkflowParams struct {
		BacktestID uuid.UUID
	}

	// GetBacktestWorkflowResults is the results of the GetBacktestWorkflow workflow.
	GetBacktestWorkflowResults struct {
		Backtest backtest.Backtest
	}
)

// ListBacktestsWorkflowName is the name of the workflow to list backtests.
const ListBacktestsWorkflowName = "ListBacktestsWorkflow"

type (
	// ListBacktestsWorkflowParams is the parameters of the ListBacktestsWorkflow workflow.
	ListBacktestsWorkflowParams struct{}

	// ListBacktestsWorkflowResults is the results of the ListBacktestsWorkflow workflow.
	ListBacktestsWorkflowResults struct {
		Backtests []backtest.Backtest
	}
)

// GetBacktestAccountsWorkflowName is the name of the workflow to get the accounts of a backtest.
const GetBacktestAccountsWorkflowName = "GetBacktestAccountsWorkflow"

type (
	// GetBacktestAccountsWorkflowParams is the parameters of the GetBacktestAccountsWorkflow workflow.
	GetBacktestAccountsWorkflowParams struct {
		BacktestID uuid.UUID
	}

	// GetBacktestAccountsWorkflowResults is the results of the GetBacktestAccountsWorkflow workflow.
	GetBacktestAccountsWorkflowResults struct {
		Accounts map[string]account.Account
	}
)

// CreateBacktestOrderWorkflowName is the name of the workflow to create an order for a backtest.
const CreateBacktestOrderWorkflowName = "CreateBacktestOrderWorkflow"

type (
	// CreateBacktestOrderWorkflowParams is the parameters of the CreateBacktestOrderWorkflow workflow.
	CreateBacktestOrderWorkflowParams struct {
		BacktestID uuid.UUID
		Order      order.Order
	}

	// CreateBacktestOrderWorkflowResults is the results of the CreateBacktestOrderWorkflow workflow.
	CreateBacktestOrderWorkflowResults struct{}
)

// GetBacktestOrdersWorkflowName is the name of the workflow to get the orders of a backtest.
const GetBacktestOrdersWorkflowName = "GetBacktestOrdersWorkflow"

type (
	// GetBacktestOrdersWorkflowParams is the parameters of the GetBacktestOrdersWorkflow workflow.
	GetBacktestOrdersWorkflowParams struct {
		BacktestID uuid.UUID
	}

	// GetBacktestOrdersWorkflowResults is the results of the GetBacktestOrdersWorkflow workflow.
	GetBacktestOrdersWorkflowResults struct {
		Orders []order.Order
	}
)

// SubscribeToPriceWorkflowName is the name of the workflow to subscribe to prices.
const SubscribeToPriceWorkflowName = "SubscribeToPriceWorkflow"

type (
	// SubscribeToPriceWorkflowParams is the parameters of the SubscribeToPriceWorkflow workflow.
	SubscribeToPriceWorkflowParams struct {
		BacktestID uuid.UUID
		Exchange   string
		Pair       string
	}

	// SubscribeToPriceWorkflowResults is the results of the SubscribeToPriceWorkflow workflow.
	SubscribeToPriceWorkflowResults struct{}
)

const (
	// ServiceInfoWorkflowName is the name of the workflow to get the service info.
	ServiceInfoWorkflowName = "ServiceInfoWorkflow"
)

type (
	// ServiceInfoParams contains the parameters of the service info workflow.
	ServiceInfoParams struct{}

	// ServiceInfoResults contains the result of the service info workflow.
	ServiceInfoResults struct {
		Version string
	}
)
