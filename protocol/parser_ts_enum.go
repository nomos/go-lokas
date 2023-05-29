package protocol

import (
	"github.com/nomos/go-lokas/util/promise"
)

type TsEnumObject struct {
	DefaultGeneratorObj
}

func NewTsEnumObject(file GeneratorFile) *TsEnumObject {
	ret := &TsEnumObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_TS_ENUM, file)
	return ret
}

func (this *TsEnumObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_ENUM_CLOSURE_START) {
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
		if this.TryAddLine(line, LINE_TS_ENUM_OBJ) {
			return true
		}
		if this.TryAddLine(line, LINE_CLOSURE_END) {
			this.state = 2
			return true
		}
		this.GetLogger().Panic("parse TsEnumObject error")
	} else if this.state == 2 {
		return false
	}
	this.GetLogger().Panic("parse TsEnumObject error")
	return false
}

type TsEnumFile struct {
	*DefaultGeneratorFile
}

var _ GeneratorFile = (*TsEnumFile)(nil)

func NewTsEnumFile(generator *Generator) *TsEnumFile {
	ret := &TsEnumFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_TS_ENUM
	return ret
}

func (this *TsEnumFile) Generate() *promise.Promise[interface{}] {
	return nil
}

func (this *TsEnumFile) Parse() *promise.Promise[interface{}] {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_TS_IMPORTS)
		this.GetLogger().Warnf("parseTsImports", offset, success)
		offset, success = this.parse(offset, OBJ_TS_ENUM, OBJ_EMPTY)
		this.GetLogger().Warnf("parseTsEnum", offset, success)
		//offset, success = this.parseGoMain(offset, nil)
		//log.Warnf("parseGoMain finish", offset, success)
		resolve(nil)
	})
}
