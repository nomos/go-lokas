package protocol

import (
	"errors"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/promise"
	"github.com/nomos/go-lokas/util"
	"os"
	"path"
	"strconv"
)

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
					panic("duplicated class:"+addedObj.ClassName)
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

func (this *Generator) processModelPackages()error{
	err:=this.processModelObjects()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	for _,m:=range this.Models {
		f:=m.(*ModelFile)
		pack:=f.ProcessPackages()
		p:=this.ModelPackages[pack.PackageName]
		if p!=nil {
			if p.GoPackageName!=pack.GoPackageName {
				return errors.New("wrong go package")
			}
			if p.CsPackageName!=pack.CsPackageName {
				return errors.New("wrong cs package")
			}
			if p.TsPackageName!=pack.TsPackageName {
				return errors.New("wrong ts package")
			}
		}
		this.ModelPackages[pack.PackageName]= pack
	}
	for _,m:=range this.Models {
		f:=m.(*ModelFile)
		ids:=f.ProcessIds()
		errs:=f.ProcessErrors()
		imports:=f.ProcessImports()
		importPacks:=[]*ModelPackageObject{}
		for _,p:=range imports {
			pa,ok:=this.ModelPackages[p]
			if !ok {
				log.Error(err.Error())
				return errors.New("cant found pack name:"+p)
			}
			importPacks = append(importPacks, pa)

		}
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
	return nil
}