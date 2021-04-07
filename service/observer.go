package service

import (
	"log"

	"github.com/kasaderos/lcongra/exchange"
)

type Observer struct {
}

func NewObserver(exchange exchange.Exchanger, logger *log.Logger, chanMsg chan<- string) *Observer {
	return &Observer{}
}

func (b *Observer) Observe() {

}
