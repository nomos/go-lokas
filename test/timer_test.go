package test

import (
	"log"
	"testing"
	"time"

	"github.com/nomos/go-lokas/lox"
)

func TestTimerActor(t *testing.T) {
	actor1 := lox.NewActor()
	actor1.StartMessagePump()

	actor1.GetTimeHandler().After(3*time.Second, func() {
		log.Println("actor1 after!!")
	})

	actor1.GetTimeHandler().Schedule(2*time.Second, func() {
		log.Println("actor1 schedule!!")
	})

	log.Println("start!!")

	time.Sleep(20 * time.Second)
}
