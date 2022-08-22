package test

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/timer"
	"github.com/nomos/go-lokas/util"
	"testing"
	"time"
)

type testTimeHandler struct {
	timeOffset time.Duration
}

func (this *testTimeHandler) SetOffset(duration time.Duration) {
	this.timeOffset = duration
}

func (this *testTimeHandler) Now() time.Time {
	return time.Now().Add(this.timeOffset)
}

func (this *testTimeHandler) Get10Ms() time.Duration {
	return time.Duration(this.Now().UnixNano() / int64(time.Millisecond) / 10)
}

func TestCron(t *testing.T) {
	timeHandler := &testTimeHandler{
		timeOffset: 0,
	}
	now := time.Now()
	timeHandler.SetOffset(time.Date(now.Year(), 9, 23, 23, 59, 59, 0, time.Local).Sub(now))
	timer.SetTimeTestHandler(timeHandler)
	handler1 := timer.NewHandler()
	go func() {
		for {
			select {
			case data := <-handler1.EventChan():
				msg := data.(*timer.TimeEventMsg)
				msg.Callback(msg.TimeNoder)
			}
		}
	}()
	first := 0
	handler1.Cron("0", "0", "0", "?", "8-9", "5-6", func(noder timer.TimeNoder) {
		log.Warnf("TimeTest", first, util.FormatTimeToString(timeHandler.Now()))
		if first == 0 {
			timeHandler.SetOffset(time.Date(2022, 9, 29, 23, 59, 59, 0, time.Local).Sub(time.Now()))
			first++
		}
	})

	util.WaitForTerminate()
}
