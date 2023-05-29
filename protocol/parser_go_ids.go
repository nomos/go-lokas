package protocol

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/promise"
)

type GoIdRegObject struct {
	DefaultGeneratorObj
}

func NewGoIdRegObject(file GeneratorFile) *GoIdRegObject {
	ret := &GoIdRegObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_ID_REG, file)
	return ret
}

func (this *GoIdRegObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_GO_INIT_FUNC_HEADER) {
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
		if this.TryAddLine(line, LINE_GO_TAG_REGISTRY) {
			tagName := line.GetTagName()
			structName := line.GetStructName()
			pkgName := line.GetPkgName()
			typ := line.GetTypeName()
			this.file.(*GoIdsFile).AssignedTypes = append(this.file.(*GoIdsFile).AssignedTypes, &GoAssignedId{
				Tag:     tagName,
				Struct:  structName,
				Value:   0,
				Package: pkgName,
				Line:    line.LineNum,
			})
			log.WithFields(log.Fields{
				"line":       line.LineNum,
				"text":       line.Text,
				"tagName":    tagName,
				"structName": structName,
				"type":       typ,
			}).Info("LINE_GO_TAG_REGISTRY")
			return true
		}
		log.Panic("parse GoIdRegObject Body error")
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse GoIdRegObject error")
	return false
}

type GoAssignedId struct {
	Tag     string
	Struct  string
	Package string
	Value   int
	Line    int
}

type GoIdObject struct {
	DefaultGeneratorObj
}

func NewGoIdObject(file GeneratorFile) *GoIdObject {
	ret := &GoIdObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_GO_ID, file)
	return ret
}

func (this *GoIdObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
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
		if this.TryAddLine(line, LINE_GO_TAG_DEFINER) {
			structName := line.GetStructName()
			tagName := line.GetTagName()
			value := line.GetValue()
			log.WithFields(log.Fields{
				"line":       line.LineNum,
				"text":       line.Text,
				"structName": structName,
				"tagName":    tagName,
				"value":      value,
			}).Info("LINE_GO_TAG_DEFINER")
			if value > this.GetFile().(*GoIdsFile).MaxId {
				this.GetFile().(*GoIdsFile).MaxId = value
			}
			if value < this.GetFile().(*GoIdsFile).MinId {
				this.GetFile().(*GoIdsFile).MinId = value
			}
			this.GetFile().(*GoIdsFile).AssignedIds = append(this.GetFile().(*GoIdsFile).AssignedIds, &GoAssignedId{
				Tag:    tagName,
				Struct: structName,
				Value:  value,
				Line:   line.LineNum,
			})
			return true
		}
		log.Panicf("parse GoIdObject Body error:", line.Text)
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse GoIdObject error")
	return false
}

type GoIdsFile struct {
	*DefaultGeneratorFile
	TagName       string
	Offset        int
	AssignedIds   []*GoAssignedId
	AssignedTypes []*GoAssignedId
	MinId         int
	MaxId         int
}

var _ GeneratorFile = (*GoIdsFile)(nil)

func NewGoIdsFile(generator *Generator) *GoIdsFile {
	ret := &GoIdsFile{DefaultGeneratorFile: NewGeneratorFile(generator), AssignedIds: []*GoAssignedId{}, AssignedTypes: []*GoAssignedId{}}
	ret.GeneratorFile = ret
	ret.FileType = FILE_GO_IDS
	ret.GetLogger().Warnf("GoIdsFile", ret.FileType.String())
	return ret
}

func (this *GoIdsFile) Parse() *promise.Promise[interface{}] {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_GO_PACKAGE)
		this.GetLogger().Infof("parseGoModelPackage", offset, success)
		offset, success = this.parse(offset, OBJ_GO_IMPORTS)
		this.GetLogger().Infof("parseGoImports", offset, success)
		offset, success = this.parse(offset, OBJ_GO_ID)
		this.GetLogger().Infof("parseGoTagId", offset, success)
		offset, success = this.parse(offset, OBJ_GO_ID_REG)
		this.GetLogger().Infof("parseGoTagReg", offset, success)
		offset, success = this.parseEmpty(offset, nil)
		this.GetLogger().Infof("parseEmpty", offset, success)
		resolve(nil)
	})
}

func (this *GoIdsFile) Generate() *promise.Promise[interface{}] {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		resolve(nil)
	})
}

func (this *GoIdsFile) ProcessIds() {

}
