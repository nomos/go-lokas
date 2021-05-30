package xmath

import (
	"log"
	"math"
)

func PowInt(x,y int)int {
	ret:=1
	for y>0 {
		ret*=x
		y--
	}
	return ret
}

func PowUInt64(x,y uint64)uint64 {
	ret:=uint64(1)
	for y>0 {
		ret*=x
		y--
	}
	return ret
}

func PowInt64(x,y int64)int64 {
	ret:=int64(1)
	for y>0 {
		ret*=x
		y--
	}
	return ret
}

func MaxInt(x,y int)int{
	if x>y {
		return x
	}
	return y
}

func MinInt(x,y int)int{
	if x<y {
		return x
	}
	return y
}

func Lerp(t,a,b float64)float64{
	return a + t * (b - a);
}

func Range(start,end float64)[]float64 {
	var s1,e1 float64
	if start>end {
		s1 = start
		e1 = end
	} else {
		s1 = end
		e1 = start
	}
	leng := e1 - s1
	if leng < 0 {
		log.Panic("arr length must >= 0")
	}
	ret:=make([]float64,0)

	for i:=s1;i<e1;i++ {
		ret = append(ret, i)
	}
	return ret
}

func RangeInt(start,end int)[]int {
	var s1,e1 int
	if start>end {
		s1 = start
		e1 = end
	} else {
		s1 = end
		e1 = start
	}
	leng := e1 - s1
	if leng < 0 {
		log.Panic("arr length must >= 0")
	}
	ret:=make([]int,0)

	for i:=s1;i<e1;i++ {
		ret = append(ret, i)
	}
	return ret
}

func Step (edge float64, x float64)float64 {
	if x<edge {
		return 0.0
	}
	return 1.0
}

func StepInt(edge int, x int)int {
	if x<edge {
		return 0
	}
	return 1
}

func Clamp (v float64, min float64, max float64)float64 {
	return math.Min(math.Max(v, min), max)
}

func ClampInt (v int, min int, max int)int {
	return MinInt(MaxInt(v, min), max)
}

func Sign (v float64)float64 {
	if v>0 {
		return 1.0
	}
	if v<0 {
		return -1.0
	}
	return 0
}

func Fract (v float64)float64 {
	return v - math.Floor(v)
}


