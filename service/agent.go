package service

import (
	"context"
	"sync"
)

type Agent struct {
	// *MQ
	// READ ONLY
	mu            sync.RWMutex
	ID            string
	baseCurrency  string
	quoteCurrency string
	interval      string
	bot           *Bot
	queue         *OrderQueue
	ctx           context.Context
	cancel        context.CancelFunc

	apikey    string
	apisecret string
}

type AgentInfo struct {
	ID       string  `json:"id"`
	Pair     string  `json:"pair"`
	Interval string  `json:"interval"`
	State    string  `json:"state"`
	Cache    float64 `json:"cache"`
	// TODO
}
