package conn

import (
	"container/list"
)

func NewMessageChan(useNoneBlocking bool, size int, done chan struct{}) MessageChan {
	if size == 0 {
		size = 100
	}
	if useNoneBlocking {
		return NewNonBlockingMessageChan(size, done)
	} else {
		return NewDefaultMessageChan(size, done)
	}
}

type DefaultMessageChan struct {
	In   chan<- []byte
	Out  <-chan []byte
	done chan struct{}
}

func NewDefaultMessageChan(size int, done chan struct{}) *DefaultMessageChan {
	if size < 10 {
		size = 10
	}
	msgChan := make(chan []byte, size)
	return &DefaultMessageChan{
		In:   msgChan,
		Out:  msgChan,
		done: done,
	}
}

func (this *DefaultMessageChan) GetInChan() chan<- []byte {
	return this.In
}

func (this *DefaultMessageChan) GetOutChan() <-chan []byte {
	return this.Out
}

func (this *DefaultMessageChan) Len() int {
	return len(this.In)
}

// reference: go-nonblockingchan

// Special type that mimics the behavior of a channel but does not block when
// items are sent. Items are stored internally until received. Closing the Send
// channel will cause the Recv channel to be closed after all items have been
// received.
type NonBlockingMessageChan struct {
	*DefaultMessageChan
	items     *list.List
	itemCount int
}

// OnCreate a new non-blocking channel.
func NewNonBlockingMessageChan(size int, done chan struct{}) *NonBlockingMessageChan {
	if size < 10 {
		size = 10
	}
	var in = make(chan []byte, size)
	var out = make(chan []byte, size)
	var n = &NonBlockingMessageChan{
		DefaultMessageChan: &DefaultMessageChan{
			In:   in,
			Out:  out,
			done: done,
		},
		items: list.New(),
	}
	go n.run(in, out)
	return n
}

// Loop for buffering items between the Send and Recv channels until the Send
// channel is closed.
func (this *NonBlockingMessageChan) run(in <-chan []byte, out chan<- []byte) {
	for {
		if in == nil && this.items.Len() == 0 {
			close(out)
			break
		}
		var (
			outChan chan<- []byte
			outVal  []byte
		)
		if this.items.Len() > 0 {
			outChan = out
			outVal = this.items.Front().Value.([]byte)
		}
		select {
		case i, ok := <-in:
			if ok {
				this.items.PushBack(i)
				this.itemCount++
			} else {
				in = nil
			}
		case outChan <- outVal:
			this.items.Remove(this.items.Front())
			this.itemCount--
		case <-this.done:
			return
		}
	}
}

// Retrieve the number of items waiting to be received.
func (this *NonBlockingMessageChan) Len() int {
	return this.itemCount
}
