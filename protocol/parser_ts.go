package protocol

import (
	"github.com/nomos/go-lokas/util/promise"
)

const (
	TS_CLASSOBJ_HEADER = iota
	TS_CLASSOBJ_DEFINER
	TS_CLASSOBJ_MEMBER_FIELD
	TS_CLASSOBJ_CONSTRUCTOR
	TS_CLASSOBJ_GETTER_SETTER
	TS_CLASSOBJ_FUNCTION
	TS_CLASSOBJ_CLASS_END
)

type TsVarObject struct {
	DefaultGeneratorObj
}

func NewTsVarObject(file GeneratorFile) *TsVarObject {
	ret := &TsVarObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_TS_VAR, file)
	return ret
}

func (this *TsVarObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_VAR_SINGLELINE) {
			this.state = 3
			return true
		}
		if this.TryAddLine(line, LINE_TS_VAR_CLOSURE_START) {
			this.state = 1
			return true
		}
		if this.TryAddLine(line, LINE_TS_VAR_ARRAY_START) {
			this.state = 2
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
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 3
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		this.GetLogger().Panic("parse TsVarObject error")
	} else if this.state == 2 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_ARRAY_END) {
			this.state = 3
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		this.GetLogger().Panic("parse TsVarObject error")
	} else if this.state == 3 {
		return false
	}
	this.GetLogger().Panic("parse TsVarObject error")
	return false
}

type TsFuncObject struct {
	DefaultGeneratorObj
}

func NewTsFuncObject(file GeneratorFile) *TsFuncObject {
	ret := &TsFuncObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_TS_FUNC, file)
	return ret
}

func (this *TsFuncObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_TS_FUNC_HEADER) {
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
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		this.GetLogger().Panic("parse TsFuncObject error")
	} else if this.state == 2 {
		return false
	}
	this.GetLogger().Panic("parse TsFuncObject error")
	return false
}

type TsClassMember struct {
	IsGetter bool
	IsSetter bool
	IsPublic bool
	Name     string
	Type     string
	Line     *LineText
}

type TsClassObject struct {
	DefaultGeneratorObj
	Package string
	ClassName      string
	IsComponent    bool
	IsRenderComponent    bool
	IsSerializable bool
	members        []*TsClassMember
	LongStringTag  map[string]bool

}

func NewTsClassObject(file GeneratorFile) *TsClassObject {
	ret := &TsClassObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
		members: []*TsClassMember{},
		LongStringTag: map[string]bool{},
	}
	ret.DefaultGeneratorObj.init(OBJ_TS_CLASS, file)
	return ret
}

func (this *TsClassObject) GetClassMember(s string) *TsClassMember {
	for _,member:=range this.members {
		if member.Name == s {
			return member
		}
	}
	return nil
}

func (this *TsClassObject) Name() string {
	return this.ClassName
}

func (this *TsClassObject) GetClassName() string {
	return this.ClassName
}

func (this *TsClassObject) IsModel() bool {
	return this.IsSerializable || this.IsComponent || this.IsRenderComponent
}

func (this *TsClassObject) CheckLongString(mName string)bool{
	_,ok:=this.LongStringTag[mName]
	return ok
}

func (this *TsClassObject) CheckLine(line *LineText) bool {
	if this.state == TS_CLASSOBJ_HEADER {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_DEFINE_SINGLELINE) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_DECORATOR_COCOS) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_DEFINE_START) {
			this.state = TS_CLASSOBJ_DEFINER
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_HEADER) {
			this.state = TS_CLASSOBJ_MEMBER_FIELD
			this.ClassName = line.GetStructName()
			extendStr := LINE_TS_CLASS_HEADER.RegExp().ReplaceAllString(line.Text, "$6")
			if extendStr == "BaseComponent" {
				this.IsComponent = true
			}
			if extendStr == "RenderComponent" {
				this.IsRenderComponent = true
			}
			if extendStr == "ISerializable" {
				this.IsSerializable = true
			}
			return true
		}
		return false
	} else if this.state == TS_CLASSOBJ_DEFINER {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_DEFINE_OBJ) {
			if line.IsLongStringTag() {
				this.LongStringTag[line.GetName()] = true
			}
			return true
		}
		if this.TryAddLine(line, LINE_TS_DEFINE_END) {
			this.state = TS_CLASSOBJ_HEADER
			return true
		}
		this.GetLogger().Panicf("parse TsClassObject error")
	} else if this.state == TS_CLASSOBJ_MEMBER_FIELD {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FUNC_HEADER) {
			this.state = TS_CLASSOBJ_FUNCTION
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FIELD_PRIVATE) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FIELD) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FIELD_PROTECTED) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FIELD_PUBLIC) {
			this.members = append(this.members, &TsClassMember{
				IsGetter: false,
				IsSetter: false,
				IsPublic: true,
				Name:     line.GetName(),
				Type:     line.GetTypeName(),
				Line:     line,
			})
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_GETTER_HEADER) {
			this.members = append(this.members, &TsClassMember{
				IsGetter: true,
				IsSetter: false,
				IsPublic: false,
				Name:     line.GetName(),
				Type:     "",
			})
			this.state = TS_CLASSOBJ_GETTER_SETTER
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_SETTER_HEADER) {
			this.members = append(this.members, &TsClassMember{
				IsGetter: false,
				IsSetter: true,
				IsPublic: false,
				Name:     line.GetName(),
				Type:     "",
			})
			this.state = TS_CLASSOBJ_GETTER_SETTER
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_CONSTRUCTOR_HEADER) {
			this.state = TS_CLASSOBJ_CONSTRUCTOR
			return true
		}
		if this.TryAddLine(line,LINE_CLOSURE_END) {
			this.state = TS_CLASSOBJ_CLASS_END
			return true
		}
	} else if this.state == TS_CLASSOBJ_GETTER_SETTER {
		if this.TryAddLine(line, LINE_TS_CLASS_FUNC_END) {
			this.state = TS_CLASSOBJ_MEMBER_FIELD
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		return false
	} else if this.state == TS_CLASSOBJ_CONSTRUCTOR {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FUNC_END) {
			this.state = TS_CLASSOBJ_MEMBER_FIELD
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		this.GetLogger().Panic("parse TsClassObject TS_CLASSOBJ_CONSTRUCTOR error")
	} else if this.state == TS_CLASSOBJ_FUNCTION {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_CLASS_FUNC_END) {
			this.state = TS_CLASSOBJ_MEMBER_FIELD
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		return false
	} else if this.state == TS_CLASSOBJ_CLASS_END {
		return true
	}
	this.GetLogger().Panic("parse TsClassObject error")
	return false
}

type TsImportObject struct {
	DefaultGeneratorObj
}

func NewTsImportObject(file GeneratorFile) *TsImportObject {
	ret := &TsImportObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_TS_IMPORTS, file)
	return ret
}

func (this *TsImportObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_IMPORT_SINGLELINE) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_IMPORT_CLOSURE_START) {
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
		if this.TryAddLine(line, LINE_TS_IMPORT_OBJ) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_IMPORT_CLOSURE_END) {
			this.state = 0
			return true
		}
		this.GetLogger().Panic("parse TsImportObject error")
	}
	this.GetLogger().Panic("parse TsImportObject error")
	return false
}

type TsModelFile struct {
	*DefaultGeneratorFile
	Package string
	ClassName string
}

var _ GeneratorFile = (*TsModelFile)(nil)

func NewTsModelFile(generator *Generator) *TsModelFile {
	ret := &TsModelFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_TS_MODEL
	return ret
}

func (this *TsModelFile) Generate() *promise.Promise {
	return nil
}

func (this *TsModelFile) Parse() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_TS_IMPORTS)
		this.GetLogger().Warnf("parseTsImports", offset, success)
		offset, success = this.parse(offset, OBJ_COMMENT, OBJ_TS_CLASS, OBJ_TS_FUNC, OBJ_TS_ENUM, OBJ_TS_VAR)
		this.GetLogger().Warnf("parse TsModelFile Finish")
		resolve(nil)
	})
}

func (this *TsModelFile) ProcessClasses() []*TsClassObject {

	ret := make([]*TsClassObject, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_TS_CLASS {
			objStruct := obj.(*TsClassObject)
			if objStruct.IsModel() {
				ret = append(ret, objStruct)
			}
		}
	}
	return ret
}
