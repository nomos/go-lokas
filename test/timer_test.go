package test

import (
	"testing"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox"
	"github.com/nomos/go-lokas/timer"
)

func init() {
	log.InitDefault(true, true, "")
}

func TestTimer(t *testing.T) {

	handler1 := timer.NewHandler()
	handler2 := timer.NewHandler()

	log.Info("handler init completed")

	go func() {
		for {
			select {
			case data := <-handler1.EventChan():

				msg := data.(*timer.TimeEventMsg)
				msg.Callback(msg.TimeNoder)
			}
		}
	}()

	go func() {
		for {
			select {
			case data := <-handler2.EventChan():
				msg := data.(*timer.TimeEventMsg)
				msg.Callback(msg.TimeNoder)
			}
		}
	}()

	handler1.PrintDebug()
	handler2.PrintDebug()

	handler1.After(3*time.Second, func(tn timer.TimeNoder) {
		log.Infof("handler1 event1  cb", tn.GetDelay(), tn.GetInterval())
	})

	handler1.Schedule(1*time.Second, func(tn timer.TimeNoder) {
		log.Infof("handler1 event2 cb")
	})

	handler2.Schedule(2*time.Second, func(tn timer.TimeNoder) {
		log.Infof("handler2 event1 cb")
	})

	handler1.PrintDebug()
	time.Sleep(5 * time.Second)
	handler1.PrintDebug()
	// handler2.PrintDebug()
	handler1.StopTimer()
	// Stop()
	log.Infof("handler1 stop")
	handler1.PrintDebug()

	time.Sleep(5 * time.Second)

	time.Sleep(20 * time.Second)

}

func TestTimerActor(t *testing.T) {
	actor1 := lox.NewActor()
	actor1.StartMessagePump()

	var ia lokas.IActor = actor1

	ia.After(3*time.Second, func(tn timer.TimeNoder) {
		log.Infof("actor1 after!!", tn.GetDelay(), tn.GetInterval())
	})

	node := ia.Schedule(1*time.Second, func(tn timer.TimeNoder) {
		log.Infof("actor1 schedule!!", tn.GetDelay(), tn.GetInterval())
	})

	log.Infof("start!!")

	time.Sleep(10 * time.Second)
	node.Stop()

	time.Sleep(20 * time.Second)
}

func TestTimerOption(t *testing.T) {

	h1 := timer.NewHandler()

	repeat := 0
	h1.Schedule(2*time.Second, func(tn timer.TimeNoder) {
		repeat++
		log.Infof("time interval ", repeat)
	}, timer.WithDelay(5*time.Second), timer.WithLoop(5))

	go func() {
		for {
			select {
			case data := <-h1.EventChan():
				msg := data.(*timer.TimeEventMsg)
				msg.Callback(msg.TimeNoder)
			}
		}
	}()

	log.Infof("time started")

	time.Sleep(30 * time.Second)
}
