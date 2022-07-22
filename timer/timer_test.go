package timer

import (
	"log"
	"testing"
	"time"
)

type InputValue struct {
	a int
	b string
	c uint32
}

func TestNewTimer(t *testing.T) {

	handler1 := NewHandler()
	handler2 := NewHandler()

	log.Printf("handler init completed")

	go func() {
		for {
			select {
			case data := <-handler1.EventChan():

				msg := data.(*TimeEventMsg)
				msg.Callback()
			}
		}
	}()

	go func() {
		for {
			select {
			case data := <-handler2.EventChan():
				msg := data.(*TimeEventMsg)
				msg.Callback()
			}
		}
	}()

	handler1.PrintDebug()
	handler2.PrintDebug()

	handler1.After(15*time.Second, func() {
		log.Println("handler1 event1  cb")
	})

	handler1.Schedule(1*time.Second, func() {
		log.Println("handler1 event2 cb")
	})

	handler2.Schedule(2*time.Second, func() {
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
