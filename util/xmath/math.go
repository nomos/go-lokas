package xmath

import (
	"log"
	"math"
)

type Number interface {
	Signed | UnSigned
}

type SignedInt interface {
	int8 | int16 | int32 | int64 | int
}

type Signed interface {
	SignedInt | Float
}

type UnSigned interface {
	uint8 | uint16 | uint32 | uint64 | uint
}

type Float interface {
	float32 | float64
}

func Ternary[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}
	return b
}

func Abs[T Signed](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

func Sqrt[T Number](a T) T {
	return T(math.Sqrt(float64(a)))
}

func Pow[T Number](x, y T) T {
	var ret T = 1
	for y > 0 {
		ret *= x
		y--
	}
	return ret
}

func MaxArr[T Number](v []T) T {
	var max T
	for i := 0; i < len(v); i++ {
		if i == 0 {
			max = v[i]
		} else {
			if v[i] > max {
				max = v[i]
			}
		}
	}
	return max
}

func MinArr[T Number](v []T) T {
	var min T
	for i := 0; i < len(v); i++ {
		if i == 0 {
			min = v[i]
		} else {
			if v[i] < min {
				min = v[i]
			}
		}
	}
	return min
}

func Max[T Number](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func Min[T Number](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func Lerp[T Number, T1 Float](t T1, a, b T) T {
	return T(float64(a) + float64(t)*float64(b-a))
}

func Range[T Number](start, end T) []T {
	var s1, e1 T
	if start > end {
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
	ret := make([]T, 0)

	for i := s1; i < e1; i++ {
		ret = append(ret, i)
	}
	return ret
}

func Step[T Number](edge T, x T) T {
	if x < edge {
		return 0
	}
	return 1
}

func Clamp[T Number](v T, min T, max T) T {
	return Min(Max(v, min), max)
}

func Sign[T Signed](v T) T {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}

func Floor[T Float](v T) T {
	return T(math.Floor(float64(v)))
}

func Ceil[T Float](v T) T {
	return T(math.Ceil(float64(v)))
}

func Fract[T Float](v T) T {
	return v - Floor(v)
}
