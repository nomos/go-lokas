package protocol

import (
	"github.com/nomos/go-lokas/log"
	"reflect"
	"strconv"
	"sync"
)

var system_tags = make(map[BINARY_TAG]string)
var system_ts_types = make(map[BINARY_TAG]string)
var system_go_types = make(map[BINARY_TAG]string)
var system_cs_types = make(map[BINARY_TAG]string)

func init(){
	system_tags[TAG_End] = "End"
	system_tags[TAG_Bool] = "Bool"
	system_tags[TAG_Byte] = "Byte"
	system_tags[TAG_Short] = "Short"
	system_tags[TAG_UShort] = "UShort"
	system_tags[TAG_Int] = "Int"
	system_tags[TAG_UInt] = "UInt"
	system_tags[TAG_Long] = "Long"
	system_tags[TAG_ULong] = "ULong"
	system_tags[TAG_Float] = "Float"
	system_tags[TAG_Double] = "Double"
	system_tags[TAG_String] = "String"
	system_tags[TAG_BoolArray] = "Bool_Array"
	system_tags[TAG_ByteArray] = "Byte_Array"
	system_tags[TAG_ShortArray] = "Short_Array"
	system_tags[TAG_UShortArray] = "UShort_Array"
	system_tags[TAG_IntArray] = "Int_Array"
	system_tags[TAG_UIntArray] = "UInt_Array"
	system_tags[TAG_LongArray] = "Long_Array"
	system_tags[TAG_ULongArray] = "ULong_Array"
	system_tags[TAG_FloatArray] = "Float_Array"
	system_tags[TAG_DoubleArray] = "Double_Array"
	system_tags[TAG_List] = "List"
	system_tags[TAG_Map] = "Map"
	system_tags[TAG_Buffer] = "Buffer"
	system_tags[TAG_Time] = "Time"
	system_tags[TAG_Decimal] = "Decimal"
	system_tags[TAG_Proto] = "Proto"
	system_tags[TAG_Null] = "Null"


	system_ts_types[TAG_Bool] = "boolean"
	system_ts_types[TAG_Byte] = "number"
	system_ts_types[TAG_Short] = "number"
	system_ts_types[TAG_UShort] = "number"
	system_ts_types[TAG_Int] = "number"
	system_ts_types[TAG_UInt] = "number"
	system_ts_types[TAG_Long] = "number"
	system_ts_types[TAG_ULong] = "number"
	system_ts_types[TAG_Float] = "number"
	system_ts_types[TAG_Double] = "number"
	system_ts_types[TAG_String] = "string"
	system_ts_types[TAG_BoolArray] = "bool[]"
	system_ts_types[TAG_ByteArray] = "number[]"
	system_ts_types[TAG_ShortArray] = "number[]"
	system_ts_types[TAG_UShortArray] = "number[]"
	system_ts_types[TAG_IntArray] = "number[]"
	system_ts_types[TAG_UIntArray] = "number[]"
	system_ts_types[TAG_LongArray] = "number[]"
	system_ts_types[TAG_ULongArray] = "number[]"
	system_ts_types[TAG_FloatArray] = "number[]"
	system_ts_types[TAG_DoubleArray] = "number[]"
	system_ts_types[TAG_Buffer] = "ByteBuffer"
	system_ts_types[TAG_Time] = "Date"
	system_ts_types[TAG_Decimal] = "bigDecimal"


	system_cs_types[TAG_Bool] = "bool"
	system_cs_types[TAG_Byte] = "byte"
	system_cs_types[TAG_Short] = "short"
	system_cs_types[TAG_UShort] = "ushort"
	system_cs_types[TAG_Int] = "int"
	system_cs_types[TAG_UInt] = "uint"
	system_cs_types[TAG_Long] = "long"
	system_cs_types[TAG_ULong] = "ulong"
	system_cs_types[TAG_Float] = "float"
	system_cs_types[TAG_Double] = "double"
	system_cs_types[TAG_String] = "string"
	system_cs_types[TAG_BoolArray] = "List<bool>"
	system_cs_types[TAG_ByteArray] = "byte[]"
	system_cs_types[TAG_ShortArray] = "List<short>"
	system_cs_types[TAG_UShortArray] = "List<ushort>"
	system_cs_types[TAG_IntArray] = "List<int>"
	system_cs_types[TAG_UIntArray] = "List<uint>"
	system_cs_types[TAG_LongArray] = "List<long>"
	system_cs_types[TAG_ULongArray] = "List<ulong>"
	system_cs_types[TAG_FloatArray] = "List<float>"
	system_cs_types[TAG_DoubleArray] = "List<double>"
	system_cs_types[TAG_Buffer] = "MemoryStream"
	system_cs_types[TAG_Time] = "DateTime"
	system_cs_types[TAG_Decimal] = "decimal"


	system_go_types[TAG_Bool] = "bool"
	system_go_types[TAG_Byte] = "byte"
	system_go_types[TAG_Short] = "int16"
	system_go_types[TAG_UShort] = "uint16"
	system_go_types[TAG_Int] = "int32"
	system_go_types[TAG_UInt] = "uint32"
	system_go_types[TAG_Long] = "int64"
	system_go_types[TAG_ULong] = "uint64"
	system_go_types[TAG_Float] = "float32"
	system_go_types[TAG_Double] = "float64"
	system_go_types[TAG_String] = "string"
	system_go_types[TAG_BoolArray] = "[]bool"
	system_go_types[TAG_ByteArray] = "[]byte"
	system_go_types[TAG_ShortArray] = "[]int16"
	system_go_types[TAG_UShortArray] = "[]uint16"
	system_go_types[TAG_IntArray] = "[]int32"
	system_go_types[TAG_UIntArray] = "[]uint32"
	system_go_types[TAG_LongArray] = "[]int64"
	system_go_types[TAG_ULongArray] = "[]uint64"
	system_go_types[TAG_FloatArray] = "[]float32"
	system_go_types[TAG_DoubleArray] = "[]float64"
	system_go_types[TAG_Buffer] = "bytes.Buffer"
	system_go_types[TAG_Time] = "time.Time"
	system_go_types[TAG_Decimal] = "decimal.Decimal"
	GetTypeRegistry()
}

type Enum int32

func (this Enum) Enum() Enum {
	return this
}

type IEnum interface {
	ToString()string
	Enum() Enum
}

type IEnumCollection []IEnum

func (this IEnumCollection) GetEnumByString(s string)IEnum{
	for _,v:=range this {
		if v.ToString()==s {
			return v
		}
	}
	return nil
}

func (this IEnumCollection) HasInt32(i int32)bool{
	for _,v:=range this {
		if int32(v.Enum()) == i {
			return true
		}
	}
	return false
}

func (this IEnumCollection) GetEnumByValue(s Enum)IEnum{
	for _,v:=range this {
		if v.Enum()==s {
			return v
		}
	}
	return nil
}

var once sync.Once
var singleton *TypeRegistry

type TypeRegistry struct {
	tagMap map[BINARY_TAG]reflect.Type
	typeMap map[reflect.Type]BINARY_TAG
	nameMap map[reflect.Type]string
	tagNameMap map[BINARY_TAG]string
	nameTagMap map[string]BINARY_TAG
}

func SetTypeRegistry(registry *TypeRegistry) {
	singleton = registry
}

//单例,进程唯一
func GetTypeRegistry()*TypeRegistry {
	if singleton == nil {
		once.Do(func() {
			//log.Info("init type registry")
			singleton = &TypeRegistry{
				tagMap:  make(map[BINARY_TAG]reflect.Type),
				typeMap: make(map[reflect.Type]BINARY_TAG),
				nameMap: make(map[reflect.Type]string),
				tagNameMap: make(map[BINARY_TAG]string),
				nameTagMap: make(map[string]BINARY_TAG),
			}
			for k,v:=range system_tags {
				singleton.RegistrySystemTag(k,v)
			}
			singleton.RegistryType(TAG_BinaryMessage,reflect.TypeOf((*BinaryMessage)(nil)).Elem())
			singleton.RegistryType(TAG_Error,reflect.TypeOf((*ErrMsg)(nil)).Elem())
			singleton.RegistryType(TAG_Compose,reflect.TypeOf((*ComposeData)(nil)).Elem())
			singleton.RegistryType(TAG_Ping,reflect.TypeOf((*Ping)(nil)).Elem())
			singleton.RegistryType(TAG_Pong,reflect.TypeOf((*Pong)(nil)).Elem())
			singleton.RegistryType(TAG_HandShake,reflect.TypeOf((*HandShake)(nil)).Elem())
		})
	}
	return singleton
}

func (this *TypeRegistry) RegistryTemplateTag(t BINARY_TAG,s string){
	this.tagNameMap[t] = s
}

func (this *TypeRegistry) RegistrySystemTag(t BINARY_TAG,s string){
	this.tagNameMap[t] = s
	this.nameTagMap[s] = t
}

func (this *TypeRegistry) GetNameByType(t reflect.Type)string {
	name,ok:= this.nameMap[t]
	if !ok {
		name = "Unknown"
	}
	return name
}

func (this *TypeRegistry) GetTagName(t BINARY_TAG)string {
	name,ok:= this.tagNameMap[t]
	if !ok {
		name = "Unknown"
	}
	return name
}

func (this *TypeRegistry) RegistryType(t BINARY_TAG,p reflect.Type) {
	if t<=TAG_Null {
		log.Panic("tag id must > "+strconv.Itoa(int(TAG_Null)))
	}
	if t>=128&&t>65535 {
		log.Panic("tag id must be <128 or >65535")
	}
	//log.Infof("RegistryType",t,p.Name())
	this.tagMap[t] = p
	this.typeMap[p] = t
	this.nameMap[p] = p.Name()
	this.tagNameMap[t] = p.Name()
	this.nameTagMap[p.Name()] = t
}

func (this *TypeRegistry) GetTagByName(s string) BINARY_TAG{
	if ret,ok:=this.nameTagMap[s];ok{
		return ret
	} else {
		return 0
	}
}

func (this *TypeRegistry) GetTagByType(p reflect.Type) (BINARY_TAG,error){
	if ret,ok:=this.typeMap[p];ok{
		return ret,nil
	} else {
		return 0, ERR_TYPE_NOT_FOUND
	}
}

func (this *TypeRegistry) GetTypeByTag(t BINARY_TAG) (reflect.Type,error){
	if ret,ok:=this.tagMap[t];ok{
		return ret,nil
	} else {
		return nil, ERR_TYPE_NOT_FOUND
	}
}

func (this *TypeRegistry) GetInterfaceByTag(t BINARY_TAG) (ISerializable,error){
	if ret,ok:=this.tagMap[t];ok{
		return reflect.New(ret).Interface().(ISerializable),nil
	} else {
		return nil, ERR_TYPE_NOT_FOUND
	}
}