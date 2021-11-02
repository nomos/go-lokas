package protocol

import "github.com/nomos/go-lokas/log"

type ObjectType int

const (
	OBJ_UNDEFINED ObjectType = iota
	OBJ_GO_PACKAGE
	OBJ_GO_IMPORTS
	OBJ_GO_INTERFACE
	OBJ_GO_STRUCT
	OBJ_GO_STRUCT_FUNC
	OBJ_GO_FUNC
	OBJ_GO_VAR
	OBJ_GO_ENUM
	OBJ_GO_ID
	OBJ_GO_ID_REG
	OBJ_GO_DEFINE
	OBJ_CONF

	OBJ_EMPTY
	OBJ_COMMENT

	OBJ_TS_CLASS
	OBJ_TS_IMPORTS
	OBJ_TS_VAR
	OBJ_TS_FUNC
	OBJ_TS_ENUM
	OBJ_TS_IDS

	OBJ_MODEL_PACKAGE
	OBJ_MODEL_IMPORTS
	OBJ_MODEL_IDS
	OBJ_MODEL_ERRORS
	OBJ_MODEL_CLASS
	OBJ_MODEL_ENUM

	OBJ_PROTO_PACKAGE
	OBJ_PROTO_SYNTAX
	OBJ_PROTO_MSG
	OBJ_PROTO_IDS
	OBJ_PROTO_ENUM

	OBJ_PROTO_TS_INTERFACE
)

var obj_string_map = make(map[ObjectType]string)
//([A-Z|_]+)
//obj_string_map[$1] = "$1"
func init(){
	obj_string_map[OBJ_UNDEFINED] = "OBJ_UNDEFINED"
	obj_string_map[OBJ_GO_PACKAGE] = "OBJ_GO_PACKAGE"
	obj_string_map[OBJ_GO_IMPORTS] = "OBJ_GO_IMPORTS"
	obj_string_map[OBJ_GO_STRUCT] = "OBJ_GO_STRUCT"
	obj_string_map[OBJ_GO_STRUCT_FUNC] = "OBJ_GO_STRUCT_FUNC"
	obj_string_map[OBJ_GO_FUNC] = "OBJ_GO_FUNC"
	obj_string_map[OBJ_GO_VAR] = "OBJ_GO_VAR"
	obj_string_map[OBJ_GO_ENUM] = "OBJ_GO_ENUM"
	obj_string_map[OBJ_GO_ID] = "OBJ_GO_ID"
	obj_string_map[OBJ_GO_ID_REG] = "OBJ_GO_ID_REG"
	obj_string_map[OBJ_GO_DEFINE] = "OBJ_GO_DEFINE"
	obj_string_map[OBJ_CONF] = "OBJ_CONF"

	obj_string_map[OBJ_EMPTY] = "OBJ_EMPTY"
	obj_string_map[OBJ_COMMENT] = "OBJ_COMMENT"

	obj_string_map[OBJ_TS_CLASS] = "OBJ_TS_CLASS"
	obj_string_map[OBJ_TS_IMPORTS] = "OBJ_TS_IMPORTS"
	obj_string_map[OBJ_TS_VAR] = "OBJ_TS_VAR"
	obj_string_map[OBJ_TS_FUNC] = "OBJ_TS_FUNC"
	obj_string_map[OBJ_TS_ENUM] = "OBJ_TS_ENUM"
	obj_string_map[OBJ_TS_IDS] = "OBJ_TS_IDS"

	obj_string_map[OBJ_MODEL_PACKAGE] = "OBJ_MODEL_PACKAGE"
	obj_string_map[OBJ_MODEL_IMPORTS] = "OBJ_MODEL_IMPORTS"
	obj_string_map[OBJ_MODEL_IDS] = "OBJ_MODEL_IDS"
	obj_string_map[OBJ_MODEL_ERRORS] = "OBJ_MODEL_ERRORS"
	obj_string_map[OBJ_MODEL_CLASS] = "OBJ_MODEL_CLASS"
	obj_string_map[OBJ_MODEL_ENUM] = "OBJ_MODEL_ENUM"

	obj_string_map[OBJ_PROTO_PACKAGE] = "OBJ_PROTO_PACKAGE"
	obj_string_map[OBJ_PROTO_SYNTAX] = "OBJ_PROTO_SYNTAX"
	obj_string_map[OBJ_PROTO_MSG] = "OBJ_PROTO_MSG"
	obj_string_map[OBJ_PROTO_ENUM] = "OBJ_PROTO_ENUM"
	obj_string_map[OBJ_PROTO_IDS] = "OBJ_PROTO_IDS"
}

func (this ObjectType) String()string{
	ret,ok:=obj_string_map[this]
	if ok {
		return ret
	}
	return "OBJ_UNDEFINED"
}

type GeneratorObject interface {
	Name() string
	File() GeneratorFile
	ObjectType() ObjectType
	GenerateToObject(typ ObjectType) GeneratorObject
	Lines() []*LineText
	GetLines(lineType LineType)[]*LineText
	CheckLine(line *LineText) bool
	AddLine(line *LineText, lineType LineType)*LineText
	TryAddLine(line *LineText, lineType LineType)bool
	RemoveLineType(lineType LineType)[]*LineText
	InsertLine(pos int, line *LineText)*LineText
	InsertAfter(line *LineText, after *LineText)*LineText
	String()string
}

var _ GeneratorObject = (*DefaultGeneratorObj)(nil)

type DefaultGeneratorObj struct {
	file       GeneratorFile
	objectType ObjectType
	lines      []*LineText
	state int
}

func (this *DefaultGeneratorObj) GetFile()GeneratorFile {
	return this.file
}

func (this *DefaultGeneratorObj) GetLogger()log.ILogger {
	return this.GetFile().GetLogger()
}

func (this *DefaultGeneratorObj) init(typ ObjectType,file GeneratorFile) {
	this.objectType = typ
	if file.GetFile()!=nil {
		this.file = file.GetFile()
	} else {
		this.file = file
	}

	this.lines = make([]*LineText, 0)
}

func (this *DefaultGeneratorObj) GetLines(lineType LineType)[]*LineText{
	ret:=make([]*LineText,0)
	for _,line:=range this.lines {
		if line.LineType == lineType {
			ret = append(ret, line)
		}
	}
	return ret
}

func (this *DefaultGeneratorObj) File() GeneratorFile{
	return this.file
}

func (this *DefaultGeneratorObj) Name() string {
	return this.objectType.String()
}

func (this *DefaultGeneratorObj) ObjectType() ObjectType {
	return this.objectType
}

func (this *DefaultGeneratorObj) GenerateToObject(typ ObjectType) GeneratorObject {
	return nil
}

func (this *DefaultGeneratorObj) RemoveLineType(lineType LineType)[]*LineText{
	result:=make([]*LineText,0)
	ret:=make([]*LineText,0)
	for _,obj:=range this.lines {
		if obj.LineType== lineType {
			ret = append(ret, obj)
			continue
		}
		result = append(result, obj)
	}
	this.lines = result
	return ret
}

func (this *DefaultGeneratorObj) String()string{
	ret:="\n"
	for _,line:=range this.Lines() {
		ret+=line.Parse()
		ret+="\n"
	}
	return ret
}

func (this *DefaultGeneratorObj) Lines() []*LineText {
	return this.lines
}

func (this *DefaultGeneratorObj) CheckLine(line *LineText) bool {
	return false
}

func (this *DefaultGeneratorObj) AddLine(line *LineText, lineType LineType)*LineText  {
	line.LineType = lineType
	line.Obj = this
	this.lines = append(this.lines, line)
	return line
}


func (this *DefaultGeneratorObj) InsertLine(pos int, line *LineText)*LineText {
	linesCopy:=make([]*LineText,len(this.lines))
	copy(linesCopy,this.lines)
	rPart := linesCopy[pos:]
	this.lines = this.lines[0:pos]
	this.lines = append(this.lines, line)
	for _, iter := range rPart {
		this.lines = append(this.lines, iter)
	}
	return line
}

func (this *DefaultGeneratorObj) InsertAfter(line *LineText, after *LineText)*LineText {
	for index, iter := range this.lines {
		if after == iter {
			this.InsertLine(index+1, line)
			return line
		}
	}
	this.GetLogger().Panic("cant found line")
	return nil
}


func (this *DefaultGeneratorObj) TryAddLine(line *LineText, lineType LineType)bool {
	removeComment := COMMENT_REGEXP.ReplaceAllString(line.Text, "$1")
	comment := COMMENT_REGEXP.ReplaceAllString(line.Text,"$2")
	if removeComment == comment {
		comment = ""
	}
	if lineType.RegMatch(removeComment) {
		line.Comment = comment
		this.AddLine(line,lineType)
		log.WithFields(log.Fields{
			"lineType":    line.LineType.String(),
			"lineNum": line.LineNum,
			"text": line.Text,
			"comment": line.Comment,
		}).Info("find "+line.ObjName())
		return true
	}
	return false
}

func (this *DefaultGeneratorObj) TryAddLineStrict(line *LineText, lineType LineType)bool {
	if lineType.RegMatch(line.Text) {
		this.AddLine(line,lineType)
		log.WithFields(log.Fields{
			"lineType":    line.LineType.String(),
			"lineNum": line.LineNum,
			"text": line.Text,
		}).Info("find "+line.ObjName())
		return true
	}
	return false
}

type EmptyObject struct {
	DefaultGeneratorObj
}

func NewEmptyObject(file GeneratorFile) *EmptyObject {
	ret := &EmptyObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_EMPTY, file)
	return ret
}

func (this *EmptyObject) CheckLine(line *LineText) bool {
	if this.TryAddLineStrict(line, LINE_EMPTY) {
		return true
	}
	return false
}

type CommentObject struct {
	DefaultGeneratorObj
	state int
}

func NewCommentObject(file GeneratorFile) *CommentObject {
	ret := &CommentObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_COMMENT, file)
	return ret
}

func (this *CommentObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLineStrict(line, LINE_COMMENT) {
			this.state = 2
			return true
		}
		if this.TryAddLineStrict(line, LINE_COMMENT_START) {
			this.state = 1
			return true
		}
		return false
	} else if this.state == 1 {
		if this.TryAddLineStrict(line, LINE_COMMENT_END) {
			this.state = 2
			return true
		}
		if this.TryAddLineStrict(line, LINE_ANY) {
			return true
		}
		this.GetLogger().Panic("parse CommentObject Body error")
	} else if this.state == 2 {
		return false
	}
	this.GetLogger().Panic("parse CommentObject error")
	return false
}




func (this ObjectType) Create(file GeneratorFile)GeneratorObject  {
	switch this {
	case OBJ_GO_PACKAGE:
		return NewGoPackageObject(file)
	case OBJ_GO_IMPORTS:
		return NewGoImportObject(file)
	case OBJ_GO_INTERFACE:
		return NewGoInterfaceObject(file)
	case OBJ_GO_STRUCT:
		return NewGoStructObject(file)
	case OBJ_GO_STRUCT_FUNC:
		return NewGoStructFuncObject(file)
	case OBJ_GO_FUNC:
		return NewGoFuncObject(file)
	case OBJ_GO_VAR:
		return NewGoVarObject(file)
	case OBJ_GO_ENUM:
		return NewGoEnumObject(file)
	case OBJ_GO_ID:
		return NewGoIdObject(file)
	case OBJ_GO_ID_REG:
		return NewGoIdRegObject(file)
	case OBJ_GO_DEFINE:
		return NewGoDefineObject(file)
	case OBJ_CONF:
		return NewConfObject(file)

	case OBJ_EMPTY:
		return NewEmptyObject(file)
	case OBJ_COMMENT:
		return NewCommentObject(file)

	case OBJ_TS_CLASS:
		return NewTsClassObject(file)
	case OBJ_TS_IMPORTS:
		return NewTsImportObject(file)
	case OBJ_TS_VAR:
		return NewTsVarObject(file)
	case OBJ_TS_FUNC:
		return NewTsFuncObject(file)
	case OBJ_TS_ENUM:
		return NewTsEnumObject(file)
	case OBJ_TS_IDS:
		return NewTsIdsObject(file)

	case OBJ_MODEL_PACKAGE:
		return NewModelPackageObject(file)
	case OBJ_MODEL_IMPORTS:
		return NewModelImportObject(file)
	case OBJ_MODEL_ERRORS:
		return NewModelErrorsObject(file)
	case OBJ_MODEL_IDS:
		return NewModelIdsObject(file)
	case OBJ_MODEL_CLASS:
		return NewModelClassObject(file)
	case OBJ_MODEL_ENUM:
		return NewModelEnumObject(file)

	case OBJ_PROTO_MSG:
		return NewProtoMsgObject(file)
	case OBJ_PROTO_ENUM:
		return NewProtoEnumObject(file)
	case OBJ_PROTO_PACKAGE:
		return NewProtoPackageObj(file)
	case OBJ_PROTO_SYNTAX:
		return NewProtoSyntaxObj(file)
	case OBJ_PROTO_IDS:
		return NewProtoIdsObj(file)
	default:
		log.Panicf("unregistered obj type",this,this.String())
	}
	return nil
}

