package timer

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
	"strconv"
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

func (this *timeHandler) At(t time.Time, cb func(TimeNoder)) TimeNoder {
	delay := t.Sub(this.wheel.Now())
	return this.Schedule(delay, cb, WithLoop(1))
}

func (this *timeHandler) After(delay time.Duration, cb func(TimeNoder)) TimeNoder {
	return this.Schedule(delay, cb, WithLoop(1))
}

func convertIntToString(i interface{}) (string, error) {
	if util.IsString(i) {
		return i.(string), nil
	}
	if util.IsInt(i) {
		return strconv.Itoa(i.(int)), nil
	}
	if util.IsBool(i) {
		if i.(bool) {
			return "*", nil
		} else {
			return "?", nil
		}
	}
	return "", log.Error("type is not correct")
}

func (this *timeHandler) Cron(i_second, i_minute, i_hour, i_day, i_month, i_weekday interface{}, cb func(TimeNoder)) TimeNoder {
	second, err := convertIntToString(i_second)
	if err != nil {
		log.Panic(err.Error())
	}
	minute, err := convertIntToString(i_minute)
	if err != nil {
		log.Panic(err.Error())
	}
	hour, err := convertIntToString(i_hour)
	if err != nil {
		log.Panic(err.Error())
	}
	day, err := convertIntToString(i_day)
	if err != nil {
		log.Panic(err.Error())
	}
	month, err := convertIntToString(i_month)
	if err != nil {
		log.Panic(err.Error())
	}
	weekday, err := convertIntToString(i_weekday)
	if err != nil {
		log.Panic(err.Error())
	}
	jiffies := atomic.LoadUint64(&this.wheel.jiffies)

	node := &timeNode{
		expire:       0,
		callback:     cb,
		delay:        uint64(0),
		loopCur:      0,
		loopMax:      0,
		handler:      this,
		isCron:       true,
		lastMonthDay: -1,
	}
	err = node.parseCron(second, minute, hour, day, month, weekday)
	if err != nil {
		//TODO:err handler
		log.Panic(err.Error())
	}
	expire, _ := node.initCronExpireFunc(this.wheel)
	node.expire = expire/(uint64(time.Millisecond*10)) + jiffies
	tn := this.wheel.add(node, jiffies)

	this.noders.Store(tn, 1)

	return tn
}

func (this *timeHandler) Schedule(interval time.Duration, cb func(TimeNoder), opts ...TimerOption) TimeNoder {
	jiffies := atomic.LoadUint64(&this.wheel.jiffies)

	node := &timeNode{
		expire:   0,
		callback: cb,
		delay:    uint64(0),
		interval: uint64(interval),
		loopCur:  0,
		loopMax:  0,
		handler:  this,
		isCron:   false,
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

	tn := this.wheel.add(node, jiffies)

	this.noders.Store(tn, 1)

	return tn
}

func (this *timeHandler) EventChan() <-chan TypeEventChan {
	return this.eventChan
}

func (this *timeHandler) DelTimer() {

	// delete time event
	this.StopTimer()

	// close channel
	close(this.eventChan)
	this.eventChan = nil

	// delete wheel handler map
	this.wheel.handlers.Delete(this.key)
	this.wheel = nil

}

func (this *timeHandler) StopTimer() {
	this.noders.Range(func(key, value any) bool {
		node := key.(*timeNode)
		node.Stop()
		return true
	})
}

func (this *timeHandler) PrintDebug() {

	i := 0

	this.noders.Range(func(key, value any) bool {
		i++
		return true
	})

	log.Infof("handler info, key:", this.key, " nodeCnt:", i)
}
