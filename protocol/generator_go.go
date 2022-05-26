package protocol

import (
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/slice"
	"github.com/nomos/go-lokas/util/stringutil"
	"io/ioutil"
	"path"
	"strings"
)

func (this *Generator) GenerateModel2Go() error {
	defer func() {
		if r := recover(); r != nil {
			util.Recover(r, false)
		}
	}()
	err := this.processModelPackages()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	err = this.generateModel2GoClasses()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	err = this.generateModel2GoEnums()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	err = this.generateModel2GoIds()
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	this.GetLogger().Warnf("GenerateModel2Go Finish")
	return nil
}

func keysof(m map[string]*ModelPackageObject) []string {
	ret := []string{}
	for v, _ := range m {
		ret = append(ret, v)
	}
	return ret
}

func (this *Generator) getGoImportsString(deps []string) string {
	deps = slice.RemoveDuplicate(deps)
	ret := "\n"
	for _, v := range deps {
		pack, ok := this.ModelPackages[v]
		if !ok {
			this.GetLogger().Panic("cant found imports name:" + v)
			return ""
		}
		ret += "\t"
		ret += `. "` + pack.GoPackageName + `"`
		ret += "\n"
	}
	ret = strings.TrimRight(ret, "\n")
	return ret
}

func (this *Generator) generateModel2GoClasses() error {
	this.GetLogger().Warnf("GoPath", this.GoPath)
	for _, m := range this.ModelClassObjects {
		err := this.generateModel2GoClass(m)
		if err != nil {
			this.GetLogger().Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoClass(m *ModelClassObject) error {
	p := path.Join(this.GoPath, m.Package, "model_"+stringutil.SplitCamelCaseLowerSnake(m.ClassName))
	p += ".go"
	err := ioutil.WriteFile(p, []byte(m.GoString(this)), 0644)
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	p = path.Join(this.GoPath, m.Package, "model_"+stringutil.SplitCamelCaseLowerSnake(m.ClassName))
	p += "_impl.go"
	if !util.IsFileExist(p) {
		err = ioutil.WriteFile(p, []byte(m.GoImplString(this)), 0644)
		if err != nil {
			this.GetLogger().Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoEnums() error {
	this.GetLogger().Warnf("GoPath", this.GoPath)
	for _, m := range this.ModelEnumObjects {
		err := this.generateModel2GoEnum(m)
		if err != nil {
			this.GetLogger().Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoEnum(m *ModelEnumObject) error {
	p := path.Join(this.GoPath, m.Package, "enum_"+stringutil.SplitCamelCaseLowerSnake(m.EnumName))
	p += ".go"
	err := ioutil.WriteFile(p, []byte(m.GoString(this)), 0644)
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	return nil
}

func (this *Generator) generateModel2GoIds() error {
	this.GetLogger().Warnf("generateModel2GoIds")

	for _, p := range this.ModelPackages {
		err := this.generateModel2GoPackage(p)

		if err != nil {
			this.GetLogger().Error(err.Error())
			return err
		}
	}
	return nil
}

func (this *Generator) generateModel2GoPackage(pack *ModelPackageObject) error {
	p := path.Join(this.GoPath, pack.PackageName, "ids")
	p += ".go"
	this.GetLogger().Warnf("generateModel2GoPackage", p)
	err := ioutil.WriteFile(p, []byte(pack.GoString(this)), 0644)
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	return nil
}
