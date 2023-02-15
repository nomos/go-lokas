package xmath

import (
	"github.com/nomos/go-lokas/log"
	"testing"
)

type WeightAbleFloat float64

func (this WeightAbleFloat) Weight() float64 {
	return float64(this)
}

func TestPickOneWeightSelect(t *testing.T) {
	arr := []WeightAbleFloat{1, 2, 3, 4, 5, 6}
	idx, w, remain := PickOneWeightSelect(arr)
	log.Infof(idx, w, remain)
	arr = []WeightAbleFloat{1, 2, 3, 4, 5, 6}
	var results []WeightAbleFloat
	results, remain = PickSomeWeightSelect(arr, 3)
	log.Infof(results, remain)
}

func TestWeightSelect(t *testing.T) {

	arr := []WeightAbleFloat{1, 2, 3, 4, 5, 6}
	picked := map[WeightAbleFloat]int{}
	var seed uint64 = 8
	for i := 0; i < 100; i++ {
		_, v := WeightSelect(arr, seed+uint64(i))
		picked[v]++
	}
	for k, v := range picked {
		log.Infof(k, v)
	}
}
