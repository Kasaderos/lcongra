package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/lcongra/exchange"
	ex "github.com/kasaderos/lcongra/exchange/fake"
)

type Agent struct {
	*MQ
	ID            string
	baseCurrency  string
	quoteCurrency string
	interval      string
	bot           *Bot
	queue         *OrderQueue
	exchange      exchange.Exchanger // TODO make just api without any object
}

func (ag *Agent) Run(ctx context.Context) {
	msgChan := make(chan string)
	infoChan := make(chan string)

	tradeCtx, cancel := context.WithCancel(context.Background())
	// for fake exchange, we need
	go ex.Update(tradeCtx, ag.exchange)

	go ag.bot.StartSM(tradeCtx, msgChan, infoChan)
	go Autotrade(
		tradeCtx,
		fmt.Sprintf("%s-%s", ag.baseCurrency, ag.quoteCurrency),
		ag.interval,
		ag.bot.queue,
		ag.exchange,
	)

	for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case msg := <-infoChan:
			ag.MQ.Send(Message{"root", msg})
		default:
		}

		msg := ag.MQ.Receive(ag.ID)
		switch msg.Data {
		case CmdDelete:
			cancel()
			return
		default:
			msgChan <- msg.Data
		}
		time.Sleep(10 * time.Second)
	}
}
