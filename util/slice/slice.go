package slice

import "reflect"

type KVIntString struct {
	K int
	V string
}

func StringArr(s... string)[]string{
	return s
}

func IntArr(i... int)[]int {
	return i
}

func ByteArr(i... byte)[]byte {
	return i
}

func SliceConcat(a... []interface{})[]interface{} {
	ret:=make([]interface{},0)
	for _,arr:=range a {
		for _,elem:=range arr {
			ret = append(ret, elem)
		}
	}
	return ret
}

func FlipSlice(a interface{})interface{}{
	v:=reflect.ValueOf(a)
	elemType:=reflect.TypeOf(a).Elem()
	length:=v.Len()
	for i := 0; i < length/2; i++ {
		temp:=reflect.New(elemType).Elem()
		temp.Set(v.Index(length-1-i))
		v.Index(length-1-i).Set(v.Index(i))
		v.Index(i).Set(temp)
	}
	return v.Interface()
}

func HasInt(arr []int,item int)bool{
	for _,v:=range arr {
		if v == item {
			return true
		}
	}
	return false
}

func HasFloat64(arr []float64,item float64)bool{
	for _,v:=range arr {
		if v == item {
			return true
		}
	}
	return false
}