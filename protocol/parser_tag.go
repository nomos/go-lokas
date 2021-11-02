package protocol

import (
	"regexp"
)

var tag_go_regexp_map = make(map[BINARY_TAG]*regexp.Regexp)
var tag_model_regexp_map = make(map[BINARY_TAG]*regexp.Regexp)

func init(){
	tag_go_regexp_map[TAG_Bool] = regexp.MustCompile(`(bool)`)
	tag_go_regexp_map[TAG_Byte] = regexp.MustCompile(`(uint8)`)
	tag_go_regexp_map[TAG_Short] = regexp.MustCompile(`(int16)`)
	tag_go_regexp_map[TAG_UShort] = regexp.MustCompile(`(uint16)`)
	tag_go_regexp_map[TAG_Int] = regexp.MustCompile(`(int32)`)
	tag_go_regexp_map[TAG_UInt] = regexp.MustCompile(`(uint32)`)
	tag_go_regexp_map[TAG_Long] = regexp.MustCompile(`(int64)`)
	tag_go_regexp_map[TAG_ULong] = regexp.MustCompile(`(uint64)`)
	tag_go_regexp_map[TAG_Float] = regexp.MustCompile(`(float32)`)
	tag_go_regexp_map[TAG_String] = regexp.MustCompile(`string`)
	//兼容proto3
	tag_go_regexp_map[TAG_Double] = regexp.MustCompile(`(float64|double)`)
	tag_go_regexp_map[TAG_BoolArray] = regexp.MustCompile(`(\[\]bool)`)
	tag_go_regexp_map[TAG_ByteArray] = regexp.MustCompile(`(\[\]uint8|\[\]byte)`)
	tag_go_regexp_map[TAG_ShortArray] = regexp.MustCompile(`(\[\]uint16|\[\]int16)`)
	tag_go_regexp_map[TAG_UShortArray] = regexp.MustCompile(`(\[\]uint16)`)
	tag_go_regexp_map[TAG_IntArray] = regexp.MustCompile(`(\[\]int32)`)
	tag_go_regexp_map[TAG_UIntArray] = regexp.MustCompile(`(\[\]uint32)`)
	tag_go_regexp_map[TAG_LongArray] = regexp.MustCompile(`(\[\]int64)`)
	tag_go_regexp_map[TAG_ULongArray] = regexp.MustCompile(`(\[\]uint64)`)
	tag_go_regexp_map[TAG_FloatArray] = regexp.MustCompile(`(\[\]float32)`)
	tag_go_regexp_map[TAG_DoubleArray] = regexp.MustCompile(`(\[\]float64)`)
	tag_go_regexp_map[TAG_List] = regexp.MustCompile(`\[\](([*]?(\w+[.])?(([A-Z]+\w*))|(string)))`)
	tag_go_regexp_map[TAG_Map] = regexp.MustCompile(`(map)\[(string|int64|int32|uint32)\][*]?(\w+[.])?(([A-Z]+\w*)|(string))`)
	tag_go_regexp_map[TAG_Buffer] = regexp.MustCompile(`([*]bytes.Buffer)`)
	tag_go_regexp_map[TAG_Time] = regexp.MustCompile(`time.Time`)
	tag_go_regexp_map[TAG_Decimal] = regexp.MustCompile(`decimal.Decimal`)


	tag_model_regexp_map[TAG_Bool] = regexp.MustCompile(`(bool)`)
	tag_model_regexp_map[TAG_Byte] = regexp.MustCompile(`(byte|uint8)`)
	tag_model_regexp_map[TAG_Short] = regexp.MustCompile(`(short|int16)`)
	tag_model_regexp_map[TAG_UShort] = regexp.MustCompile(`(ushort|uint16)`)
	tag_model_regexp_map[TAG_Int] = regexp.MustCompile(`(int|int32)`)
	tag_model_regexp_map[TAG_UInt] = regexp.MustCompile(`(uint|uint32)`)
	tag_model_regexp_map[TAG_Long] = regexp.MustCompile(`(long|int64)`)
	tag_model_regexp_map[TAG_ULong] = regexp.MustCompile(`(ulong|uint64)`)
	tag_model_regexp_map[TAG_Float] = regexp.MustCompile(`(float32|float)`)
	tag_model_regexp_map[TAG_String] = regexp.MustCompile(`string`)

	tag_model_regexp_map[TAG_Double] = regexp.MustCompile(`(float64|double)`)
	tag_model_regexp_map[TAG_BoolArray] = regexp.MustCompile(`(\[bool\])`)
	tag_model_regexp_map[TAG_ByteArray] = regexp.MustCompile(`(\[uint8\]|\[byte\])`)
	tag_model_regexp_map[TAG_ShortArray] = regexp.MustCompile(`(\[int16\]|\[short\])`)
	tag_model_regexp_map[TAG_UShortArray] = regexp.MustCompile(`(\[uint16\]|\[ushort\])`)
	tag_model_regexp_map[TAG_IntArray] = regexp.MustCompile(`(\[int32\]|\[int\])`)
	tag_model_regexp_map[TAG_UIntArray] = regexp.MustCompile(`(\[uint32\]|\[uint\])`)
	tag_model_regexp_map[TAG_LongArray] = regexp.MustCompile(`(\[int64\]|\[long\])`)
	tag_model_regexp_map[TAG_ULongArray] = regexp.MustCompile(`(\[uint64\]|\[ulong\])`)
	tag_model_regexp_map[TAG_FloatArray] = regexp.MustCompile(`(\[float32\]|\[float\])`)
	tag_model_regexp_map[TAG_DoubleArray] = regexp.MustCompile(`(\[float64\]|\[double\])`)
	tag_model_regexp_map[TAG_List] = regexp.MustCompile(`\[\s*(\w+)\s*\]`)
	tag_model_regexp_map[TAG_Map] = regexp.MustCompile(`\{\s*(\w+)\s*[\:|\,]\s*(\w+)\s*\}`)
	tag_model_regexp_map[TAG_Buffer] = regexp.MustCompile(`(bytes)`)
	tag_model_regexp_map[TAG_Time] = regexp.MustCompile(`time`)
	tag_model_regexp_map[TAG_Decimal] = regexp.MustCompile(`decimal`)
}

const tag_go_slice_item_reg = "$4"
const tag_go_map_key_reg = "$2"
const tag_go_map_key_item = "$4"

const tag_model_array_item_reg = "$1"
const tag_model_map_key_reg = "$1"
const tag_model_map_key_item = "$2"

func (this BINARY_TAG) MatchRegExp(str string)bool {
	return tag_go_regexp_map[this].FindString(str) == str&&str!=""
}

func MatchGoProtoTag(s string)bool {
	protoTagArr:=[]BINARY_TAG{TAG_Bool,TAG_Byte,TAG_Short,TAG_UShort,TAG_Int,TAG_UInt,TAG_Long,TAG_ULong,TAG_Float,TAG_Double,TAG_String,TAG_Time,TAG_Buffer}
	for _,tag:=range protoTagArr {
		if tag_go_regexp_map[tag].FindString(s) == s {
			return true
		}
	}
	return false
}

func MatchGoSystemTag(s string)(BINARY_TAG,string,string) {
	for tag,reg:=range tag_go_regexp_map {
		if reg.FindString(s) == s {
			if tag == TAG_List {
				return tag,reg.ReplaceAllString(s, tag_go_slice_item_reg),""
			}
			if tag== TAG_Map {
				return tag,reg.ReplaceAllString(s, tag_go_map_key_reg),reg.ReplaceAllString(s, tag_go_map_key_item)
			}
			return tag,"",""
		}
	}
	return 0,"",""
}

func MatchModelProtoTag(s string)BINARY_TAG {
	protoTagArr:=[]BINARY_TAG{TAG_Bool,TAG_Byte,TAG_Short,TAG_UShort,TAG_Int,TAG_UInt,TAG_Long,TAG_ULong,TAG_Float,TAG_Double,TAG_String,TAG_Time,TAG_BoolArray,TAG_ByteArray,TAG_UShortArray,TAG_ShortArray,TAG_UIntArray,TAG_IntArray,TAG_ULongArray,TAG_LongArray}
	for _,tag:=range protoTagArr {
		if tag_model_regexp_map[tag].FindString(s) == s {
			return tag
		}
	}
	return 0
}

func GetModelProtoTag(s string)BINARY_TAG {
	protoTagArr:=[]BINARY_TAG{TAG_Bool,TAG_Byte,TAG_Short,TAG_UShort,TAG_Int,TAG_UInt,TAG_Long,TAG_ULong,TAG_Float,TAG_Double,TAG_String,TAG_Time,TAG_BoolArray,TAG_ByteArray,TAG_UShortArray,TAG_ShortArray,TAG_UIntArray,TAG_IntArray,TAG_ULongArray,TAG_LongArray}
	for _,tag:=range protoTagArr {
		if tag_model_regexp_map[tag].FindString(s) == s {
			return tag
		}
	}
	return 0
}

func MatchModelSystemTag(s string)(BINARY_TAG,string,string) {
	for tag,reg:=range tag_model_regexp_map {
		if reg.FindString(s) == s {
			if tag == TAG_List {
				return tag,reg.ReplaceAllString(s, tag_model_array_item_reg),""
			}
			if tag== TAG_Map {
				return tag,reg.ReplaceAllString(s, tag_model_map_key_reg),reg.ReplaceAllString(s, tag_model_map_key_item)
			}
			return tag,"",""
		}
	}
	return 0,"",""
}