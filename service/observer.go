package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kasaderos/lcongra/exchange"
)

type Observer struct {
	exchange exchange.Exchanger
	pairs    []string
	chanMsg  chan<- string
	logger   *log.Logger
}

func NewObserver(exchange exchange.Exchanger, logger *log.Logger, chanMsg chan<- string, pairs []string) *Observer {
	return &Observer{exchange, pairs, chanMsg, logger}
}

func (b *Observer) Observe(ctx context.Context) {
	b.logger.Println("started")
	for {
		select {
		case <-ctx.Done():
			b.logger.Println("stopped")
		default:
		}
		for _, pair := range b.pairs {
			rate, err := b.exchange.GetRate(context.Background(), pair)
			if err != nil {
				b.logger.Println(pair, err)
				continue
			}
			msg := fmt.Sprintf("%s=%f", pair, rate)
			b.logger.Println(msg)
			b.chanMsg <- msg
		}
		time.Sleep(10 * time.Second)
	}
}
