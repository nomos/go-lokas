package protocol

import (
	"errors"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/promise"
	"io/ioutil"
	"os"
	"path"
)

func (this *Generator) LoadCsFolder(p string) *promise.Promise {
	this.CsPath = p
	err:=os.MkdirAll(this.CsPath, os.ModePerm)
	if err != nil {
		log.Error(err.Error())
		return promise.Reject(nil)
	}
	return promise.Resolve(nil)
}

func (this *Generator) GenerateModel2Cs()error{
	err:=this.processModelPackages()
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
