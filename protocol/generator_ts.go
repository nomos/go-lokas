package protocol

import (
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/promise"
	"github.com/nomos/go-lokas/util/stringutil"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func (this *Generator) GenerateModel2Ts() error {
	err := this.processModelPackages()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	err = this.generateModel2TsEnums()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	err = this.generateModel2TsIds()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	err = this.generateModel2TsClasses()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	this.GetLogger().Warnf("GenerateModel2Ts Finish")
	return nil
}

func (this *Generator) LoadTsFolder(p string) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		err := this.LoadTsIds(p)
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

func (this *Generator) LoadTsIds(p string) error {
	p = strings.ReplaceAll(p, "\\", "/")
	baseName := path.Base(p)
	this.TsPath = p
	idsPath := util.FindFile(p, baseName+"_ids.ts", false)
	if idsPath == "" {
		idsPath = path.Join(p, baseName+"_ids.ts")
		util.CreateFile(idsPath)
	}
	this.TsIds = NewTsIdsFile(this)
	this.GetLogger().Warnf("ts idsPath", idsPath)
	_, err := this.TsIds.Load(idsPath).Await()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	_, err = this.TsIds.Parse().Await()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) LoadGo2TsModels(p string) error {
	baseName := path.Base(p)
	var err error
	util.WalkDirFilesWithFunc(p, func(filePath string, file os.FileInfo) bool {
		if path.Ext(filePath) != "ts" {
			return false
		}
		fileName := path.Base(filePath)
		if fileName == baseName+"_ids.ts" || strings.HasPrefix(fileName, "enum_") {
			return false
		}
		_, err = this.LoadAndParseTsFile(filePath)
		if err != nil {
			this.GetLogger().Error(err.Error())
			return true
		}
		return false
	}, true)
	return err
}

func (this *Generator) LoadAndParseTsFile(modelsPath string) (*TsModelFile, error) {
	this.GetLogger().Warnf("ts modelsPath", modelsPath)
	file := NewTsModelFile(this)
	_, err := file.Load(modelsPath).Await()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return nil, err
	}
	_, err = file.Parse().Await()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return nil, err
	}
	for _, tsClass := range file.ProcessClasses() {
		if tsClass.IsModel() {
			this.TsClassObjects = append(this.TsClassObjects, tsClass)
		}
	}
	this.TsModels = append(this.TsModels, file)
	this.processTsClassObjects()
	return file, nil
}

func (this *Generator) generateModel2TsIds() error {
	this.GetLogger().Info("generateModel2TsIds")
	this.LoadTsIds(this.TsPath)
	strs := auto_gen_header
	importObjs := this.TsIds.GetObj(OBJ_TS_IMPORTS)
	for _, obj := range importObjs {
		obj.RemoveLineType(LINE_EMPTY)
		strs += obj.String()
	}

	packName := ""
	for _, p := range this.ModelPackages {
		if p.TsPackageName == "" {
			continue
		}
		packName = p.TsPackageName
	}

	errIds := []*ModelError{}
	for _, p := range this.ModelPackages {
		if p.TsPackageName == "" {
			continue
		}
		for _, id := range p.Errors {
			errIds = append(errIds, id)
		}
	}
	sort.Slice(errIds, func(i, j int) bool {
		return errIds[i].ErrorId < errIds[j].ErrorId
	})
	strs += "\n"
	strs += "export namespace " + packName + " {"

	for _, id := range errIds {
		strs += "\n"
		strs += id.TsString(this)
	}
	strs += "\n"
	strs += "}\n\n"
	strs += "(function () {\n"
	strs += "\t" + `if (window["CC_EDITOR"]) {` + "\n"
	strs += "\t\treturn;\n"
	strs += "\t}\n"
	ids := []*ModelId{}
	for _, p := range this.ModelPackages {
		if p.TsPackageName == "" {
			continue
		}
		for _, id := range p.Ids {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Id < ids[j].Id
	})
	for _, id := range ids {
		strs += "\tTypeRegistry.getInstance().RegisterCustomTag(\"" + id.Name + "\"," + strconv.Itoa(id.Id) + ")\n"
	}
	strs += "})()\n"
	err := ioutil.WriteFile(this.TsIds.FilePath, []byte(strs), 0644)
	if err != nil {
		this.GetLogger().Errorf(err.Error())
	}
	return nil
}

func (this *Generator) generateModel2TsClasses() error {
	//this.GetLogger().Warnf("models", this.GoStructObjects)
	//this.GetLogger().Warnf("GoPath",this.GoPath)
	//for _,m:=range this.ModelClassObjects {
	//	err:=this.generateModel2TsClass(m)
	//	if err != nil {
	//		this.GetLogger().Error(err.Error())
	//		return err
	//	}
	//}
	this.processTsClassObjects()
	for i := len(this.ModelClassObjects) - 1; i >= 0; i-- {
		schema := this.ModelClassObjects[i]
		tsFile := this.getTsModelFileByModel(schema)
		if tsFile != nil {
			tsFile.Package = schema.TsPackage
			tsClass := this.getTsClassByName(schema.ClassName)
			if tsClass != nil {
				this.regenTsClass(schema, tsClass)
			} else {
				this.genTsClass(tsFile, schema)
			}
		} else {
			tsFile.Package = schema.TsPackage
			this.genTsClass(tsFile, schema)
		}
	}
	for _, modelFile := range this.TsModels {
		this.GetLogger().Infof("modelFile.FileName", modelFile.FileName)
		strs := auto_gen_header
		modelFile.RemoveAutoGenHeader()
		modelFile.RemoveObjType(OBJ_EMPTY)
		imports := modelFile.GetObj(OBJ_TS_IMPORTS)
		for _, impo := range imports {
			impo.RemoveLineType(LINE_EMPTY)
		}
		if len(imports) == 0 {
			imports := NewTsImportObject(modelFile)
			modelFile.InsertObject(0, imports)
		}
		for _, obj := range modelFile.Objects {
			strs += obj.String()
		}

		strs = stringutil.TrimEnd(strs)
		p := path.Join(this.TsPath, modelFile.Package+"_"+stringutil.SplitCamelCaseLowerSnake(modelFile.ClassName))
		p += ".ts"
		ioutil.WriteFile(p, []byte(strs), 0644)
	}
	return nil
}

func (this *Generator) genTsClass(tsFile *TsModelFile, schema *ModelClassObject) {
	var tsClass *TsClassObject
	tsClass = NewTsClassObject(tsFile)
	tsClass.Package = schema.TsPackage
	imports := tsFile.GetObj(OBJ_TS_IMPORTS)
	for _, impor := range imports {
		impor.RemoveLineType(LINE_EMPTY)
	}
	this.GetLogger().Warnf("tsFile", tsFile.Objects)
	this.GetLogger().Warnf("tsFile", len(imports))
	if len(imports) == 0 {
		tsFile.InsertObject(0, tsClass)
	} else {
		tsFile.InsertAfter(tsClass, imports[len(imports)-1])
	}
	this.genTsClassDefine(schema, tsClass)
	this.genTsClassOther(schema, tsClass)
}

func (this *Generator) regenTsClass(schema *ModelClassObject, tsClass *TsClassObject) {
	if schema == nil {
		this.GetLogger().Panic("schema is nil")
	}
	this.GetLogger().Info(tsClass.String())
	tsClass.RemoveLineType(LINE_TS_DEFINE_SINGLELINE)
	tsClass.RemoveLineType(LINE_TS_DEFINE_START)
	tsClass.RemoveLineType(LINE_TS_DEFINE_OBJ)
	tsClass.RemoveLineType(LINE_TS_DEFINE_END)
	this.genTsClassDefine(schema, tsClass)
	this.regenTsClassField(schema, tsClass)
	this.GetLogger().Info(tsClass.String())
}

func (this *Generator) regenTsClassField(schema *ModelClassObject, tsClass *TsClassObject) {
	for _, body := range schema.Fields {
		member := tsClass.GetClassMember(body.Name)
		if member != nil {
			if member.IsPublic {
				if strings.ReplaceAll(strings.TrimSpace(member.Type), " ", "") == "Uint8Array" && body.TsPublicType(this) == "number[]" {
					continue
				}
				if strings.ReplaceAll(strings.TrimSpace(member.Type), " ", "") != strings.TrimSpace(body.TsPublicType(this)) {
					this.GetLogger().Warnf("ADSASDASDSDADSDS", member.Type, body.TsPublicType(this), member.Name)

				}
				if strings.ReplaceAll(strings.TrimSpace(member.Type), " ", "") == strings.TrimSpace(body.TsPublicType(this)) {
					continue
				}
				member.Line.Text = body.TsPublicString(this)
			}
		} else {
			tsClass.InsertAfter(&LineText{
				Obj:      tsClass,
				LineNum:  0,
				Text:     body.TsPublicString(this),
				LineType: 0,
			}, tsClass.GetLines(LINE_TS_CLASS_HEADER)[0])
		}
	}
}

func (this *Generator) genTsClassOther(schema *ModelClassObject, tsClass *TsClassObject) {
	tsClass.AddLine(&LineText{
		Obj:      tsClass,
		LineNum:  0,
		Text:     schema.ToTsClassHeader(this, tsClass),
		LineType: 0,
	}, LINE_TS_CLASS_HEADER)
	for _, field := range schema.Fields {
		tsClass.AddLine(&LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     field.TsPublicString(this),
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

func (this *Generator) genTsClassDefine(schema *ModelClassObject, tsClass *TsClassObject) {
	if len(schema.Fields) == 0 {
		tsClass.InsertLine(0, &LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     schema.TsDefineSingleLine(this),
			LineType: LINE_TS_DEFINE_SINGLELINE,
		})
	} else {
		line := tsClass.InsertLine(0, &LineText{
			Obj:      tsClass,
			LineNum:  0,
			Text:     schema.TsDefineStart(this),
			LineType: LINE_TS_DEFINE_START,
		})
		bodyTexts := schema.TsDefineLines(this, tsClass)
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
			Text:     schema.TsDefineEnd(this),
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
	this.GetLogger().Info("NIL")
	return nil
}

func (this *Generator) getTsModelFileByModel(schema *ModelClassObject) *TsModelFile {
	tsPath := path.Join(this.TsPath, schema.TsPackage+"_"+stringutil.SplitCamelCaseLowerSnake(schema.ClassName)) + ".ts"

	for _, file := range this.TsModels {
		if file.FilePath == tsPath {
			return file
		}
	}
	err := util.CreateFile(tsPath)
	if err != nil {
		this.GetLogger().Error(err.Error())
		return nil
	}
	file, err := this.LoadAndParseTsFile(tsPath)
	if err != nil {
		this.GetLogger().Error(err.Error())
		return nil
	}
	file.ClassName = schema.ClassName
	return file
}

func (this *Generator) processTsClassObjects() {
	ret := make([]*TsClassObject, 0)
	for _, file := range this.TsModels {
		objects := file.ProcessClasses()
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
	this.TsClassObjects = ret
}

func (this *Generator) generateModel2TsEnums() error {
	//strs := auto_gen_header
	//importObjs := this.TsEnums.GetObj(OBJ_TS_IMPORTS)
	//for _, obj := range importObjs {
	//	strs += obj.String()
	//}
	//strs += "\n"
	for _, enum := range this.ModelEnumObjects {
		name := stringutil.SplitCamelCaseLowerSnake(enum.EnumName)
		err := ioutil.WriteFile(path.Join(this.TsPath, enum.TsPackage+"_"+"enum_"+name+".ts"), []byte(stringutil.TrimEnd(enum.TsString(this))), 0644)
		if err != nil {
			this.GetLogger().Error(err.Error())
			return err
		}
	}
	return nil
}
