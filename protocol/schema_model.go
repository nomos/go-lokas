package protocol

import (
	"github.com/nomos/go-log/log"
	"strings"
)

type ContainerType int

const (
	CTYPE_NONE  ContainerType = iota
	CTYPE_PROTO
	CTYPE_ARRAY
	CTYPE_MAP
)

type ModelSchema struct {
	Index         int
	Name          string
	Path          string
	EnumName      string
	Component     bool
	IsLongString  bool
	Type          BINARY_TAG
	ContainerType BINARY_TAG
	KeyType       BINARY_TAG
	Body          []*ModelSchema
	Depends       []string
	model  		*GoStructObject
}

func (this *ModelSchema) ToTsPublicType()string{
	if this.ContainerType == 0 {
		if this.Type.IsProto() {
			return this.Type.String()
		} else {
			return this.Type.TsTypeString()
		}
	} else if this.ContainerType == TAG_List {
		if this.Type.IsProto() {
			return this.Type.String()+"[]"
		} else {
			return this.Type.TsTypeString()+"[]"
		}

	} else if this.ContainerType == TAG_Map {
		mapStr:=""
		if this.KeyType == TAG_String||this.KeyType==TAG_Long{
			mapStr = "Dict<"
		} else {
			mapStr = "NumbericDict<"
		}

		if this.Type.IsProto() {
			return mapStr+this.Type.String()+">"
		} else {
			return mapStr+this.Type.TsTypeString()+">"
		}
	} else {
		log.Panicf("get type error",this)
		return ""
	}
}

func (this *ModelSchema) ToTsPublicSingleLine()string{
	ret:="\tpublic "
	ret+= this.Name+":"
	ret += this.ToTsPublicType()
	return ret
}

func (this *ModelSchema) ToTsClassHeader()string{
	compStr:="ISerializable"
	if this.Component {
		compStr = "IComponent"
	}
	ret:="export class "+this.Name+" extends "+compStr+" {"
	return ret
}

func (this *ModelSchema) ToSingleLine() string {
	ret := `@define("` + this.Name + `"`
	if len(this.Body) > 0 {
		log.Panicf("single line must not have members")
	}
	if len(this.Depends) > 0 {
		ret += ",[]"
		for _, depend := range this.Depends {
			ret += `,"` + depend + `"`
		}
	}
	ret += `)`
	return ret
}

func (this *ModelSchema) ToLineStart() string {
	return `@define("` + this.Name + `"` + `,[`
}

func (this *ModelSchema) ToLineObject() []string {
	ret := make([]string, 0)
	for _, body := range this.Body {
		str := "\t[" + `"`
		str += body.Name + `",`
		if body.ContainerType == TAG_List {
			str += "Tag.List"
			str += "," + body.Type.TsTagString()
		} else if body.ContainerType == TAG_Map {
			str += "Tag.Map"
			str += "," + body.KeyType.TsTagString()
			str += "," + body.Type.TsTagString()
		} else {
			str += body.Type.TsTagString()
		}
		if body.IsLongString {
			str+=",Tag.LongString"
		}
		str += "],"
		ret = append(ret, str)
	}
	return ret
}

func (this *ModelSchema) ToLineEnd() string {
	ret := "]"
	for _, depend := range this.Depends {
		ret += `,"` + depend + `"`
	}
	ret += `)`
	return ret
}

func (this *ModelSchema) ToJsonSchema() string {
	return ""
}

func (this *ModelSchema) SetGoRelativePath(goPath string) {
	this.Path = strings.Replace(goPath, ".go", ".ts", -1)
}
