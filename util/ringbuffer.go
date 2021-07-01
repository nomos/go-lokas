package util

import (
	"fmt"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/slice"
	"sync"
)

type RingBuffer struct {
	elements []interface{}
	first    int
	last     int
	size     int
	capacity int
	end      int
	mutex    sync.RWMutex
}

func NewRingBuffer(capacity int) *RingBuffer {
	ret := &RingBuffer{
		elements: make([]interface{}, capacity, capacity),
		first:    0,
		last:     0,
		size:     0,
		capacity: capacity,
		end:      0,
		mutex:    sync.RWMutex{},
	}
	return ret
}

func (this *RingBuffer) Capacity() int {
	return this.capacity
}

func (this *RingBuffer) IsEmpty() bool {
	return this.size == 0
}

func (this *RingBuffer) IsFull() bool {
	return this.size == this.Capacity()
}

func (this *RingBuffer) Peek() interface{} {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if this.IsEmpty() {
		log.Warn("isEmpty")
		return nil
	}
	return this.elements[this.first]
}

func (this *RingBuffer) PeekN(count int) interface{} {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if count > this.size {
		count = this.size
	}
	end := Ternary((this.first+count) > this.capacity, this.capacity, this.first+count).(int)
	firstHalf := this.elements[this.first:end]
	if end < this.capacity {
		return firstHalf
	}
	secondHalf := this.elements[0 : count-len(firstHalf)]
	return slice.SliceConcat(secondHalf)
}

func (this *RingBuffer) Deq() interface{} {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	element := this.Peek()
	if element != nil {
		this.size--
		this.first = (this.first + 1) % this.capacity
		return element
	}
	return nil
}

func (this *RingBuffer) DeqN(count int) interface{} {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	elements := this.PeekN(count)
	count = len(elements.([]interface{}))
	this.size-=count
	this.first = (this.first+count)%this.capacity
	return nil
}

func (this *RingBuffer) Enq(element interface{})int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.end = (this.first+this.size)%this.capacity
	full:=this.IsFull()
	if full {
		fmt.Println("RingBuffer is Full")
	}
	this.elements[this.end] = element
	if full {
		this.first = (this.first+1)%this.capacity
	} else {
		this.size++
	}
	return this.size
}

func (this RingBuffer) Size()int {
	return this.size
}