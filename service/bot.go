package service

import (
	"context"
	"log"
	"sync"
	"time"
	"math"
	"errors"

	"github.com/kasaderos/lcongra/exchange"
)

type State int

const (
	Start State = iota
	GetSignal
	CreateOrder
	CheckOrder
	WaitOrder
	Nothing
	CancelOrder
	OpenPosition
	ClosePosition
	ClosePositionNow
)

const (
	AttemptsNumber = 3
	MinSum         = 11
)

var states = []string{"Start", "GetSignal", "CreateOrder", "CheckOrder", "WaitOrder", "Nothing", "CancelOrder", "OpenPosition", "ClosePosition", "ClosePositionNow"}

func (s State) String() string {
	return states[s]
}

type Bot struct {
	ms    sync.RWMutex // TODO
	state State

	mc       sync.RWMutex // TODO
	pair     string
	cache    float64
	count    int
	interval time.Duration

	exchange   exchange.Exchanger
	logger     *log.Logger
	info       *exchange.Information
	lastSignal Signal
	exCtx      context.Context
}

func NewBot(ex exchange.Exchanger, logger *log.Logger, pair string, exCtx context.Context, interval time.Duration) *Bot {

	return &Bot{
		exchange: ex,
		logger:   logger,
		pair:     pair,
		exCtx:    exCtx,
		interval: interval,
	}
}

func (b *Bot) GetState() State {
	b.ms.RLock()
	defer b.ms.RUnlock()
	return b.state
}

func (b *Bot) SetState(s State) {
	b.ms.Lock()
	defer b.ms.Unlock()
	b.state = s
}

func (b *Bot) GetCache() (baseSum float64, quoteSum float64, all float64) {
	base, quote := exchange.Currencies(b.pair)
	baseAmount, err := b.exchange.GetBalance(b.exCtx, base)
	if err != nil {
		b.logger.Println("error: get base amount", err)
	}
	quoteAmount, err := b.exchange.GetBalance(b.exCtx, quote)
	if err != nil {
		b.logger.Println("error: get quote amount", err)
	}
	price, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		b.logger.Println("balance: error: can't get rate", err)
	}
	b.ms.Lock()
	defer b.ms.Unlock()
	baseSum = price * baseAmount
	b.cache = baseSum + quoteAmount
	return baseSum, quoteAmount, b.cache
}

/*
			v---------------------------------------------------------------------------------------
	Start -> GetSignal |                                                                                    |
			    -> OpenPosition -> CreateOrder -> CheckOrder |                                      |
			    					          -> ClosePosition -> CreateOrder -> CheckOrder

*/
func (b *Bot) StartSM(ctx context.Context, msgChan <-chan string, signalChannel <-chan Signal) {
	b.SetState(Start)

	var currentOrder *exchange.Order
	var err error
	closed := true
	b.logger.Println("SM started")
SM:
	for {
		b.logger.Println("state", b.state)

		switch b.state {
		case Start:
			info, err := b.exchange.GetInformation(b.exCtx, b.pair)
			if err != nil {
				b.logger.Println("can't get info", err)
				time.Sleep(time.Minute)
				continue
			}
			b.info = info
			b.state = GetSignal

		case GetSignal:
			select {
			case b.lastSignal = <-signalChannel:
				if b.lastSignal.Dir == Up && time.Since(b.lastSignal.Time) < time.Second*5 {
					b.SetState(OpenPosition)
				}
			case <-ctx.Done():
				b.logger.Println("deleted")
				break SM
			}
		case OpenPosition:
			_, quote, _ := b.GetCache()
			if quote < MinSum {
				b.logger.Printf("not enough money in balance quote %v", quote)
				b.SetState(Nothing)
				continue
			}
			currentOrder, err = b.createBuyOrder()
			if err != nil {
				b.logger.Println(err)
				continue
			}
			b.SetState(CreateOrder)
		case ClosePosition:
			// bought currency let's sell
			base, _ := exchange.Currencies(b.pair)
			amount, err := b.exchange.GetBalance(b.exCtx, base)
			if err != nil {
				b.logger.Println(err)
				continue
			}
			if amount < 1e-12 {
				b.logger.Printf("not enough money in balance, base %v", amount)
				b.SetState(Nothing)
				continue
			}
			amount = roundDown(amount, b.info.BasePrecision)
			if closed {
				currentOrder = b.createSellOrder(currentOrder.Price, amount)
			} else {
				currentOrder, err = b.createMarketSellOrder(amount)
				if err != nil {
					b.logger.Println(err)
					continue
				}
			}
			b.SetState(CreateOrder)

		case CreateOrder:
			var id string
			for attempts := 0; attempts < AttemptsNumber; attempts++ {
				id, err = b.exchange.CreateOrder(b.exCtx, currentOrder)
				currentOrder.ID = id
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				b.SetState(GetSignal)
				b.logger.Printf("can't create order %+v, %+v", err, currentOrder)
				continue
			}

			b.logger.Printf("order created %+v\n", currentOrder)
			b.SetState(CheckOrder)

		case CheckOrder:
			var (
				orders []exchange.Order
				err    error
			)
			for attempts := 0; attempts < AttemptsNumber; attempts++ {
				orders, err = b.exchange.OpenedOrders(b.exCtx, currentOrder.Pair)
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				b.logger.Println("can't check order")
				time.Sleep(5 * time.Second)
				continue
			}

			exist := false
			for _, order := range orders {
				if order.ID == currentOrder.ID {
					exist = true
				}
			}

			if !exist {
				if currentOrder.Side == "BUY" {
					b.SetState(ClosePosition)
				} else {
					closed = true
					b.SetState(GetSignal)
				}
				b.logger.Printf("order finished %+v\n", currentOrder)
				continue
			}
			b.logger.Println("order exist", currentOrder.ID, currentOrder.Side, currentOrder.Price)

			// check order time
			if currentOrder.Side == "BUY" {
				if time.Since(currentOrder.OrderTime) > b.interval {
					b.logger.Println("order not completed: side=buy")
					b.SetState(CancelOrder)
				} else {
					time.Sleep(3 * time.Second)
				}
			} else if currentOrder.Side == "SELL" {
				if time.Since(currentOrder.OrderTime) > b.interval*10 {
					b.logger.Println("order not completed: side=sell")
					b.SetState(CancelOrder)
				} else {
					time.Sleep(b.interval / 3)
				}
			}

		case CancelOrder:
			err = b.exchange.CancelOrder(b.exCtx, currentOrder.Pair, currentOrder.ID)
			if err != nil {
				b.logger.Println(err)
				continue
			}
			b.SetState(GetSignal)
			b.logger.Println("order cancelled:", currentOrder.ID)
			if currentOrder.Side == "SELL" {
				closed = false
				b.SetState(ClosePosition)
			}
		case Nothing:
			time.Sleep(b.interval * 10)
			b.SetState(GetSignal)
		}
	}
}

func (b *Bot) createBuyOrder() (*exchange.Order, error) {
	rate, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		return nil, err
	}
	eps := rate * 0.0005
	// TODO add stop loss. When SELL order not created we need close position
	buyOrder := &exchange.Order{
		CreatedTime: time.Now(),
		OrderTime:   time.Now().Add(30 * time.Second),
		Pair:        b.pair,
		Type:        "LIMIT", // todo get from exchange
		Side:        "BUY",
		Price:       round(rate+eps, b.info.PricePrecision),
		Amount:      round(MinSum/(rate+eps), b.info.BasePrecision),
	}
	return buyOrder, nil
}

func (b *Bot) createMarketSellOrder(amount float64) (*exchange.Order, error) {
	rate1, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		return nil, err
	}
	rate2, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		return nil, err
	}
	// TODO
	if math.Abs(rate2 - rate1) > rate1 * 0.05 {
		return nil, errors.New("too expensive order when close position")
	}

	var price float64
	if rate1 > rate2 {
		price = rate1
	} else {
		price = rate2
	}
	//eps := price * 0.0005
	eps := 0.0
	sellOrder := &exchange.Order{
		CreatedTime: time.Now(),
		OrderTime:   time.Now().Add(30 * time.Second),
		Pair:        b.pair,
		Type:        "LIMIT", // todo get from exchange
		Side:        "SELL",
		Price:       round(price-eps, b.info.PricePrecision),
		Amount:      amount,
	}
	return sellOrder, nil
}

func (b *Bot) createSellOrder(boughtRate float64, boughtAmount float64) *exchange.Order {
	eps := boughtRate * 0.0025
	order := &exchange.Order{
		CreatedTime: time.Now(),
		OrderTime:   time.Now().Add(b.interval * 20), // todo OrderTime???
		Pair:        b.pair,
		Type:        "LIMIT", // todo get from exchange
		Side:        "SELL",
		Price:       round(boughtRate+eps, b.info.PricePrecision),
		//StopPrice:   round(boughtRate-2*eps, b.info.PricePrecision),
		Amount:      round(boughtAmount*0.999, b.info.BasePrecision),
	}
	return order
}
