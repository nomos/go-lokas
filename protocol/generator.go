package protocol

import (
	"errors"
	"github.com/iancoleman/orderedmap"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/stringutil"
	"github.com/nomos/promise"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	GameDataFolder string
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

	Schemas []*ModelSchema

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
	this.Schemas = make([]*ModelSchema, 0)
	this.ProtoMsgObjects = make([]*ProtoMsgObject, 0)
	this.ModelClassObjects = make([]*ModelClassObject, 0)
	this.ModelEnumObjects = make([]*ModelEnumObject, 0)
	this.ModelIdsObjects = make(map[uint16]*ModelId)
	this.ModelErrorObjects = make(map[int]*ModelError)
	this.ModelPackages = make(map[string]*ModelPackageObject)
	this.InitDefaultSchemas()
}

func (this *Generator) InitDefaultSchemas() {

}

func (this *Generator) SetOption(opt GeneratorOption) {

}

func (this *Generator) LoadCsFolder(p string) *promise.Promise {
	this.CsPath = p
	err:=os.MkdirAll(this.CsPath, os.ModePerm)
	if err != nil {
		log.Error(err.Error())
		return promise.Reject(nil)
	}
	return promise.Resolve(nil)
}

func (this *Generator) LoadModelFolder(p string) *promise.Promise {
	this.ModelPath = p
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		err := this.LoadModels(p)
		if err != nil {
			reject(err)
			return
		}
		resolve(nil)
	})
}

func (this *Generator) LoadModels(p string) error {
	log.Warnf("LoadModels")
	var err error
	_,err=util.WalkDirFilesWithFunc(p, func(filePath string, file os.FileInfo) bool {
		if path.Ext(filePath) == ".model" {
			log.Warnf("filePath", filePath)
			file := NewModelFile(this)
			_, err = file.Load(filePath).Await()
			if err != nil {
				log.Error(err.Error())
				return true
			}
			_, err = file.Parse().Await()
			if err != nil {
				log.Error(err.Error())
				return true
			}
			this.Models[filePath] = file
		}
		return false
	}, true)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Warnf("load "+strconv.Itoa(len(this.Models))+" models")
	return nil
}

func (this *Generator) processModelObjects() error{
	ret := make([]*ModelClassObject, 0)
	for _, file := range this.Models {
		objects := file.(*ModelFile).ProcessModels()
		for _, toAddObj := range objects {
			foundSame := false
			for _, addedObj := range ret {
				if addedObj.ClassName == toAddObj.ClassName {
					foundSame = true
					panic("duplicated class")
					break
				}
			}
			if foundSame {
				continue
			}
			ret = append(ret, toAddObj)
		}
	}
	this.ModelClassObjects = ret
	defer func() {
		r:=recover()
		if err,ok:=r.(error);ok {
			log.Error(err.Error())
		}
	}()
	ret2 := make(map[uint16]*ModelId)
	for _, file := range this.Models {
		objects := file.(*ModelFile).ProcessIds()
		for _, id := range objects {
			for _, addedObj := range ret2 {
				if id.Name == addedObj.Name {
					return errors.New("duplicated class")
				}
			}
			ret2[uint16(id.Id)] = id
		}
	}
	this.ModelIdsObjects = ret2
	ret3 := make([]*ModelEnumObject,0)
	for _, file := range this.Models {
		objects := file.(*ModelFile).ProcessEnums()
		for _, obj := range objects {
			for _, addedObj := range ret3 {
				if obj.EnumName == addedObj.EnumName {
					return errors.New("duplicated enum")
				}
			}
			for _, addedObj := range ret {
				if obj.EnumName == addedObj.ClassName {
					return errors.New("duplicated enum")
				}
			}
			ret3 = append(ret3, obj)
		}
	}
	this.ModelEnumObjects = ret3
	for idx,ids:=range this.ModelIdsObjects {
		for _,v:=range this.ModelClassObjects {
			log.Warnf(ids.Name,v.ClassName)
			if ids.Name==v.ClassName {
				v.TagId = BINARY_TAG(idx)
				ids.ClassObj = v
			}
			if ids.Resp == v.ClassName {
				ids.RespClassObj = v
			}
		}
	}
	log.Warnf("ModelIdsObjects", this.ModelIdsObjects)
	log.Warnf("ModelClassObjects", this.ModelClassObjects)
	log.Warnf("ModelEnumObjects", this.ModelEnumObjects)
	return nil
}

func (this *Generator) LoadGoFolder(p string) *promise.Promise {
	this.GoPath = p
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		err := this.LoadConf(p)
		if err != nil {
			reject(err)
			return
		}
		err = this.LoadGoIds(p)
		if err != nil {
			reject(err)
			return
		}
		err = this.LoadGoModels(p)
		if err != nil {
			reject(err)
			return
		}
		resolve(nil)
	})
}

func (this *Generator) LoadGoModels(p string) error {
	var err error
	util.WalkDirFilesWithFunc(p, func(filePath string, file os.FileInfo) bool {
		if file.Name() == "Ids.go" {
			return false
		}
		if path.Ext(filePath) == ".go" {
			log.Warnf("filePath", filePath)
			file := NewGoModelFile(this)
			_, err = file.Load(filePath).Await()
			if err != nil {
				log.Error(err.Error())
				return true
			}
			_, err = file.Parse().Await()
			if err != nil {
				log.Error(err.Error())
				return true
			}
			this.GoModels[filePath] = file
		}
		return false
	}, true)
	if err != nil {
		return err
	}
	return nil
}

func (this *Generator) LoadConf(p string) error {
	confPath := util.FindFile(p, "tag.ini", false)
	if confPath == "" {
		log.Error("tag.ini file is not exist")
		return errors.New("tag.ini file is not exist")
	}
	log.Warnf("confPath", confPath)
	this.Conf = NewConfFile(this)
	_, err := this.Conf.Load(confPath).Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = this.Conf.Parse().Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) LoadGoIds(p string) error {
	idsPath := util.FindFile(p, "Ids.go", false)
	if idsPath == "" {
		idsPath = path.Join(p, "Ids.go")
		util.CreateFile(idsPath)
	}
	this.GoIds = NewGoIdsFile(this)
	log.Warnf("go idsPath", idsPath)
	_, err := this.GoIds.Load(idsPath).Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = this.GoIds.Parse().Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) LoadProtoFolder(p string) *promise.Promise {
	this.ProtoPath = p
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		err := this.LoadConf(p)
		if err != nil {
			reject(err)
			return
		}
		err = this.LoadProtoIds(p)
		if err != nil {
			reject(err)
			return
		}
		err = this.LoadProtos(p)
		if err != nil {
			log.Panic(err.Error())
			reject(err)
			return
		}
		resolve(nil)
	})
}

func (this *Generator) LoadProtos(p string) error {
	var err error
	util.WalkDirFilesWithFunc(p, func(filePath string, file os.FileInfo) bool {
		if file.Name() == "proto.id" {
			return true
		}
		if path.Ext(filePath) == ".proto" {
			log.Warnf("filePath", filePath)
			file := NewProtoFile(this)
			_, err = file.Load(filePath).Await()
			if err != nil {
				log.Error(err.Error())
				return true
			}
			_, err = file.Parse().Await()
			if err != nil {
				log.Error(err.Error())
				return true
			}
			this.Protos[filePath] = file
		}
		return false
	}, true)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) processProtoObjects() {
	ret := make([]*ProtoMsgObject, 0)
	for _, file := range this.Protos {
		objects := file.(*ProtoFile).ProcessProtos()
		for _, toAddObj := range objects {
			foundSame := false
			for _, addedObj := range ret {
				if addedObj.ProtoName == toAddObj.ProtoName {
					foundSame = true
					break
				}
			}
			if foundSame {
				continue
			}
			ret = append(ret, toAddObj)
		}
	}
	this.ProtoMsgObjects = ret
	log.Warnf("ProtoMsgObjects", this.ProtoMsgObjects)
}

func (this *Generator) LoadProtoIds(p string) error {
	pidsPath := util.FindFile(p, "proto.id", false)
	if pidsPath == "" {
		pidsPath = path.Join(p, "proto.id")
		util.CreateFile(pidsPath)
	}
	log.Warnf("pidsPsth", pidsPath)
	this.ProtoIds = NewProtoIdsFile(this)
	log.Warnf("proto idsPath", pidsPath)
	_, err := this.ProtoIds.Load(pidsPath).Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = this.ProtoIds.Parse().Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) getSortedProtoIds() *orderedmap.OrderedMap {
	existIds := make(map[string]int)
	exportIds := orderedmap.New()
	maxId := this.Conf.Offset + 1
	for id, name := range this.ProtoIds.Ids {
		if id > this.Conf.Offset {
			existIds[name] = id
			if id >= maxId {
				maxId = id + 1
			}
		}
	}
	for _, proto := range this.ProtoMsgObjects {
		if _, ok := existIds[proto.ProtoName]; ok {
			continue
		}
		proto.ProtoId = maxId
		maxId++
		exportIds.Set(proto.ProtoName, proto.ProtoId)
	}
	exportIds.Sort(func(a *orderedmap.Pair, b *orderedmap.Pair) bool {
		return a.Value().(int) < b.Value().(int)
	})
	return exportIds
}

func (this *Generator) GenerateProto2Ids() {
	log.Warnf("GenerateProto2Ids")
	this.processProtoObjects()

	strs := auto_gen_header
	strs += "package " + this.Conf.PackageName + "\n\n"
	exportIds := this.getSortedProtoIds()
	for _, key := range exportIds.Keys() {
		id, _ := exportIds.Get(key)
		strs += strconv.Itoa(id.(int)) + " " + key + "\n"
	}
	err := ioutil.WriteFile(this.ProtoIds.FilePath, []byte(strs), 0)
	if err != nil {
		log.Errorf(err.Error())
	}
}

func (this *Generator) GenerateProto2TsJsImport() {
	jsImportPath := util.FindFile(this.TsPath, this.Conf.PackageName+".js", false)
	if jsImportPath == "" {
		log.Errorf("proto js file not found:", jsImportPath)
		return
	}
	istrs, _ := ioutil.ReadFile(jsImportPath)
	estrs := string(istrs)
	estrs = strings.Replace(estrs, `var $protobuf = require("protobufjs/light");`, `var $protobuf = require("protobufjs/light");
const `+this.Conf.PackageName+` = {}
if (!CC_EDITOR) {
`, 1)
	estrs = strings.Replace(estrs, "module.exports = $root;", `
  for (let i in $root.nested.`+this.Conf.PackageName+`.nested) {
    let ctor = $root.lookup(i).ctor
    if (!ctor) {
      ctor = $root.lookup(i).values
    }
    `+this.Conf.PackageName+`[i] = ctor
  }
}
module.exports = {
	`+this.Conf.PackageName+` : `+this.Conf.PackageName+`
}
`, 1)
	err := ioutil.WriteFile(jsImportPath, []byte(estrs), 0)
	if err != nil {
		log.Errorf(err.Error())
	}
}

func (this *Generator) GenerateProto2TsIds() {
	strs := auto_gen_header
	importObjs := this.TsIds.GetObj(OBJ_TS_IMPORTS)
	for _, obj := range importObjs {
		obj.RemoveLineType(LINE_EMPTY)
		strs += obj.String()
	}
	relative, _ := filepath.Rel(path.Dir(this.TsIds.FilePath), this.TsDependPath)
	if len(importObjs) == 0 && relative != "" {

		strs += "\n"
		strs += "import {TypeRegistry} from \"" + relative + "/protocol/types\"\n"
	}
	strs += "\n"
	strs += "(function () {\n"
	exportIds := this.getSortedProtoIds()
	for _, key := range exportIds.Keys() {
		id, _ := exportIds.Get(key)
		strs += "\tTypeRegistry.getInstance().RegisterProtoTag(" + this.Conf.PackageName + "." + key + "," + strconv.Itoa(id.(int)) + ")\n"
	}
	strs += "})()\n"
	err := ioutil.WriteFile(this.TsIds.FilePath, []byte(strs), 0)
	if err != nil {
		log.Errorf(err.Error())
	}
	this.LoadGo2TsIds(this.TsPath)
}

func (this *Generator) GenerateProto2Ts() error {
	this.GenerateProto2Ids()
	this.GenerateProto2TsIds()
	if this.Proto2TsCmdLinExec != nil {
		err := this.Proto2TsCmdLinExec(this.Conf.PackageName, path.Join(this.ProtoPath, "proto"), this.TsPath)
		if err != nil {
			return err
		}
	}

	this.GenerateProto2TsJsImport()
	time.Sleep(1)
	return nil
}

func (this *Generator) GenerateProto2GoIds() {
	this.GenerateGoIds(true)
}

func (this *Generator) GenerateProto2Go() error {
	this.GenerateProto2Ids()
	if this.Proto2GoCmdLinExec != nil {
		err := this.Proto2GoCmdLinExec(this.Conf.PackageName, path.Join(this.ProtoPath, "proto"), this.GoPath)
		if err != nil {
			return err
		}
	}
	this.GenerateProto2GoIds()
	return nil
}

func (this *Generator) GenerateModel2Go()error{
	err:=this.processModelObjects()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.generateModel2GoClasses()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.generateModel2GoEnums()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.generateModel2GoIds()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Warnf("Finish")
	return nil
}

func (this *Generator) generateModel2GoClasses()error{
	log.Warnf("GoPath",this.GoPath)
	for _,m:=range this.ModelClassObjects {
		err:=this.generateModel2GoClass(m)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoClass(m *ModelClassObject)error{
	p:=path.Join(this.GoPath,"model_"+stringutil.SplitCamelCaseLowerSlash(m.ClassName))
	p+=".go"
	err:=ioutil.WriteFile(p, []byte(m.GoString(this)), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	p=path.Join(this.GoPath,"model_"+stringutil.SplitCamelCaseLowerSlash(m.ClassName))
	p+="_impl.go"
	if !util.IsFileExist(p) {
		err=ioutil.WriteFile(p, []byte(m.GoImplString(this)), 0644)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoEnums()error{
	log.Warnf("GoPath",this.GoPath)
	for _,m:=range this.ModelEnumObjects {
		err:=this.generateModel2GoEnum(m)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoEnum(m *ModelEnumObject)error{
	p:=path.Join(this.GoPath,"enum_"+stringutil.SplitCamelCaseLowerSlash(m.EnumName))
	p+=".go"
	err:=ioutil.WriteFile(p, []byte(m.GoString(this)), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) generateModel2GoIds()error{
	log.Warnf("generateModel2GoIds")
	for _,m:=range this.Models {
		f:=m.(*ModelFile)
		pack:=f.ProcessPackages()
		p:=this.ModelPackages[pack.PackageName]
		if p!=nil {
			if p.GoPackageName!=pack.GoPackageName {
				return errors.New("wrong cs package")
			}
		}
		this.ModelPackages[pack.PackageName]= pack
	}
	for _,m:=range this.Models {
		f:=m.(*ModelFile)
		ids:=f.ProcessIds()
		errs:=f.ProcessErrors()
		for _,e:=range errs {
			this.ModelErrorObjects[e.ErrorId] = e
		}
		for _,v:=range ids {
			pack :=this.ModelPackages[v.PackageName]
			if pack==nil {
				return errors.New("package not found:"+v.PackageName)
			}
			pack.Ids[BINARY_TAG(v.Id)] = v
		}
		for _,v:=range errs {
			pack :=this.ModelPackages[v.PackageName]
			if pack==nil {
				return errors.New("package not found:"+v.PackageName)
			}
			pack.Errors[v.ErrorId] = v
		}
	}

	for _,p:=range this.ModelPackages {
		err:=this.generateModel2GoPackage(p)

		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoPackage(pack *ModelPackageObject)error{
	p:=path.Join(this.GoPath,"ids")
	p+=".go"
	log.Warnf("generateModel2GoPackage",p)
	err:=ioutil.WriteFile(p, []byte(pack.GoString(this)), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) GenerateModel2Cs()error{
	err:=this.processModelObjects()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.generateModel2CsClasses()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.generateModel2CsEnums()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.generateModel2CsIds()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) generateModel2CsClasses()error{
	log.Warnf("CsPath",this.CsPath)
	for _,m:=range this.ModelClassObjects {
		err:=this.generateModel2CsClass(m)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2CsClass(m *ModelClassObject)error{
	p:=path.Join(this.CsPath,m.ClassName)
	p+=".cs"
	err:=ioutil.WriteFile(p, []byte(m.CsString(this)), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) generateModel2CsEnums()error{
	log.Warnf("CsPath",this.CsPath)
	for _,m:=range this.ModelEnumObjects {
		err:=this.generateModel2CsEnum(m)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2CsEnum(m *ModelEnumObject)error{
	p:=path.Join(this.CsPath,m.EnumName)
	p+=".cs"
	err:=ioutil.WriteFile(p, []byte(m.CsString(this)), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) generateModel2CsIds()error{
	for _,m:=range this.Models {
		f:=m.(*ModelFile)
		pack:=f.ProcessPackages()
		p:=this.ModelPackages[pack.PackageName]
		if p!=nil {
			if p.CsPackageName!=pack.CsPackageName {
				return errors.New("wrong cs package")
			}
		}
		this.ModelPackages[pack.PackageName]= pack
	}

	for _,m:=range this.Models {
		f:=m.(*ModelFile)
		ids:=f.ProcessIds()
		errs:=f.ProcessErrors()
		for _,e:=range errs {
			this.ModelErrorObjects[e.ErrorId] = e
		}
		for _,v:=range ids {
			pack :=this.ModelPackages[v.PackageName]
			if pack==nil {
				return errors.New("package not found:"+v.PackageName)
			}
			pack.Ids[BINARY_TAG(v.Id)] = v
		}
		for _,v:=range errs {
			pack :=this.ModelPackages[v.PackageName]
			if pack==nil {
				return errors.New("package not found:"+v.PackageName)
			}
			pack.Errors[v.ErrorId] = v
		}
	}

	for _,p:=range this.ModelPackages {
		err:=this.generateModel2CsPackage(p)
		if err != nil {
			log.Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2CsPackage(pack *ModelPackageObject)error{
	className:=pack.GetCsPackageName()
	log.Warnf(className)
	p:=path.Join(this.CsPath,className)
	p+=".cs"
	err:=ioutil.WriteFile(p, []byte(pack.CsString(this)), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) IsEnum(s string)bool{
	for _,v:=range this.ModelEnumObjects {
		if v.EnumName == s {
			return true
		}
	}
	return false
}

func (this *Generator) GenerateModel2Ts(){
	this.processModelObjects()

}

func (this *Generator) LoadGo2TsFolder(p string, individual bool) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		err := this.LoadGo2TsIds(p)
		if err != nil {
			reject(err)
			return
		}
		err = this.LoadTsEnums(p)
		if err != nil {
			reject(err)
			return
		}
		err = this.LoadGo2TsModels(p)
		if err != nil {
			reject(err)
			return
		}
		resolve(nil)
	})
}

func (this *Generator) LoadTsEnums(p string) error {
	baseName := path.Base(p)
	enumsPath := util.FindFile(p, baseName+"_enums.ts", false)
	if enumsPath == "" {
		enumsPath = path.Join(p, baseName+"_enums.ts")
		util.CreateFile(enumsPath)
	}
	this.TsEnums = NewTsEnumFile(this)
	log.Warnf("ts enumsPath", enumsPath)
	_, err := this.TsEnums.Load(enumsPath).Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = this.TsEnums.Parse().Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) LoadGo2TsIds(p string) error {
	baseName := path.Base(p)
	this.TsPath = p
	idsPath := util.FindFile(p, baseName+"_ids.ts", false)
	if idsPath == "" {
		idsPath = path.Join(p, baseName+"_ids.ts")
		util.CreateFile(idsPath)
	}
	this.TsIds = NewTsIdsFile(this)
	log.Warnf("ts idsPath", idsPath)
	_, err := this.TsIds.Load(idsPath).Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = this.TsIds.Parse().Await()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) LoadGo2TsModels(p string) error {
	baseName := path.Base(p)
	if this.Individual {
		var err error
		util.WalkDirFilesWithFunc(p, func(filePath string, file os.FileInfo) bool {
			if path.Ext(filePath) != "ts" {
				return false
			}
			fileName := path.Base(filePath)
			switch fileName {
			case baseName + "_ids.ts", baseName + "_models.ts", baseName + "_enums.ts":
				return false
			default:
				_, err := this.LoadAndParseTsFile(filePath)
				if err != nil {
					log.Error(err.Error())
					return true
				}
				return false
			}
		}, true)
		return err
	} else {
		modelsPath := util.FindFile(p, baseName+"_models.ts", false)
		if modelsPath == "" {
			modelsPath = path.Join(p, baseName+"_models.ts")
			util.CreateFile(modelsPath)
		}
		_, err := this.LoadAndParseTsFile(modelsPath)
		return err
	}
}

func (this *Generator) LoadAndParseTsFile(modelsPath string) (*TsModelFile, error) {
	log.Warnf("ts modelsPath", modelsPath)
	file := NewTsModelFile(this)
	_, err := file.Load(modelsPath).Await()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	_, err = file.Parse().Await()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	for _, tsClass := range file.ProcessClasses() {
		if tsClass.IsModel() {
			this.TsClassObjects = append(this.TsClassObjects, tsClass)
		}
	}
	this.TsModels = append(this.TsModels, file)
	return file, nil
}

func (this *Generator) processGoObjects() {
	for _, file := range this.GoModels {
		objects := file.(*GoModelFile).ProcessStruct()
		for _, object := range objects {
			this.GoStructObjects = append(this.GoStructObjects, object)
		}
	}
	for _, file := range this.GoModels {
		objects := file.(*GoModelFile).ProcessEnum()
		for _, object := range objects {
			this.GoEnumObjects = append(this.GoEnumObjects, object)
		}
	}
}

func (this *Generator) GenerateGoIds(fromProto bool) {

	ids := make([]*GoAssignedId, 0)

	if !fromProto {
		this.processGoObjects()
		for _, id := range this.GoIds.AssignedIds {
			if id.Value < this.Conf.Offset {
				continue
			}
			for _, id2 := range this.GoIds.AssignedTypes {
				if id.Struct == id2.Struct && id.Tag == id2.Tag {
					ids = append(ids, id)
				}
			}
		}
		log.Warnf("this.Conf", this.Conf.PackageName, this.Conf.Offset)
		log.Warnf("this.GoStructObjects", this.GoStructObjects)
		log.Warnf("this.GoIds.AssignedIds", this.GoIds.AssignedIds)
		log.Warnf("this.GoIds.AssignedTypes", this.GoIds.AssignedTypes)
		log.Warnf("Ids", ids)
		for _, structObj := range this.GoStructObjects {
			find := false
			for _, id := range ids {
				if id.Struct == structObj.StructName {
					id.Package = structObj.Package
					find = true
					break
				}
			}
			if !find {
				ids = append(ids, &GoAssignedId{
					Tag:     this.Conf.TagName,
					Struct:  structObj.StructName,
					Package: structObj.Package,
					Value:   0,
					Line:    0,
				})
			}
		}
		log.Warnf("Ids", ids)
		for _, line := range this.GoIds.Lines {
			if line.LineType == LINE_GO_TAG_DEFINER {
				log.Warnf("LINE_GO_TAG_DEFINER", line.LineNum, line.Text)
			}
		}
		for i := 0; i < len(ids); i++ {
			for j := 0; j < len(ids); j++ {
				if i == j {
					continue
				}
				if ids[i].Value == ids[j].Value && ids[i].Value != 0 {
					ids[i].Value = 0
				}
			}
		}
		maxId := this.Conf.Offset
		for _, id := range ids {
			if id.Value > maxId {
				maxId = id.Value
			}
		}
		for index, id := range ids {
			if id.Value == 0 {
				maxId++
				ids[index].Value = maxId
			}
			this.getGoStructByName(id.Struct).TagId = BINARY_TAG(id.Value)
		}
		this.GoIds.RemoveAutoGenHeader()
		this.GoIds.RemoveObjType(OBJ_EMPTY)
	} else {
		sortedIds := this.getSortedProtoIds()
		for _, key := range sortedIds.Keys() {
			value, _ := sortedIds.Get(key)
			id := value.(int)
			ids = append(ids, &GoAssignedId{
				Tag:     this.Conf.TagName,
				Struct:  key,
				Package: this.Conf.PackageName,
				Value:   id,
				Line:    0,
			})
		}
	}

	importObjs := this.GoIds.GetObj(OBJ_GO_IMPORTS)
	var importObj GeneratorObject
	if len(importObjs) > 0 {
		importObj = importObjs[0]
	}

	strs := auto_gen_header
	strs += "package " + this.Conf.PackageName + "\n"
	strs += "\n"
	if importObj != nil {
		importObj.RemoveLineType(LINE_EMPTY)
		strs += importObj.String()
	}
	strs += "\n"
	strs += "const (\n"
	for _, id := range ids {
		strs += "\t" + id.Tag + "_" + id.Struct + " " + "protocol.BINARY_TAG = " + strconv.Itoa(id.Value) + "\n"
	}
	strs += ")\n\n"
	strs += "func init() {\n"
	for _, id := range ids {
		typeName := id.Package + "." + id.Struct
		if id.Package == this.Conf.PackageName {
			typeName = id.Struct
		}
		strs += "\t" + "protocol.GetTypeRegistry().RegistryType(" + id.Tag + "_" + id.Struct + ",reflect.TypeOf((*" + typeName + ")(nil)).Elem())\n"
	}
	strs += "}\n"
	err := ioutil.WriteFile(this.GoIds.FilePath, []byte(strs), 0)
	if err != nil {
		log.Errorf(err.Error())
	}
}

func (this *Generator) GenerateGo2Ts() {
	this.GenerateGoIds(false)
	err := this.LoadGoIds(this.GoPath)
	if err != nil {
		log.Error(err.Error())
		return
	}
	this.GenerateTsEnum()
	this.GenerateGo2TsIds()
	this.GenerateTsModels()
}

func (this *Generator) GenerateGo2TsIds() {
	strs := auto_gen_header
	importObjs := this.TsIds.GetObj(OBJ_TS_IMPORTS)
	for _, obj := range importObjs {
		obj.RemoveLineType(LINE_EMPTY)
		strs += obj.String()
	}
	relative, _ := filepath.Rel(path.Dir(this.TsIds.FilePath), this.TsDependPath)
	if len(importObjs) == 0 && relative != "" {

		strs += "\n"
		strs += "import {TypeRegistry} from \"" + relative + "/protocol/types\"\n"
	}
	strs += "\n"
	strs += "(function () {\n"
	for _, id := range this.GoIds.AssignedIds {
		strs += "\tTypeRegistry.getInstance().RegisterCustomTag(\"" + id.Struct + "\"," + strconv.Itoa(id.Value) + ")\n"
	}
	strs += "})()\n"
	err := ioutil.WriteFile(this.TsIds.FilePath, []byte(strs), 0)
	if err != nil {
		log.Errorf(err.Error())
	}
	this.LoadGo2TsIds(this.TsPath)
}

func (this *Generator) getRelativeTagId(name string) BINARY_TAG {
	schema := this.getSchemaByName(name)
	if schema != nil {
		return schema.Type
	}
	structObj := this.getGoStructByName(name)
	if structObj != nil {
		return structObj.TagId
	}
	return 0
}

func (this *Generator) getSchemaByName(name string) *ModelSchema {
	for _, iter := range this.Schemas {
		if iter.Name == name {
			return iter
		}
	}
	return nil
}

func (this *Generator) getGoStructByName(name string) *GoStructObject {
	for _, iter := range this.GoStructObjects {
		if iter.StructName == name {
			return iter
		}
	}
	return nil
}

func (this *Generator) GenerateTsModels() {
	log.Warnf("models", this.GoStructObjects)

	for _, id := range this.GoIds.AssignedIds {
		goModel := this.getGoStructByName(id.Struct)
		if goModel == nil {
			log.Panicf("go struct not found", id.Struct)
			continue
		}
		log.Warnf("go_model tag id ", goModel.TagId, goModel.StructName)
		this.ProcessGoModel2Schema(goModel)
	}
	this.PostProcessAllGoSchemas()
	if !this.Individual {
		if len(this.TsModels) == 0 {
			log.Panic("err read models.ts")
		}
		tsFile := this.TsModels[0]
		//importObjs:=tsFile.GetObj(OBJ_TS_IMPORTS)
		for i := len(this.Schemas) - 1; i >= 0; i-- {
			schema := this.Schemas[i]
			this.getTsModelFileBySchema(schema)
			tsClass := this.getTsClassByName(schema.Name)
			if tsClass == nil {
				this.genTsClass(tsFile, schema)
			} else {
				for _, body := range schema.Body {
					if tsClass.CheckLongString(body.Name) {
						body.IsLongString = true
					}
				}
				this.regenTsClass(schema, tsClass)
			}
		}
	} else {
		for i := len(this.Schemas) - 1; i >= 0; i-- {
			schema := this.Schemas[i]
			tsFile := this.getTsModelFileBySchema(schema)
			tsClass := this.getTsClassByName(schema.Name)
			if tsClass == nil {
				this.genTsClass(tsFile, schema)
			} else {
				for _, body := range schema.Body {
					if tsClass.CheckLongString(body.Name) {
						body.IsLongString = true
					}
				}
				this.regenTsClass(schema, tsClass)
			}
		}
	}
	for _, modelFile := range this.TsModels {
		log.Infof("modelFile.FileName", modelFile.FileName)
		strs := auto_gen_header
		modelFile.RemoveAutoGenHeader()
		modelFile.RemoveObjType(OBJ_EMPTY)
		imports := modelFile.GetObj(OBJ_TS_IMPORTS)
		for _, impo := range imports {
			impo.RemoveLineType(LINE_EMPTY)
		}
		relative, _ := filepath.Rel(path.Dir(modelFile.FilePath), this.TsDependPath)
		if len(imports) == 0 && relative != "" {
			imports := NewTsImportObject(modelFile)
			imports.AddLine(&LineText{
				Obj:     imports,
				LineNum: 0,
				Text:    "import {define,Tag} from \"" + relative + "/protocol/types\"",
			}, LINE_TS_IMPORT_SINGLELINE)
			imports.AddLine(&LineText{
				Obj:     imports,
				LineNum: 0,
				Text:    "import {ISerializable} from \"" + relative + "/protocol/protocol\"",
			}, LINE_TS_IMPORT_SINGLELINE)
			modelFile.InsertObject(0, imports)
		}
		for _, obj := range modelFile.Objects {
			strs += obj.String()
		}
		ioutil.WriteFile(modelFile.FilePath, []byte(strs), 0644)
	}
}

func (this *Generator) genTsClass(tsFile *TsModelFile, schema *ModelSchema) {
	var tsClass *TsClassObject
	tsClass = NewTsClassObject(tsFile)
	imports := tsFile.GetObj(OBJ_TS_IMPORTS)
	for _, impor := range imports {
		impor.RemoveLineType(LINE_EMPTY)
	}
	log.Warnf("tsFile", tsFile.Objects)
	log.Warnf("tsFile", len(imports))
	if len(imports) == 0 {
		tsFile.InsertObject(0, tsClass)
	} else {
		tsFile.InsertAfter(tsClass, imports[len(imports)-1])
	}
	this.genTsClassDefine(schema, tsClass)
	this.genTsClassOther(schema, tsClass)
}

func (this *Generator) regenTsClass(schema *ModelSchema, tsClass *TsClassObject) {
	if schema == nil {
		log.Panic("schema is nil")
	}
	log.Info(tsClass.String())
	tsClass.RemoveLineType(LINE_TS_DEFINE_SINGLELINE)
	tsClass.RemoveLineType(LINE_TS_DEFINE_START)
	tsClass.RemoveLineType(LINE_TS_DEFINE_OBJ)
	tsClass.RemoveLineType(LINE_TS_DEFINE_END)
	this.genTsClassDefine(schema, tsClass)
	this.regenTsClassField(schema, tsClass)
	log.Info(tsClass.String())
}

func (this *Generator) regenTsClassField(schema *ModelSchema, tsClass *TsClassObject) {
	for _, body := range schema.Body {
		member := tsClass.GetClassMember(body.Name)
		if member != nil {
			if member.IsPublic {
				if member.Type != body.ToTsPublicType() {
					log.Warnf(member.Type, body.ToTsPublicType(), member.Name)

				}
				if member.Type == body.ToTsPublicType() {
					continue
				}
				member.Line.Text = body.ToTsPublicSingleLine()
			}
		} else {
			tsClass.InsertAfter(&LineText{
				Obj:      tsClass,
				LineNum:  0,
				Text:     body.ToTsPublicSingleLine(),
				LineType: 0,
			}, tsClass.GetLines(LINE_TS_CLASS_HEADER)[0])
		}
	}
}

func (this *Generator) genTsClassOther(schema *ModelSchema, tsClass *TsClassObject) {
	tsClass.AddLine(&LineText{
		Obj:      tsClass,
		LineNum:  0,
		Text:     schema.ToTsClassHeader(),
		LineType: 0,
	}, LINE_TS_CLASS_HEADER)
	for _, body := range schema.Body {
		tsClass.AddLine(&LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     body.ToTsPublicSingleLine(),
			LineType: 0,
		}, LINE_TS_CLASS_FIELD_PUBLIC)
	}
	tsClass.AddLine(&LineText{
		Obj:      tsClass,
		LineNum:  0,
		Text:     "\tconstructor() {",
		LineType: 0,
	}, LINE_TS_CLASS_CONSTRUCTOR_HEADER)
	tsClass.AddLine(&LineText{
		Obj:      tsClass,
		LineNum:  0,
		Text:     "\t\tsuper()",
		LineType: 0,
	}, LINE_ANY)
	tsClass.AddLine(&LineText{
		Obj:      tsClass,
		LineNum:  0,
		Text:     "\t}",
		LineType: 0,
	}, LINE_TS_CLASS_FUNC_HEADER)
	tsClass.AddLine(&LineText{
		Obj:      tsClass,
		LineNum:  0,
		Text:     "}",
		LineType: 0,
	}, LINE_CLOSURE_END)
}

func (this *Generator) genTsClassDefine(schema *ModelSchema, tsClass *TsClassObject) {
	if len(schema.Body) == 0 {
		tsClass.InsertLine(0, &LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     schema.ToSingleLine(),
			LineType: LINE_TS_DEFINE_SINGLELINE,
		})
	} else {
		line := tsClass.InsertLine(0, &LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     schema.ToLineStart(),
			LineType: LINE_TS_DEFINE_START,
		})
		bodyTexts := schema.ToLineObject()
		for _, bodyText := range bodyTexts {
			line = tsClass.InsertAfter(&LineText{
				Obj:      tsClass,
				LineNum:  0,
				Text:     bodyText,
				LineType: LINE_TS_DEFINE_OBJ,
			}, line)
		}
		line = tsClass.InsertAfter(&LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     schema.ToLineEnd(),
			LineType: LINE_TS_DEFINE_END,
		}, line)
	}
}

func (this *Generator) getTsClassByName(s string) *TsClassObject {
	for _, class := range this.TsClassObjects {
		if class.ClassName == s {
			return class
		}
	}
	return nil
}

func (this *Generator) getTsModelFileBySchema(schema *ModelSchema) *TsModelFile {
	if !this.Individual {
		return this.TsModels[0]
	}
	tsPath := strings.Replace(schema.Path, ".go", ".ts", -1)
	tsPath = path.Join(this.TsPath, tsPath)

	log.Infof("tsPath", tsPath)
	for _, file := range this.TsModels {
		if file.FilePath == tsPath {
			return file
		}
	}
	err := util.CreateFile(tsPath)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	file, err := this.LoadAndParseTsFile(tsPath)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return file
}

func (this *Generator) ProcessGoModel2Schema(model *GoStructObject) {
	if !model.IsModel() {
		return
	}
	schema := &ModelSchema{
		Name:          model.StructName,
		Path:          model.file.GetGoRelativePath(),
		Type:          model.TagId,
		ContainerType: TAG_Proto,
		KeyType:       0,
		Body:          make([]*ModelSchema, 0),
		Depends:       make([]string, 0),
		model:         model,
	}
	if model.IComponent {
		schema.Component = true
	}
	GetTypeRegistry().RegistryTemplateTag(model.TagId, model.StructName)
	this.Schemas = append(this.Schemas, schema)
}

func (this *Generator) PostProcessAllGoSchemas() {
	for _, schema := range this.Schemas {
		for index, member := range schema.model.Fields {
			schema1 := &ModelSchema{
				Index:         index,
				Name:          member.Name,
				Path:          "",
				ContainerType: 0,
				KeyType:       0,
				Body:          make([]*ModelSchema, 0),
				Depends:       make([]string, 0),
			}
			tag, isEnum, tagstr1, tagstr2 := this.MatchGoExistTag(member.Type)
			if isEnum {
				schema1.Type = TAG_Int
				schema1.EnumName = member.Type
			} else if tag == TAG_List {
				schema1.ContainerType = TAG_List
				if tag1, _, _ := MatchGoSystemTag(tagstr1); tag1 != 0 {
					schema1.Type = tag1
				} else {
					if tagId := this.getRelativeTagId(tagstr1); tagId != 0 {
						schema1.Type = tagId
					} else {
						continue
					}
				}
			} else if tag == TAG_Map {
				schema1.ContainerType = TAG_Map
				switch tagstr1 {
				case "string":
					schema1.KeyType = TAG_String
				case "int64":
					schema1.KeyType = TAG_Long
				case "int32", "uint32":
					schema1.KeyType = TAG_Int
				default:
					continue
				}
				if tag1, _, _ := MatchGoSystemTag(tagstr2); tag1 != 0 {
					schema1.Type = tag1
				} else {
					if tagId := this.getRelativeTagId(tagstr2); tagId != 0 {
						schema1.Type = tagId
					} else {
						continue
					}
				}
			} else if tag > 0 {
				schema1.Type = tag
			} else {
				continue
			}
			schema.Body = append(schema.Body, schema1)
		}
	}
}

func (this *Generator) MatchGoExistTag(s string) (tag BINARY_TAG, isEnum bool, tagstr1 string, tagstr2 string) {
	log.Warnf("MatchGoExistTag", s)
	sysTag, tagstr1, tagstr2 := MatchGoSystemTag(s)
	if sysTag == TAG_End {
		if this.MatchGoEnum(s) {
			return TAG_Int, true, tagstr1, tagstr2
		}
		if tag := this.MatchGoStruct(s); tag != 0 {
			return tag, false, tagstr1, tagstr2
		}
		if tagId := this.getRelativeTagId(s); tagId != 0 {
			return tag, false, tagstr1, tagstr2
		}
	} else {
		log.Warnf("MatchGoExistTag", sysTag.String(), tagstr1, tagstr2, s)
		return sysTag, false, tagstr1, tagstr2
	}
	return 0, false, tagstr1, tagstr2
}

func (this *Generator) MatchGoEnum(s string) bool {
	for _, enum := range this.GoEnumObjects {
		if enum.Type() == s {
			return true
		}
	}
	return false
}

func (this *Generator) MatchGoStruct(s string) BINARY_TAG {
	for _, obj := range this.GoStructObjects {
		if obj.StructName == s {
			return obj.TagId
		}
	}
	return 0
}

func (this *Generator) GenerateTsEnum() {
	strs := auto_gen_header
	importObjs := this.TsEnums.GetObj(OBJ_TS_IMPORTS)
	for _, obj := range importObjs {
		strs += obj.String()
	}
	strs += "\n"
	for _, enum := range this.GoEnumObjects {
		typ := enum.Type()
		strs += "export enum " + typ + " {\n"
		for _, line := range enum.lines {
			//改成1去掉前缀
			name := strings.Replace(line.GetName(), typ+"_", "", 0)
			value := line.GetValue()
			switch line.LineType {
			case LINE_GO_ENUM_VARIABLE_IOTA:
				strs += "\t" + name + " = " + strconv.Itoa(value) + ",\n"
			case LINE_GO_ENUM_VARIABLE:
				strs += "\t" + name + " = " + strconv.Itoa(value) + ",\n"
			case LINE_GO_ENUM_AUTO:
				strs += "\t" + name + ",\n"
			}
		}
		strs += "}\n"
		strs += "\n"
	}
	err := ioutil.WriteFile(this.TsEnums.FilePath, []byte(strs), 0)
	if err != nil {
		log.Errorf(err.Error())
	}
}
