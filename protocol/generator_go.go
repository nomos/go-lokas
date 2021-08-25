package protocol

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/promise"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/stringutil"
	"io/ioutil"
	"os"
	"path"
)

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

func (this *Generator) GenerateModel2Go()error{
	err:=this.processModelPackages()
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
