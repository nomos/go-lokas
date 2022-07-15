package timer

import (
	"sync"
	"time"
)

// 定时器接口
type Timer interface {
	// 一次性定时器
	// AfterFunc(expire time.Duration, callback func()) TimeNoder

	// 周期性定时器
	// ScheduleFunc(expire time.Duration, callback func()) TimeNoder

	// 运行
	// Run()

	After(delay time.Duration, value interface{}) TimeNoder

	// Schedule(interal time.Duration, value interface{}, delay time.Duration, loop uint64) TimeNoder
	Schedule(interal time.Duration, value interface{}) TimeNoder

	// 停止所有定时器
	Stop()

	Start()

	TimeEventChan() <-chan interface{}
}

// 停止单个定时器
type TimeNoder interface {
	Stop()
}

var once sync.Once
var ins Timer

// 定时器构造函数
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
