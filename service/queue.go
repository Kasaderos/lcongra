package service

import (
	"sync"

	"github.com/kasaderos/lcongra/exchange"
)

type OrderQueue struct {
	list []exchange.Order
	mu   sync.RWMutex
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{
		list: make([]exchange.Order, 0, 10),
	}
}

func (q *OrderQueue) Push(a exchange.Order) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.list = append(q.list, a)
}

func (q *OrderQueue) Pop() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.list = q.list[1:]
}

func (q *OrderQueue) Front() exchange.Order {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.list[0]
}

func (q *OrderQueue) Empty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.list) == 0
}
