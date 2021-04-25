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
)

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

func (b *Bot) StartSM(ctx context.Context) {
	b.state = Start
	var currentOrder exchange.Order
SM:
	for {
		select {
		case <-ctx.Done():
			break SM
		default:
		}

		switch b.state {
		case Start:
			b.state = PopOrder

		case PopOrder:
			if b.queue.Empty() {
				continue
			}
			now := time.Now()
			currentOrder = b.queue.Front()
			// if open position expired we decline creating order
			if currentOrder.Side == "BUY" && now.After(currentOrder.OrderTime) {
				b.queue.Pop()
				continue
			}

			// if sell expired then we sell by current price
			if currentOrder.Side == "SELL" && now.After(currentOrder.OrderTime) {
				currentPrice, err := b.exchange.GetRate(currentOrder.Pair)
				if err != nil {
					b.logger.Println("close expired position failed", err)
					b.queue.Pop()
					continue
				}
				currentOrder.Price = currentPrice
			}

			if !currentOrder.OrderTime.IsZero() && currentOrder.OrderTime.Before(now) {
				time.Sleep(now.Sub(currentOrder.OrderTime))
			}

			b.queue.Pop()
			b.state = CreateOrder

		case CreateOrder:
			id, err := b.exchange.CreateOrder(&currentOrder)
			currentOrder.ID = id
			if err != nil {
				b.logger.Println(err)
				continue
			}
			b.state = CheckOrder

		case CheckOrder:
			orders, err := b.exchange.OpenedOrders(currentOrder.Pair)
			if err != nil {
				b.logger.Println(err)
				continue
			}
			for _, order := range orders {
				if order.ID == currentOrder.ID {
					b.state = PopOrder
					continue SM
				}
			}
			b.state = WaitOrder
		case WaitOrder:
			orders, err := b.exchange.OpenedOrders(currentOrder.Pair)
			if err != nil {
				b.logger.Println(err)
				continue
			}
			if len(orders) == 0 {
				b.state = PopOrder
			}
		}
	}
}
