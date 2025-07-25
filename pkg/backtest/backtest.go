package backtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cryptellation/candlesticks/pkg/candlestick"
	"github.com/cryptellation/candlesticks/pkg/period"
	"github.com/cryptellation/runtime"
	"github.com/cryptellation/runtime/account"
	"github.com/cryptellation/runtime/order"
	"github.com/cryptellation/ticks/pkg/tick"
	"github.com/google/uuid"
)

var (
	// ErrTickSubscriptionAlreadyExists is the error for a tick subscription already existing.
	ErrTickSubscriptionAlreadyExists = errors.New("tick subscription already exists")
	// ErrInvalidExchange is the error for an invalid exchange.
	ErrInvalidExchange = errors.New("invalid exchange")
	// ErrNoDataForOrderValidation is the error for no data for order validation.
	ErrNoDataForOrderValidation = errors.New("no data for order validation")
	// ErrStartAfterEnd is the error for starting after ending.
	ErrStartAfterEnd = errors.New("start after end")
	// ErrInvalidPricePeriod is the error for an invalid price period.
	ErrInvalidPricePeriod = errors.New("invalid price period")
)

// CurrentCandlestick represent the current price based on candlestick step.
type CurrentCandlestick struct {
	Time  time.Time
	Price candlestick.PriceType
}

// Backtest is the struct for a backtest.
type Backtest struct {
	ID                  uuid.UUID                  `json:"id"`
	StartTime           time.Time                  `json:"start_time"`
	EndTime             time.Time                  `json:"end_time"`
	Mode                Mode                       `json:"mode"`
	PricePeriod         period.Symbol              `json:"price_period"`
	CurrentCandlestick  CurrentCandlestick         `json:"current_candlestick"`
	Accounts            map[string]account.Account `json:"accounts"`
	PricesSubscriptions []tick.Subscription        `json:"tick_subscriptions"`
	Orders              []order.Order              `json:"orders"`
	Callbacks           runtime.Callbacks          `json:"callbacks"`
}

// Parameters is the struct for the backtest parameters.
type Parameters struct {
	Accounts    map[string]account.Account
	StartTime   time.Time
	EndTime     *time.Time
	Mode        *Mode
	PricePeriod *period.Symbol
}

// EmptyFieldsToDefault sets empty fields to default values.
func (params *Parameters) EmptyFieldsToDefault() *Parameters {
	if params.EndTime == nil {
		params.EndTime = defaultEndTime()
	}

	if params.PricePeriod == nil {
		params.PricePeriod = period.M1.Opt()
	}

	if params.Mode == nil {
		m := ModeIsCloseOHLC
		params.Mode = &m
	}

	return params
}

// Validate validates the backtest parameters.
func (params Parameters) Validate() error {
	if !params.StartTime.Before(*params.EndTime) {
		return ErrStartAfterEnd
	}

	if params.PricePeriod == nil {
		return fmt.Errorf("%w: nil", ErrInvalidPricePeriod)
	}

	if params.Mode == nil {
		return ErrInvalidMode
	}

	for exchange, a := range params.Accounts {
		if exchange == "" {
			return fmt.Errorf("error with exchange %q in new backtest params: %w", exchange, ErrInvalidExchange)
		}

		if err := a.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func defaultEndTime() *time.Time {
	t := time.Now()
	return &t
}

// New creates a new backtest.
func New(params Parameters, callbacks runtime.Callbacks) (Backtest, error) {
	// Set default fields params and validate it
	if err := params.EmptyFieldsToDefault().Validate(); err != nil {
		return Backtest{}, err
	}

	// Set current candlestick based on mode
	cc := CurrentCandlestick{
		Time: params.StartTime,
	}
	switch *params.Mode {
	case ModeIsCloseOHLC:
		cc.Price = candlestick.PriceTypeIsClose
	case ModeIsFullOHLC:
		cc.Price = candlestick.PriceTypeIsOpen
	}

	return Backtest{
		ID:                  uuid.New(),
		StartTime:           params.StartTime,
		EndTime:             *params.EndTime,
		Mode:                *params.Mode,
		PricePeriod:         *params.PricePeriod,
		CurrentCandlestick:  cc,
		Accounts:            params.Accounts,
		PricesSubscriptions: make([]tick.Subscription, 0),
		Orders:              make([]order.Order, 0),
		Callbacks:           callbacks,
	}, nil
}

// CurrentTime returns the current time of the backtest.
func (bt Backtest) CurrentTime() string {
	return fmt.Sprintf("%s [%s]", bt.CurrentCandlestick.Time, bt.CurrentCandlestick.Price)
}

// MarshalBinary marshals a backtest to binary data.
func (bt Backtest) MarshalBinary() ([]byte, error) {
	return json.Marshal(bt)
}

// UnmarshalBinary unmarshals a backtest from binary data.
func (bt *Backtest) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, bt)
}

// Advance advances the backtest to the next candlestick.
func (bt *Backtest) Advance() (done bool, err error) {
	switch bt.Mode {
	case ModeIsCloseOHLC:
		bt.advanceWithModeIsCloseOHLC()
	case ModeIsFullOHLC:
		bt.advanceWithModeIsFullOHLC()
	default:
		return false, fmt.Errorf("error with backtest mode %q: %w", bt.Mode, ErrInvalidMode)
	}

	return bt.Done(), nil
}

func (bt *Backtest) advanceWithModeIsCloseOHLC() {
	bt.CurrentCandlestick.Time = bt.CurrentCandlestick.Time.Add(bt.PricePeriod.Duration())
}

func (bt *Backtest) advanceWithModeIsFullOHLC() {
	switch bt.CurrentCandlestick.Price {
	case candlestick.PriceTypeIsOpen:
		bt.CurrentCandlestick.Price = candlestick.PriceTypeIsHigh
	case candlestick.PriceTypeIsHigh:
		bt.CurrentCandlestick.Price = candlestick.PriceTypeIsLow
	case candlestick.PriceTypeIsLow:
		bt.CurrentCandlestick.Price = candlestick.PriceTypeIsClose
	case candlestick.PriceTypeIsClose:
		bt.CurrentCandlestick.Time = bt.CurrentCandlestick.Time.Add(bt.PricePeriod.Duration())
	default:
		bt.CurrentCandlestick.Price = candlestick.PriceTypeIsOpen
	}
}

// Done returns true if the backtest is done.
func (bt Backtest) Done() bool {
	return !bt.CurrentCandlestick.Time.Before(bt.EndTime)
}

// SetCurrentTime sets the current time of the backtest.
func (bt *Backtest) SetCurrentTime(ts time.Time) {
	// Set new time
	bt.CurrentCandlestick.Time = ts

	// Starting the time on open if mode is full OHLC
	if bt.Mode == ModeIsFullOHLC {
		bt.CurrentCandlestick.Price = candlestick.PriceTypeIsOpen
	}
}

// CreateTickSubscription creates a new tick subscription for the backtest.
func (bt *Backtest) CreateTickSubscription(exchange string, pair string) (tick.Subscription, error) {
	for _, ts := range bt.PricesSubscriptions {
		if ts.Exchange == exchange && ts.Pair == pair {
			return tick.Subscription{}, ErrTickSubscriptionAlreadyExists
		}
	}

	s := tick.Subscription{
		Exchange: exchange,
		Pair:     pair,
	}
	bt.PricesSubscriptions = append(bt.PricesSubscriptions, s)

	return s, nil
}

// AddOrder adds an order to the backtest.
func (bt *Backtest) AddOrder(ord order.Order, cs candlestick.Candlestick) error {
	// Get exchange account
	exchangeAccount, ok := bt.Accounts[ord.Exchange]
	if !ok {
		return fmt.Errorf("error with orders exchange %q: %w", ord.Exchange, ErrInvalidExchange)
	}

	// Execute the order
	price := cs.Price(bt.CurrentCandlestick.Price)
	if err := exchangeAccount.ApplyOrder(price, ord); err != nil {
		return err
	}
	bt.Accounts[ord.Exchange] = exchangeAccount

	// Update and save the order
	ord.ExecutionTime = &bt.CurrentCandlestick.Time
	ord.Price = price
	bt.Orders = append(bt.Orders, ord)

	return nil
}
