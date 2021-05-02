package service

import (
	"testing"

	"github.com/kasaderos/lcongra/exchange"
)

func TestQueue(t *testing.T) {
	queue := NewOrderQueue()
	queue.Push(exchange.Order{ID: "A"})
	queue.Push(exchange.Order{ID: "B"})
	if queue.Empty() {
		t.Errorf("queue empty")
	}
	if queue.Front().ID != "A" {
		t.Errorf("first invalid")
	}
	queue.Pop()
	queue.Pop()
	if !queue.Empty() {
		t.Errorf("queue non empty")
	}
}
