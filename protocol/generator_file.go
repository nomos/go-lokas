package protocol

import (
	"bufio"
	"errors"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/promise"
	"io"
	"os"
	path2 "path"
	"strings"
)

type FileType int

const (
	FILE_UNDEFINED FileType = iota
	FILE_CONFIG
	FILE_GO_MODEL
	FILE_GO_IDS
	FILE_TS_MODEL
	FILE_TS_ENUM
	FILE_TS_IDS
	FILE_PROTO
	FILE_PROTO_IDS
	FILE_MODEL
)

var file_str_map = make(map[FileType]string)

const auto_gen_header = "//this is a generate file,do not edit it\n"

func init() {
	file_str_map[FILE_UNDEFINED] = "FILE_UNDEFINED"
	file_str_map[FILE_CONFIG] = "FILE_CONFIG"
	file_str_map[FILE_GO_MODEL] = "FILE_GO_MODEL"
	file_str_map[FILE_GO_IDS] = "FILE_GO_IDS"
	file_str_map[FILE_TS_MODEL] = "FILE_TS_MODEL"
	file_str_map[FILE_TS_ENUM] = "FILE_TS_ENUM"
	file_str_map[FILE_TS_IDS] = "FILE_TS_IDS"
	file_str_map[FILE_PROTO] = "FILE_PROTO"
	file_str_map[FILE_PROTO_IDS] = "FILE_PROTO_IDS"
	file_str_map[FILE_MODEL] = "FILE_MODEL"
}

func (this FileType) String() string {
	return file_str_map[this]
}

type GeneratorFile interface {
	GetLogger()log.ILogger
	Parse() *promise.Promise
	Load(string) *promise.Promise
	GetFilePath() string
	GetTsRelativePath() string
	GetGoRelativePath() string
	SetGenerator(generator *Generator)
	GetLines() []*LineText
	GetFile() GeneratorFile
	AddObject(obj GeneratorObject)
	InsertObject(pos int, obj GeneratorObject)
}

type DefaultGeneratorFile struct {
	GeneratorFile
	file       *os.File
	FileType   FileType
	FilePath   string
	IsDir      bool
	Generator    *Generator
	DirPath    string
	FileName   string
	LineLength int
	Lines      []*LineText
	Objects    []GeneratorObject
}

func NewGeneratorFile(generator    *Generator) *DefaultGeneratorFile {
	ret := &DefaultGeneratorFile{}
	ret.Generator = generator
	ret.reset()
	return ret
}

func (this *DefaultGeneratorFile) CheckFinish(offset int)bool {
	return len(this.Lines) == offset
}

func (this *DefaultGeneratorFile) GetFilePath() string {
	return this.FilePath
}

func (this *DefaultGeneratorFile) SetGenerator(generator *Generator) {
	this.Generator = generator
}

func (this *DefaultGeneratorFile) GetTsRelativePath()string {
	return strings.Replace(this.FilePath,this.Generator.TsPath,"",-1)
}

func (this *DefaultGeneratorFile) GetGoRelativePath()string {
	return strings.Replace(this.FilePath,this.Generator.GoPath,"",-1)
}

func (this *DefaultGeneratorFile) GetFile() GeneratorFile {
	return this.GeneratorFile
}

func (this *DefaultGeneratorFile) GetLines() []*LineText {
	return this.Lines
}

func (this *DefaultGeneratorFile) reset() {
	this.FilePath = ""
	this.LineLength = 0
	this.Lines = make([]*LineText, 0)
	this.Objects = make([]GeneratorObject, 0)
}

func (this *DefaultGeneratorFile) GetLogger()log.ILogger{
	return this.Generator.GetLogger()
}

func (this *DefaultGeneratorFile) load(path string) error {
	this.GetLogger().Warnf("DefaultGeneratorFile load ", path)
	this.reset()
	this.FilePath = path
	file, err := os.OpenFile(path, os.O_CREATE, 0666)
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	this.file = file
	stat, err := file.Stat()
	if stat.IsDir() {
		this.GetLogger().Panic("path is dir")
	} else {
		this.IsDir = false
		this.DirPath = path2.Dir(path)
		this.FileName = path2.Base(path)
	}
	if err != nil {
		this.GetLogger().Error(err.Error())
		return err
	}
	this.GetLogger().Infof("size", stat.Size())
	buf := bufio.NewReader(file)
	lineNo := 0
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		//line = strings.TrimSpace(line)
		this.Lines = append(this.Lines, &LineText{
			Obj:     nil,
			LineNum: lineNo,
			Text:    line,
		})
		if err != nil {
			if err == io.EOF {
				this.GetLogger().Info("read file ok")
				break
			} else {
				this.GetLogger().Error("read file error")
				return errors.New("read file error")
			}
		}
		lineNo++
	}
	this.LineLength = lineNo
	return nil
}

func (this *DefaultGeneratorFile) parseEmpty(num int, lastObj GeneratorObject) (int, bool) {
	if num > len(this.Lines)-1 {
		this.GetLogger().Warn("reach end of file")
		if lastObj!=nil&&len(lastObj.Lines()) > 0 {
			this.AddObject(lastObj)
		}
		return num, lastObj != nil
	}
	line := this.Lines[num]
	if lastObj == nil || lastObj.ObjectType() == OBJ_EMPTY {
		if lastObj == nil {
			lastObj = NewEmptyObject(this)
		}
		if lastObj.CheckLine(line) {
			num++
			return this.parseEmpty(num, lastObj)
		} else {
			if len(lastObj.Lines()) > 0 {
				this.AddObject(lastObj)
				return num, true
			}
		}
	}
	return num, false
}

func (this *DefaultGeneratorFile) parseComment(num int, lastObj GeneratorObject) (int, bool) {
	if num > len(this.Lines)-1 {
		this.GetLogger().Warn("reach end of file")
		if lastObj!=nil&&len(lastObj.Lines()) > 0 {
			this.AddObject(lastObj)
		}
		return num, lastObj != nil
	}
	line := this.Lines[num]
	if lastObj == nil || lastObj.ObjectType() == OBJ_COMMENT {
		if lastObj == nil {
			lastObj = NewCommentObject(this)
		}
		if lastObj.CheckLine(line) {
			num++
			return this.parseComment(num, lastObj)
		} else {
			if len(lastObj.Lines()) > 0 {
				this.AddObject(lastObj)
				return num, true
			}
		}
	}
	return num, false
}

func (this *DefaultGeneratorFile) Load(path string) *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		go func() {
			err := this.load(path)
			if err != nil {
				this.GetLogger().Error(err.Error())
				reject(err)
				return
			}
			resolve(nil)
		}()
	})
}

func (this *DefaultGeneratorFile) Parse() *promise.Promise {
	return promise.Reject(errors.New("have no implement"))
}

func (this *DefaultGeneratorFile) GetObj(objType ObjectType) []GeneratorObject {
	ret := make([]GeneratorObject, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == objType {
			ret = append(ret, obj)
		}
	}
	return ret
}

type GenerateObjRemoveCondition func(object GeneratorObject)bool

func (this *DefaultGeneratorFile) RemoveObjByCondition(condition GenerateObjRemoveCondition) {
	result := make([]GeneratorObject, 0)
	for _, obj := range this.Objects {
		if condition(obj) {
			continue
		}
		result = append(result, obj)
	}
	this.Objects = result
}

func (this *DefaultGeneratorFile) RemoveObjType(objType ObjectType) {
	result := make([]GeneratorObject, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == objType {
			continue
		}
		result = append(result, obj)
	}
	this.Objects = result
}

func (this *DefaultGeneratorFile) AddObject(obj GeneratorObject) {
	this.Objects = append(this.Objects, obj)
}

func (this *DefaultGeneratorFile) InsertObject(pos int, obj GeneratorObject) {
	objsCopy:=make([]GeneratorObject,len(this.Objects))
	copy(objsCopy,this.Objects)
	rPart := objsCopy[pos:]
	this.Objects = this.Objects[0:pos]
	this.Objects = append(this.Objects, obj)
	for _, iter := range rPart {
		this.Objects = append(this.Objects, iter)
	}
}

func (this *DefaultGeneratorFile) InsertAfter(obj GeneratorObject, after GeneratorObject) {
	for index, iter := range this.Objects {
		if after == iter {
			this.InsertObject(index+1, obj)
			return
		}
	}
	this.GetLogger().Panic("cant found obj")
}

func (this *DefaultGeneratorFile) RemoveAutoGenHeader(){
	this.RemoveObjByCondition(func(object GeneratorObject) bool {
		if object.ObjectType() == OBJ_COMMENT {
			lines:=object.GetLines(LINE_COMMENT)
			if len(lines) ==1 {
				if lines[0].Text == "//this is a generate file,do not edit it" {
					return true
				}
			}
			return false
		}
		return false
	})
}

func (this *DefaultGeneratorFile) Generate(typ FileType) *promise.Promise {
	return nil
}

func (this *DefaultGeneratorFile) MarshalToFile(path string) *promise.Promise {
	return nil
}

func (this *DefaultGeneratorFile) parseMultiple(num int, lastObj GeneratorObject, objs ... ObjectType) (int, bool) {
	if num > len(this.Lines)-1 {
		this.GetLogger().Warn("reach end of file")
		if lastObj != nil && len(lastObj.Lines()) > 0 {
			this.AddObject(lastObj)
		}
		return num, lastObj != nil
	}
	line := this.Lines[num]
	if lastObj == nil {
		num, success := this.parseEmpty(num, nil)
		if success {
			return this.parseMultiple(num, nil, objs...)
		}
		num, success = this.parseComment(num, nil)
		if success {
			return this.parseMultiple(num, nil, objs...)
		}

		for _, obj := range objs {
			lastObj = obj.Create(this)
			if lastObj.CheckLine(line) {
				num++
				return this.parseMultiple(num, lastObj, objs...)
			}
		}
		this.GetLogger().Panicf("parse multiple error",line.Text,line.LineNum)
	} else {
		if lastObj.CheckLine(line) {
			num++
			return this.parseMultiple(num, lastObj, objs...)
		} else {
			if len(lastObj.Lines()) > 0 {
				this.AddObject(lastObj)
			}
			return this.parseMultiple(num, nil, objs...)
		}
	}
	return num, lastObj != nil
}

func (this *DefaultGeneratorFile) parseOne(num int, lastObj GeneratorObject, objType ObjectType) (int, bool) {
	if num > len(this.Lines)-1 {
		this.GetLogger().Warn("reach end of file")
		if lastObj != nil && len(lastObj.Lines()) > 0 {
			this.AddObject(lastObj)
		}
		return num, lastObj != nil
	}
	line := this.Lines[num]
	if lastObj == nil {
		num, success := this.parseEmpty(num, nil)
		if success {
			return this.parseOne(num, nil, objType)
		}
		num, success = this.parseComment(num, nil)
		if success {
			return this.parseOne(num, nil, objType)
		}
		lastObj = objType.Create(this)
		return this.parseOne(num, lastObj, objType)
	} else {
		if lastObj.CheckLine(line) {
			num++
			return this.parseOne(num, lastObj, objType)
		} else {
			if len(lastObj.Lines()) > 0 {
				this.AddObject(lastObj)
			}
		}
	}
	return num, lastObj != nil
}

func (this *DefaultGeneratorFile) parse(num int, objs ...ObjectType) (int, bool) {

	if len(objs) == 0 {
		this.GetLogger().Panic("params error")
		return 0, false
	} else if len(objs) == 1 {
		return this.parseOne(num, nil, objs[0])
	} else {
		return this.parseMultiple(num, nil, objs...)
	}
}
