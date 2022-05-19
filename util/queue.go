package util

import "sync"

type Queue[T any] struct {
	first *node[T]
	last  *node[T]
	n     int
	mutex sync.Mutex
}

type node[T any] struct {
	item T
	next *node[T]
}

func NewQueue[T any]() Queue[T] {
	return Queue[T]{}
}

func (q Queue[T]) IsEmpty() bool {
	return q.n == 0
}

func (q Queue[T]) Size() int {
	return q.n
}

func (q *Queue[T]) EnQueue(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	oldlast := q.last
	q.last = &node[T]{}
	q.last.item = item
	q.last.next = nil
	if q.IsEmpty() {
		q.first = q.last
	} else {
		oldlast.next = q.last
	}
	q.n++
}

func (q *Queue[T]) DeQueue() T {
	var t T
	if q.IsEmpty() {
		return t
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
