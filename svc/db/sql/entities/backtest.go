package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cryptellation/backtests/pkg/backtest"
	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
	"github.com/google/uuid"
)

// Parameters is the entity for the parameters of a backtest.
type Parameters struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Mode        string    `json:"mode"`
	PricePeriod string    `json:"price_period"`
}

// BacktestData is the entity for the backtest data.
type BacktestData struct {
	StartTime         time.Time          `json:"start_time"`
	EndTime           time.Time          `json:"end_time"`
	Mode              string             `json:"mode"`
	PricePeriod       string             `json:"price_period"`
	CurrentTime       time.Time          `json:"current_time"`
	CurrentPriceType  string             `json:"current_price_type"`
	Balances          []Balance          `json:"balances"`
	Orders            []Order            `json:"orders"`
	TickSubscriptions []TickSubscription `json:"tick_subscriptions"`
	Callbacks         Callbacks          `json:"callbacks"`
}

// Backtest is the entity for a backtest.
type Backtest struct {
	ID   string `db:"id"`
	Data []byte `db:"data"`
}

// ToModel converts the entity to a model.
func (bt Backtest) ToModel() (backtest.Backtest, error) {
	// Get the backtest data.
	var data BacktestData
	if err := json.Unmarshal(bt.Data, &data); err != nil {
		return backtest.Backtest{}, err
	}

	priceType := candlestick.PriceType(data.CurrentPriceType)
	if err := priceType.Validate(); err != nil {
		wrappedErr := fmt.Errorf("error when validating current price type, got %q: %w", data.CurrentPriceType, err)
		return backtest.Backtest{}, wrappedErr
	}

	periodBetweenEvents := period.Symbol(data.PricePeriod)
	if err := periodBetweenEvents.Validate(); err != nil {
		return backtest.Backtest{}, err
	}

	mode := backtest.Mode(data.Mode)
	if err := mode.Validate(); err != nil {
		return backtest.Backtest{}, err
	}

	orders, err := ToOrderModels(data.Orders)
	if err != nil {
		return backtest.Backtest{}, err
	}

	id, err := uuid.Parse(bt.ID)
	if err != nil {
		return backtest.Backtest{}, err
	}

	return backtest.Backtest{
		ID:          id,
		StartTime:   data.StartTime,
		EndTime:     data.EndTime,
		Mode:        mode,
		PricePeriod: periodBetweenEvents,
		CurrentCandlestick: backtest.CurrentCandlestick{
			Time:  data.CurrentTime,
			Price: priceType,
		},
		Accounts:            ToAccountModels(data.Balances),
		Orders:              orders,
		PricesSubscriptions: ToTickSubscriptionModels(data.TickSubscriptions),
		Callbacks:           data.Callbacks.ToCallbacksModel(),
	}, nil
}

// FromBacktestModel converts a model into an entity.
func FromBacktestModel(bt backtest.Backtest) (Backtest, error) {
	// Create the backtest data.
	data := BacktestData{
		StartTime:         bt.StartTime,
		EndTime:           bt.EndTime,
		Mode:              bt.Mode.String(),
		PricePeriod:       bt.PricePeriod.String(),
		CurrentTime:       bt.CurrentCandlestick.Time,
		CurrentPriceType:  bt.CurrentCandlestick.Price.String(),
		Balances:          FromAccountModels(bt.Accounts),
		Orders:            FromOrderModels(bt.Orders),
		TickSubscriptions: FromTickSubscriptionModels(bt.PricesSubscriptions),
		Callbacks:         FromCallbacksModel(bt.Callbacks),
	}

	// Marshal the backtest data.
	dataByte, err := json.Marshal(data)
	if err != nil {
		return Backtest{}, err
	}

	return Backtest{
		ID:   bt.ID.String(),
		Data: dataByte,
	}, nil
}
