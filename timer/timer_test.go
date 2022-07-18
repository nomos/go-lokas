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

	handler1 := GetHandler("h1")
	handler2 := GetHandler("h2")

	log.Printf("handler init completed")

	go func() {
		for {
			select {
			case data := <-handler1.EventChan():

				msg := data.(*TimeEventMsg)
				msg.cb()
			}
		}
	}()

	go func() {
		for {
			select {
			case data := <-handler2.EventChan():
				msg := data.(*TimeEventMsg)
				msg.cb()
			}
		}
	}()

	handler1.PrintDebug()
	// handler2.PrintDebug()

	handler1.After(15*time.Second, func(...interface{}) {
		log.Println("handler1 event1  cb")
	})

	handler1.Schedule(1*time.Second, func(params ...interface{}) {
		log.Println("handler1 event2 cb")
	})

	handler2.Schedule(2*time.Second, func(i ...interface{}) {
		log.Println("handler2 event1 cb")
	})

	handler1.PrintDebug()
	time.Sleep(5 * time.Second)
	handler1.PrintDebug()
	// handler2.PrintDebug()
	handler1.DelSelf()
	// Stop()
	log.Println("handler1 del")
	handler1.PrintDebug()

	time.Sleep(5 * time.Second)

	time.Sleep(20 * time.Second)

}
