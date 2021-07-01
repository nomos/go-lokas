package protocol

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/promise"
	"strconv"
)

type ProtoSyntaxObj struct {
	DefaultGeneratorObj
}

func NewProtoSyntaxObj(file GeneratorFile) *ProtoSyntaxObj {
	ret := &ProtoSyntaxObj{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_PROTO_SYNTAX, file)
	return ret
}

func (this *ProtoSyntaxObj) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_SYNTAX) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		return false
	}
	log.Panic("parse ProtoSyntaxObj error")
	return false
}

type ProtoPackageObj struct {
	DefaultGeneratorObj
}

func NewProtoPackageObj(file GeneratorFile) *ProtoPackageObj {
	ret := &ProtoPackageObj{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_PROTO_PACKAGE, file)
	return ret
}

func (this *ProtoPackageObj) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_PACKAGE) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		return false
	}
	log.Panic("parse ProtoPackageObj error")
	return false
}

type ProtoField struct {
	Repeated bool
	Map bool
	Proto bool
	Enum bool
	KeyType string
	Type string
	Id int
}

type ProtoMsgObject struct {
	DefaultGeneratorObj
	ProtoName string
	ProtoId int
	Fields []*ProtoField
}

func NewProtoMsgObject(file GeneratorFile) *ProtoMsgObject {
	ret := &ProtoMsgObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
		Fields: []*ProtoField{},
	}
	ret.DefaultGeneratorObj.init(OBJ_PROTO_MSG, file)
	return ret
}

func (this *ProtoMsgObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_HEADER) {
			this.ProtoName = line.GetName()
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_FIELD) {
			this.CheckField(line)
			return true
		}
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		log.Panic("parse ProtoMsgObject error")
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse ProtoMsgObject error")
	return false
}

func (this *ProtoMsgObject) CheckField(line *LineText){
	s := COMMENT_REGEXP.ReplaceAllString(line.Text, "$1")
	field:=&ProtoField{
		Repeated: false,
		Map:      false,
		Proto:    false,
		KeyType:  "",
		Type:     "",
		Id:       0,
	}
	isMap:=LINE_PROTO_FIELD.RegExp().ReplaceAllString(s,"$2")!=""
	if isMap {
		field.Map = true
		field.KeyType = LINE_PROTO_FIELD.RegExp().ReplaceAllString(s,"$3")
		field.Type = LINE_PROTO_FIELD.RegExp().ReplaceAllString(s,"$4")
		if !MatchGoProtoTag(field.Type) {
			field.Proto = true
		}
		this.Fields = append(this.Fields,  field)
		return
	}
	field.Repeated = LINE_PROTO_FIELD.RegExp().ReplaceAllString(s,"$7")!=""
	field.Type = LINE_PROTO_FIELD.RegExp().ReplaceAllString(s,"$9")
	if !MatchGoProtoTag(field.Type) {
		field.Proto = true
	}
	this.Fields = append(this.Fields,  field)
}

type ProtoEnumObject struct {
	DefaultGeneratorObj
	Alias bool
	Fields []*ProtoField
}

func NewProtoEnumObject(file GeneratorFile) *ProtoEnumObject {
	ret := &ProtoEnumObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
		Alias: false,
		Fields: []*ProtoField{},
	}
	ret.DefaultGeneratorObj.init(OBJ_PROTO_ENUM, file)
	return ret
}

func (this *ProtoEnumObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_ENUM_HEADER) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_ENUM_FIELD) {
			this.CheckField(line)
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_ENUM_ALIAS) {
			this.Alias = true
			return true
		}
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		log.Panic("parse ProtoEnumObject error")
	} else if this.state == 2 {
		log.Warnf("field",this.Fields)
		return false
	}
	log.Panic("parse ProtoEnumObject error")
	return false
}

func (this *ProtoEnumObject) CheckField(line *LineText) {
	s := COMMENT_REGEXP.ReplaceAllString(line.Text, "$1")
	field:=&ProtoField{
		Repeated: false,
		Map:      false,
		Proto:    false,
		KeyType:  "",
		Type:     "",
		Id:       0,
	}
	field.KeyType = LINE_PROTO_ENUM_FIELD.RegReplaceName(s)
	field.Id,_ = strconv.Atoi(LINE_PROTO_ENUM_FIELD.RegReplaceValue(s))
	this.Fields = append(this.Fields, field)
}


type ProtoFile struct {
	*DefaultGeneratorFile
}

var _ GeneratorFile = (*ProtoFile)(nil)

func NewProtoFile(generator *Generator) *ProtoFile {
	ret := &ProtoFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_PROTO
	return ret
}

func (this *ProtoFile) Generate() *promise.Promise {
	return nil
}

func (this *ProtoFile) Parse() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0,OBJ_PROTO_SYNTAX)
		log.Warnf("parseProtoSyntax", offset, success)
		offset, success = this.parse(offset, OBJ_PROTO_PACKAGE)
		log.Warnf("parseProtoPackage", offset, success)
		offset, success = this.parse(offset, OBJ_COMMENT,OBJ_PROTO_ENUM, OBJ_PROTO_MSG)
		//offset, success = this.parseGoImports(offset, nil)
		//log.Warnf("parseGoImports", offset, success)
		//offset, success = this.parseGoMain(offset, nil)
		//log.Warnf("parseGoMain finish", offset, success)
		resolve(nil)
	})
}

func (this *ProtoFile) ProcessProtos()[]*ProtoMsgObject {
	ret:=make([]*ProtoMsgObject,0)
	for _,obj:=range this.Objects {
		if obj.ObjectType() == OBJ_PROTO_MSG {
			ret = append(ret, obj.(*ProtoMsgObject))
		}
	}
	return ret
}
