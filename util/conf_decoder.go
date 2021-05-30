package util

import (
	"encoding/json"
	"fmt"
	"github.com/nomos/go-log/log"
	"reflect"
	"regexp"
	"strings"
)

var conf_regexp_struct = regexp.MustCompile(`([(]{0,1}(([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}){1}([_]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}))*){1}([,]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}){1}([_]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}))*)*[)]{0,1})`)

var conf_regexp_array = regexp.MustCompile(`(([(](.+)[)]|[0-9]){1}([_]([(](.+)[)]|[0-9]))*){1}([,](([(](.+)[)]|[0-9]){1}([_]([(](.+)[)]|[0-9]))*))+`)

var conf_regexp = regexp.MustCompile(`([(]{0,1}(([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}){1}([_]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}))*){1}([,]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}){1}([_]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}))*)*[)]{0,1}){0,1}([|]([(]{0,1}(([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}){1}([_]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}))*){1}([,]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}){1}([_]([(]{0,1}([0-9]+([_][0-9]+)*){1}([,][0-9]+([_][0-9]+)*)*[)]{0,1}))*)*[)]{0,1}))*`)

var bline_arr_numberic_regexp = regexp.MustCompile(`[0-9]+([_][0-9]+)+`)
var bline_arr_regexp = regexp.MustCompile(`([\[][^_]+[\]]|[0-9]+){1}([_]([\[][^_]+[\]]|[0-9]+))+`)
var arr_regexp = regexp.MustCompile(`([\[][^_]+[\]]|[0-9]+){1}([,]([\[][^_]+[\]]|[0-9]+))+`)

func ConfFormat(str string)string{
	str  = strings.Replace(str,"(","[",-1)
	str  = strings.Replace(str,")","]",-1)
	str = bline_arr_numberic_regexp.ReplaceAllStringFunc(str, func(s string) string {
		splits:=strings.Split(s,"_")
		ret:="["
		for index,v:=range splits {
			ret+=v
			if index<len(splits)-1 {
				ret+=","
			}
		}
		return ret+"]"
	})
	log.Warnf(str)
	str = bline_arr_regexp.ReplaceAllStringFunc(str, func(s string) string {
		log.Warnf(s)
		splits:=strings.Split(s,"_")
		ret:="["
		for index,v:=range splits {
			ret+=v
			if index<len(splits)-1 {
				ret+=","
			}
		}
		return ret+"]"
	})
	return str
}

func parseStruct(v reflect.Value,t reflect.Type) ([]reflect.Value,[]reflect.Type) {
	parsedValue := make([]reflect.Value, 0)
	parsedType := make([]reflect.Type, 0)
	if !v.IsValid() {
		return parsedValue,parsedType
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			continue
		}
		name := f.Name
		if tag := f.Tag.Get("bt"); tag != "" {
			name = tag
		}
		if name == "-" {
			continue
		}
		parsedValue = append(parsedValue, v.Field(i))
		parsedType = append(parsedType, f.Type)
	}
	return parsedValue,parsedType
}

func decodeAllocate(t reflect.Type)reflect.Value {
	switch t.Kind() {
	case reflect.Int:
		return reflect.ValueOf(new(int)).Elem()
	case reflect.Map:
		return reflect.New(t).Elem()
	case reflect.Slice:
		return reflect.New(t).Elem()
	case reflect.Struct:
		return reflect.Zero(t)
	case reflect.Ptr:
		return reflect.New(t.Elem())
	default:
		panic(fmt.Errorf("unsupport type:%s",t.Name()))
	}
}

func decodeFromJsonObj(t reflect.Type,v reflect.Value,jsonObj interface{}) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = reflect.Indirect(v)
	}
	switch t.Kind() {
	case reflect.Slice:
		length:=len(jsonObj.([]interface{}))
		if uint32(v.Cap()) < uint32(length) {
			v.Set(reflect.MakeSlice(v.Type(), 0, int(length)))
		}
		v.Set(reflect.MakeSlice(t,0,length))
		for _,iter:=range jsonObj.([]interface{}) {
			value:= decodeAllocate(t.Elem())
			decodeFromJsonObj(t.Elem(),value,iter)
			v.Set(reflect.Append(v,value))
		}


	case reflect.Map:
		v.Set(reflect.MakeMap(t))
		for _,iter:=range jsonObj.([]interface{}) {
			key :=reflect.ValueOf(new(int)).Elem()
			mapElem:=iter.([]interface{})
			decodeFromJsonObj(t.Key(),key,mapElem[0])
			value:= decodeAllocate(t.Elem())
			decodeFromJsonObj(t.Elem(),value,mapElem[1])
			v.SetMapIndex(key, value)
		}
	case reflect.Struct:
		valueFields,typeFields:= parseStruct(v,t)
		log.WithFields(log.Fields{
			"jsonObj":jsonObj,
			"obj len":len(jsonObj.([]interface{})),
			"field len":len(typeFields),
			"value":v,
		}).Warn("struct")
		if len(typeFields)!=len(jsonObj.([]interface{})) {
			panic(fmt.Errorf("struct type:%s not match conf str:%v",t.Name(),jsonObj))
		}
		for index,field:=range jsonObj.([]interface{}){
			decodeFromJsonObj(typeFields[index],valueFields[index],field)
		}
	case reflect.Int:
		v.SetInt(int64(jsonObj.(float64)))
	default:
		panic(fmt.Errorf("unsupport type:%v",t))
	}
}

func ConfDecode(out interface{},str string) error{
	ret:= conf_regexp_struct.FindAllString(str,-1)
	var formatStr string
	if len(ret) == 1 {
		formatStr = ConfFormat(ret[0])
	} else {
		for index,v:=range ret {
			if conf_regexp_array.FindString(v) == v {
				v = "("+v+")"
			}
			formatStr+= ConfFormat(v)
			if index<len(ret)-1 {
				formatStr+=","
			}
		}
	}
	formatStr = "["+formatStr+"]"
	log.Warnf(formatStr)
	t:=reflect.TypeOf(out)
	v:=reflect.ValueOf(out)
	var jsonObj interface{}
	json.Unmarshal(([]byte)(formatStr),&jsonObj)
	defer func() {
		if rec:=recover();rec!=nil {
			log.Error(fmt.Sprintf("type %s decode err:%v",t.Name(),rec))
		}
	}()
	decodeFromJsonObj(t,v,jsonObj)
	return nil
}