package protocol

import (
	"github.com/nomos/go-lokas/util/promise"
)

type ConfObject struct {
	DefaultGeneratorObj
}

func NewConfObject(file GeneratorFile) *ConfObject {
	ret := &ConfObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_CONF, file)
	return ret
}

func (this *ConfObject) CheckLine(line *LineText) bool {
	if this.TryAddLine(line, LINE_EMPTY) {
		return true
	}
	if this.TryAddLine(line, LINE_CONF_TAG) {
		tagName := LINE_CONF_TAG.RegReplaceValue(line.Text)
		this.file.(*ConfFile).TagName = tagName
		return true
	}
	if this.TryAddLine(line, LINE_CONF_PACKAGE) {
		packageName := LINE_CONF_PACKAGE.RegReplaceValue(line.Text)
		this.file.(*ConfFile).PackageName = packageName
		return true
	}

	if this.TryAddLine(line, LINE_CONF_OFFSET) {
		this.file.(*ConfFile).Offset = line.GetValue()
		return true
	}
	if this.TryAddLine(line, LINE_CONF_RECUR) {
		recurStr := line.GetValue()
		this.file.(*ConfFile).Recursive = recurStr > 0
		return true
	}
	this.GetLogger().Panic("parse ConfObject error")
	return false
}

type ConfFile struct {
	*DefaultGeneratorFile
	TagName     string
	PackageName string
	Offset      int
	Recursive   bool
}

var _ GeneratorFile = (*ConfFile)(nil)

func NewConfFile(generator *Generator) *ConfFile {
	ret := &ConfFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_CONFIG
	return ret
}

func (this *ConfFile) Parse() *promise.Promise[interface{}] {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parseConfObject(0, nil)
		this.GetLogger().Infof("parseConf", offset, success)
		resolve(nil)
	})
}

func (this *ConfFile) parseConfObject(num int, lastObj GeneratorObject) (int, bool) {
	if num > len(this.Lines)-1 {
		this.GetLogger().Warn("reach end of file")
		return num, lastObj != nil
	}
	line := this.Lines[num]
	if lastObj == nil {
		lastObj = NewConfObject(this)
	}
	if lastObj.CheckLine(line) {
		num++
	}
	return this.parseConfObject(num, lastObj)
}
