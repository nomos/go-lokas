package timer

import (
	"fmt"
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
				cb := data.(func(...interface{}))
				log.Printf("time event handler1 call")
				cb()
			}
		}
	}()

	go func() {
		for {
			select {
			case data := <-handler2.EventChan():
				cb := data.(func(...interface{}))
				log.Printf("time event handler2 call")
				cb()
			}
		}
	}()

	handler1.PrintDebug()
	handler2.PrintDebug()

	handler1.After(3*time.Second, func(params ...interface{}) {
		fmt.Println("handler1 event1  cb")
	})

	tn1 := handler1.Schedule(1*time.Second, func(params ...interface{}) {
		fmt.Println("handler1 event2 cb")
	})

	time.Sleep(8 * time.Second)
	handler1.PrintDebug()
	handler2.PrintDebug()
	tn1.Stop()

	handler1.PrintDebug()

	time.Sleep(20 * time.Second)

}
