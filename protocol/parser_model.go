package protocol

import (
	"fmt"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas/util/stringutil"
	"github.com/nomos/promise"
	"sort"
	"strconv"
	"strings"
)

type ModelPackageObject struct {
	DefaultGeneratorObj
	PackageName   string
	GoPackageName string
	CsPackageName string
	TsPackageName string
	Ids map[BINARY_TAG]*ModelId
}

func NewModelPackageObject(file GeneratorFile) *ModelPackageObject {
	ret := &ModelPackageObject{DefaultGeneratorObj: DefaultGeneratorObj{},Ids: map[BINARY_TAG]*ModelId{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_PACKAGE, file)
	return ret
}

//TODO
func (this *ModelPackageObject) GoString()string {
	ret:=`//this is a generated file,do not modify it!!!
package {PackageName}

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"reflect"
)

const (
{Ids}
)

func init() {
{IdRegister}
}

{Protocols}
`
	ret = strings.Replace(ret,`{PackageName}`,this.GoPackageName,1)
	ret = strings.Replace(ret,`{Ids}`,this.GetGoIdAssignString(),1)
	ret = strings.Replace(ret,`{IdRegister}`,this.GetGoIdRegString(),1)
	ret = strings.Replace(ret,`{Protocols}`,this.GetGoFuncString(),1)
	return ret
}

func (this *ModelPackageObject) CsString()string {
	ret:=`//this is a generated file,do not modify it!!!
using System;
using System.Threading.Tasks;
using Funnel.Client;
using Funnel.Protocol;
using Funnel.Protocol.Abstractions;
#if UNITY_2017_1_OR_NEWER
    using UnityEngine;
#endif
using static Funnel.FunnelGlobal;

namespace {CsPackageName}
{
    public static class {CsClassName}
    {
{Protocols}

        private static Client _client = null;

        //方法注册

        public static void Init(Client client)
        {
            setClient(client);
            registerIds();
            registerMessages();
        }
        private static void setClient(Client client)
        {
            _client = client;
        }

        private static void registerMessages()
        {
{HandlerRegister}
        }

        private static void registerIds()
        {
{IdRegister}
        }
    }
}`
	ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackageName,1)
	ret = strings.Replace(ret,`{CsClassName}`,this.GetCsPackageName(),1)
	ret = strings.Replace(ret,`{Protocols}`,this.GetCsFuncString(),1)
	ret = strings.Replace(ret,`{HandlerRegister}`,this.GetCsMessageRegString(),1)
	ret = strings.Replace(ret,`{IdRegister}`,this.GetCsIdRegString(),1)
	return ret
}

func (this *ModelPackageObject) GetCsFuncString()string {
	ret := ""
	for _,id:=range this.Ids {
		if id.Type!="" {
			s:=id.GetCsProtocolFuncString()
			s+="\n"
			ret+=s
		}
	}
	return ret
}

func (this *ModelPackageObject) GetCsMessageRegString()string {
	ret := ""
	for _,id:=range this.Ids {
		if id.Type!="" {
			s:=id.GetCsMessageRegisterString()
			s+="\n"
			ret+=s
		}
	}
	return ret
}

func (this *ModelPackageObject) GetCsIdRegString()string {
	ret := ""
	for _,id:=range this.Ids {
		s:=id.GetCsIdRegisterString()
		s+="\n"
		ret+=s
	}
	return ret
}

func (this *ModelPackageObject) GetGoFuncString()string {
	ret := ""
	for _,id:=range this.Ids {
		if id.Type!="" {
			s:=id.GetGoProtocolFuncString()
			s+="\n"
			s+="\n"
			ret+=s
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelId) GetGoProtocolFuncString()string{
	switch this.Type {
	case "REQ":
		ret:=`type OnRequest{ClassB}Func func(avatar lokas.IEntityActor, actorId util.ID, transId uint32,req *{ClassB}) (*{ClassA},error)`
		ret = strings.ReplaceAll(ret,"{ClassB}",this.Name)
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Resp)
		return ret
	case "NTF":
		ret:=`type OnNotify{ClassA}Func func(avatar lokas.IEntityActor, actorId util.ID,notify *{ClassA})`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		return ret
	case "EVT":
		ret:=`type SendEvent{ClassA}Func func(evt *{ClassA})error`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		return ret
	default:
		return ""
	}
}

func (this *ModelPackageObject) GetGoIdAssignString()string {
	ret := ""
	for _,id:=range this.Ids {
		s:=id.GetGoIdAssignString()
		s+="\n"
		ret+=s
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelPackageObject) GetGoIdRegString()string {
	ret := ""
	for _,id:=range this.Ids {
		s:=id.GetGoIdRegisterString()
		s+="\n"
		ret+=s
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelPackageObject) GetCsPackageName()string {
	split := strings.Split(this.CsPackageName,".")
	className:=split[len(split)-1]
	return className
}

func (this *ModelPackageObject) GetGoPackageName()string {
	return this.GoPackageName
}

func (this *ModelPackageObject) CheckLine(line *LineText) bool {
	log.Warnf(line.Text)
	if this.TryAddLine(line, LINE_COMMENT) {
		return true
	}
	if this.TryAddLine(line, LINE_MODEL_PACKAGE) {
		this.PackageName = line.GetPkgName()
		return true
	}
	if this.TryAddLine(line, LINE_MODEL_GOPACKAGE) {
		this.GoPackageName = line.GetPkgName()
		return true
	}
	if this.TryAddLine(line, LINE_MODEL_CSPACKAGE) {
		this.CsPackageName = line.GetPkgName()
		return true
	}
	if this.TryAddLine(line, LINE_MODEL_TSPACKAGE) {
		this.TsPackageName = line.GetPkgName()
		return true
	}
	if this.GoPackageName == "" {
		this.GoPackageName = this.PackageName
	}
	if this.TsPackageName == "" {
		this.TsPackageName = this.PackageName
	}
	this.file.(*ModelFile).Package = this.PackageName
	this.file.(*ModelFile).GoPackage = this.GoPackageName
	this.file.(*ModelFile).TsPackage = this.TsPackageName
	this.file.(*ModelFile).CsPackage = this.CsPackageName

	return false
}

type ModelImportObject struct {
	DefaultGeneratorObj
	imports []string
}

func NewModelImportObject(file GeneratorFile) *ModelImportObject {
	ret := &ModelImportObject{DefaultGeneratorObj: DefaultGeneratorObj{}, imports: []string{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_IMPORTS, file)
	return ret
}

func (this *ModelImportObject) CheckLine(line *LineText) bool {
	if this.TryAddLine(line, LINE_MODEL_IMPORTS) {
		this.imports = append(this.imports, line.GetPkgName())
		return true
	}
	if this.TryAddLine(line, LINE_EMPTY) {
		return true
	}
	if this.TryAddLine(line, LINE_COMMENT) {
		return true
	}
	return false
}

type ModelId struct {
	Name string
	Id   int
	Type string
	Resp string
	PackageName string
	ClassObj *ModelClassObject
	RespClassObj *ModelClassObject
}

func (this *ModelId) GetCsProtocolFuncString()string{
	log.Warnf("this.Type",this.Type)
	switch this.Type {
	case "REQ":
		ret:=`		public static async Task<{ClassA}> Request{ClassB}({ClassB} msg)
		{
			return await _client.Request<{ClassA}, {ClassB}>(msg);
		}`
		ret = strings.ReplaceAll(ret,"{ClassB}",this.Name)
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Resp)
		return ret
	case "NTF":
		ret:=`		public static FunnelError Send{ClassA}({ClassA} msg)
        {
            return _client.SendMessage(msg);
        }`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		return ret
	case "EVT":
		ret:=`		public static Action<{ClassA}> OnEvent{ClassA} { get; set; }`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		return ret
	default:
		return ""
	}
}

func (this *ModelId) GetCsMessageRegisterString()string{
	switch this.Type {
	case "EVT":
		ret:=`			_client.RegisterHandler({MsgId}, ((sender, serializable) =>
            {
                if (OnEvent{ClassA} == null)
                {
                    throw ERR_UNHANDLED_MESSAGE;
                }

                OnEvent{ClassA}(serializable as {ClassA});
            }));`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		ret = strings.ReplaceAll(ret,"{MsgId}",strconv.Itoa(this.Id))
		return ret
	default:
		return ""
	}
}

func (this *ModelId) GetCsIdRegisterString()string {
	ret := `			FunnelSerializable.RegisterType(new {ClassA}(), {MsgId});`
	ret = strings.ReplaceAll(ret, "{ClassA}", this.Name)
	ret = strings.ReplaceAll(ret, "{MsgId}", strconv.Itoa(this.Id))
	return ret
}

func (this *ModelId) GetGoIdAssignString()string {
	ret:= "\t{TagName}  protocol.BINARY_TAG = {TagId}"
	ret = strings.ReplaceAll(ret,"{TagName}","TAG_"+stringutil.SplitCamelCaseUpperSlash(this.Name))
	ret = strings.ReplaceAll(ret,"{TagId}",strconv.Itoa(this.Id))
	return ret
}

func (this *ModelId) GetGoIdRegisterString()string {
	ret:= line_parse_map[LINE_GO_TAG_REGISTRY]
	ret = strings.ReplaceAll(ret,"{$type}",this.Name)
	ret = strings.ReplaceAll(ret,"{$name}","TAG_"+stringutil.SplitCamelCaseUpperSlash(this.Name))
	return ret
}

type ModelIdsObject struct {
	DefaultGeneratorObj
	state int
	PackageName string
	Ids   map[int]*ModelId
}

func NewModelIdsObject(file GeneratorFile) *ModelIdsObject {
	ret := &ModelIdsObject{DefaultGeneratorObj: DefaultGeneratorObj{}, Ids: map[int]*ModelId{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_IDS, file)
	return ret
}

func (this *ModelIdsObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_MODEL_IDS_HEADER) {
			this.state = 1
			return true
		}
		log.Panicf("parse ModelIdsObject Error", this.state)
		return false
	} else if this.state == 1 {
		if this.TryAddLine(line, LINE_MODEL_ID) {
			id:=line.GetValue()
			if id <=0 || id >65535 {
				log.Panicf("id must >0 and <65535",id)
			}
			p:=&ModelId{
				Name: line.GetName(),
				Id:   id,
				Type: "",
				Resp: "",
			}
			if line.GetTypeName() != "" {
				p.Type = strings.TrimSpace(line.GetTypeName())
			}
			if line.GetTagName() != "" {
				p.Resp = line.GetTagName()
			}
			this.Ids[line.GetValue()] = p
			return true
		}
		if this.TryAddLine(line, LINE_EMPTY) {
			return true
		}
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		return false
	}
	log.Panicf("parse NewModelIdsObject Error", this.state)
	return false
}

type ModelEnumObject struct {
	DefaultGeneratorObj
	state     int
	Package   string
	CsPackage string
	GoPackage string
	TsPackage string
	EnumName  string
}

func NewModelEnumObject(file GeneratorFile) *ModelEnumObject {
	ret := &ModelEnumObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_ENUM, file)
	return ret
}

func (this *ModelEnumObject) CsString()string{
	ret:=`//this is a generate file,do not modify it!
using Funnel.Protocol.Abstractions;

namespace {CsPackageName}
{
    public enum {EnumName}
    {
{ClassBody}
    }
}
`
	ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackage,1)
	ret = strings.Replace(ret,`{EnumName}`,stringutil.SplitCamelCaseUpperSlash(this.EnumName),1)
	ret = strings.Replace(ret,`{ClassBody}`,this.csFields(),1)
	return ret
}

func (this *ModelEnumObject) csFields()string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			ret+="\t\t"
			ret+=stringutil.SplitCamelCaseUpperSlash(l.Name)
			ret+= " = "
			ret+=strconv.Itoa(l.GetValue())
			ret+=","
			ret+="\n"
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelEnumObject) GoString()string{
	ret:=`//this is a generate file,do not modify it!
package {PackageName}

import "github.com/nomos/go-lokas/protocol"

type {EnumName} protocol.Enum

const (
{ClassBody}
)
`
	ret = strings.Replace(ret,`{PackageName}`,this.GoPackage,1)
	ret = strings.Replace(ret,`{EnumName}`,stringutil.SplitCamelCaseUpperSlash(this.EnumName),1)
	ret = strings.Replace(ret,`{ClassBody}`,this.goFields(),1)
	return ret
}

func (this *ModelEnumObject) goFields()string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			ret+="\t"
			ret+=stringutil.SplitCamelCaseUpperSlash(this.EnumName)
			ret+="_"
			ret+=stringutil.SplitCamelCaseUpperSlash(l.Name)
			ret+=" "
			ret+=stringutil.SplitCamelCaseUpperSlash(this.EnumName)
			ret+=" "
			ret+= " = "
			ret+=strconv.Itoa(l.GetValue())
			ret+="\n"
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelEnumObject) SetPackage(pack string) {
	this.Package = pack
}

func (this *ModelEnumObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_MODEL_ENUM_HEADER) {
			this.EnumName = line.GetStructName()
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
		if this.TryAddLine(line, LINE_MODEL_ENUM_FIELD) {
			if this.EnumName != "" {
				line.GetName()
				line.GetValue()
			}
			return true
		}
		return false
	}
	log.Panicf("parse ModelClassObject Error", this.state)
	return false
}

type ModelClassFields struct {
	Name    string
	Type    string
	Index   int
}

func (this *ModelClassFields) csString(lower bool)string {
	name:=this.Name
	if lower {
		name=stringutil.FirstToLower(this.Name)
	}
	ret:=""
	t:= MatchModelProtoTag(this.Type)
	if t!=0 {
		t:=GetModelProtoTag(this.Type)
		ret = t.CsTypeString()+" "+name
	} else if t,s1,s2:=MatchModelSystemTag(this.Type);t!=0 {
		if t==TAG_List {
			t = MatchModelProtoTag(s1)
			if t!=0 {
				ret = "List<"+t.CsTypeString()+"> "+name
			} else {
				ret = "List<"+s1+"> "+name
			}
		} else if t==TAG_Map {
			t1 := MatchModelProtoTag(s1)
			t2 := MatchModelProtoTag(s2)
			type1 := s1
			if t1!=0 {
				type1 = t1.CsTypeString()
			}
			type2 := s2
			if t2!=0 {
				type2 = t2.CsTypeString()
			}
			ret =  "Dictionary<"+type1+","+type2+"> "+name

		}
	} else {
		ret = this.Type+" "+name
	}
	return ret

}

func (this *ModelClassFields) CsString()string {
	return "\t\tpublic "+this.csString(false)+"{ get;set; }"
}


func (this *ModelClassFields) ParamString()string {
	return this.csString(true)
}


func (this *ModelClassFields) ParamAssignString()string {
	return "\t\t\t"+this.Name+" = "+stringutil.FirstToLower(this.Name)+";"
}

func (this *ModelClassFields) GoString()string {
	ret:=""
	t:= MatchModelProtoTag(this.Type)
	if t!=0 {
		t:=GetModelProtoTag(this.Type)
		ret = "\t"+this.Name+" "+t.GoTypeString()
	} else if t,s1,s2:=MatchModelSystemTag(this.Type);t!=0 {
		if t==TAG_List {
			t = MatchModelProtoTag(s1)
			if t!=0 {
				ret = "\t"+this.Name+" []"+t.GoTypeString()
			} else {
				ret = "\t"+this.Name+" []*"+s1
			}
		} else if t==TAG_Map {
			t1 := MatchModelProtoTag(s1)
			t2 := MatchModelProtoTag(s2)
			type1 := "*"+s1
			if t1!=0 {
				type1 = t1.GoTypeString()
			}
			type2 := "*"+s2
			if t2!=0 {
				type2 = t2.GoTypeString()
			}
			ret =  "\t"+this.Name+" map["+type1+"]"+type2

		}
	} else {
		ret = "\t"+this.Name+" *"+this.Type
	}
	return ret
}

type ModelClassObject struct {
	DefaultGeneratorObj
	TagId     BINARY_TAG
	state     int
	Fields    []*ModelClassFields
	Package   string
	CsPackage string
	GoPackage string
	TsPackage string
	ClassName string
}

func NewModelClassObject(file GeneratorFile) *ModelClassObject {
	ret := &ModelClassObject{DefaultGeneratorObj: DefaultGeneratorObj{}, Fields: []*ModelClassFields{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_CLASS, file)
	return ret
}

func (this *ModelClassObject) CsString()string{
	ret:=`//this is a generate file,do not modify it!
using System.Collections.Generic;
using Funnel.Protocol.Abstractions;

namespace {CsPackageName}
{
    public class {ClassName}:FunnelSerializable
    {
{ClassBody}


		public {ClassName}() {

		}

		public {ClassName}({CsParams}) {
{ParamAssign}
		}
    }
}
`

	if len(this.Fields) == 0 {
		ret = `//this is a generate file,do not modify it!
using System.Collections.Generic;
using Funnel.Protocol.Abstractions;

namespace {CsPackageName}
{
    public class {ClassName}:FunnelSerializable
    {
{ClassBody}
    }
}
`
		ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackage,-1)
		ret = strings.Replace(ret,`{ClassName}`,this.ClassName,-1)
		ret = strings.Replace(ret,`{ClassBody}`,this.csFields(),-1)
	} else {
		ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackage,-1)
		ret = strings.Replace(ret,`{ClassName}`,this.ClassName,-1)
		ret = strings.Replace(ret,`{ClassBody}`,this.csFields(),-1)
		ret = strings.Replace(ret,`{CsParams}`,this.csParams(),-1)
		ret = strings.Replace(ret,`{ParamAssign}`,this.csParamAssign(),-1)
	}
	return ret
}

func (this *ModelClassObject) csFields()string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.CsString()
		ret+="\n"
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelClassObject) csParams()string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.ParamString()
		ret+=","
	}
	ret = strings.TrimRight(ret,",")
	return ret
}

func (this *ModelClassObject) csParamAssign()string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.ParamAssignString()
		ret+="\n"
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelClassObject) GoImplString()string{
	ret:=`//this is a generate file,edit implement on this file only!
package {PackageName}

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

func New{ClassName}()*{ClassName}{
	ret:=&{ClassName}{
	}
	return ret
}

func (this *{ClassName}) OnAdd(e lokas.IEntity, r lokas.IRuntime) {
	
}

func (this *{ClassName}) OnRemove(e lokas.IEntity, r lokas.IRuntime) {
	
}

func (this *{ClassName}) OnCreate(r lokas.IRuntime) {
	
}

func (this *{ClassName}) OnDestroy(r lokas.IRuntime) {
	
}`
	ret = strings.Replace(ret,`{PackageName}`,this.GoPackage,-1)
	ret = strings.Replace(ret,`{ClassName}`,this.ClassName,-1)
	return ret
}

func (this *ModelClassObject) GoString()string{
	ret:=`//this is a generate file,do not modify it!
package {PackageName}

import (
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

var _ lokas.IComponent = (*{EnumName})(nil)

type {EnumName} struct {
	ecs.Component `+"`json:\"-\"`"+`
{ClassBody}
}

func (this *{EnumName}) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *{EnumName}) Serializable()protocol.ISerializable {
	return this
}
`
	ret = strings.Replace(ret,`{PackageName}`,this.GoPackage,-1)
	ret = strings.Replace(ret,`{EnumName}`,this.ClassName,-1)
	ret = strings.Replace(ret,`{ClassBody}`,this.goFields(),-1)
	return ret
}

func (this *ModelClassObject) goFields()string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.GoString()
		ret+="\n"
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelClassObject) SetPackage(pack string) {
	this.Package = pack
}

func (this *ModelClassObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_MODEL_CLASS_HEADER) {
			this.ClassName = line.GetStructName()
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
		if this.TryAddLine(line, LINE_COMMENT) {
			return true
		}
		if this.TryAddLine(line, LINE_MODEL_CLASS_FIELD) {
			if this.ClassName != "" {
				fieldName := line.GetName()
				fieldType := line.GetTypeName()
				this.Fields = append(this.Fields, &ModelClassFields{
					Name:    fieldName,
					Type:    fieldType,
					Index:   len(this.Fields),
				})
			}
			return true
		}
		return false
	}
	log.Panicf("parse ModelClassObject Error", this.state)
	return false
}

var _ GeneratorFile = &ModelFile{}

func NewModelFile(generator *Generator) *ModelFile {
	ret := &ModelFile{DefaultGeneratorFile: NewGeneratorFile(generator)}
	ret.GeneratorFile = ret
	ret.FileType = FILE_MODEL
	return ret
}

type ModelFile struct {
	*DefaultGeneratorFile
	Package string
	GoPackage string
	CsPackage string
	TsPackage string
}

func (this *ModelFile) Generate() *promise.Promise {
	return nil
}

func (this *ModelFile) Parse() *promise.Promise {
	return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
		offset, success := this.parse(0, OBJ_MODEL_PACKAGE)
		log.Infof("parseModelPackage", offset, success)
		offset, success = this.parse(offset, OBJ_MODEL_IMPORTS)
		log.Infof("parseModelImports", offset, success)
		offset, success = this.parse(offset, OBJ_MODEL_IDS)
		log.Infof("parseModelIds", offset, success)
		offset, success = this.parse(offset, OBJ_MODEL_CLASS, OBJ_MODEL_ENUM)
		log.Warnf("isFinish", len(this.Lines), offset)
		if !this.CheckFinish(offset) {
			reject(fmt.Sprintf("file not finish %d", offset))
			return
		}
		log.Infof("parseModelClass", offset, success)
		log.Infof("parseModel finish", offset, success)
		resolve(nil)
	})
}

func (this *ModelFile) ProcessModels() []*ModelClassObject {
	ret := make([]*ModelClassObject, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_CLASS {
			o:=obj.(*ModelClassObject)
			o.Package = this.Package
			o.TsPackage = this.TsPackage
			o.GoPackage = this.GoPackage
			o.CsPackage = this.CsPackage
			ret = append(ret,o)
		}
	}
	return ret
}

func (this *ModelFile) ProcessEnums() []*ModelEnumObject {
	ret := make([]*ModelEnumObject, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_ENUM {
			o:=obj.(*ModelEnumObject)
			o.Package = this.Package
			o.TsPackage = this.TsPackage
			o.GoPackage = this.GoPackage
			o.CsPackage = this.CsPackage
			ret = append(ret,o)
		}
	}
	return ret
}

func (this *ModelFile) ProcessIds() []*ModelId {
	ret := make([]*ModelId, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_IDS {
			for _,v:=range obj.(*ModelIdsObject).Ids {
				ret = append(ret, v)
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Id<ret[j].Id
	})
	return ret
}

func (this *ModelFile) ProcessPackages() *ModelPackageObject {
	var ret *ModelPackageObject = nil
	defer func() {
		r:=recover()
		if r!=nil {
			log.Errorf(r)
		}
	}()
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_PACKAGE {
			log.Warnf("this.Package",this.Package)
			obj.(*ModelPackageObject).PackageName = this.Package
			ret =  obj.(*ModelPackageObject)
		}
		if obj.ObjectType() ==  OBJ_MODEL_IDS{
			for _,id:=range obj.(*ModelIdsObject).Ids {
				id.PackageName = this.Package
			}
		}
	}
	log.Warnf("ProcessPackages")
	return ret
}
