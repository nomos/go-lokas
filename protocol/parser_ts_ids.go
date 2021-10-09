package protocol

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/promise"
)

type TsIdsObject struct {
	DefaultGeneratorObj
}

func NewTsIdsObject(file GeneratorFile) *TsIdsObject {
	ret := &TsIdsObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_TS_IDS, file)
	return ret
}

func (this *TsIdsObject) CheckLine(line *LineText) bool {

	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_INIT_FUNC_HEADER) {
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

		if this.TryAddLine(line, LINE_TS_INIT_FUNC_COCOS) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_ID_REG) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_PROTO_ID_REG) {
			return true
		}
		if this.TryAddLine(line, LINE_TS_INIT_FUNC_END) {
			this.state = 2
			return true
		}
		log.Panic("parse TsIdsObject error")
	} else if this.state == 2 {
		return false
	}
	log.Panic("parse TsIdsObject error")
	return false
}

type TsIdsFile struct {
	*DefaultGeneratorFile
}

var _ GeneratorFile = (*TsIdsFile)(nil)

func NewTsIdsFile(generator *Generator) *TsIdsFile {
	ret := &TsIdsFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_TS_IDS
	return ret
}

func (this *TsIdsFile) Generate() *promise.Promise {
	return nil
}

func (this *TsIdsFile) Parse() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_TS_IMPORTS)
		log.Infof("parseTsImports", offset, success)
		offset, success = this.parse(offset, OBJ_TS_IDS, OBJ_EMPTY)
		log.Infof("parseTsIds", offset, success)
		//offset, success = this.parseGoMain(offset, nil)
		//log.Warnf("parseGoMain finish", offset, success)
		resolve(nil)
	})
}
