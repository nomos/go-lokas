package protocol

import (
	"github.com/nomos/go-lokas/util/promise"
	"strconv"
)

type ProtoIdsObj struct {
	DefaultGeneratorObj
	ProtoName string
	Id        int
}

var _ GeneratorObject = (*ProtoIdsObj)(nil)

func NewProtoIdsObj(file GeneratorFile) *ProtoIdsObj {
	ret := &ProtoIdsObj{
		DefaultGeneratorObj: DefaultGeneratorObj{},
	}
	ret.DefaultGeneratorObj.init(OBJ_PROTO_IDS, file)
	return ret
}

func (this *ProtoIdsObj) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_PROTO_ID) {
			this.Check(line)
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		return false
	}
	this.GetLogger().Panic("parse ProtoIdsObj error")
	return false
}

func (this *ProtoIdsObj) Check(line *LineText) {
	this.Id, _ = strconv.Atoi(LINE_PROTO_ID.RegReplaceValue(line.Text))
	this.ProtoName = LINE_PROTO_ID.RegReplaceName(line.Text)
	this.file.(*ProtoIdsFile).AddId(this.ProtoName, this.Id)
}

type ProtoIdsFile struct {
	*DefaultGeneratorFile
	Ids map[int]string
}

var _ GeneratorFile = (*ProtoIdsFile)(nil)

func NewProtoIdsFile(generator *Generator) *ProtoIdsFile {
	ret := &ProtoIdsFile{
		DefaultGeneratorFile: NewGeneratorFile(generator),
		Ids:                  make(map[int]string),
	}
	ret.GeneratorFile = ret
	ret.FileType = FILE_PROTO_IDS
	return ret
}

func (this *ProtoIdsFile) Generate() *promise.Promise[interface{}] {
	return nil
}

func (this *ProtoIdsFile) AddId(s string, id int) {
	this.Ids[id] = s
}

func (this *ProtoIdsFile) Parse() *promise.Promise[interface{}] {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_PROTO_PACKAGE)
		this.GetLogger().Warnf("parseProtoPackage", offset, success)
		offset, success = this.parse(offset, OBJ_COMMENT, OBJ_PROTO_IDS)
		//offset, success = this.parseGoImports(offset, nil)
		//this.GetLogger().Warnf("parseGoImports", offset, success)
		//offset, success = this.parseGoMain(offset, nil)
		//this.GetLogger().Warnf("parseGoMain finish", offset, success)
		resolve(nil)
	})
}
