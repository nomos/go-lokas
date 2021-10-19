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

func RemoveDuplicateString(a []string)[]string {
	ret:=[]string{}
	for _,v:=range a {
		found:=false
		for _,r:=range ret {
			if v==r {
				found = true
			}
		}
		if !found {
			ret = append(ret, v)
		}
	}
	return ret
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

func AppendOnceInt(arr []int,item int)[]int{
	if !HasInt(arr,item) {
		arr = append(arr, item)
	}
	return arr
}

func HasInt32(arr []int32,item int32)bool{
	for _,v:=range arr {
		if v == item {
			return true
		}
	}
	return false
}

func AppendOnceInt32(arr []int32,item int32)[]int32{
	if !HasInt32(arr,item) {
		arr = append(arr, item)
	}
	return arr
}

func HasInt64(arr []int64,item int64)bool{
	for _,v:=range arr {
		if v == item {
			return true
		}
	}
	return false
}

func AppendOnceInt64(arr []int64,item int64)[]int64{
	if !HasInt64(arr,item) {
		arr = append(arr, item)
	}
	return arr
}

func HasString(arr []string,item string)bool{
	for _,v:=range arr {
		if v == item {
			return true
		}
	}
	return false
}

func AppendOnceString(arr []string,item string)[]string{
	if !HasString(arr,item) {
		arr = append(arr, item)
	}
	return arr
}

func HasFloat64(arr []float64,item float64)bool{
	for _,v:=range arr {
		if v == item {
			return true
		}
	}
	return false
}

func RemoveInt (arr []int,id int)[]int{
	idx:=-1
	for i,v:=range arr {
		if v==id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return arr
	}
	return append(arr[:idx], arr[idx+1:]...)
}

func RemoveInt32 (arr []int32,id int32)[]int32{
	idx:=-1
	for i,v:=range arr {
		if v==id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return arr
	}
	return append(arr[:idx], arr[idx+1:]...)
}

func RemoveInt64 (arr []int64,id int64)[]int64{
	idx:=-1
	for i,v:=range arr {
		if v==id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return arr
	}
	return append(arr[:idx], arr[idx+1:]...)
}

func RemoveString (arr []string,id string)[]string{
	idx:=-1
	for i,v:=range arr {
		if v==id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return arr
	}
	return append(arr[:idx], arr[idx+1:]...)
}