package util

import (
	"fmt"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/slice"
	"sync"
)

type RingBuffer[T any] struct {
	elements []T
	first    int
	last     int
	size     int
	capacity int
	end      int
	mutex    sync.RWMutex
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	ret := &RingBuffer[T]{
		elements: make([]T, capacity, capacity),
		first:    0,
		last:     0,
		size:     0,
		capacity: capacity,
		end:      0,
		mutex:    sync.RWMutex{},
	}
	return ret
}

func (this *RingBuffer[T]) Capacity() int {
	return this.capacity
}

func (this *RingBuffer[T]) IsEmpty() bool {
	return this.size == 0
}

func (this *RingBuffer[T]) IsFull() bool {
	return this.size == this.Capacity()
}

func (this *RingBuffer[T]) Peek() T {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if this.IsEmpty() {
		log.Warn("isEmpty")
		return Nil[T]()
	}
	return this.elements[this.first]
}

func (this *RingBuffer[T]) PeekN(count int) []T {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if count > this.size {
		count = this.size
	}
	end := Ternary((this.first+count) > this.capacity, this.capacity, this.first+count)
	firstHalf := this.elements[this.first:end]
	if end < this.capacity {
		return firstHalf
	}
	secondHalf := this.elements[0 : count-len(firstHalf)]
	return slice.Concat(secondHalf)
}

func (this *RingBuffer[T]) Deq() T {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	element := this.Peek()
	if !IsNil(element) {
		this.size--
		this.first = (this.first + 1) % this.capacity
		return element
	}
	return Nil[T]()
}

func (this *RingBuffer[T]) DeqN(count int) T {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	elements := this.PeekN(count)
	count = len(elements)
	this.size -= count
	this.first = (this.first + count) % this.capacity
	return Nil[T]()
}

func (this *RingBuffer[T]) Enq(element T) int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.end = (this.first + this.size) % this.capacity
	full := this.IsFull()
	if full {
		fmt.Println("RingBuffer is Full")
	}
	this.elements[this.end] = element
	if full {
		this.first = (this.first + 1) % this.capacity
	} else {
		this.size++
	}
	return this.size
}

func (this RingBuffer[T]) Size() int {
	return this.size
}
