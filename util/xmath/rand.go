package xmath

import (
	"github.com/nomos/go-lokas/util/slice"
	"math"
	"math/rand"
)

const JSMAXSAFE = 1342177

func SRandom(seed... uint64)float64{
	if len(seed)==0 {
		return rand.Float64()
	}
	s1:=seed[0]
	s1%=JSMAXSAFE
	s1 = (s1 + 1627) * (s1 + 125383)+s1*2
	s1%=JSMAXSAFE
	s1+=s1 * 9311 + 149297
	s1%=JSMAXSAFE

	s2:=uint64(1)
	y:=13
	for y>0 {
		s2*=s1
		s2%=JSMAXSAFE
		y--
	}
	s2 %= 233173
	return float64(s2) / 233173.0
}

func Rng(val1,val2 float64,seed...uint64)float64{
	var minVal,maxVal float64
	if val1<val2 {
		minVal = val1
		maxVal = val2
	} else {
		minVal = val2
		maxVal = val1
	}
	return minVal+(maxVal-minVal)*SRandom(seed...)
}

func RngInt(val1,val2 int,seed...uint64)int {
	return int(math.Floor(Rng(float64(val1),float64(val2+1),seed...)))
}

func Porn(seed...uint64)float64{
	if OneIn(2.0,seed...) {
		return -1.0
	} else {
		return 1.0
	}
}

func PornInt(seed...uint64)int{
	if OneIn(2.0,seed...) {
		return -1
	} else {
		return 1
	}
}

func OneIn(chance float64,seed...uint64)bool {
	return chance<=1||Rng(0,chance,seed...)<1.0
}

func OneInInt(chance int,seed...uint64)bool {
	return chance<=1||RngInt(0,chance,seed...)<1.0
}

func XInY(x float64,y float64,seed...uint64)bool {
	return SRandom(seed...)<x/y
}

func XInYInt(x int,y int,seed...uint64)bool {
	return SRandom(seed...)<float64(x)/float64(y)
}

type WeightAble interface {
	Weight()float64
}

func WeightSelect(weightList []WeightAble,seed...uint64)int {
	sum := 0.0
	for i := 0; i < len(weightList); i++ {
		sum += weightList[i].Weight()
	}
	r := SRandom(seed...) * sum
	var i int
	for i = 0; i < len(weightList); i++ {
		weight := weightList[i].Weight()
		if weight > r {
			break
		} else {
			r -= weight
		}
	}
	return i
}

func Shuffle(arr []interface{},seed...uint64)[]interface{}{
	sr := true
	if len(seed)==0 {
		sr = false
	}
	leng:=len(arr)
	var temp interface{}
	var random int
	seedMod := 0
	collection:=slice.SliceConcat(arr)
	for leng>0 {
		seedMod++
		if sr {
			random = int(math.Floor(SRandom(seed[0]+uint64(seedMod))*float64(leng)))
		} else {
			random = int(math.Floor(rand.Float64()*float64(leng)))
		}
		leng-=1
		temp = collection[leng]
		collection[leng] = collection[random]
		collection[random] = temp
	}
	return collection
}

func GetOne(arr []interface{},seed...uint64)interface{}{
	index := RngInt(0,len(arr)-1,seed...)
	return arr[index]
}

func PickOne(arr []interface{},seed...uint64)(interface{},[]interface{}){
	leng:=len(arr)
	index:=int(math.Floor(float64(leng)*SRandom(seed...)))
	var ret interface{}
	retOrigin:=make([]interface{},0)
	for i:=0;i<leng;i++ {
		if i==index {
			ret = arr[i]
			continue
		}
		retOrigin = append(retOrigin, arr[i])
	}
	return ret,retOrigin
}

func PickSome(arr []interface{},num int,seed...uint64)([]interface{},[]interface{}){
	sr := true
	if len(seed)==0 {
		sr = false
	}
	leng:=len(arr)
	var index int
	retOrigin:=make([]interface{},0)
	ret:=make([]interface{},0)
	spliceArr :=make([]int,0)
	for i:=0;i<num;i++ {
		if sr {
			index =int(math.Floor(float64(leng)*rand.Float64()))
		} else {
			index = int(math.Floor(float64(leng)*SRandom(seed[0]+uint64(i))))
		}
		spliceArr = append(spliceArr, index)
	}
 	for i:=0;i<leng;i++ {
 		if slice.HasInt(spliceArr,i) {
			ret = append(ret, arr[i])
 			continue
		}
		retOrigin = append(retOrigin, arr[i])
	}
	return ret,retOrigin
}