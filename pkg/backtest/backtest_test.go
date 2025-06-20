//go:build unit
// +build unit

package backtest

import (
	"testing"
	"time"

	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/runtime/account"
	"github.com/stretchr/testify/suite"
)

func TestBacktestSuite(t *testing.T) {
	suite.Run(t, new(BacktestSuite))
}

type BacktestSuite struct {
	suite.Suite
}

func (suite *BacktestSuite) TestMarshalUnMarshalBinary() {
	bt := Backtest{
		StartTime: time.Unix(0, 0).UTC(),
		EndTime:   time.Unix(200, 0).UTC(),
		CurrentCandlestick: CurrentCandlestick{
			Time: time.Unix(100, 0).UTC(),
		},
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"USDC": 10000,
				},
			},
		},
	}

	content, err := bt.MarshalBinary()
	suite.Require().NoError(err)
	nbt := Backtest{}
	suite.Require().NoError(nbt.UnmarshalBinary(content))
	suite.Require().Equal(bt, nbt)
}

func (suite *BacktestSuite) TestIncrementPriceID() {
	// TODO(#3): Implement TestIncrementPriceID
}

func (suite *BacktestSuite) TestBacktestCreateWithModeFullOHLC() {
	mode := ModeIsCloseOHLC
	per := period.M1
	params := Parameters{
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"USDC": 10000,
				},
			},
		},
		StartTime:   time.Unix(0, 0).UTC(),
		EndTime:     nil,
		Mode:        &mode,
		PricePeriod: &per,
	}

	bt, err := New(params, runtime.Callbacks{})
	suite.Require().NoError(err)
	suite.Require().Equal(ModeIsCloseOHLC, bt.Mode)
	suite.Require().Equal(candlestick.PriceTypeIsClose, bt.CurrentCandlestick.Price)
}

func (suite *BacktestSuite) TestBacktestCreateWithModeCloseOHLC() {
	mode := ModeIsCloseOHLC
	per := period.M1
	params := Parameters{
		Accounts: map[string]account.Account{
			"exchange": {
				Balances: map[string]float64{
					"USDC": 10000,
				},
			},
		},
		StartTime:   time.Unix(0, 0).UTC(),
		EndTime:     nil,
		Mode:        &mode,
		PricePeriod: &per,
	}

	bt, err := New(params, runtime.Callbacks{})
	suite.Require().NoError(err)
	suite.Require().Equal(ModeIsCloseOHLC, bt.Mode)
	suite.Require().Equal(candlestick.PriceTypeIsClose, bt.CurrentCandlestick.Price)
}

func (suite *BacktestSuite) TestBacktestSetNewTimeWithFullOHLCMode() {
	bt := Backtest{
		StartTime: time.Unix(0, 0).UTC(),
		EndTime:   time.Unix(200, 0).UTC(),
		Mode:      ModeIsFullOHLC,
		CurrentCandlestick: CurrentCandlestick{
			Time:  time.Unix(100, 0).UTC(),
			Price: candlestick.PriceTypeIsClose,
		},
	}

	bt.SetCurrentTime(time.Unix(150, 0).UTC())
	suite.Require().Equal(time.Unix(150, 0).UTC(), bt.CurrentCandlestick.Time)
	suite.Require().Equal(candlestick.PriceTypeIsOpen, bt.CurrentCandlestick.Price)
}

func (suite *BacktestSuite) TestBacktestSetNewTimeWithCloseOHLCMode() {
	bt := Backtest{
		StartTime: time.Unix(0, 0).UTC(),
		EndTime:   time.Unix(200, 0).UTC(),
		Mode:      ModeIsCloseOHLC,
		CurrentCandlestick: CurrentCandlestick{
			Time:  time.Unix(100, 0).UTC(),
			Price: candlestick.PriceTypeIsClose,
		},
	}

	bt.SetCurrentTime(time.Unix(150, 0).UTC())
	suite.Require().Equal(time.Unix(150, 0).UTC(), bt.CurrentCandlestick.Time)
	suite.Require().Equal(candlestick.PriceTypeIsClose, bt.CurrentCandlestick.Price)
}
