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
