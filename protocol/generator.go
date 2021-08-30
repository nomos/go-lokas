package protocol

import (
	"github.com/nomos/go-lokas/log"
)

type GenType int

const (
	GEN_UNDEFINED GenType = iota
	GEN_GO
	GEN_TS
	GEN_JSON
	GEN_CS
)

type GeneratorOption struct {
	SourcePath string
	DestPath   string
}

type Generator struct {
	GenType           GenType
	Models            map[string]GeneratorFile
	GoModels          map[string]GeneratorFile
	Protos            map[string]GeneratorFile
	GoIds             *GoIdsFile
	ProtoIds          *ProtoIdsFile
	Conf              *ConfFile
	GoStructObjects   []*GoStructObject
	GoEnumObjects     []*GoEnumObject
	ProtoMsgObjects   []*ProtoMsgObject
	ModelClassObjects []*ModelClassObject
	ModelEnumObjects []*ModelEnumObject
	ModelIdsObjects map[uint16]*ModelId
	ModelErrorObjects map[int]*ModelError
	ModelPackages map[string]*ModelPackageObject

	TsModels       []*TsModelFile
	TsIds          *TsIdsFile
	TsEnums        *TsEnumFile
	TsClassObjects []*TsClassObject
	TsEnumObjects  []*TsEnumObject

	//Schemas []*ModelClassObject

	Individual   bool
	TsDependPath string
	GoPath       string
	TsPath       string
	ProtoPath    string
	ModelPath    string
	CsPath       string

	Proto2GoCmdLinExec func(pack, protoPath, GoPath string) error
	Proto2TsCmdLinExec func(pack, protoPath, GoPath string) error
}

func NewGenerator(typ GenType) *Generator {
	ret := &Generator{
		GenType: typ,
	}
	ret.Clear()
	return ret
}

func (this *Generator) SetProto2GoCmdLine(f func(pack, protoPath, GoPath string) error) {
	this.Proto2GoCmdLinExec = f
}

func (this *Generator) SetProto2TsCmdLine(f func(pack, protoPath, TsPath string) error) {
	this.Proto2TsCmdLinExec = f
}

func (this *Generator) GetErrorName(id int)string{
	e,ok:=this.ModelErrorObjects[id]
	if ok {
		return e.ErrorName
	}
	return ""
}

func (this *Generator) IsErrorName(s string)bool{
	for _,v:=range this.ModelErrorObjects {
		if v.ErrorName == s {
			return true
		}
	}
	return false
}

func (this *Generator) Clear() {
	log.Warnf("Generator Clear")
	this.Models = make(map[string]GeneratorFile)
	this.GoModels = make(map[string]GeneratorFile)
	this.Protos = make(map[string]GeneratorFile)
	this.GoStructObjects = make([]*GoStructObject, 0)
	this.GoEnumObjects = make([]*GoEnumObject, 0)
	this.TsModels = make([]*TsModelFile, 0)
	this.TsClassObjects = make([]*TsClassObject, 0)
	this.TsEnumObjects = make([]*TsEnumObject, 0)
	//this.Schemas = make([]*ModelClassObject, 0)
	this.ProtoMsgObjects = make([]*ProtoMsgObject, 0)
	this.ModelClassObjects = make([]*ModelClassObject, 0)
	this.ModelEnumObjects = make([]*ModelEnumObject, 0)
	this.ModelIdsObjects = make(map[uint16]*ModelId)
	this.ModelErrorObjects = make(map[int]*ModelError)
	this.ModelPackages = make(map[string]*ModelPackageObject)
}

func (this *Generator) GetModelByName(s string)*ModelClassObject{
	for _,v:=range this.ModelClassObjects {
		if v.ClassName == s {
			return v
		}
	}
	return nil
}

func (this *Generator) GetEnumByName(s string)*ModelEnumObject{
	for _,v:=range this.ModelEnumObjects {
		if v.EnumName == s {
			return v
		}
	}
	return nil
}

func (this *Generator) SetOption(opt GeneratorOption) {

}

func (this *Generator) IsEnum(s string)bool{
	for _,v:=range this.ModelEnumObjects {
		if v.EnumName == s {
			return true
		}
	}
	return false
}


