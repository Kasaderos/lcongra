package exchange

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

var (
	ErrBalanceLocked   = errors.New("balance locked")
	ErrNotFound        = errors.New("not found")
	ErrIncompleteData  = errors.New("error incomplete data")
	ErrBrokenData      = errors.New("error broken data")
	ErrOrderNotCreated = errors.New("order not created")
	ErrKeysNotFound    = errors.New("keys not found")
)

type Order struct {
	ID         string
	OrderTime  time.Time
	PushedTime time.Time
	Pair       string
	Type       string
	Side       string
	Price      float64
	Amount     float64
}

// pair format BTC-USDT
type Exchanger interface {
	CreateOrder(ctx context.Context, order *Order) (id string, err error)
	OpenedOrders(ctx context.Context, pair string) (orders []Order, err error)
	GetRate(ctx context.Context, pair string) (rate float64, err error)
	// BTC-USDT -> exchange format
	PairFormat(ctx context.Context, pair string) (expair string)
	GetBalance(ctx context.Context, currency string) (amount float64, err error)
	GetInformation(ctx context.Context, pair string) (info *Information, err error)
}

func GetPrecision(f float64) int {
	i := 0
	for math.Round(f) <= 0 {
		f *= 10
		i++
	}
	return i
}

type Information struct {
	TakerCommission    float64
	MakerCommission    float64
	CanTrade           bool
	Pair               string
	BasePrecision      int
	FreeQuotedCurrency float64 // balance
	QuotePrecision     int
	PricePrecision     int
}

type Orderbook struct {
	asks, bids []Order
}

func NewOrderbook(asks, bids []Order) *Orderbook {
	return &Orderbook{
		asks: asks,
		bids: bids,
	}
}

func Currencies(pair string) (base string, quoated string) {
	currs := strings.Split(pair, "-")
	return currs[0], currs[1]
}

func (b *Orderbook) CalcRate() (rate float64) {
	sort.Sort(ByPrice(b.asks))
	sort.Sort(sort.Reverse(ByPrice(b.bids)))
	rate = (b.bids[0].Price + b.asks[0].Price) / 2.0
	return
}

type ByPrice []Order

func (b ByPrice) Less(i, j int) bool {
	return b[i].Price < b[j].Price
}

func (b ByPrice) Len() int {
	return len(b)
}

func (b ByPrice) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
