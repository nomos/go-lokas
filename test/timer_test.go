package test

import (
	"log"
	"testing"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/lox"
)

func TestTimerActor(t *testing.T) {
	actor1 := lox.NewActor()
	actor1.StartMessagePump()

	var ia lokas.IActor = actor1

	ia.After(3*time.Second, func() {
		log.Println("actor1 after!!")
	})

	node := ia.Schedule(1*time.Second, func() {
		log.Println("actor1 schedule!!")
	})

	log.Println("start!!")

	time.Sleep(10 * time.Second)
	node.Stop()

	time.Sleep(20 * time.Second)
}
