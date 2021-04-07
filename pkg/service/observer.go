package service

import (
	"github.com/kasaderos/lcongra/pkg/exchange"
	"go.uber.org/zap"
)

type Observer struct {
}

func NewObserver(exchange exchange.Exchanger, logger *zap.Logger, chanMsg chan<- string) *Observer {
	return &Observer{}
}

func (b *Observer) Observe() {

}
