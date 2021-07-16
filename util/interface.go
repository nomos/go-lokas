package util

import "reflect"

func IsNil(i interface{}) bool {
	if i==nil {
		return false
	}
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}