package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/colors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func MarshalMessage(transId uint32, data interface{}, t TYPE) (ret []byte, err error) {
	if t == JSON {
		return MarshalJsonMessage(transId, data)
	} else if t == BINARY {
		return MarshalBinaryMessage(transId, data)
	} else {
		return nil, errors.New("unidentified protocol")
	}
}

func MarshalJsonMessage(transId uint32, data interface{}) (ret []byte, err error) {
	var out bytes.Buffer
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
			} else {
				err = r.(error)
			}
		}
	}()
	writeJsonMessage(&out, transId, data)
	if out.Len() > 65536 {
		ret = out.Bytes()
		binary.LittleEndian.PutUint16(ret[0:2], uint16(0))
		return ret, nil
	}
	ret = out.Bytes()
	binary.LittleEndian.PutUint16(ret[0:2], uint16(out.Len()))
	return ret, nil
}

func MarshalBinaryMessage(transId uint32, data interface{}) (ret []byte, err error) {
	var out bytes.Buffer
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
			} else {
				err = r.(error)
			}
		}
	}()
	writeBinaryMessage(&out, transId, data)
	if out.Len() > 65536 {
		ret = out.Bytes()
		binary.LittleEndian.PutUint16(ret[0:2], uint16(0))
		return ret, nil
	}
	ret = out.Bytes()
	binary.LittleEndian.PutUint16(ret[0:2], uint16(out.Len()))
	return ret, nil
}

func MarshalBinary(s interface{}) ([]byte, error) {
	var out bytes.Buffer
	tag, err := GetTypeRegistry().GetTagByType(reflect.TypeOf(s).Elem())
	if err != nil {
		return nil, err
	}
	v := reflect.ValueOf(s).Elem()
	t := reflect.TypeOf(s).Elem()
	w(&out, uint16(tag))
	writeComplex(&out, tag, v, t)
	return out.Bytes(), nil
}

func writeJsonMessage(out io.Writer, transId uint32, s interface{}) {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	defer func() {
		if r := recover(); r != nil {
			log.Error("recover", zap.Any("recover", r))
			log.Panic(fmt.Sprintf("%v at struct field", r))
		}
	}()
	tag, v, t := getTagId(v, t)
	if tag <= TAG_Null {
		log.Panic("not a binary message type")
	}
	w(out, uint16(0))
	w(out, transId)
	w(out, uint16(tag))
	data, _ := json.Marshal(s)
	w(out, data)
}

func writeBinaryMessage(out io.Writer, transId uint32, s interface{}) {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	defer func() {
		if r := recover(); r != nil {
			log.Error("recover", zap.Any("recover", r))
			log.Panic(fmt.Sprintf("%v at struct field", r))
		}
	}()
	tag, v, t := getTagId(v, t)
	if tag <= TAG_Null {
		log.Panic("not a binary message type")
	}
	w(out, uint16(0))
	w(out, transId)
	w(out, uint16(tag))
	writeValue(out, tag, v, t)
}

func w(out io.Writer, v interface{}) {
	err := binary.Write(out, binary.LittleEndian, v)
	if err != nil {
		log.Panic(err.Error())
	}
}

func writeBaseValue(out io.Writer, v reflect.Value, kind reflect.Kind) {
	switch kind {
	case reflect.Bool:
		w(out, v.Bool())
	case reflect.Int8:
		w(out, int8(v.Int()))
	case reflect.Uint8:
		w(out, byte(v.Uint()))
	case reflect.Int16:
		w(out, int16(v.Int()))
	case reflect.Uint16:
		w(out, uint16(v.Uint()))
	case reflect.Int32:
		w(out, int32(v.Int()))
	case reflect.Uint32:
		w(out, uint32(v.Uint()))
	case reflect.Int64:
		w(out, int64(v.Int()))
	case reflect.Uint64:
		w(out, uint64(v.Uint()))
	case reflect.Float32:
		w(out, float32(v.Float()))
	case reflect.Float64:
		w(out, v.Float())
	default:
		log.Panic(fmt.Sprintf("illegal base type %v", kind))
	}
}

func writeTime(out io.Writer, v reflect.Value) {
	w(out, v.Interface().(time.Time).UnixNano()/time.Millisecond.Nanoseconds())
}

func writeDecimal(out io.Writer, v reflect.Value) {
	data := []byte(v.Interface().(decimal.Decimal).String())
	w(out, uint16(len(data)))
	w(out, data)
}

func writeColor(out io.Writer, v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		data := v.Interface().(*colors.Color)
		w(out, data.Uint32())
	} else {
		data := v.Interface().(colors.Color)
		w(out, data.Uint32())
	}
}

func writeBuffer(out io.Writer, v reflect.Value) {
	data := v.Interface().(*bytes.Buffer).Bytes()
	w(out, uint16(len(data)))
	w(out, data)
}

func writeString(out io.Writer, v interface{}) {
	w(out, uint16(len(v.(string))))
	_, err := out.Write([]byte(v.(string)))
	if err != nil {
		log.Panic(err.Error())
	}
}

func isArrayValue(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool:
		return true
	case reflect.Int8, reflect.Uint8:
		return true
	case reflect.Int16, reflect.Uint16:
		return true
	case reflect.Int32, reflect.Uint32:
		return true
	case reflect.Int64, reflect.Uint64:
		return true
	case reflect.Float32:
		return true
	case reflect.Float64:
		return true
	}
	return false
}

func getArrayTypeByKind(kind reflect.Kind) BINARY_TAG {
	switch kind {
	case reflect.Bool:
		return TAG_BoolArray
	case reflect.Uint8:
		return TAG_ByteArray
	case reflect.Int16:
		return TAG_ShortArray
	case reflect.Uint16:
		return TAG_UShortArray
	case reflect.Int32:
		return TAG_IntArray
	case reflect.Uint32:
		return TAG_UIntArray
	case reflect.Int64:
		return TAG_LongArray
	case reflect.Uint64:
		return TAG_ULongArray
	case reflect.Float32:
		return TAG_FloatArray
	case reflect.Float64:
		return TAG_DoubleArray
	default:
		log.Panic("not a ArrayType kind")
		return 0
	}
}

func getListTypeByType(t reflect.Type) BINARY_TAG {
	if t.Kind() == reflect.Ptr {
		log.Panic("出错了 不能是ptr")
	}
	switch t.Elem().Kind() {
	case reflect.String:
		return TAG_String
	case reflect.Array, reflect.Slice:
		k := t.Elem().Elem().Kind()
		if !isArrayValue(k) {
			return TAG_List
		}
		return getArrayTypeByKind(k)
	case reflect.Map:
		return TAG_Map
	case reflect.Struct:
		if id, err := GetTypeRegistry().GetTagByType(t.Elem()); err == nil {
			return id
		}
		return TAG_Proto
	case reflect.Interface:
		switch t.Elem().Elem().Kind() {
		case reflect.Struct:
			if id, err := GetTypeRegistry().GetTagByType(t.Elem().Elem()); err == nil {
				return id
			}
			return TAG_Proto
		default:
			log.Panic("non struct interface")
		}
	case reflect.Ptr:
		switch t.Elem().Elem().Kind() {
		case reflect.Struct:
			if id, err := GetTypeRegistry().GetTagByType(t.Elem().Elem()); err == nil {
				return id
			}
			return TAG_Proto
		default:
			log.Panic("non struct ptr2")
		}
	default:
		log.Panic(fmt.Sprintf("Unhandled type:%v ", t))
	}
	return 0
}

func getTagId(v reflect.Value, t reflect.Type) (BINARY_TAG, reflect.Value, reflect.Type) {
	kind := t.Kind()
	switch kind {
	case reflect.Bool:
		return TAG_Bool, v, t
	case reflect.Uint8:
		return TAG_Byte, v, t
	case reflect.Int16:
		return TAG_Short, v, t
	case reflect.Uint16:
		return TAG_UShort, v, t
	case reflect.Int32:
		return TAG_Int, v, t
	case reflect.Uint32:
		return TAG_UInt, v, t
	case reflect.Int64:
		return TAG_Long, v, t
	case reflect.Uint64:
		return TAG_ULong, v, t
	case reflect.Float32:
		return TAG_Float, v, t
	case reflect.Float64:
		return TAG_Double, v, t
	case reflect.String:
		return TAG_String, v, t
	case reflect.Array, reflect.Slice:
		k := t.Elem().Kind()
		if !isArrayValue(k) {
			return TAG_List, v, t
		}
		return getArrayTypeByKind(k), v, t
	case reflect.Map:
		return TAG_Map, v, t
	case reflect.Struct:
		if t == reflect.TypeOf((*time.Time)(nil)).Elem() {
			return TAG_Time, v, t
		}
		if t == reflect.TypeOf((*colors.Color)(nil)).Elem() {
			return TAG_Color, v, t
		}
		if t == reflect.TypeOf((*decimal.Decimal)(nil)).Elem() {
			return TAG_Decimal, v, t
		}
		tag, err := GetTypeRegistry().GetTagByType(t)
		if err != nil {
			log.WithFields(log.Fields{
				"value": v,
				"type":  t,
			}).Panic("tag is not registered")
		}
		return tag, v, t
	case reflect.Ptr:
		switch t.Elem().Kind() {
		case reflect.Struct:
			if t.Elem() == reflect.TypeOf((*time.Time)(nil)).Elem() {
				log.Warn("时间格式")
				return TAG_Time, v.Elem(), t.Elem()
			}
			if t.Elem() == reflect.TypeOf((*colors.Color)(nil)).Elem() {
				log.Warn("颜色格式")
				return TAG_Color, v.Elem(), t.Elem()
			}
			if t.Elem() == reflect.TypeOf((*decimal.Decimal)(nil)).Elem() {
				log.Warn("Decimal格式")
				return TAG_Decimal, v.Elem(), t.Elem()
			}
			if t.Elem() == reflect.TypeOf((*bytes.Buffer)(nil)).Elem() {
				log.Warn("Buffer格式")
				return TAG_Buffer, v, t
			}
			tag, err := GetTypeRegistry().GetTagByType(t.Elem())
			if err != nil {
				log.WithFields(log.Fields{
					"value": v.Elem(),
					"type":  t.Elem(),
				}).Panic("tag is not registered")
			}
			if v.Kind() == reflect.Map {
				return tag, v, t.Elem()
			} else {
				return tag, v.Elem(), t.Elem()
			}
		default:
			log.Panic("non struct ptr2", zap.Any("value", v), zap.Any("type", t))
		}
	case reflect.Interface:
		log.Panic("interface is unsupported")
	default:
		log.Panic(fmt.Sprintf("Unhandled type: %v (%v)", v.Type(), v.Interface()))
	}
	return 0, v, t
}

func writeTag(out io.Writer, tag BINARY_TAG) {
	w(out, uint16(tag))
}

func writeValue(out io.Writer, tag BINARY_TAG, v reflect.Value, t reflect.Type) {
	var tagCustom BINARY_TAG = 0
	if tag > TAG_Null {
		tagCustom = tag
		tag = TAG_Proto
	}
	if tag.IsBaseValue() {
		writeBaseValue(out, v, v.Kind())
	} else if tag == TAG_String {
		writeString(out, v.String())
	} else if tag == TAG_Time {
		writeTime(out, v)
	} else if tag == TAG_Color {
		writeColor(out, v)
	} else if tag == TAG_Decimal {
		writeDecimal(out, v)
	} else if tag == TAG_Buffer {
		writeBuffer(out, v)
	} else if tag.IsArrayLike() {
		writeBaseArray(out, v)
	} else if tag == TAG_List {
		writeList(out, v, t)
	} else if tag == TAG_Map {
		writeMap(out, v, t)
	} else if tag == TAG_Proto {
		writeComplex(out, tagCustom, v, t)
	}
}

func writeByteArray(out io.Writer, v interface{}) {
	w(out, uint32(len(v.([]byte))))
	_, err := out.Write(v.([]byte))
	if err != nil {
		log.Panic(err.Error())
	}
}

func writeBaseArray(out io.Writer, v reflect.Value) {
	switch v.Type().Elem().Kind() {
	case reflect.Int8, reflect.Uint8:
		writeByteArray(out, v.Slice(0, v.Len()).Bytes())

	case reflect.Int16, reflect.Uint16, reflect.Int32, reflect.Uint32, reflect.Int64, reflect.Uint64, reflect.Float32, reflect.Float64:
		w(out, uint32(v.Len()))
		for i := 0; i < v.Len(); i++ {
			w(out, v.Index(i).Interface())
		}

	case reflect.Bool:
		w(out, uint32(v.Len()))
		trueLen := int(math.Ceil(float64(v.Len()) / 8.0))
		u256s := make([]byte, 0)
		for i := 0; i < v.Len(); i++ {
			bi := int(math.Floor(float64(i) / 8.0))
			si := i % 8
			if len(u256s) < bi+1 {
				u256s = append(u256s, 0)
			}
			data := v.Index(i).Interface().(bool)
			if data {
				u256s[bi] += 0x01 << si
			}
		}
		for i := 0; i < trueLen; i++ {
			w(out, u256s[i])
		}
	}
}

func writeList(out io.Writer, v reflect.Value, t reflect.Type) {
	tag := getListTypeByType(t)
	writeTag(out, tag)
	w(out, uint32(v.Len()))

	for i := 0; i < v.Len(); i++ {
		t0, v1, t1 := getTagId(v.Index(i), t.Elem())
		writeValue(out, t0, v1, t1)
	}
}

func writeMap(out io.Writer, v reflect.Value, t reflect.Type) {
	keyType := t.Key()
	keyKind := keyType.Kind()
	switch keyKind {
	case reflect.String:
		writeTag(out, TAG_String)
	case reflect.Int32:
		writeTag(out, TAG_Int)
	case reflect.Uint32:
		writeTag(out, TAG_UInt)
	case reflect.Int64:
		writeTag(out, TAG_Long)
	default:
		log.Panic(fmt.Sprintf("illegal key type %v", keyKind))
	}
	//这里临时建立一个空指针的reflect.Value
	elemTag, _, _ := getTagId(reflect.ValueOf(&t), t.Elem())
	writeTag(out, elemTag)
	length := uint32(len(v.MapKeys()))
	w(out, length)
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		typ := t.Elem()
		tag, v1, t1 := getTagId(value, typ)
		switch keyKind {
		case reflect.String:
			writeString(out, key.String())
		case reflect.Uint32, reflect.Int32, reflect.Int64:
			writeBaseValue(out, key, keyKind)
		default:
			log.Panic(fmt.Sprintf("illegal key type %v", keyKind))
		}
		writeValue(out, tag, v1, t1)
	}
}

func writeComplex(out io.Writer, tag BINARY_TAG, v reflect.Value, t reflect.Type) {
	if v.Kind() == reflect.Invalid {
		w(out, uint8(0))
		w(out, TAG_End)
		return
	}
	fields := parseStruct(v, t)
	w(out, uint8(len(fields)))
	for _, value := range fields {
		tag1, v1, t1 := getTagId(value, value.Type())
		writeTag(out, tag1)
		writeValue(out, tag1, v1, t1)
	}
	w(out, TAG_End)
}

func parseStruct(v reflect.Value, t reflect.Type) []reflect.Value {
	parsed := make([]reflect.Value, 0)
	if !v.IsValid() {
		return parsed
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			continue
			log.Panic("cannot have anonymous field")
		}
		name := f.Name

		if tag := f.Tag.Get("json"); tag == "-" {
			continue
		}
		if tag := f.Tag.Get("bson"); tag == "-" {
			continue
		}
		if tag := f.Tag.Get("bt"); tag != "" {
			name = tag
		}
		if name[0] < 65 || name[0] > 90 {
			continue
		}
		if name == "-" {
			continue
		}
		parsed = append(parsed, v.Field(i))
	}
	return parsed
}

//func writeCompound(out io.Writer, v reflect.Value, t reflect.LineType) {
//	v = reflect.Indirect(v)
//	fields := parseStruct2Map(v)
//	for name, value := range fields {
//
//		typ := t.Elem()
//		tag, v1, t1 := getTagId(value, typ)
//		writeTag(out, tag)
//		writeString(out, name)
//		writeValue(out, tag, v1, t1)
//	}
//	w(out, TAG_End)
//}

//func parseStruct2Map(v reflect.Value) map[string]reflect.Value {
//	parsed := make(map[string]reflect.Value)
//	t := v.LineType()
//
//	for i := 0; i < t.NumField(); i++ {
//		f := t.Field(i)
//		if f.Anonymous {
//			continue
//		}
//		name := f.StructName
//		if tag := f.Tag.Get("bt"); tag != "" {
//			name = tag
//		}
//		if name == "-" {
//			continue
//		}
//		if _, exists := parsed[name]; exists {
//			log.Panic(fmt.Sprintf("Multiple fields with name %#v", name))
//		}
//		parsed[name] = reflect.Indirect(v.Field(i))
//	}
//
//	return parsed
//}

func MarshalRouteMsg(msg *RouteMessage, t TYPE) (ret []byte, err error) {
	if t == JSON {
		return MarshalJsonRouteMsg(msg)
	} else if t == BINARY {
		return nil, errors.New("todo marsh binary")
	} else {
		return nil, errors.New("unidentified protocol")
	}
}

func MarshalBody(body ISerializable, t TYPE) (bodyData []byte, err error) {
	if t == JSON {
		data, err := json.Marshal(body)
		return data, err
	} else if t == BINARY {
		return nil, errors.New("todo marshal binary")
	} else {
		return nil, errors.New("unidentified protocol")
	}
}

func MarshalJsonRouteMsg(msg *RouteMessage) (ret []byte, err error) {
	var out bytes.Buffer
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
			} else {
				err = r.(error)
			}
		}
	}()

	v := reflect.ValueOf(msg.Body)
	t := reflect.TypeOf(msg.Body)

	tag, _, _ := getTagId(v, t)
	if tag <= TAG_Null {
		log.Panic("not a binary message type")
	}
	w(&out, uint16(0))
	w(&out, uint16(msg.InnerId))
	w(&out, msg.TransId)
	w(&out, uint64(msg.ToActor))
	w(&out, uint64(msg.FromActor))
	w(&out, msg.ReqType)
	data, _ := json.Marshal(msg.Body)
	w(&out, data)

	if out.Len() > 65535 {
		ret = out.Bytes()
		binary.LittleEndian.PutUint16(ret[0:2], uint16(0))
		return ret, nil
	}
	ret = out.Bytes()
	binary.LittleEndian.PutUint16(ret[0:2], uint16(out.Len()))
	return ret, nil
}
