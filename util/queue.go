package util

import "sync"

type Queue struct {
	first *node
	last  *node
	n     int
	mutex sync.Mutex
}

type node struct {
	item interface{}
	next *node
}

func NewQueue() Queue {
	return Queue{}
}

func (q Queue) IsEmpty() bool {
	return q.n == 0
}

func (q Queue) Size() int {
	return q.n
}

func (q *Queue) EnQueue(item interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	oldlast := q.last
	q.last = &node{}
	q.last.item = item
	q.last.next = nil
	if q.IsEmpty() {
		q.first = q.last
	} else {
		oldlast.next = q.last
	}
	q.n++
}

func (q *Queue) DeQueue() interface{} {
	if q.IsEmpty() {
		return nil
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	item := q.first.item
	q.first = q.first.next
	if q.IsEmpty() {
		q.last = nil
	}
	q.n--
	return item
}