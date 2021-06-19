package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/kasaderos/lcongra/exchange"
)

type State int

const (
	Start State = iota
	PopOrder
	CreateOrder
	CheckOrder
	WaitOrder
	Nothing
	CancelOrder
	ClosePositions
)

const (
	AttemptsNumber = 3
	MinSum         = 11
)

var states = []string{"Start", "PopOrder", "CreateOrder", "CheckOrder", "WaitOrder", "Nothing"}

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

	queue    *OrderQueue
	exchange exchange.Exchanger
	logger   *log.Logger

	exCtx context.Context
}

func NewBot(queue *OrderQueue, ex exchange.Exchanger, logger *log.Logger, pair string, exCtx context.Context) *Bot {
	return &Bot{
		queue:    queue,
		exchange: ex,
		logger:   logger,
		pair:     pair,
		exCtx:    exCtx,
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
		b.logger.Println(err)
		return 0.0, 0.0, 0.0
	}
	quoteAmount, err := b.exchange.GetBalance(b.exCtx, quote)
	if err != nil {
		b.logger.Println(err)
		return 0.0, 0.0, 0.0
	}
	price, err := b.exchange.GetRate(b.exCtx, b.pair)
	if err != nil {
		b.logger.Println(err)
		return 0.0, 0.0, 0.0
	}
	b.ms.Lock()
	defer b.ms.Unlock()
	baseSum = price * baseAmount
	b.cache = baseSum + quoteAmount
	return baseSum, quoteAmount, b.cache
}

func (b *Bot) StartSM(ctx context.Context, msgChan <-chan string) {
	b.SetState(Start)
	var currentOrder exchange.Order
	b.logger.Println("SM started")
SM:
	for {
		// b.logger.Println("state", b.state)
		select {
		case <-ctx.Done():
			b.logger.Println("deleted")
			break SM
		case cmd := <-msgChan:
			if cmd == CmdStop {
				b.SetState(Nothing)
				b.logger.Println("stopped")
			}
		default:
		}

		switch b.state {
		case Start:
			b.state = PopOrder

		case PopOrder:
			b.logger.Println("state", b.state)
			if b.queue.Empty() {
				// b.logger.Println("queue empty")
				time.Sleep(time.Second * 10)
				continue
			}
			now := time.Now()
			currentOrder = b.queue.Front()
			// if open position expired we decline creating order
			if currentOrder.Side == "BUY" && now.After(currentOrder.OrderTime) {
				b.logger.Println("BUY order time after now")
				b.queue.Pop()
				continue
			}

			// if sell expired then we sell by current price
			if currentOrder.Side == "SELL" && now.After(currentOrder.OrderTime) {
				currentPrice, err := b.exchange.GetRate(b.exCtx, currentOrder.Pair)
				if err != nil {
					b.logger.Println("close expired position failed", err)
					continue
				}
				b.logger.Println("SELL order time after now")
				currentOrder.Price = currentPrice
			}

			if !currentOrder.OrderTime.IsZero() && currentOrder.OrderTime.Before(now) {
				b.logger.Println("order time befor now, sleep")
				time.Sleep(now.Sub(currentOrder.OrderTime))
			}

			// b.logger.Printf("got order %+v\n", currentOrder)
			b.SetState(CreateOrder)

		case CreateOrder:
			b.logger.Println("state", b.state)
			var (
				id  string
				err error
			)
			_, quote, _ := b.GetCache()
			if currentOrder.Side == "BUY" && quote < MinSum {
				b.logger.Println("not enough money in balance")
				b.SetState(Nothing)
				continue
			}
			for attempts := 0; attempts < AttemptsNumber; attempts++ {
				id, err = b.exchange.CreateOrder(b.exCtx, &currentOrder)
				currentOrder.ID = id
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				b.SetState(ClosePositions)
				b.logger.Println("can't create order", err)
				continue
			}

			b.logger.Printf("order created %+v\n", currentOrder)
			b.SetState(CheckOrder)

		case CheckOrder:
			b.logger.Println("state", b.state)
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
			if exist {
				b.SetState(WaitOrder)
			}
		case WaitOrder:
			// b.logger.Println("state", b.state)
			var (
				orders []exchange.Order
				err    error
			)
			for attempts := 0; attempts < AttemptsNumber; attempts++ {
				orders, err = b.exchange.OpenedOrders(b.exCtx, currentOrder.Pair)
				if err != nil {
					b.logger.Println(err)
					continue
				}
			}
			if err != nil {
				ctx.Done()
				return
			}

			if currentOrder.Side == "BUY" && time.Now().Sub(currentOrder.OrderTime) > b.interval/2 {
				b.logger.Println("order not completed: side=buy")
				b.SetState(CancelOrder)
				continue
			}

			if currentOrder.Side == "BUY" && time.Now().Sub(currentOrder.OrderTime) > b.interval/2 {
				b.logger.Println("order not created: side=buy")
				b.SetState(CancelOrder)
				continue
			}
			if len(orders) == 0 {
				b.queue.Pop()
				b.SetState(PopOrder)
				b.logger.Printf("order finished %+v\n", currentOrder)
				continue
			}

			time.Sleep(5 * time.Second)
		case CancelOrder:
			// cancel BUY, and next SELL
			rate, _ := b.exchange.GetRate(b.exCtx, b.pair)
			if rate > 1e-3 {
				err := b.exchange.CancelOrder(b.exCtx, currentOrder.Pair, currentOrder.ID)
				if err != nil {
					b.logger.Println(err)
					b.SetState(PopOrder)
					continue
				}
				b.logger.Println("order cancelled:", currentOrder.ID)
				// pop SELL order
				if !b.queue.Empty() {
					currentOrder = b.queue.Front()
					if currentOrder.Side == "SELL" {
						b.queue.Pop()
					}
					b.SetState(PopOrder)
				}
			}
		case Nothing:
			time.Sleep(10 * time.Second)
		}
	}
}
