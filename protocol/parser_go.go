package protocol

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/promise"
	"strings"
)

type GoPackageObject struct {
	DefaultGeneratorObj
	packageName string
}

func NewGoPackageObject(file GeneratorFile) *GoPackageObject {
	ret := &GoPackageObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_PACKAGE, file)
	return ret
}

func (this *GoPackageObject) CheckLine(line *LineText) bool {
	if this.TryAddLine(line, LINE_GO_PACKAGE) {
		this.packageName = line.GetPkgName()
		return true
	}
	return false
}

type GoImportObject struct {
	DefaultGeneratorObj
}

func NewGoImportObject(file GeneratorFile) *GoImportObject {
	ret := &GoImportObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_GO_IMPORTS, file)
	return ret
}

func (this *GoImportObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_IMPORT_SINGLELINE) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_IMPORT_HEADER) {
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
		if this.TryAddLine(line, LINE_GO_IMPORT_BODY) {
			return true
		}
		if this.TryAddLine(line, LINE_BRACKET_END) {
			this.state = 0
			return true
		}
		log.Panic("parse GoImportObject error")
	}
	log.Panic("parse GoPackageObject error")
	return false
}



type GoInterfaceObject struct {
	DefaultGeneratorObj
	state         int
}

func NewGoInterfaceObject(file GeneratorFile) *GoInterfaceObject {
	ret := &GoInterfaceObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_INTERFACE, file)
	return ret
}


func (this *GoInterfaceObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		//HEADER
		if this.TryAddLine(line, LINE_GO_INTERFACE_HEADER) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		//BODY
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
	} else if this.state == 2 {
		return false
	}
	log.Panicf("parse GoInterfaceObject Error", this.state)
	return false
}

type GoStructFields struct {
	Name    string
	Type    string
	Package string
	Index   int
}

type GoStructObject struct {
	DefaultGeneratorObj
	TagId         BINARY_TAG
	state         int
	Fields        []*GoStructFields
	Package       string
	StructName    string
	ISerializable bool
	IComponent    bool
}

func NewGoStructObject(file GeneratorFile) *GoStructObject {
	ret := &GoStructObject{DefaultGeneratorObj: DefaultGeneratorObj{}, Fields: make([]*GoStructFields, 0)}
	ret.DefaultGeneratorObj.init(OBJ_GO_STRUCT, file)
	return ret
}

func (this *GoStructObject) IsModel() bool {
	return (this.ISerializable || this.IComponent) && this.StructName != ""
}

func (this *GoStructObject) SetPackage(pack string) {
	this.Package = pack
	for _, field := range this.Fields {
		field.Package = pack
	}
}

func (this *GoStructObject) CheckLine(line *LineText) bool {
	removeComment := COMMENT_REGEXP.ReplaceAllString(line.Text, "$1")
	if this.state == 0 {
		//HEADER
		if this.TryAddLine(line, LINE_GO_PUBLIC_STRUCT_HEADER) {
			this.state = 1
			this.StructName = line.GetStructName()
			return true
		}
		if this.TryAddLine(line, LINE_GO_PRIVATE_STRUCT_HEADER) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		//BODY
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_STRUCT_FIELD_PUBLIC) {
			if this.StructName != "" {
				fieldName := line.GetName()
				fieldType := line.GetTypeName()
				this.Fields = append(this.Fields, &GoStructFields{
					Name:    fieldName,
					Type:    fieldType,
					Package: this.Package,
					Index:   len(this.Fields),
				})
			}
			return true
		}
		if this.TryAddLine(line, LINE_GO_STRUCT_FIELD_PRIVATE) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_STRUCT_FIELD_INHERIT) {
			this.IComponent = ICOMPONENT_REGEXP.MatchString(removeComment)
			this.ISerializable = ISERIALIZABLE_REGEXP.MatchString(removeComment)
			return true
		}
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
	} else if this.state == 2 {
		return false
	}
	log.Panicf("parse GoStructObject Error", this.state)
	return false
}

type GoFuncObject struct {
	DefaultGeneratorObj
}

func NewGoFuncObject(file GeneratorFile) *GoFuncObject {
	ret := &GoFuncObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_FUNC, file)
	return ret
}

func (this *GoFuncObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_GO_FUNC_HEADER) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		log.Panic("parse GoFuncObject Body error")
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse GoFuncObject error")
	return false
}

type GoStructFuncObject struct {
	DefaultGeneratorObj
}

func NewGoStructFuncObject(file GeneratorFile) *GoStructFuncObject {
	ret := &GoStructFuncObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_STRUCT_FUNC, file)
	return ret
}

func (this *GoStructFuncObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_GO_STRUCT_FUNC_HEADER) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_ANY) {
			return true
		}
		log.Panic("parse GoStructFuncObject Body error")
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse GoStructFuncObject error")
	return false
}

type GoDefineObject struct {
	DefaultGeneratorObj
}

func NewGoDefineObject(file GeneratorFile) *GoDefineObject {
	ret := &GoDefineObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_DEFINE, file)
	return ret
}

func (this *GoDefineObject) CheckLine(line *LineText) bool {
	if this.TryAddLine(line, LINE_GO_DEFINER) {
		return true
	}
	return false
}

type GoVarObject struct {
	DefaultGeneratorObj
}

func NewGoVarObject(file GeneratorFile) *GoVarObject {
	ret := &GoVarObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_VAR, file)
	return ret
}

func (this *GoVarObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_GO_VARIABLE) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_GO_CONST) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_GO_VAR_CLOSURE_START) {
			this.state = 1
			return true
		}
		if this.TryAddLine(line, LINE_GO_CONST_CLOSURE_START) {
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
		if this.TryAddLine(line, LINE_BRACKET_END) {
			this.state = 2
			return true
		}
		if this.TryAddLine(line, LINE_GO_ENUM_VARIABLE_IOTA) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_ENUM_AUTO) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_ENUM_VARIABLE) {
			return true
		}
		log.Panic("parse GoEnumObject Body error")
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse GoVarObject error")
	return false
}

type GoEnumObject struct {
	DefaultGeneratorObj
	state     int
	typ       string
	lastValue int
	lastType  string
}

func NewGoEnumObject(file GeneratorFile) *GoEnumObject {
	ret := &GoEnumObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_ENUM, file)
	return ret
}

func (this *GoEnumObject) Type() string {
	return this.typ
}

func (this *GoEnumObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_GO_ENUM_DEFINER) {
			this.typ = line.GetTypeName()
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
		if this.TryAddLine(line, LINE_GO_CONST_CLOSURE_START) {
			this.state = 2
			return true
		}
	} else if this.state == 2 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_BRACKET_END) {
			this.state = 3
			return true
		}
		if this.TryAddLine(line, LINE_GO_ENUM_VARIABLE_IOTA) {
			this.lastType = line.Type
			this.lastValue = line.Value
			return true
		}
		if this.TryAddLine(line, LINE_GO_ENUM_VARIABLE) {
			return true
		}
		if this.TryAddLine(line, LINE_GO_ENUM_AUTO) {
			line.GetName()
			this.lastValue++
			line.Value = this.lastValue
			line.Type = this.lastType
			return true
		}
		log.Panic("parse GoEnumObject Body error")
	} else if this.state == 3 {
		return false
	}
	log.Panic("parse GoEnumObject error")
	return false
}

type GoModelFile struct {
	*DefaultGeneratorFile
}

var _ GeneratorFile = (*GoModelFile)(nil)

func NewGoModelFile(generator *Generator) *GoModelFile {
	ret := &GoModelFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_GO_MODEL
	return ret
}

func (this *GoModelFile) Generate() *promise.Promise {
	return nil
}

func (this *GoModelFile) Parse() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_GO_PACKAGE)
		log.Infof("parseGoModelPackage", offset, success)
		offset, success = this.parse(offset, OBJ_GO_IMPORTS)
		log.Infof("parseGoImports", offset, success)
		offset, success = this.parse(offset,
			OBJ_GO_STRUCT,
			OBJ_GO_INTERFACE,
			OBJ_GO_FUNC,
			OBJ_GO_STRUCT_FUNC,
			OBJ_GO_ENUM,
			OBJ_GO_VAR,
			OBJ_GO_DEFINE,
		)
		log.Infof("parseGoMain finish", offset, success)
		resolve(nil)
	})
}

func (this *GoModelFile) ProcessStruct() []*GoStructObject {
	ret := make([]*GoStructObject, 0)
	packName := ""
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_GO_PACKAGE {
			packName = obj.(*GoPackageObject).packageName
			continue
		}
		if obj.ObjectType() == OBJ_GO_STRUCT {
			objStruct := obj.(*GoStructObject)
			if objStruct.IsModel() {
				objStruct.SetPackage(packName)
				ret = append(ret, objStruct)
			}
		}
	}
	return ret
}

func (this *GoModelFile) ProcessEnum() []*GoEnumObject {
	ret := make([]*GoEnumObject, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_GO_ENUM {
			objEnum := obj.(*GoEnumObject)
			for _, line := range objEnum.Lines() {
				switch line.LineType {
				case LINE_GO_ENUM_VARIABLE_IOTA, LINE_GO_ENUM_VARIABLE, LINE_GO_ENUM_AUTO:
					line.Name = strings.Replace(line.Name, line.Type+"_", "", -1)
				}
			}
			ret = append(ret, objEnum)
		}
	}
	return ret
}
