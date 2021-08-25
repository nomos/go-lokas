package protocol

import (
	"github.com/iancoleman/orderedmap"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/promise"
	"github.com/nomos/go-lokas/util"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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