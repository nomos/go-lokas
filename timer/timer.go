package timer

import (
	"sync"
	"time"
)

// type TypeEventChan func(...interface{})
type TypeEventChan interface{}

// 定时器接口
type Timer interface {
	// 一次性定时器
	// AfterFunc(expire time.Duration, callback func()) TimeNoder

	// 周期性定时器
	// ScheduleFunc(expire time.Duration, callback func()) TimeNoder

	// 运行
	// Run()

	// After(delay time.Duration, value interface{}) TimeNoder

	// Schedule(interal time.Duration, value interface{}, delay time.Duration, loop uint64) TimeNoder
	// Schedule(interal time.Duration, value interface{}) TimeNoder

	// 停止所有定时器
	Stop()

	Start()

	// TimeEventChan() <-chan interface{}

	NewHandler() TimeHandler
}

// 停止单个定时器
type TimeNoder interface {
	Stop()
}

type TimeHandler interface {
	EventChan() <-chan TypeEventChan

	After(delay time.Duration, cb func()) TimeNoder

	Schedule(interval time.Duration, cb func()) TimeNoder

	// 停止所有定时器
	Stop()

	// 删除handler
	DelSelf()

	PrintDebug()
}

type TimeEventMsg struct {
	Callback func()
}

var once sync.Once
var ins Timer

func NewTimer() Timer {
	return newTimeWheel()
}

func Instance() Timer {
	once.Do(func() {
		if ins == nil {
			ins = newTimeWheel()
			// go ins.Run()
			ins.Start()
		}
	})
	return ins
}

func NewHandler() TimeHandler {
	return Instance().NewHandler()
}

func Stop() {
	Instance().Stop()
}