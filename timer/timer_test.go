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
	// tm := Instance()

	tm := NewTimer()
	tm.Start()

	log.Printf("tm init completed")

	go func() {
		for {
			select {
			case value := <-tm.TimeEventChan():
				data := value.(*InputValue)
				log.Printf("time event call, %d, %s, %d", data.a, data.b, data.c)
			}
		}
	}()

	var v *InputValue = &InputValue{
		a: -12,
		b: "fdfdf",
		c: 666,
	}
	tm.After(3*time.Second, v)

	v2 := &InputValue{
		a: 463,
		b: "ggfgfg",
		c: 123,
	}
	tm.Schedule(2*time.Second, v2)

	time.Sleep(20 * time.Second)

}
