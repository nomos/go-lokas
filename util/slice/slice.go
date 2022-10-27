package slice

import (
	"reflect"
)

func Nil[T any]() T {
	var t T
	return t
}

type KVIntString struct {
	K int
	V string
}

func New[T any](v ...T) []T {
	return v
}

func RemoveDuplicate[T comparable](arr []T) []T {
	ret := []T{}
	for _, v := range arr {
		found := false
		for _, r := range ret {
			if v == r {
				found = true
			}
		}
		if !found {
			ret = append(ret, v)
		}
	}
	return ret
}

type Number interface {
	uint8 | int8 | uint16 | int16 | uint32 | int32 | uint64 | int64 | int | float32 | float64
}

type Signed interface {
	int8 | int16 | int32 | int64 | int | float32 | float64
}

type Float interface {
	float32 | float64
}

func NumberConvert[T1, T2 Number](arr []T1) []T2 {
	ret := []T2{}
	for _, v := range arr {
		ret = append(ret, T2(v))
	}
	return ret
}

func RemoveAt[T any](arr []T, index int) []T {
	if len(arr) == 0 {
		return arr
	}
	return append(arr[:index], arr[index+1:]...)
}

func Remove[T comparable](arr []T, a ...T) []T {
	ret := []T{}
	for _, v := range arr {
		found := false
		for _, r := range a {
			if v == r {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, v)
		}
	}
	return ret
}

func Concat[T any](a ...[]T) []T {
	ret := make([]T, 0)
	for _, arr := range a {
		for _, elem := range arr {
			ret = append(ret, elem)
		}
	}
	return ret
}

func Flip[T any](a []T) []T {
	v := reflect.ValueOf(a)
	elemType := reflect.TypeOf(a).Elem()
	length := v.Len()
	for i := 0; i < length/2; i++ {
		temp := reflect.New(elemType).Elem()
		temp.Set(v.Index(length - 1 - i))
		v.Index(length - 1 - i).Set(v.Index(i))
		v.Index(i).Set(temp)
	}
	return v.Interface().([]T)
}

func HasInterface(arr interface{}, item ...interface{}) bool {
	arr1 := arr.([]interface{})
	for _, i := range item {
		found := false
		for _, v := range arr1 {
			if v == i {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func Has[T comparable](arr []T, item ...T) bool {
	for _, i := range item {
		found := false
		for _, v := range arr {
			if v == i {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func AppendOnce[T comparable](arr []T, item T) []T {
	if !Has(arr, item) {
		arr = append(arr, item)
	}
	return arr
}

func IndexOf[T comparable](arr []T, id T) int {
	idx := -1
	for i, v := range arr {
		if v == id {
			idx = i
			break
		}
	}
	return idx
}

func GetOneWithCondition[T any](slice []T, f func(index int, elem T) bool) T {
	for i, v := range slice {
		if f(i, v) {
			return v
		}
	}
	return Nil[T]()
}

func RemoveWithCondition[T any](slice []T, f func(index int, elem T) bool) ([]T, []T) {
	ret := append(slice[:0])
	remove := append(slice[:0])
	for i, v := range slice {
		if !f(i, v) {
			ret = append(ret, v)
		} else {
			remove = append(remove, v)
		}
	}
	return ret, remove
}

func RemoveOneWithCondition[T any](slice []T, f func(index int, elem T) bool) ([]T, []T) {
	ret := append(slice[:0])
	remove := append(slice[:0])
	for i, v := range slice {
		if !f(i, v) {
			ret = append(slice, v)
		} else {
			remove = append(remove, v)
			return ret, remove
		}
	}
	return ret, remove
}

func RemoveOne[T comparable](arr []T, id T) []T {
	idx := -1
	for i, v := range arr {
		if v == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return arr
	}
	return append(arr[:idx], arr[idx+1:]...)
}
