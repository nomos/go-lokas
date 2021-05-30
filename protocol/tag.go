package protocol

import (
	"fmt"
	"github.com/nomos/go-log/log"
)

const (
	//基础类型 0-40
	TAG_End BINARY_TAG = iota
	TAG_Bool
	TAG_Byte
	TAG_Short
	TAG_UShort
	TAG_Int
	TAG_UInt
	TAG_Long
	TAG_ULong
	TAG_Float
	TAG_Double
	TAG_String
	TAG_BoolArray
	TAG_ByteArray
	TAG_ShortArray
	TAG_UShortArray
	TAG_IntArray
	TAG_UIntArray
	TAG_LongArray
	TAG_ULongArray
	TAG_FloatArray
	TAG_DoubleArray
	TAG_List
	TAG_Map
	TAG_Buffer
	TAG_Time
	TAG_Decimal
	TAG_Proto
	TAG_Null
	TAG_LongString
	//Ecs
	TAG_EntityRef BINARY_TAG = 32
	//系统预留类型 40-127
	TAG_BinaryMessage BINARY_TAG = 40 //保留基本传输类型
	TAG_Error         BINARY_TAG = 41
	TAG_Compose       BINARY_TAG = 42
	TAG_Ping          BINARY_TAG = 43
	TAG_Pong          BINARY_TAG = 44
	TAG_HandShake     BINARY_TAG = 45
	TAG_RouteMessage  BINARY_TAG = 46
	//内建类型
)

type BINARY_TAG uint16
type BINARY_NAME_SIZE int16
type BINARY_LENGTH uint32

func (tag BINARY_TAG) IsProto() bool {
	return tag > TAG_Null
}

func (tag BINARY_TAG) IsArrayLike() bool {
	switch tag {
	case TAG_BoolArray, TAG_ByteArray, TAG_ShortArray,TAG_UShortArray, TAG_IntArray,TAG_UIntArray, TAG_LongArray,TAG_ULongArray, TAG_FloatArray, TAG_DoubleArray:
		return true
	default:
		return false
	}
}
func (tag BINARY_TAG) IsBaseValue() bool {
	switch tag {
	case TAG_Bool, TAG_Byte, TAG_Short,TAG_UShort, TAG_Int,TAG_UInt, TAG_Long,TAG_ULong, TAG_Float, TAG_Double:
		return true
	default:
		return false
	}
}

func (tag BINARY_TAG) getArrayBaseType() BINARY_TAG {
	switch tag {
	case TAG_BoolArray:
		return TAG_Bool
	case TAG_ByteArray:
		return TAG_Byte
	case TAG_ShortArray:
		return TAG_Short
	case TAG_UShortArray:
		return TAG_UShort
	case TAG_IntArray:
		return TAG_Int
	case TAG_UIntArray:
		return TAG_UInt
	case TAG_LongArray:
		return TAG_Long
	case TAG_ULongArray:
		return TAG_ULong
	case TAG_FloatArray:
		return TAG_Float
	case TAG_DoubleArray:
		return TAG_Double
	default:
		panic(fmt.Sprintf("arr type %d not recognized", tag))
	}
}

func (tag BINARY_TAG) String() string {
	name, ok := GetTypeRegistry().tagNameMap[tag]
	if !ok {
		name = "Tag_Unknown"
	}
	return name
}

func (tag BINARY_TAG) TsTagString() string {
	name, ok := GetTypeRegistry().tagNameMap[tag]
	if !ok {
		log.Panicf("not valid type", tag, GetTypeRegistry().tagNameMap)
	}
	if !tag.IsProto() {
		return "Tag." + name
	}
	return name
}

func (tag BINARY_TAG) TsTypeString() string {
	return system_ts_types[tag]
}

func (tag BINARY_TAG) CsTypeString() string {
	return system_cs_types[tag]
}

func (tag BINARY_TAG) GoTypeString() string {
	return system_go_types[tag]
}

type Compression byte

const (
	Uncompressed Compression = 0
	GZip         Compression = 1
	ZLib         Compression = 2
)
