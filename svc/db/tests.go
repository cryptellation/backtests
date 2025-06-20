package db

import (
	"context"
	"time"

	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/runtime/account"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// BacktestSuite is a suite of tests for the backtest activity database.
type BacktestSuite struct {
	suite.Suite
	DB DB
}

// TestCreateRead tests that creatin then reading a backtest activity works.
func (suite *BacktestSuite) TestCreateRead() {
	bt := backtest.Backtest{
		ID:          uuid.New(),
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(120, 0),
		Mode:        backtest.ModeIsFullOHLC,
		PricePeriod: period.M1,
		CurrentCandlestick: backtest.CurrentCandlestick{
			Time:  time.Unix(60, 0),
			Price: candlestick.PriceTypeIsLow,
		},
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"DAI": 1000,
				},
			},
		},
		Callbacks: runtime.Callbacks{
			OnInitCallback: runtime.CallbackWorkflow{
				Name:          "test-init-workflow",
				TaskQueueName: "test-queue",
			},
			OnNewPricesCallback: runtime.CallbackWorkflow{
				Name:          "test-prices-workflow",
				TaskQueueName: "test-queue",
			},
			OnExitCallback: runtime.CallbackWorkflow{
				Name:          "test-exit-workflow",
				TaskQueueName: "test-queue",
			},
		},
	}
	_, err := suite.DB.CreateBacktestActivity(context.Background(), CreateBacktestActivityParams{
		Backtest: bt,
	})
	suite.Require().NoError(err)
	resp, err := suite.DB.ReadBacktestActivity(context.Background(), ReadBacktestActivityParams{
		ID: bt.ID,
	})
	suite.Require().NoError(err, bt.ID.String())

	suite.Require().Equal(bt.ID, resp.Backtest.ID)
	suite.Require().Len(resp.Backtest.Accounts, 1)
	suite.Require().Len(resp.Backtest.Accounts["exchange"].Balances, 1)
	suite.Require().Equal(bt.Accounts["exchange"].Balances["DAI"], resp.Backtest.Accounts["exchange"].Balances["DAI"])
	suite.Require().Equal(backtest.ModeIsFullOHLC, resp.Backtest.Mode)
	suite.Require().Equal(bt.Callbacks.OnInitCallback, resp.Backtest.Callbacks.OnInitCallback)
	suite.Require().Equal(bt.Callbacks.OnNewPricesCallback, resp.Backtest.Callbacks.OnNewPricesCallback)
	suite.Require().Equal(bt.Callbacks.OnExitCallback, resp.Backtest.Callbacks.OnExitCallback)
}

// createTestBacktest creates a test backtest with the given ID and workflow names.
func (suite *BacktestSuite) createTestBacktest(id uuid.UUID, initName, pricesName, exitName string) backtest.Backtest {
	return backtest.Backtest{
		ID:          id,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(120, 0),
		Mode:        backtest.ModeIsFullOHLC,
		PricePeriod: period.M1,
		CurrentCandlestick: backtest.CurrentCandlestick{
			Time:  time.Unix(60, 0),
			Price: candlestick.PriceTypeIsLow,
		},
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"DAI": 1000,
				},
			},
		},
		Callbacks: runtime.Callbacks{
			OnInitCallback: runtime.CallbackWorkflow{
				Name:          initName,
				TaskQueueName: "test-queue",
			},
			OnNewPricesCallback: runtime.CallbackWorkflow{
				Name:          pricesName,
				TaskQueueName: "test-queue",
			},
			OnExitCallback: runtime.CallbackWorkflow{
				Name:          exitName,
				TaskQueueName: "test-queue",
			},
		},
	}
}

// TestList tests that listing backtests returns the correct number of backtests.
func (suite *BacktestSuite) TestList() {
	bt1 := suite.createTestBacktest(uuid.New(), "test-init-workflow-1", "test-prices-workflow-1", "test-exit-workflow-1")
	bt2 := suite.createTestBacktest(uuid.New(), "test-init-workflow-2", "test-prices-workflow-2", "test-exit-workflow-2")

	_, err := suite.DB.CreateBacktestActivity(context.Background(), CreateBacktestActivityParams{
		Backtest: bt1,
	})
	suite.Require().NoError(err)
	_, err = suite.DB.CreateBacktestActivity(context.Background(), CreateBacktestActivityParams{
		Backtest: bt2,
	})
	suite.Require().NoError(err)

	resp, err := suite.DB.ListBacktestsActivity(context.Background(), ListBacktestsActivityParams{})
	suite.Require().NoError(err)

	suite.Require().Len(resp.Backtests, 2)
	suite.Require().Equal(bt1.ID, resp.Backtests[0].ID)
	suite.Require().Equal(bt2.ID, resp.Backtests[1].ID)
}

// createUpdatedTestBacktest creates an updated test backtest with different values.
func (suite *BacktestSuite) createUpdatedTestBacktest(id uuid.UUID) backtest.Backtest {
	return backtest.Backtest{
		ID:          id,
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(120, 0),
		Mode:        backtest.ModeIsFullOHLC,
		PricePeriod: period.M1,
		CurrentCandlestick: backtest.CurrentCandlestick{
			Time:  time.Unix(60, 0),
			Price: candlestick.PriceTypeIsClose,
		},
		Accounts: map[string]account.Account{
			"exchange2": {
				Balances: map[string]float64{
					"USDC": 1500,
				},
			},
		},
		Callbacks: runtime.Callbacks{
			OnInitCallback: runtime.CallbackWorkflow{
				Name:          "test-init-workflow-updated",
				TaskQueueName: "test-queue",
			},
			OnNewPricesCallback: runtime.CallbackWorkflow{
				Name:          "test-prices-workflow-updated",
				TaskQueueName: "test-queue",
			},
			OnExitCallback: runtime.CallbackWorkflow{
				Name:          "test-exit-workflow-updated",
				TaskQueueName: "test-queue",
			},
		},
	}
}

// TestUpdate tests that updating a backtest works.
func (suite *BacktestSuite) TestUpdate() {
	bt := suite.createTestBacktest(uuid.New(),
		"test-init-workflow-original",
		"test-prices-workflow-original",
		"test-exit-workflow-original")

	_, err := suite.DB.CreateBacktestActivity(context.Background(), CreateBacktestActivityParams{
		Backtest: bt,
	})
	suite.Require().NoError(err)

	bt2 := suite.createUpdatedTestBacktest(bt.ID)
	bt2.ID = bt.ID // Ensure same ID

	_, err = suite.DB.UpdateBacktestActivity(context.Background(), UpdateBacktestActivityParams{
		Backtest: bt2,
	})
	suite.Require().NoError(err)

	resp, err := suite.DB.ReadBacktestActivity(context.Background(), ReadBacktestActivityParams{
		ID: bt.ID,
	})
	suite.Require().NoError(err)

	suite.Require().Equal(bt.ID, resp.Backtest.ID)
	suite.Require().Equal(bt2.ID, resp.Backtest.ID)
	suite.Require().Len(resp.Backtest.Accounts, 1)
	suite.Require().Len(resp.Backtest.Accounts["exchange2"].Balances, 1)
	suite.Require().Equal(bt2.Accounts["exchange2"].Balances["USDC"], resp.Backtest.Accounts["exchange2"].Balances["USDC"])
	suite.Require().Equal(bt2.Callbacks.OnInitCallback, resp.Backtest.Callbacks.OnInitCallback)
	suite.Require().Equal(bt2.Callbacks.OnNewPricesCallback, resp.Backtest.Callbacks.OnNewPricesCallback)
	suite.Require().Equal(bt2.Callbacks.OnExitCallback, resp.Backtest.Callbacks.OnExitCallback)
}

// TestDelete tests that deleting a backtest works.
func (suite *BacktestSuite) TestDelete() {
	bt := backtest.Backtest{
		ID:          uuid.New(),
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(120, 0),
		Mode:        backtest.ModeIsFullOHLC,
		PricePeriod: period.M1,
		CurrentCandlestick: backtest.CurrentCandlestick{
			Time:  time.Unix(60, 0),
			Price: candlestick.PriceTypeIsLow,
		},
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"ETH": 1000,
				},
			},
		},
		Callbacks: runtime.Callbacks{
			OnInitCallback: runtime.CallbackWorkflow{
				Name:          "test-init-workflow-delete",
				TaskQueueName: "test-queue",
			},
			OnNewPricesCallback: runtime.CallbackWorkflow{
				Name:          "test-prices-workflow-delete",
				TaskQueueName: "test-queue",
			},
			OnExitCallback: runtime.CallbackWorkflow{
				Name:          "test-exit-workflow-delete",
				TaskQueueName: "test-queue",
			},
		},
	}
	_, err := suite.DB.CreateBacktestActivity(context.Background(), CreateBacktestActivityParams{
		Backtest: bt,
	})
	suite.Require().NoError(err)
	_, err = suite.DB.DeleteBacktestActivity(context.Background(), DeleteBacktestActivityParams{
		ID: bt.ID,
	})
	suite.Require().NoError(err)
	_, err = suite.DB.ReadBacktestActivity(context.Background(), ReadBacktestActivityParams{
		ID: bt.ID,
	})
	suite.Error(err)
}

// TestDeleteInexistant tests that deleting an inexistant backtest does not return an error.
func (suite *BacktestSuite) TestDeleteInexistant() {
	bt := backtest.Backtest{
		ID:          uuid.New(),
		StartTime:   time.Unix(0, 0),
		EndTime:     time.Unix(120, 0),
		Mode:        backtest.ModeIsFullOHLC,
		PricePeriod: period.M1,
		CurrentCandlestick: backtest.CurrentCandlestick{
			Time:  time.Unix(60, 0),
			Price: candlestick.PriceTypeIsLow,
		},
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"ETH": 1000,
				},
			},
		},
		Callbacks: runtime.Callbacks{
			OnInitCallback: runtime.CallbackWorkflow{
				Name:          "test-init-workflow-inexistant",
				TaskQueueName: "test-queue",
			},
			OnNewPricesCallback: runtime.CallbackWorkflow{
				Name:          "test-prices-workflow-inexistant",
				TaskQueueName: "test-queue",
			},
			OnExitCallback: runtime.CallbackWorkflow{
				Name:          "test-exit-workflow-inexistant",
				TaskQueueName: "test-queue",
			},
		},
	}
	_, err := suite.DB.CreateBacktestActivity(context.Background(), CreateBacktestActivityParams{
		Backtest: bt,
	})
	suite.Require().NoError(err)
	_, err = suite.DB.DeleteBacktestActivity(context.Background(), DeleteBacktestActivityParams{
		ID: bt.ID,
	})
	suite.Require().NoError(err)
	_, err = suite.DB.DeleteBacktestActivity(context.Background(), DeleteBacktestActivityParams{
		ID: bt.ID,
	})
	suite.Require().NoError(err)
}
