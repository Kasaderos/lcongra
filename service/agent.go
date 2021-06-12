package service

import (
	"context"
	"sync"

	"github.com/kasaderos/lcongra/exchange"
)

type Agent struct {
	*MQ
	// READ ONLY
	mu            sync.RWMutex
	ID            string
	baseCurrency  string
	quoteCurrency string
	interval      string
	bot           *Bot
	queue         *OrderQueue
	exchange      exchange.Exchanger // TODO make just api without any object
	tradeCtx      context.Context
	cancel        context.CancelFunc
}

type AgentInfo struct {
	ID       string  `json:"id"`
	Pair     string  `json:"pair"`
	Interval string  `json:"interval"`
	State    string  `json:"state"`
	Cache    float64 `json:"cache"`
	// TODO
}
