package util

import (
	"errors"
	"fmt"
	"github.com/nomos/go-lokas/util/stringutil"
	"reflect"
	"unsafe"
)

func MapToMap(s interface{}) map[string]interface{} {
	v := reflect.ValueOf(s)
	result := map[string]interface{}{}
	for _, k := range v.MapKeys() {
		result[k.Interface().(string)] = v.MapIndex(k).Interface()
	}
	return result
}

func Struct2Map(s interface{}, fields ...string) (map[string]interface{}, error) {
	if s == nil {
		return nil, errors.New("struct cannot be empty")
	}
	enableFields := map[string]interface{}{}
	for _, v := range fields {
		enableFields[v] = 1
	}
	result := map[string]interface{}{}
	valueOf := reflect.Indirect(reflect.ValueOf(s))
	numField := valueOf.NumField()
	for i := 0; i < numField; i++ {
		key := valueOf.Type().Field(i).Tag.Get("json")
		if key == "" {
			key = valueOf.Type().Field(i).Name
			if !stringutil.StartWithCapital(key) {
				continue
			}
		}
		value := valueOf.Field(i).Interface()
		if len(fields) == 0 {
			result[key] = value
		} else {
			if enableFields[key] == nil {
				continue
			}
			result[key] = value
		}
	}

	if len(result) == 0 {
		result = nil
	}

	return result, nil
}

func SizeStruct(data interface{}) int {
	return sizeof(reflect.ValueOf(data))
}

func sizeof(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Map:
		sum := 0
		keys := v.MapKeys()
		for i := 0; i < len(keys); i++ {
			mapkey := keys[i]
			s := sizeof(mapkey)
			if s < 0 {
				return -1
			}
			sum += s
			s = sizeof(v.MapIndex(mapkey))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum
	case reflect.Slice, reflect.Array:
		sum := 0
		for i, n := 0, v.Len(); i < n; i++ {
			s := sizeof(v.Index(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.String:
		sum := 0
		for i, n := 0, v.Len(); i < n; i++ {
			s := sizeof(v.Index(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Ptr, reflect.Interface:
		p := (*[]byte)(unsafe.Pointer(v.Pointer()))
		if p == nil {
			return 0
		}
		return sizeof(v.Elem())
	case reflect.Struct:
		sum := 0
		for i, n := 0, v.NumField(); i < n; i++ {
			s := sizeof(v.Field(i))
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Int:
		return int(v.Type().Size())

	default:
		fmt.Println("t.Kind() no found:", v.Kind())
	}

	return -1
}
