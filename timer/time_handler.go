package timer

import (
	"fmt"
	"sync"
	"time"
)

type timeHandler struct {
	key       string
	eventChan chan TypeEventChan

	wheel  *timeWheel
	noders sync.Map
}

func (t *timeHandler) After(delay time.Duration, cb func(...interface{})) TimeNoder {

	tn := t.wheel.After(delay, cb)

	t.noders.Store(&tn, 1)

	return tn
}

func (t *timeHandler) Schedule(interval time.Duration, cb func(...interface{})) TimeNoder {
	tn := t.wheel.Schedule(interval, cb)

	t.noders.Store(&tn, 1)

	return tn
}

func (t *timeHandler) EventChan() <-chan TypeEventChan {
	return t.eventChan
}

func (t *timeHandler) DelSelf() {

	// delete time event
	t.noders.Range(func(key, value any) bool {
		node := key.(timeNode)
		node.Stop()
		return true
	})

	// close channel
	close(t.eventChan)
	t.eventChan = nil

	// delete wheel handler map
	t.wheel.handlers.Delete(t.key)
	t.wheel = nil

}

func (t *timeHandler) PrintDebug() {

	i := 0

	t.noders.Range(func(key, value any) bool {
		i++
		return true
	})

	fmt.Println("handler info, key:", t.key, " nodeCnt:", i)
}
