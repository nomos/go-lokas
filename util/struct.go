package util

import (
	"errors"
	"github.com/nomos/go-lokas/util/stringutil"
	"reflect"
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
