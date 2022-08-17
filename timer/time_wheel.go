package timer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	nearShift = 8

	nearSize = 1 << nearShift

	levelShift = 6

	levelSize = 1 << levelShift

	nearMask = nearSize - 1

	levelMask = levelSize - 1
)

type timeWheel struct {
	// 单调递增累加值, 走过一个时间片就+1
	jiffies uint64

	// 256个槽位
	t1 [nearSize]*Time

	// 4个64槽位, 代表不同的刻度
	t2Tot5 [4][levelSize]*Time

	// 时间只精确到10ms
	// curTimePoint 为1就是10ms 为2就是20ms
	curTimePoint time.Duration

	// 上下文
	ctx context.Context

	// 取消函数
	cancel context.CancelFunc

	// timeEvent chan interface{}

	handlers sync.Map

	handlerSeq uint64
}

func newTimeWheel() *timeWheel {

	ctx, cancel := context.WithCancel(context.Background())

	t := &timeWheel{
		ctx:        ctx,
		cancel:     cancel,
		handlerSeq: 0,
	}

	t.init()

	return t
}

func (t *timeWheel) Now() time.Time {
	return time.UnixMilli(int64(t.curTimePoint * 10)).Local()
}

func (t *timeWheel) init() {

	// t.timeEvent = make(chan interface{})

	for i := 0; i < nearSize; i++ {
		t.t1[i] = newTimeHead(1, uint64(i))

	}

	for i := 0; i < 4; i++ {
		for j := 0; j < levelSize; j++ {
			t.t2Tot5[i][j] = newTimeHead(uint64(i+2), uint64(j))
		}
	}

	t.curTimePoint = get10Ms()
}

func maxVal() uint64 {
	return (1 << (nearShift + 4*levelShift)) - 1
}

func levelMax(index int) uint64 {
	return 1 << (nearShift + index*levelShift)
}

func (t *timeWheel) index(n int) uint64 {
	return (t.jiffies >> (nearShift + levelShift*n)) & levelMask
}

func (t *timeWheel) add(node *timeNode, jiffies uint64) *timeNode {

	var head *Time
	expire := node.expire
	idx := expire - jiffies

	level, index := uint64(1), uint64(0)

	if idx < nearSize {

		index = uint64(expire) & nearMask
		head = t.t1[index]

	} else {

		max := maxVal()
		for i := 0; i <= 3; i++ {

			if idx > max {
				idx = max
				expire = idx + jiffies
			}

			if uint64(idx) < levelMax(i+1) {
				index = uint64(expire >> (nearShift + i*levelShift) & levelMask)
				head = t.t2Tot5[i][index]
				level = uint64(i) + 2
				break
			}
		}
	}

	if head == nil {
		panic("not found head")
	}

	head.lockPushBack(node, level, index)

	return node
}

// func (t *timeWheel) AfterFunc(expire time.Duration, callback func()) TimeNoder {

// 	jiffies := atomic.LoadUint64(&t.jiffies)

// 	expire = expire/(time.Millisecond*10) + time.Duration(jiffies)

// 	node := &timeNode{
// 		expire:   uint64(expire),
// 		callback: callback,
// 	}

// 	return t.add(node, jiffies)
// }

func getExpire(expire time.Duration, jiffies uint64) time.Duration {
	return expire/(time.Millisecond*10) + time.Duration(jiffies)
}

// func (t *timeWheel) ScheduleFunc(userExpire time.Duration, callback func()) TimeNoder {

// 	jiffies := atomic.LoadUint64(&t.jiffies)

// 	expire := getExpire(userExpire, jiffies)

// 	node := &timeNode{
// 		userExpire: userExpire,
// 		expire:     uint64(expire),
// 		callback:   callback,
// 		isSchedule: true,
// 	}

// 	return t.add(node, jiffies)
// }

// func (t *timeWheel) After(delay time.Duration, cb func()) TimeNoder {
// 	jiffies := atomic.LoadUint64(&t.jiffies)

// 	expire := delay/(time.Millisecond*10) + time.Duration(jiffies)

// 	node := &timeNode{
// 		expire: uint64(expire),

// 		callback: cb,
// 	}

// 	return t.add(node, jiffies)
// }

// func (t *timeWheel) Schedule(interval time.Duration, cb func()) TimeNoder {
// 	jiffies := atomic.LoadUint64(&t.jiffies)

// 	expire := getExpire(interval, jiffies)

// 	node := &timeNode{
// 		userExpire: interval,
// 		expire:     uint64(expire),
// 		callback:   cb,
// 		isSchedule: true,

// 		delay:    uint64(0),
// 		interval: uint64(interval),
// 		loop:     0,
// 	}

// 	return t.add(node, jiffies)
// }

func (t *timeWheel) Stop() {
	t.cancel()
}

func (t *timeWheel) NewHandler() TimeHandler {

	th := &timeHandler{
		key:       atomic.AddUint64(&t.handlerSeq, 1),
		eventChan: make(chan TypeEventChan),
		wheel:     t,
	}

	value, _ := t.handlers.LoadOrStore(th.key, th)

	return value.(TimeHandler)
}

func (t *timeWheel) cascade(levelIndex int, index int) {

	tmp := newTimeHead(0, 0)

	l := t.t2Tot5[levelIndex][index]
	l.Lock()
	if l.Len() == 0 {
		l.Unlock()
		return
	}

	l.ReplaceInit(&tmp.Head)

	atomic.AddUint64(&l.version, 1)
	l.Unlock()

	offset := unsafe.Offsetof(tmp.Head)
	tmp.ForEachSafe(func(pos *Head) {
		node := (*timeNode)(pos.Entry(offset))
		t.add(node, atomic.LoadUint64(&t.jiffies))
	})

}

func (t *timeWheel) moveAndExec() {

	// 这里时间溢出
	if uint32(t.jiffies) == 0 {
		// TODO
		// return
	}

	//如果本层的盘子没有定时器，这时候从上层的盘子移动一些过来
	index := t.jiffies & nearMask
	if index == 0 {
		for i := 0; i <= 3; i++ {
			index2 := t.index(i)
			t.cascade(i, int(index2))
			if index2 != 0 {
				break
			}
		}
	}

	atomic.AddUint64(&t.jiffies, 1)

	t.t1[index].Lock()
	if t.t1[index].Len() == 0 {
		t.t1[index].Unlock()
		return
	}

	head := newTimeHead(0, 0)
	t1 := t.t1[index]
	t1.ReplaceInit(&head.Head)
	atomic.AddUint64(&t1.version, 1)
	t.t1[index].Unlock()

	// 执行,链表中的定时器
	offset := unsafe.Offsetof(head.Head)

	head.ForEachSafe(func(pos *Head) {
		val := (*timeNode)(pos.Entry(offset))
		head.Del(pos)

		val.loopCur++
		var interval uint64
		var isLoop bool
		if !val.isCron {
			interval, isLoop = val.intervalExpireFunc()
		} else {
			interval, isLoop = val.cronExpireFunc(t)
		}
		val.interval = interval
		if !isLoop {
			val.handler.noders.Delete(val)
		}

		if atomic.LoadUint32(&val.stop) == haveStop {
			return
		}

		msg := &TimeEventMsg{
			Callback:  val.callback,
			TimeNoder: val,
		}
		val.handler.eventChan <- msg
		if isLoop {
			jiffies := t.jiffies
			// 这里的jiffies必须要减去1
			// 当前的callback被调用，已经包含一个时间片,如果不把这个时间片减去，
			// 每次多一个时间片，就变成累加器, 最后周期定时器慢慢会变得不准

			// val.expire = uint64(getExpire(val.userExpire, jiffies-1))
			val.expire = val.interval/(uint64(time.Millisecond)*10) + jiffies - 1
			t.add(val, jiffies)
		}

	})

}

// get10Ms函数通过参数传递，为了方便测试
func (t *timeWheel) run(get10Ms func() time.Duration) {
	// 先判断是否需要更新
	// 内核里面实现使用了全局jiffies和本地的jiffies比较,应用层没有jiffies，直接使用时间比较
	// 这也是skynet里面的做法

	ms10 := get10Ms()

	if ms10 < t.curTimePoint {

		fmt.Printf("timer:Time has been called back?from(%d)(%d)\n",
			ms10, t.curTimePoint)

		t.curTimePoint = ms10
		return
	}

	diff := ms10 - t.curTimePoint
	t.curTimePoint = ms10

	for i := 0; i < int(diff); i++ {
		t.moveAndExec()
	}

}

func (t *timeWheel) Run() {

	// 10ms精度
	tk := time.NewTicker(time.Millisecond * 10)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			t.run(get10Ms)
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *timeWheel) Start() {

	go func() {
		tk := time.NewTicker(time.Millisecond * 10)
		defer tk.Stop()

		for {
			select {
			case <-tk.C:
				t.run(get10Ms)

			case <-t.ctx.Done():
				return

			}
		}
	}()
}
