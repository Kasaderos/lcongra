package service

import (
	"container/list"
	"sync"
)

type OrderQueue struct {
	lst *list.List
	mu  sync.Mutex
}

func (q *OrderQueue) Push(a Order) {
	q.mu.Lock()
	defer q.mu.Lock()
	q.lst.PushBack(a)
}

func (q *OrderQueue) Pop() {
	q.mu.Lock()
	defer q.mu.Lock()
	q.lst.Remove(q.lst.Front())
}

func (q *OrderQueue) Front() Order {
	q.mu.Lock()
	defer q.mu.Lock()
	order, _ := q.lst.Front().Value.(Order)
	return order
}
