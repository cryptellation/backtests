package entities

import (
	"time"

	"github.com/cryptellation/runtime"
)

// CallbackWorkflow is the entity for a callback workflow.
type CallbackWorkflow struct {
	Name             string        `json:"name"`
	TaskQueueName    string        `json:"task_queue_name"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
}

// Callbacks is the entity for callbacks.
type Callbacks struct {
	OnInitCallback      CallbackWorkflow `json:"on_init_callback"`
	OnNewPricesCallback CallbackWorkflow `json:"on_new_prices_callback"`
	OnExitCallback      CallbackWorkflow `json:"on_exit_callback"`
}

// ToCallbackWorkflowModel converts a CallbackWorkflow entity to a runtime.CallbackWorkflow model.
func (cw CallbackWorkflow) ToCallbackWorkflowModel() runtime.CallbackWorkflow {
	return runtime.CallbackWorkflow{
		Name:             cw.Name,
		TaskQueueName:    cw.TaskQueueName,
		ExecutionTimeout: cw.ExecutionTimeout,
	}
}

// FromCallbackWorkflowModel converts a runtime.CallbackWorkflow model to a CallbackWorkflow entity.
func FromCallbackWorkflowModel(cw runtime.CallbackWorkflow) CallbackWorkflow {
	return CallbackWorkflow{
		Name:             cw.Name,
		TaskQueueName:    cw.TaskQueueName,
		ExecutionTimeout: cw.ExecutionTimeout,
	}
}

// ToCallbacksModel converts a Callbacks entity to a runtime.Callbacks model.
func (c Callbacks) ToCallbacksModel() runtime.Callbacks {
	return runtime.Callbacks{
		OnInitCallback:      c.OnInitCallback.ToCallbackWorkflowModel(),
		OnNewPricesCallback: c.OnNewPricesCallback.ToCallbackWorkflowModel(),
		OnExitCallback:      c.OnExitCallback.ToCallbackWorkflowModel(),
	}
}

// FromCallbacksModel converts a runtime.Callbacks model to a Callbacks entity.
func FromCallbacksModel(c runtime.Callbacks) Callbacks {
	return Callbacks{
		OnInitCallback:      FromCallbackWorkflowModel(c.OnInitCallback),
		OnNewPricesCallback: FromCallbackWorkflowModel(c.OnNewPricesCallback),
		OnExitCallback:      FromCallbackWorkflowModel(c.OnExitCallback),
	}
}
