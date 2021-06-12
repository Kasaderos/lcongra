package service

import (
	"container/list"
	"sync"
)

type queue struct {
	sync.Mutex
	lst *list.List
}

type Message struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

func (q *queue) Push(msg Message) {
	q.lst.PushBack(msg)
}

func (q *queue) Pop() {
	q.lst.Remove(q.lst.Front())
}

func (q *queue) Front() Message {
	fr := q.lst.Front()
	f := fr.Value.(Message)
	return f
}

func (q *queue) Empty() bool {
	return q.lst.Len() == 0
}

type MQ struct {
	sync.RWMutex
	queues map[string]*queue
}

func NewMQ() *MQ {
	return &MQ{
		queues: make(map[string]*queue),
	}
}

func (mq *MQ) Send(msg Message) {
	mq.Lock()
	defer mq.Unlock()
	mq.queues[msg.ID].Push(msg)
}

func (mq *MQ) Receive(id string) Message {
	mq.Lock()
	defer mq.Unlock()
	q := mq.queues[id]
	if q.Empty() {
		return Message{id, "no message"}
	}
	msg := q.Front()
	q.Pop()
	return msg
}

func (mq *MQ) AddQueue(id string) {
	mq.Lock()
	defer mq.Unlock()
	mq.queues[id] = &queue{}
}

func (mq *MQ) Delete(id string) {
	mq.Lock()
	defer mq.Unlock()
	delete(mq.queues, id)
}
