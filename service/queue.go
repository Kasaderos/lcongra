package service

import (
	"container/list"
	"sync"

	"github.com/kasaderos/lcongra/exchange"
)

type OrderQueue struct {
	lst *list.List
	mu  sync.Mutex
}

func NewOrderQueue() *OrderQueue {
	return &OrderQueue{
		lst: list.New(),
	}
}

func (q *OrderQueue) Push(a *exchange.Order) {
	q.mu.Lock()
	defer q.mu.Lock()
	q.lst.PushBack(a)
}

func (q *OrderQueue) Pop() {
	q.mu.Lock()
	defer q.mu.Lock()
	q.lst.Remove(q.lst.Front())
}

func (q *OrderQueue) Front() *exchange.Order {
	q.mu.Lock()
	defer q.mu.Lock()
	order, _ := q.lst.Front().Value.(*exchange.Order)
	return order
}

func (q *OrderQueue) Empty() bool {
	return q.lst.Len() == 0
}
