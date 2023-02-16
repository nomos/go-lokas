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
	arr := []WeightAbleFloat{1, 12, 3, 4, 5, 6, 0}
	idx, w, remain := PickOneWeightSelect(arr)
	log.Infof(idx, w, remain)
	arr = []WeightAbleFloat{1, 2, 3, 4, 5, 6}
	var results []WeightAbleFloat
	results, remain = PickSomeWeightSelect(arr, 3)
	log.Infof(results, remain)
}

func TestWeightSelect(t *testing.T) {

	arr := []WeightAbleFloat{1, 12, 3, 4, 5, 0, 6}
	picked := map[WeightAbleFloat]int{}
	picked1 := map[WeightAbleFloat]int{}
	var seed uint64 = 8
	for i := 0; i < 100; i++ {
		_, v := WeightSelect(arr, seed+uint64(i))
		picked[v]++
	}
	for k, v := range picked {
		log.Infof("11111", k, v)
	}
	for i := 0; i < 100; i++ {
		_, v := WeightSelect1(arr, seed+uint64(i))
		picked1[v]++
	}
	for k, v := range picked1 {
		log.Infof(k, v)
	}
}

func WeightSelect1[T WeightAble](weightList []T, seed ...uint64) (int, T) {
	if len(weightList) == 0 {
		log.Panic("weight list must > 0")
	}
	sum := 0.0
	for i := 0; i < len(weightList); i++ {
		sum += weightList[i].Weight()
	}
	r := SRandom(seed...) * sum
	var i int

	for i = 0; i < len(weightList); i++ {
		weight := weightList[i].Weight()
		if weight == 0 {
			continue
		}
		if weight >= r {
			break
		} else {
			r -= weight
		}
	}
	return i, weightList[i]
}
