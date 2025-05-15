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
	"github.com/lerenn/cryptellation/pkg/utils"
	"go.temporal.io/sdk/workflow"
)

type testRobotRun struct {
	Suite *EndToEndSuite

	BacktestID     uuid.UUID
	BacktestParams api.CreateBacktestWorkflowParams

	WfClient clients.WfClient

	OnInitCalls      int
	OnNewPricesCalls int
	OnExitCalls      int
}

func (r *testRobotRun) OnInit(ctx workflow.Context, params runtime.OnInitCallbackWorkflowParams) error {
	checkBacktestRunContext(r.Suite, params.RunCtx, r.BacktestID)
	r.Suite.Require().WithinDuration(r.BacktestParams.BacktestParameters.StartTime, params.RunCtx.Now, time.Second)

	_, err := r.WfClient.SubscribeToPrice(ctx, api.SubscribeToPriceWorkflowParams{
		BacktestID: r.BacktestID,
		Exchange:   "binance",
		Pair:       "BTC-USDT",
	})
	r.Suite.Require().NoError(err)

	r.OnInitCalls++
	return err
}

func (r *testRobotRun) OnNewPrices(_ workflow.Context, params runtime.OnNewPricesCallbackWorkflowParams) error {
	checkBacktestRunContext(r.Suite, params.Run, r.BacktestID)

	// TODO: test order passing in OnNewPrices

	r.OnNewPricesCalls++
	return nil
}

func (r *testRobotRun) OnExit(_ workflow.Context, params runtime.OnExitCallbackWorkflowParams) error {
	checkBacktestRunContext(r.Suite, params.Run, r.BacktestID)
	r.Suite.Require().WithinDuration(*r.BacktestParams.BacktestParameters.EndTime, params.Run.Now, time.Second)

	r.OnExitCalls++
	return nil
}

func (suite *EndToEndSuite) TestBacktestRun() {
	// WHEN creating a new backtest

	params := api.CreateBacktestWorkflowParams{
		BacktestParameters: backtest.Parameters{
			Accounts: map[string]account.Account{
				"binance": {
					Balances: map[string]float64{
						"BTC": 1,
					},
				},
			},
			StartTime: utils.Must(time.Parse(time.RFC3339, "2023-02-26T12:00:00Z")),
			EndTime:   utils.ToReference(utils.Must(time.Parse(time.RFC3339, "2023-02-26T12:02:00Z"))),
		},
	}
	backtest, err := suite.client.NewBacktest(context.Background(), params)

	// THEN no error is returned

	suite.Require().NoError(err)

	// WHEN running the backtest with a robot

	r := &testRobotRun{
		BacktestParams: params,
		BacktestID:     backtest.ID,
		Suite:          suite,
		WfClient:       clients.NewWfClient(),
	}
	err = backtest.Run(context.Background(), r)

	// THEN no error is returned

	suite.Require().NoError(err)

	// AND the robot callbacks are called
	suite.Require().Equal(1, r.OnInitCalls)
	suite.Require().Equal(2, r.OnNewPricesCalls)
	suite.Require().Equal(1, r.OnExitCalls)
}

func checkBacktestRunContext(suite *EndToEndSuite, ctx runtime.Context, backtestID uuid.UUID) {
	suite.Require().Equal(backtestID, ctx.ID)
	suite.Require().Equal(runtime.ModeBacktest, ctx.Mode)
	suite.Require().NotEmpty(ctx.TaskQueue)
}
