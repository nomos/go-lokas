package test

import (
	"log"
	"testing"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/lox"
	"github.com/nomos/go-lokas/timer"
)

func TestTimer(t *testing.T) {

	handler1 := timer.NewHandler()
	handler2 := timer.NewHandler()

	log.Printf("handler init completed")

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
		log.Println("handler1 event1  cb", tn.GetDelay(), tn.GetInterval())
	})

	handler1.Schedule(1*time.Second, func(tn timer.TimeNoder) {
		log.Println("handler1 event2 cb")
	})

	handler2.Schedule(2*time.Second, func(tn timer.TimeNoder) {
		log.Println("handler2 event1 cb")
	})

	handler1.PrintDebug()
	time.Sleep(5 * time.Second)
	handler1.PrintDebug()
	// handler2.PrintDebug()
	handler1.StopTimer()
	// Stop()
	log.Println("handler1 stop")
	handler1.PrintDebug()

	time.Sleep(5 * time.Second)

	time.Sleep(20 * time.Second)

}

func TestTimerActor(t *testing.T) {
	actor1 := lox.NewActor()
	actor1.StartMessagePump()

	var ia lokas.IActor = actor1

	ia.After(3*time.Second, func(tn timer.TimeNoder) {
		log.Println("actor1 after!!", tn.GetDelay(), tn.GetInterval())
	})

	node := ia.Schedule(1*time.Second, func(tn timer.TimeNoder) {
		log.Println("actor1 schedule!!", tn.GetDelay(), tn.GetInterval())
	})

	log.Println("start!!")

	time.Sleep(10 * time.Second)
	node.Stop()

	time.Sleep(20 * time.Second)
}
