package protocol

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nomos/go-lokas/protocol/encoding/number_json"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/colors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func UnmarshalFromBytes(data []byte, v interface{}) error {
	reader := bytes.NewReader(data)
	err := unmarshalBinaryMessage(reader, v)
	if err != nil {
		return err
	}
	return nil
}

func UnmarshalMessage(data []byte, t TYPE) (*BinaryMessage, error) {
	if t == JSON {
		return UnmarshalJsonMessage(data)
	} else if t == BINARY {
		return UnmarshalBinaryMessage(data)
	} else {
		return nil, errors.New("unidentified protocol")
	}
}

func UnmarshalJsonMessage(data []byte) (*BinaryMessage, error) {
	header, err := unmarshalHeader(data)
	if err != nil {
		log.Error("unmarshalBinaryMessage header error",
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	if header.Len != 0 {
		data = data[:header.Len]
	}
	bodyData := data[HEADER_SIZE+2:]
	body, err := GetTypeRegistry().GetInterfaceByTag(header.CmdId)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	dec := number_json.NewDecoder(bytes.NewBuffer(bodyData))
	dec.UseNumber()
	err = dec.Decode(body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &BinaryMessage{
		CmdId:   header.CmdId,
		TransId: header.TransId,
		Len:     header.Len,
		Body:    body,
	}, nil
}

func UnmarshalBinaryMessage(data []byte) (*BinaryMessage, error) {
	header, err := unmarshalHeader(data)
	if err != nil {
		log.Error("unmarshalBinaryMessage header error",
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	body, err := unmarshalBodyByTag(header.CmdId, data)
	if err != nil {
		log.Error("unmarshalBinaryMessage body error",
			zap.Any("cmdId", header.CmdId),
			zap.Uint32("transId", header.TransId),
			zap.String("error", err.Error()),
		)
		return nil, err
	}
	return &BinaryMessage{
		CmdId:   header.CmdId,
		TransId: header.TransId,
		Len:     header.Len,
		Body:    body.(ISerializable),
	}, nil
}

func unmarshalHeader(data []byte) (*BinaryMessage, error) {
	length := binary.LittleEndian.Uint16(data[0:2])
	transId := binary.LittleEndian.Uint32(data[2:6])
	var cmdId BINARY_TAG
	cmdId = BINARY_TAG(binary.LittleEndian.Uint16(data[6:8]))
	//log.Infof("unmarshalHeader",cmdId,cmdId.String())
	return &BinaryMessage{
		CmdId:   cmdId,
		TransId: transId,
		Len:     length,
		Body:    nil,
	}, nil
}

func Unmarshal(data []byte, v interface{}) (err error) {
	reader := bytes.NewReader(data)
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
			} else {
				err = r.(error)
			}
		}
	}()
	new(decodeState).init(reader).unmarshal(v)
	return
}

func unmarshalBinaryMessage(in io.Reader, v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
			} else {
				err = r.(error)
			}
		}
	}()
	new(decodeState).init(in).unmarshalMessage(v)
	return
}

func unmarshalBodyByTag(cmdId BINARY_TAG, data []byte) (interface{}, error) {
	ret, err := GetTypeRegistry().GetInterfaceByTag(cmdId)
	if err != nil {
		log.Warn(err.Error())
		return nil, err
	}
	reader := bytes.NewReader(data)
	err = unmarshalBinaryMessage(reader, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type decodeState struct {
	in io.Reader
}

func (this *decodeState) init(in io.Reader) *decodeState {
	if in == nil {
		log.Panic(fmt.Sprintf("binary: Input stream is nil"))
	}
	this.in = in
	return this
}

func (this *decodeState) unmarshal(v interface{}) {
	tag := this.readTag()
	e := reflect.ValueOf(v).Elem()
	t := reflect.TypeOf(v).Elem()
	this.readValue(tag, e, t)
}

func (this *decodeState) unmarshalMessage(v interface{}) {
	_ = this.readLength()
	_ = this.readTransId()
	tag := this.readTag()
	//log.WithFields(log.Fields{
	//	"transId": transId,
	//	"length":  length,
	//	"tag":     tag,
	//}).Warn("")
	e := reflect.ValueOf(v).Elem()
	t := reflect.TypeOf(v).Elem()
	this.readValue(tag, e, t)
}

func (this *decodeState) r(i interface{}) {
	err := binary.Read(this.in, binary.LittleEndian, i)
	if err != nil {
		panic(err)
	}
}

func (this *decodeState) readTransId() uint32 {
	var transId uint32
	this.r(&transId)
	return transId
}

func (this *decodeState) readLength() uint16 {
	var len uint16
	this.r(&len)
	return len
}

func (this *decodeState) readTag() BINARY_TAG {
	var tag BINARY_TAG
	this.r(&tag)
	return tag
}

func (this *decodeState) allocate(tag BINARY_TAG) reflect.Value {
	switch tag {
	case TAG_Bool:
		return reflect.ValueOf(new(bool)).Elem()
	case TAG_Byte:
		return reflect.ValueOf(new(uint8)).Elem()
	case TAG_Short:
		return reflect.ValueOf(new(int16)).Elem()
	case TAG_UShort:
		return reflect.ValueOf(new(uint16)).Elem()
	case TAG_Int:
		return reflect.ValueOf(new(int32)).Elem()
	case TAG_UInt:
		return reflect.ValueOf(new(uint32)).Elem()
	case TAG_Long:
		return reflect.ValueOf(new(int64)).Elem()
	case TAG_ULong:
		return reflect.ValueOf(new(uint64)).Elem()
	case TAG_Float:
		return reflect.ValueOf(new(float32)).Elem()
	case TAG_Double:
		return reflect.ValueOf(new(float64)).Elem()
	case TAG_String:
		return reflect.ValueOf(new(string)).Elem()
	case TAG_BoolArray:
		return reflect.ValueOf(new([]bool)).Elem()
	case TAG_ByteArray:
		return reflect.ValueOf(new([]byte)).Elem()
	case TAG_ShortArray:
		return reflect.ValueOf(new([]int16)).Elem()
	case TAG_UShortArray:
		return reflect.ValueOf(new([]uint16)).Elem()
	case TAG_IntArray:
		return reflect.ValueOf(new([]int32)).Elem()
	case TAG_UIntArray:
		return reflect.ValueOf(new([]uint32)).Elem()
	case TAG_LongArray:
		return reflect.ValueOf(new([]int64)).Elem()
	case TAG_ULongArray:
		return reflect.ValueOf(new([]uint64)).Elem()
	case TAG_FloatArray:
		return reflect.ValueOf(new([]float32)).Elem()
	case TAG_DoubleArray:
		return reflect.ValueOf(new([]float64)).Elem()
	case TAG_List:
		return reflect.ValueOf(new([]interface{})).Elem()
	case TAG_Map:
		return reflect.ValueOf(new(map[string]interface{})).Elem()
	case TAG_Proto:
		return reflect.ValueOf(new(interface{})).Elem()
	case TAG_Null:
		return reflect.ValueOf(new(interface{})).Elem()
	}
	log.Panic(fmt.Sprintf("binary: Unhandled tag %s", tag))
	return reflect.Value{}
}

func (this *decodeState) readString() string {
	var length uint16
	this.r(&length)
	value := make([]byte, length)
	_, err := this.in.Read(value)
	if err != nil {
		panic(err)
	}
	return string(value)
}

func (this *decodeState) readDecimal() decimal.Decimal {
	var length uint16
	this.r(&length)
	value := make([]byte, length)
	_, err := this.in.Read(value)
	if err != nil {
		panic(err)
	}
	ret, err := decimal.NewFromString(string(value))
	if err != nil {
		panic(err)
	}
	return ret
}

func (this *decodeState) readBuffer() bytes.Buffer {
	var length uint16
	this.r(&length)
	value := make([]byte, length)
	_, err := this.in.Read(value)
	if err != nil {
		panic(err)
	}
	ret := *(bytes.NewBuffer(value))
	return ret
}

func (this *decodeState) readValue(tag BINARY_TAG, v reflect.Value, t reflect.Type) {
	kind := t.Kind()
	switch kind {
	case reflect.Int, reflect.Uint:
		log.Error("binary: int and uint types are not supported for portability reasons. Try int32 or uint32.")
		log.Panic(fmt.Sprintf("binary: int and uint types are not supported for portability reasons. Try int32 or uint32."))
	}

	if tag.IsBaseValue() {
		this.readBaseValue(tag, v)
		return
	} else if tag.IsArrayLike() {
		this.readArray(tag, v)
		return
	}
	if tag > TAG_Null {
		tag = TAG_Proto
	}
	switch tag {
	case TAG_String:
		switch kind {
		case reflect.String:
			v.SetString(this.readString())
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Decimal:
		value := this.readDecimal()
		switch v.Kind() {
		case reflect.Ptr:
			v1 := v.Elem()
			t = t.Elem()
			if v1.IsValid() {
				v = v1
			} else {
				v.Set(reflect.New(t))
			}
			v = reflect.Indirect(v)
			v.Set(reflect.ValueOf(value))
		case reflect.Struct:
			v.Set(reflect.ValueOf(value))
		}
	case TAG_Buffer:
		value := this.readBuffer()
		switch v.Kind() {
		case reflect.Ptr:
			v1 := v.Elem()
			t = t.Elem()
			if v1.IsValid() {
				v = v1
			} else {
				v.Set(reflect.New(t))
			}
			v = reflect.Indirect(v)
			v.Set(reflect.ValueOf(value))
		case reflect.Struct:
			v.Set(reflect.ValueOf(value))
		}
	case TAG_Color:
		var value uint32
		this.r(&value)
		switch v.Kind() {
		case reflect.Ptr:
			v1 := v.Elem()
			t = t.Elem()
			if v1.IsValid() {
				v = v1
			} else {
				v.Set(reflect.New(t))
			}
			v = reflect.Indirect(v)
			v.Set(reflect.ValueOf(colors.NewColorUint32(value)))
		case reflect.Struct:
			v.Set(reflect.ValueOf(colors.NewColorUint32(value)))
		}
	case TAG_Time:
		var value int64
		this.r(&value)
		switch v.Kind() {
		case reflect.Ptr:
			v1 := v.Elem()
			t = t.Elem()
			if v1.IsValid() {
				v = v1
			} else {
				v.Set(reflect.New(t))
			}
			v = reflect.Indirect(v)
			v.Set(reflect.ValueOf(time.Unix(0, value*time.Millisecond.Nanoseconds())))
		case reflect.Struct:
			v.Set(reflect.ValueOf(time.Unix(0, value*time.Millisecond.Nanoseconds())))
		}
	case TAG_List:
		this.readList(tag, v, t)
	case TAG_Proto:
		this.readComplex(tag, v, t)

	case TAG_Map:
		switch v.Kind() {

		case reflect.Map:
			if v.IsNil() {
				v.Set(reflect.MakeMap(v.Type()))
			}
			keyTag := this.readTag()
			switch keyTag {
			case TAG_String, TAG_Int, TAG_Long:
			default:
				log.Panic(fmt.Sprintf("illegal key type %s",
					fmt.Sprintf("%s (0x%02x)", keyTag, byte(keyTag))))
			}
			valueTag := this.readTag()
			var length uint32
			this.r(&length)
			defer func() {
				if r := recover(); r != nil {
					log.Panic(fmt.Sprintf("%v\n\t\tat struct field", r))
				}
			}()
			for index := 0; index < int(length); index++ {
				var keyValue reflect.Value
				switch keyTag {
				case TAG_String:
					keyValue = reflect.ValueOf(this.readString())
				case TAG_Int, TAG_Long:
					keyValue = reflect.New(v.Type().Key()).Elem()
					this.readBaseValue(keyTag, keyValue)
				}
				t1 := v.Type().Elem()
				var val reflect.Value
				if t1.Kind() == reflect.Ptr {
					val = reflect.New(t1.Elem())
					this.readValue(valueTag, val, t1)
					v.SetMapIndex(keyValue, val)
				} else {
					val = reflect.New(t1)
					this.readValue(valueTag, val.Elem(), t1)
					v.SetMapIndex(keyValue, val.Elem())
				}
			}
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_End:
		return
	default:
		log.Panic(fmt.Sprintf("binary: Unhandled tag: %s", tag))
	}
}

func (this *decodeState) readComplex(tag BINARY_TAG, v reflect.Value, t reflect.Type) {
	var length uint8
	this.r(&length)
	tag0 := this.readTag()
	if tag0 == TAG_End {
		return
	}
	switch v.Kind() {
	case reflect.Interface:
		log.Error("cannot use interface")
		panic("cannot use interface")
	case reflect.Ptr:
		v1 := v.Elem()
		t = t.Elem()
		if v1.IsValid() {
			v = v1
		} else {
			v.Set(reflect.New(t))
		}
		v = reflect.Indirect(v)
	case reflect.Struct:
		v.Set(reflect.Zero(t))

	default:
		log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %v!", tag0, t.Kind()))
	}
	fields := parseStruct(v, t)
	defer func() {
		if r := recover(); r != nil {
			log.Panic(fmt.Sprintf("%v", r))
		}
	}()
	for index, field := range fields {
		var tag1 BINARY_TAG
		if index == 0 {
			tag1 = tag0
		} else {
			tag1 = this.readTag()
		}
		if tag1 == TAG_End {
			return
		}
		this.readValue(tag1, field, field.Type())
	}
	tag1 := this.readTag()
	if tag1 != TAG_End {
		log.Error("complex read error")
	}
}

func (this *decodeState) readList(tag BINARY_TAG, v reflect.Value, t reflect.Type) {
	v = reflect.Indirect(v)
	var inner = this.readTag()
	var length uint32
	this.r(&length)

	switch v.Kind() {
	case reflect.Slice:
		if uint32(v.Cap()) < length {
			v.Set(reflect.MakeSlice(v.Type(), 0, int(length)))
		} else {
			v.Set(v.Slice(0, 0))
		}
		kind := v.Type().Elem()

		var i uint32
		defer func() {
			if r := recover(); r != nil {
				log.Panic(fmt.Sprintf("%v\n\t\tat list index %d", r, i))
			}
		}()

		for i = 0; i < length; i++ {
			var value reflect.Value
			if kind.Kind() == reflect.Ptr {
				value = reflect.New(kind.Elem())
				this.readValue(inner, value.Elem(), t.Elem().Elem())
			} else {
				if kind.Kind() == reflect.Interface {
					log.Error("cannot use interface")
					panic("cannot use interface")
				} else {
					value = reflect.New(kind).Elem()
				}
				this.readValue(inner, value, t.Elem())
			}
			v.Set(reflect.Append(v, value))
		}

	case reflect.Array:

	default:
		log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
	}
}

func (this *decodeState) readArray(tag BINARY_TAG, v reflect.Value) {
	switch tag {
	case TAG_BoolArray:
		var length uint32
		this.r(&length)

		trueLen := int(math.Ceil(float64(length) / 8.0))
		u256s := make([]byte, trueLen, trueLen)
		for i := 0; i < trueLen; i++ {
			var item uint8
			this.r(&item)
			u256s[i] = item
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			if v.Kind() == reflect.Array {
				if uint32(v.Len()) < length {
					log.Panic(fmt.Sprintf("binary: Byte array is of length %d, but only the array given is only %d long!", length, v.Len()))
				}
			} else {
				if uint32(v.Len()) < length {
					v.Set(reflect.MakeSlice(v.Type(), int(length), int(length)))
				}
			}
			for i := 0; i < int(length); i++ {
				bi := int(math.Floor(float64(i) / 8.0))
				si := i % 8
				bNumber := u256s[bi]
				data := bNumber >> si & 1
				value := v.Index(i)
				if data == 1 {
					value.SetBool(true)
				} else {
					value.SetBool(false)
				}
			}

		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}

	case TAG_ByteArray, TAG_ShortArray, TAG_UShortArray, TAG_IntArray, TAG_UIntArray, TAG_LongArray, TAG_ULongArray, TAG_FloatArray, TAG_DoubleArray:
		var length uint32
		this.r(&length)
		innerTag := tag.getArrayBaseType()
		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			if v.Kind() == reflect.Array {
				if uint32(v.Len()) < length {
					log.Panic(fmt.Sprintf("binary: Byte array is of length %d, but only the array given is only %d long!", length, v.Len()))
				}
			} else {
				if uint32(v.Len()) < length {
					v.Set(reflect.MakeSlice(v.Type(), int(length), int(length)))
				}
			}

			for i := 0; i < int(length); i++ {
				value := v.Index(i)
				this.readBaseValue(innerTag, value)
			}

		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	}
}

func (this *decodeState) readBaseValue(tag BINARY_TAG, v reflect.Value) {
	switch tag {
	case TAG_Bool:
		var value uint8
		this.r(&value)
		switch v.Kind() {
		case reflect.Bool:
			v.SetBool(value != 0)
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Byte:
		var value byte
		this.r(&value)
		switch v.Kind() {
		case reflect.Bool:
			v.SetBool(value != 0)
		case reflect.Uint8:
			v.SetUint(uint64(value))
		case reflect.Int8:
			v.SetInt(int64(value))
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Short:
		var value int16
		this.r(&value)
		switch v.Kind() {
		case reflect.Int16:
			v.SetInt(int64(value))
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_UShort:
		var value uint16
		this.r(&value)
		switch v.Kind() {
		case reflect.Uint16:
			v.SetInt(int64(value))
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Int:
		var value int32
		this.r(&value)
		switch v.Kind() {
		case reflect.Int32:
			v.SetInt(int64(value))
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_UInt:
		var value uint32
		this.r(&value)
		switch v.Kind() {
		case reflect.Uint32:
			v.SetUint(uint64(value))
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Long:
		var value int64
		this.r(&value)
		switch v.Kind() {
		case reflect.Int64:
			v.SetInt(value)
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_ULong:
		var value uint64
		this.r(&value)
		switch v.Kind() {
		case reflect.Uint64:
			v.SetUint(value)
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Float:
		var value float32
		this.r(&value)
		switch v.Kind() {
		case reflect.Float32:
			v.SetFloat(float64(value))
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	case TAG_Double:
		var value float64
		this.r(&value)
		switch v.Kind() {
		case reflect.Float64:
			v.SetFloat(value)
		default:
			log.Panic(fmt.Sprintf("binary: Tag is %s, but I don't know how to put that in a %s!", tag, v.Kind()))
		}
	}
}

// func UnmarshalBody(data []byte, t TYPE, v interface{}) error {
// 	if t == JSON {
// 		err := json.Unmarshal(data, &v)
// 		if err != nil {
// 			log.Error(err.Error())
// 			return err
// 		}
// 	} else if t == BINARY {
// 		// return UnmarshalBinaryMessage(data)
// 	} else {
// 		return errors.New("unkonw message type")
// 	}

// 	return nil
// }

func UnmarshalRouteMsg(data []byte, t TYPE) (*RouteMessage, error) {
	if t == JSON {
		return UnmarshalJsonRouteMsg(data)
	} else if t == BINARY {
		return nil, errors.New("todo binary format")
	} else {
		return nil, errors.New("unidentified protocol")
	}
}

func UnmarshalJsonRouteMsg(data []byte) (*RouteMessage, error) {
	routeMsg, headerSize, err := unmarshalRouteMsgHeader(data)
	if err != nil {
		log.Error("UnmarshalJsonRouteMsg header error", zap.String("error", err.Error()))
		return nil, err
	}

	bodyData := data[headerSize:routeMsg.Len]
	body, err := GetTypeRegistry().GetInterfaceByTag(routeMsg.CmdId)
	if err != nil {
		log.Error("not find cmd", zap.Uint16("cmd", uint16(routeMsg.CmdId)), zap.String("err", err.Error()))
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewBuffer(bodyData))
	dec.UseNumber()
	err = dec.Decode(body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	routeMsg.Body = body

	return routeMsg, nil
}

func unmarshalRouteMsgHeader(data []byte) (*RouteMessage, int, error) {
	headerSize := ROUTE_MSG_HEAD_SIZE

	if len(data) < headerSize {
		return nil, 0, ERR_MSG_FORMAT
	}

	routeMsg := &RouteMessage{}

	routeMsg.Len = binary.LittleEndian.Uint16(data[0:2])
	routeMsg.InnerId = BINARY_TAG(binary.LittleEndian.Uint16(data[2:4]))
	routeMsg.TransId = binary.LittleEndian.Uint32(data[4:8])
	routeMsg.ToActor = util.ID(binary.LittleEndian.Uint64(data[8:16]))
	routeMsg.FromActor = util.ID(binary.LittleEndian.Uint64(data[16:24]))
	routeMsg.ReqType = uint8(data[24])

	if routeMsg.ReqType == REQ_TYPE_REPLAY {
		routeMsg.Req = false
	} else {
		routeMsg.Req = true
	}

	// routeMsg.FromActor = sess.GetId()
	routeMsg.CmdId = routeMsg.InnerId

	return routeMsg, headerSize, nil
}
