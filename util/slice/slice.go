package slice

import "reflect"

type KVIntString struct {
	K int
	V string
}

func New[T any](v ...T) []T {
	return v
}

func RemoveMany[T comparable](a []T) []T {
	ret := []T{}
	for _, v := range a {
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

func Remove[T comparable](a ...T) []T {
	ret := []T{}
	for _, v := range a {
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
