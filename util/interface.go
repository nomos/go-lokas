package util

import "reflect"

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}

func Nil[T any]() T {
	var t T
	return t
}

func IsString(i interface{}) bool {
	return reflect.TypeOf(i).Kind() == reflect.String
}

func IsInt(i interface{}) bool {
	return reflect.TypeOf(i).Kind() == reflect.Int
}

func IsBool(i interface{}) bool {
	return reflect.TypeOf(i).Kind() == reflect.Bool
}
