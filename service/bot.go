package service

import (
	"context"
	"log"
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
)

var states = []string{"Start", "PopOrder", "CreateOrder", "CheckOrder", "WaitOrder", "Nothing"}

func (s State) String() string {
	return states[s]
}

type Bot struct {
	state    State
	queue    *OrderQueue
	exchange exchange.Exchanger
	logger   *log.Logger
}

func NewBot(queue *OrderQueue, ex exchange.Exchanger, logger *log.Logger) *Bot {
	return &Bot{
		queue:    queue,
		exchange: ex,
		logger:   logger,
	}
}

func (b *Bot) StartSM(ctx context.Context, msgChan <-chan string, infoChan chan<- string) {
	b.state = Start
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
				b.state = Nothing
				b.logger.Println("stopped")
			} else if cmd == CmdGetState {
				infoChan <- b.state.String()
			}
		default:
		}

		switch b.state {
		case Start:
			b.state = PopOrder

		case PopOrder:
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
				currentPrice, err := b.exchange.GetRate(currentOrder.Pair)
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

			b.logger.Printf("got order %+v\n", currentOrder)
			b.state = CreateOrder

		case CreateOrder:
			b.logger.Println("state", b.state)
			var (
				id  string
				err error
			)
			for attempts := 0; attempts < 3; attempts++ {
				id, err = b.exchange.CreateOrder(&currentOrder)
				currentOrder.ID = id
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				b.state = PopOrder
				b.logger.Println("can't create order", err)
				continue
			}

			b.logger.Printf("order created %+v\n", currentOrder)
			b.state = CheckOrder

		case CheckOrder:
			b.logger.Println("state", b.state)
			var (
				orders []exchange.Order
				err    error
			)
			for attempts := 0; attempts < 3; attempts++ {
				orders, err = b.exchange.OpenedOrders(currentOrder.Pair)
				if err == nil {
					break
				} else {
					b.logger.Println(err)
				}
			}
			if err != nil {
				ctx.Done()
				return
			}
			exist := false
			for _, order := range orders {
				if order.ID == currentOrder.ID {
					exist = true
				}
			}
			if exist {
				b.state = WaitOrder
			}
		case WaitOrder:
			b.logger.Println("state", b.state)
			var (
				orders []exchange.Order
				err    error
			)
			for attempts := 0; attempts < 3; attempts++ {
				orders, err = b.exchange.OpenedOrders(currentOrder.Pair)
				if err != nil {
					b.logger.Println(err)
					continue
				}
			}
			if err != nil {
				ctx.Done()
				return
			}
			if len(orders) == 0 {
				b.queue.Pop()
				b.state = PopOrder
				b.logger.Printf("order finished %+v\n", currentOrder)
				continue
			}
			time.Sleep(5 * time.Second)
		case Nothing:
			time.Sleep(10 * time.Second)
		}
	}
}
