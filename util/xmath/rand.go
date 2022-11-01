package xmath

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/slice"
	"math"
	"math/rand"
)

const JSMAXSAFE = 1342177

func SRandom(seed ...uint64) float64 {
	if len(seed) == 0 {
		return rand.Float64()
	}
	s1 := seed[0]
	s1 %= JSMAXSAFE
	s1 = (s1+1627)*(s1+125383) + s1*2
	s1 %= JSMAXSAFE
	s1 += s1*9311 + 149297
	s1 %= JSMAXSAFE

	s2 := uint64(1)
	y := 13
	for y > 0 {
		s2 *= s1
		s2 %= JSMAXSAFE
		y--
	}
	return float64(s2) / float64(JSMAXSAFE)
}

func Rng[T Number](val1, val2 T, seed ...uint64) T {
	var minVal, maxVal T
	if val1 < val2 {
		minVal = val1
		maxVal = val2
	} else {
		minVal = val2
		maxVal = val1
	}
	return minVal + T(float64(maxVal-minVal)*SRandom(seed...))
}

func Porn(seed ...uint64) float64 {
	if OneIn(2.0, seed...) {
		return -1.0
	} else {
		return 1.0
	}
}

func PornInt(seed ...uint64) int {
	if OneIn(2.0, seed...) {
		return -1
	} else {
		return 1
	}
}

func OneIn[T Number](chance T, seed ...uint64) bool {
	return chance <= 1 || Rng(0, 1/float64(chance), seed...) < 1.0
}

func XInY[T Number](x T, y T, seed ...uint64) bool {
	return SRandom(seed...) < float64(x)/float64(y)
}

type WeightAble interface {
	Weight() float64
}

func PickOneWeightSelect[T WeightAble](weightList []T, seed ...uint64) (int, T, []T) {
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
		if weight >= r {
			break
		} else {
			r -= weight
		}
	}
	ret := weightList[i]
	arr := slice.RemoveAt(weightList, i)
	return i, ret, arr
}
func PickSomeWeightSelect[T WeightAble](weightList []T, num int, seed ...uint64) (results []T, remain []T) {
	results = []T{}
	remain = weightList
	var ret T
	for i := 0; i < num; i++ {
		_, ret, remain = PickOneWeightSelect(remain, seed...)
		results = append(results, ret)
	}
	return results, remain
}

func WeightSelect[T WeightAble](weightList []T, seed ...uint64) (int, T) {
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
		if weight >= r {
			break
		} else {
			r -= weight
		}
	}
	return i, weightList[i]
}

func Shuffle[T any](arr []T, seed ...uint64) []T {
	sr := true
	if len(seed) == 0 {
		sr = false
	}
	leng := len(arr)
	var temp T
	var random int
	seedMod := 0
	collection := slice.Concat(arr)
	for leng > 0 {
		seedMod++
		if sr {
			random = int(math.Floor(SRandom(seed[0]+uint64(seedMod)) * float64(leng)))
		} else {
			random = int(math.Floor(rand.Float64() * float64(leng)))
		}
		leng -= 1
		temp = collection[leng]
		collection[leng] = collection[random]
		collection[random] = temp
	}
	return collection
}

func GetOne[T any](arr []T, seed ...uint64) T {
	index := Rng(0, len(arr)-1, seed...)
	return arr[index]
}

func PickOne[T any](arr []T, seed ...uint64) (T, []T) {
	leng := len(arr)
	index := int(math.Floor(float64(leng) * SRandom(seed...)))
	var ret T
	retOrigin := make([]T, 0)
	for i := 0; i < leng; i++ {
		if i == index {
			ret = arr[i]
			continue
		}
		retOrigin = append(retOrigin, arr[i])
	}
	return ret, retOrigin
}

func PickSome[T any](arr []T, num int, seed ...uint64) ([]T, []T) {
	sr := true
	if len(seed) == 0 {
		sr = false
	}
	leng := len(arr)
	var index int
	retOrigin := make([]T, 0)
	ret := make([]T, 0)
	spliceArr := make([]int, 0)
	for i := 0; i < num; i++ {
		if sr {
			index = int(math.Floor(float64(leng) * rand.Float64()))
		} else {
			index = int(math.Floor(float64(leng) * SRandom(seed[0]+uint64(i))))
		}
		spliceArr = append(spliceArr, index)
	}
	for i := 0; i < leng; i++ {
		if slice.Has(spliceArr, i) {
			ret = append(ret, arr[i])
			continue
		}
		retOrigin = append(retOrigin, arr[i])
	}
	return ret, retOrigin
}
