package timer

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type TimerOption func(node *timeNode)

type timeHandler struct {
	key       uint64
	eventChan chan TypeEventChan

	wheel  *timeWheel
	noders sync.Map
}

func WithDelay(delay time.Duration) TimerOption {
	return func(node *timeNode) {
		node.delay = uint64(delay)
	}
}

func WithLoop(loop uint64) TimerOption {
	return func(node *timeNode) {
		node.loopMax = loop
	}
}

func (t *timeHandler) After(delay time.Duration, cb func(TimeNoder)) TimeNoder {
	// jiffies := atomic.LoadUint64(&t.wheel.jiffies)

	// expire := delay/(time.Millisecond*10) + time.Duration(jiffies)

	// node := &timeNode{
	// 	expire: uint64(expire),

	// 	callback: cb,

	// 	delay:    uint64(delay),
	// 	interval: uint64(0),
	// 	loopMax:  1,
	// 	handler:  t,
	// }

	// tn := t.wheel.add(node, jiffies)

	// t.noders.Store(tn, 1)

	// return tn

	return t.Schedule(delay, cb, WithLoop(1))
}

func (t *timeHandler) Schedule(interval time.Duration, cb func(TimeNoder), opts ...TimerOption) TimeNoder {
	// tn := t.wheel.Schedule(interval, cb)

	jiffies := atomic.LoadUint64(&t.wheel.jiffies)

	// expire := getExpire(interval, jiffies)

	node := &timeNode{
		expire:   0,
		callback: cb,
		delay:    uint64(0),
		interval: uint64(interval),
		loopCur:  0,
		loopMax:  0,
		handler:  t,
	}

	for _, v := range opts {
		v(node)
	}

	if node.delay <= 0 {
		node.delay = 0
	}

	if node.delay > 0 {
		node.expire = node.delay/(uint64(time.Millisecond*10)) + jiffies
	} else {
		node.expire = node.interval/(uint64(time.Millisecond*10)) + jiffies
	}

	tn := t.wheel.add(node, jiffies)

	t.noders.Store(tn, 1)

	return tn
}

func (t *timeHandler) EventChan() <-chan TypeEventChan {
	return t.eventChan
}

func (t *timeHandler) DelTimer() {

	// delete time event
	t.StopTimer()

	// close channel
	close(t.eventChan)
	t.eventChan = nil

	// delete wheel handler map
	t.wheel.handlers.Delete(t.key)
	t.wheel = nil

}

func (t *timeHandler) StopTimer() {
	t.noders.Range(func(key, value any) bool {
		node := key.(*timeNode)
		node.Stop()
		return true
	})
}

func (t *timeHandler) PrintDebug() {

	i := 0

	t.noders.Range(func(key, value any) bool {
		i++
		return true
	})

	log.Println("handler info, key:", t.key, " nodeCnt:", i)
}
