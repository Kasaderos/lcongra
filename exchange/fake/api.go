package exchanging

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"log"

	"github.com/google/uuid"
	"github.com/kasaderos/lcongra/exchange"
	"github.com/kasaderos/lcongra/exchange/binance"
)

type fakeExchange struct {
	Name            string
	account         *Account
	makerCommission float64
	pair            string
	pricePrecision  int
	basePrecision   int
	quotePrecision  int
	mu              sync.RWMutex
	price           float64 // rate of pair
	logger          *log.Logger
	realExchange    exchange.Exchanger
}

type Account struct {
	// baseFree of fakeExchange.pair
	baseCurrency   string
	quotedCurrency string
	baseFree       float64
	quotedFree     float64
	mu             sync.RWMutex
	orders         []exchange.Order
}

func NewExchange(logger *log.Logger) exchange.Exchanger {
	return &fakeExchange{
		Name: "fake_exchange",
		account: &Account{
			baseCurrency:   "BTC",
			quotedCurrency: "USDT",
			baseFree:       0.0,   // BTC
			quotedFree:     100.0, // USDT
		},
		pair:            "BTCUSDT",
		pricePrecision:  2,
		basePrecision:   6,
		price:           0,
		logger:          logger,
		realExchange:    binance.NewExchange(logger, "", ""),
		makerCommission: 0.001,
	}
}

func Update(ctx context.Context, ex exchange.Exchanger) {
	exchange, ok := ex.(*fakeExchange)
	if !ok {
		log.Println("exchanging: error interface conversation\n")
		ctx.Done()
		return
	}
	for {
		select {
		case <-ctx.Done():
		default:
		}
		exchange.updatePrice()
		exchange.updateActiveOrders()
		time.Sleep(1000 * time.Millisecond)
	}
}

func (ex *fakeExchange) updatePrice() {
	rate, err := ex.realExchange.GetRate(ex.pair)
	if err != nil {
		ex.logger.Println(err)
		return
	}
	// websocket prices of pair
	ex.mu.Lock()
	defer ex.mu.Unlock()
	ex.price = rate
}

func (ex *fakeExchange) updateActiveOrders() {
	logger := ex.logger
	ex.mu.RLock()
	currentPrice := ex.price
	ex.mu.RUnlock()
	ex.account.mu.Lock()
	defer ex.account.mu.Unlock()
	orders := ex.account.orders

	// fmt.Printf("[\n")
	newOrders := make([]exchange.Order, len(orders))
	copy(newOrders, orders)
	for i, r := range orders {
		if r.Side == "BUY" && r.Price > currentPrice {
			logger.Println(r.ID, "completed")
			ex.account.quotedFree -= r.Price * r.Amount
			roundZero(&ex.account.quotedFree)
			ex.account.baseFree += r.Amount - r.Amount*ex.makerCommission
			logger.Println("baseFree", ex.account.baseFree, "quotedFree", ex.account.quotedFree)
			// completed
			newOrders[i], newOrders[len(newOrders)-1] = newOrders[len(newOrders)-1], newOrders[i]
			newOrders = newOrders[:len(newOrders)-1]
		} else if r.Side == "SELL" && r.Price < currentPrice {
			logger.Println(r.ID, "completed")
			ex.account.baseFree -= r.Amount
			roundZero(&ex.account.baseFree)
			sum := r.Price * r.Amount
			ex.account.quotedFree += sum - (sum * ex.makerCommission)
			logger.Println("baseFree", ex.account.baseFree, "quotedFree", ex.account.quotedFree)
			// completed
			newOrders[i], newOrders[len(newOrders)-1] = newOrders[len(newOrders)-1], newOrders[i]
			newOrders = newOrders[:len(newOrders)-1]
		}
		// fmt.Printf("%d orderID %s\n", i, r.ID)
	}
	ex.account.orders = newOrders
	// fmt.Printf("]\n")
}

func roundZero(a *float64) {
	if math.Abs(*a) < 1e-2 {
		*a = 0.0
	}
}

func (ex *fakeExchange) PairFormat(pair string) string {
	b, q := exchange.Currencies(pair)
	return b + q
}

func (ex *fakeExchange) Ping() error {
	ex.logger.Println("success")
	return nil
}

func (ex *fakeExchange) GetRate(pair string) (rate float64, err error) {
	ex.mu.RLock()
	defer ex.mu.RUnlock()
	return ex.price, nil
}

func (ex *fakeExchange) CreateOrder(order *exchange.Order) (string, error) {
	sum := order.Price * order.Amount
	ex.account.mu.Lock()
	defer ex.account.mu.Unlock()
	if order.Side == "BUY" && ex.account.quotedFree < sum {
		return "", errors.New(fmt.Sprintf("not enough money in balance baseFree %v", ex.account.quotedFree))
	} else if order.Side == "SELL" && ex.account.baseFree < order.Amount {
		return "", errors.New(fmt.Sprintf("not enough money in balance baseFree %v", ex.account.baseFree))
	}
	if sum < 10 { // dollars
		msg := fmt.Sprintf("exchanging sum not enough %f", sum)
		ex.logger.Println("error", msg)
		return "", errors.New(msg)
	}
	ex.logger.Printf("exchanging sum %f", sum)
	order.ID = uuid.New().String()
	ex.account.orders = append(ex.account.orders, *order)
	return order.ID, nil
}

func (ex *fakeExchange) GetBalance(ctx context.Context,
	curr string) (amount float64, err error) {
	ex.logger.Println("got", curr)
	if curr == ex.account.baseCurrency {
		ex.logger.Println(curr, ex.account.baseFree)
		return ex.account.baseFree, nil
	} else if curr == ex.account.quotedCurrency {
		ex.logger.Println(curr, ex.account.quotedFree)
		return ex.account.quotedFree, nil
	}

	return -1, fmt.Errorf("error unknown currency")
}

func (ex *fakeExchange) OpenedOrders(pair string) ([]exchange.Order, error) {
	ex.account.mu.RLock()
	// ex.logger.Println("orders", len(ex.account.orders))
	defer ex.account.mu.RUnlock()
	return ex.account.orders, nil
}

func (ex *fakeExchange) CancelOrder(ctx context.Context, pair string, orderID string) (err error) {
	ex.logger.Println("status", "CANCELED")
	ex.account.mu.Lock()
	defer ex.account.mu.Unlock()
	orders := ex.account.orders
	for i, r := range orders {
		if r.ID == orderID {
			orders[i], orders[len(orders)-1] = orders[len(orders)-1], orders[i]
			orders = orders[:len(orders)-1]
			ex.account.orders = orders
			return nil
		}
	}
	return fmt.Errorf("can't find orderID %s\n", orderID)
}

func (ex *fakeExchange) GetInformation(ctx context.Context, pair string) (info *exchange.Information, err error) {
	info = &exchange.Information{
		MakerCommission:    ex.makerCommission,
		CanTrade:           true,
		Pair:               ex.pair,
		PricePrecision:     ex.pricePrecision,
		BasePrecision:      ex.basePrecision,
		FreeQuotedCurrency: ex.account.quotedFree,
	}
	return info, nil
}

func getPrecision(f float64) int {
	i := 0
	for math.Round(f) <= 0 {
		f *= 10
		i++
	}
	return i
}
