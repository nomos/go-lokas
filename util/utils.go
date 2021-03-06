package util

import (
	"bytes"
	"github.com/nomos/go-lokas/log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func ExecPath()(string,error){
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "",err
	}
	if strings.HasSuffix(dir,`.app/Contents/MacOS`) {
		log.Warnf("is a macos app")
		s:=strings.Split(dir,"/")
		dir = strings.Join(s[:len(s)-3],"/")
		log.Warnf(dir)
	}
	return dir,nil
}

func Ternary(expr bool, whenTrue, whenFalse interface{}) interface{} {
	if expr == true {
		return whenTrue
	}
	return whenFalse
}

func WaitForTerminateChanCb(c chan struct{},cb func()) {
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		close(c)
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-c
	if cb!=nil {
		cb()
	}
}
func WaitForTerminateChan(c chan struct{}) {
	WaitForTerminateChanCb(c,nil)
}


func WaitForTerminate() {
	WaitForTerminateChan(make(chan struct{}))
}

func Apply(f interface{}, args []interface{})([]reflect.Value){
	fun := reflect.ValueOf(f)
	in := make([]reflect.Value, len(args))
	for k,param := range args{
		in[k] = reflect.ValueOf(param)
	}
	r := fun.Call(in)
	return r
}

func GetGoroutineID() uint64 {
	b := make([]byte, 64)
	runtime.Stack(b, false)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func IsEqual(a interface{},b interface{})bool {
	if reflect.TypeOf(a)!=reflect.TypeOf(b) {
		return false
	}
	switch a.(type) {
	case bool:
		return a.(bool) == b.(bool)
	case uint:
		return a.(uint) == b.(uint)
	case uint8:
		return a.(uint8) == b.(uint8)
	case uint16:
		return a.(uint16) == b.(uint16)
	case uint32:
		return a.(uint32) == b.(uint32)
	case uint64:
		return a.(uint64) == b.(uint64)
	case int:
		return a.(int) == b.(int)
	case int8:
		return a.(int8) == b.(int8)
	case int16:
		return a.(int16) == b.(int16)
	case int32:
		return a.(int32) == b.(int32)
	case int64:
		return a.(int64) == b.(int64)
	case float32:
		return a.(float32) == b.(float32)
	case float64:
		return a.(float64) == b.(float64)
	case string:
		return a.(string) == b.(string)
	case uintptr:
		return a.(uintptr) == b.(uintptr)
	case reflect.Type:
		return a.(reflect.Type) == b.(reflect.Type)
	case unsafe.Pointer:
		return a.(unsafe.Pointer) == b.(unsafe.Pointer)
	}

	if a==b {
		return true
	}

	return false
}

func HasInSlice(slice interface{},v interface{})bool {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		panic("data is not a slice type")
	}
	value := reflect.ValueOf(slice)
	for index:=0;index<value.Len();index++ {
		if IsEqual(value.Index(index).Interface(),v) {
			return true
		}
	}
	return false
}

func HasReflectTypeInSlice(slice []reflect.Type,v interface{})bool {
	for _, va := range slice {
		if IsEqual(va,v) {
			return true
		}
	}
	return false
}

func CheckDuplicateString(slice []string) bool {
	for i := 0; i < len(slice); i++ {
		for j := 0; j < len(slice); j++ {
			if i == j {
				continue
			}
			if slice[i] == slice[j] {
				return true
			}
		}
	}
	return false
}

func CheckDuplicate(slice interface{}) bool {
	slice1:=slice.([]interface{})
	for i := 0; i < len(slice1); i++ {
		for j := 0; j < len(slice1); j++ {
			if i == j {
				continue
			}
			if slice1[i] == slice1[j] {
				return true
			}
		}
	}
	return false
}

func GetIndexOfSlice(slice []interface{},v interface{})int {
	for i, va := range slice {
		if IsEqual(va,v) {
			return i
		}
	}
	return -1
}

func RemoveSliceElement(slice []interface{}, elem interface{}) []interface{} {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if IsEqual(v,elem) {
			slice = append(slice[:i], slice[i+1:]...)
			return RemoveSliceElement(slice, elem)
		}
	}
	return slice
}

func ToSliceInterface(arr interface{}) []interface{} {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		panic("toslice arr not slice")
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret
}

func RemoveMapElement(obj map[interface{}]interface{},elem interface{}) map[interface{}]interface{} {
	if len(obj) == 0 {
		return obj
	}
	for i, v := range obj {
		if IsEqual(v,elem) {
			delete(obj, i)
			return RemoveMapElement(obj, elem)
		}
	}
	return obj
}

func ToMapInterface(obj interface{}) map[interface{}]interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Map {
		panic("to map obj not map")
	}
	ret := map[interface{}]interface{}{}

	for _,k := range v.MapKeys() {
		value :=v.MapIndex(k)
		ret[k] = value

	}
	return ret
}

//?????????????????????
func IsNilPointer(c interface{}) bool {
	vi := reflect.ValueOf(&c)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}

type RemoveSliceCondition func(index int,elem interface{}) bool

func GetSliceElemWithCondition(slice []interface{}, f RemoveSliceCondition) interface{} {
	for i, v := range slice {
		if f(i,v) {
			return v
		}
	}
	return nil
}

func RemoveSliceWithCondition(slice []interface{}, f RemoveSliceCondition) []interface{} {
	if len(slice) == 0 {
		return slice
	}
	ret := append(slice[:0])
	for i, v := range slice {
		if !f(i,v) {
			ret = append(slice, v)
		}
	}
	return ret
}


type RemoveMapCondition func(key interface{},elem interface{}) bool

func RemoveMapWithCondition(obj map[interface{}]interface{}, f RemoveMapCondition) map[interface{}]interface{} {
	if len(obj) == 0 {
		return obj
	}
	for k, v := range obj {
		if f(k,v) {
			delete(obj, k)
			return RemoveMapWithCondition(obj, f)
		}
	}
	return obj
}