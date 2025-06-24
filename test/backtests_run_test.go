//go:build e2e
// +build e2e

package test

import (
	"context"
	"time"

	"github.com/cryptellation/backtests/api"
	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/backtests/pkg/clients"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/runtime/account"
	"github.com/google/uuid"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type testRunner struct {
	Suite *EndToEndSuite

	BacktestID uuid.UUID
	Params     backtest.Parameters

	WfClient clients.WfClient

	OnInitCallsCount      int
	OnNewPricesCallsCount int
	OnExitCallsCount      int
}

func (r *testRunner) Name() string {
	return "BacktestE2eRunner"
}

func (r *testRunner) OnInit(ctx workflow.Context, params runtime.OnInitCallbackWorkflowParams) error {
	checkBacktestRunContext(r.Suite, params.Context, r.BacktestID)
	r.Suite.Require().WithinDuration(r.Params.StartTime, params.Context.Now, time.Second)

	_, err := r.WfClient.SubscribeToPrice(ctx, api.SubscribeToPriceWorkflowParams{
		BacktestID: params.Context.ID,
		Exchange:   "binance",
		Pair:       "BTC-USDT",
	})
	r.Suite.Require().NoError(err)

	r.OnInitCallsCount++
	return err
}

func (r *testRunner) OnNewPrices(_ workflow.Context, params runtime.OnNewPricesCallbackWorkflowParams) error {
	checkBacktestRunContext(r.Suite, params.Context, r.BacktestID)

	// TODO(#6): test order passing in OnNewPrices

	r.OnNewPricesCallsCount++
	return nil
}

func (r *testRunner) OnExit(_ workflow.Context, params runtime.OnExitCallbackWorkflowParams) error {
	checkBacktestRunContext(r.Suite, params.Context, r.BacktestID)
	r.Suite.Require().WithinDuration(*r.Params.EndTime, params.Context.Now, time.Second)

	r.OnExitCallsCount++
	return nil
}

func (suite *EndToEndSuite) TestBacktestRun() {
	// GIVEN a running temporal worker

	tq := "BacktestE2eRunner-TaskQueue"
	w := worker.New(suite.temporalclient, tq, worker.Options{})
	go func() {
		if err := w.Run(nil); err != nil {
			suite.Require().NoError(err)
		}
	}()
	defer w.Stop()

	// AND a registered runnable

	start, _ := time.Parse(time.RFC3339, "2023-02-26T12:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2023-02-26T12:02:00Z")
	params := backtest.Parameters{
		Accounts: map[string]account.Account{
			"binance": {
				Balances: map[string]float64{
					"BTC": 1,
				},
			},
		},
		StartTime: start,
		EndTime:   &end,
	}
	r := &testRunner{
		Params:   params,
		Suite:    suite,
		WfClient: clients.NewWfClient(),
	}
	callbacks := runtime.RegisterRunnable(w, tq, r)

	// WHEN creating a new backtest

	backtest, err := suite.client.NewBacktest(context.Background(), params, callbacks)

	// THEN no error is returned

	suite.Require().NoError(err)

	// WHEN running the backtest with a runner on the worker

	r.BacktestID = backtest.ID // Add backtest ID to runner for checking backtest run context
	err = backtest.Run(context.Background())

	// THEN no error is returned

	suite.Require().NoError(err)

	// AND the runner callbacks are called
	suite.Require().Equal(1, r.OnInitCallsCount)
	suite.Require().Equal(2, r.OnNewPricesCallsCount)
	suite.Require().Equal(1, r.OnExitCallsCount)
}

func checkBacktestRunContext(suite *EndToEndSuite, ctx runtime.Context, backtestID uuid.UUID) {
	suite.Require().Equal(backtestID, ctx.ID)
	suite.Require().Equal(runtime.ModeBacktest, ctx.Mode)
	suite.Require().NotEmpty(ctx.ParentTaskQueue)
}
