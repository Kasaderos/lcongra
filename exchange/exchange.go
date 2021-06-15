package exchange

import (
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

type Exchanger interface {
	CreateOrder(*Order) (string, error)
	OpenedOrders(string) ([]Order, error)
	GetRate(string) (float64, error)
	PairFormat(string) string
	GetBalance(string) (float64, error)
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
