package bitset

import (
	"github.com/nomos/go-lokas/log"
	"testing"
	"time"
)

func TestBit(t *testing.T) {
	var a BitSet = 0
	a = a.Set(7, true)
	log.Infof(a.Print())
	log.Infof(a.Get(7))
	log.Infof(a.Get(6))
	a = a.Set(3, true)
	a = a.Set(4, true)
	a = a.Set(4, false)
	log.Infof(a.Print())
	log.Infof(a.Values(8))
}

func TestTime(t *testing.T) {
	testArr()
	testBitset()
}

func testArr() {
	var a, b, c bool
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		arr := [64]bool{}
		arr[3] = true
		arr[5] = true
		arr[6] = true
		a = arr[4]
		b = arr[5]
		c = arr[6]
	}
	end := time.Now()
	log.Infof(a, b, c, end.Sub(start))
}

func testBitset() {
	var a, b, c bool
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		arr := BitSet(0)
		arr = arr.Set(3, true)
		arr = arr.Set(5, true)
		arr = arr.Set(6, true)
		a = arr.Get(4)
		b = arr.Get(5)
		c = arr.Get(6)
	}
	end := time.Now()
	log.Infof(a, b, c, end.Sub(start))
}
