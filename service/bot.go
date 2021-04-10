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
	var currentOrder *exchange.Order
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
			if currentOrder.OrderTime.Before(now) {
				time.Sleep(now.Sub(currentOrder.OrderTime))
			}
			b.queue.Pop()
			b.state = CreateOrder

		case CreateOrder:
			id, err := b.exchange.CreateOrder(currentOrder)
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
		}
	}
}
