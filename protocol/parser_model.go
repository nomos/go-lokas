package protocol

import (
	"fmt"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/promise"
	"github.com/nomos/go-lokas/util/stringutil"
	"regexp"
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
	Imports map[string]*ModelPackageObject
	Ids map[BINARY_TAG]*ModelId
	Errors map[int]*ModelError
}

func NewModelPackageObject(file GeneratorFile) *ModelPackageObject {
	ret := &ModelPackageObject{
		DefaultGeneratorObj: DefaultGeneratorObj{},
		Ids: map[BINARY_TAG]*ModelId{},
		Errors: map[int]*ModelError{},
	}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_PACKAGE, file)
	return ret
}

//TODO
func (this *ModelPackageObject) GoString(g *Generator)string {
	ret0:=`//this is a generated file,do not modify it!!!
package {PackageName}

import (
	"github.com/nomos/go-lokas/protocol"{Imports}
	"reflect"
)

const (
{Ids}
)
{Errors}
func init() {
{IdRegister}
}

{Protocols}
`

	ret:=`//this is a generated file,do not modify it!!!
package {PackageName}

import ({Imports}
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"reflect"
)

const (
{Ids}
)
{Errors}
func init() {
{IdRegister}
}

{Protocols}
`
	funcStr:=this.GetGoFuncString(g)
	if funcStr == "" {
		ret = ret0
	}
	deps:=[]string{}
	for _,v:=range this.Ids {
		for _,d:=range v.Deps(g) {
			deps = append(deps, d)
		}
	}
	if len(deps)>0 {
		ret = strings.Replace(ret,`{Imports}`,g.getGoImportsString(deps),-1)
	} else {
		ret = strings.Replace(ret,`{Imports}`,"",-1)
	}
	ret = strings.Replace(ret,`{Errors}`,this.GoErrorString(g),-1)
	ret = strings.Replace(ret,`{PackageName}`,this.PackageName,-1)
	ret = strings.Replace(ret,`{Ids}`,this.GetGoIdAssignString(g),-1)
	ret = strings.Replace(ret,`{IdRegister}`,this.GetGoIdRegString(g),-1)
	ret = strings.Replace(ret,`{Protocols}`,this.GetGoFuncString(g),-1)
	ret = strings.ReplaceAll(ret,"\r","")
	return ret
}

func (this *ModelPackageObject) CsString(g *Generator)string {
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

        private static Funnel.Client.Client _client = null;
		//错误注册
{Errors}
        //方法注册
        public static void Init(Funnel.Client.Client client)
        {
            setClient(client);
            registerIds();
            registerMessages();
        }
        private static void setClient(Funnel.Client.Client client)
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
	ret = strings.Replace(ret,`{Errors}`,this.CsErrorString(g),-1)
	ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackageName,-1)
	ret = strings.Replace(ret,`{CsClassName}`,this.GetCsPackageName(),-1)
	ret = strings.Replace(ret,`{Protocols}`,this.GetCsFuncString(g),-1)
	ret = strings.Replace(ret,`{HandlerRegister}`,this.GetCsMessageRegString(g),-1)
	ret = strings.Replace(ret,`{IdRegister}`,this.GetCsIdRegString(g),-1)
	return ret
}

func (this *ModelPackageObject) TsErrorString(g *Generator)string{
	if len(this.Errors)==0 {
		return ""
	}
	ret:=`
var (
`
	ids:=make([]*ModelError,0)
	for _,v:=range this.Errors {
		ids = append(ids, v)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].ErrorId<ids[j].ErrorId
	})
	for _,id:=range ids {
		ret+="\t"
		ret+=id.GoString(g)
		ret+="\n"
	}
	ret+=`)
`
	return ret
}

func (this *ModelPackageObject) GoErrorString(g *Generator)string{
	if len(this.Errors)==0 {
		return ""
	}
	ret:=`
var (
`
	ids:=make([]*ModelError,0)
	for _,v:=range this.Errors {
		ids = append(ids, v)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].ErrorId<ids[j].ErrorId
	})
	for _,id:=range ids {
		ret+="\t"
		ret+=id.GoString(g)
		ret+="\n"
	}
	ret+=`)
`
	return ret
}

func (this *ModelPackageObject) CsErrorString(g *Generator)string{
	if len(this.Errors)==0 {
		return ""
	}
	ret:=`
`
	ids:=make([]*ModelError,0)
	for _,v:=range this.Errors {
		ids = append(ids, v)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].ErrorId<ids[j].ErrorId
	})
	for _,id:=range ids {
		ret+="\t\t"
		ret+=id.CsString(g)
		ret+="\n"
	}
	ret+=`
`
	return ret
}

func (this *ModelPackageObject) TsString(g *Generator)string {
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
		//错误注册
{Errors}
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
	ret = strings.Replace(ret,`{Errors}`,this.CsErrorString(g),-1)
	ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackageName,-1)
	ret = strings.Replace(ret,`{CsClassName}`,this.GetCsPackageName(),-1)
	ret = strings.Replace(ret,`{Protocols}`,this.GetCsFuncString(g),-1)
	ret = strings.Replace(ret,`{HandlerRegister}`,this.GetCsMessageRegString(g),-1)
	ret = strings.Replace(ret,`{IdRegister}`,this.GetCsIdRegString(g),-1)
	return ret
}

func (this *ModelPackageObject) GetCsFuncString(g *Generator)string {
	ret := ""
	ids:=make([]*ModelId,0)
	for _,id:=range this.Ids {
		ids= append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Id<ids[j].Id
	})
	for _,id:=range ids {
		if id.Type!="" {
			s:=id.GetCsProtocolFuncString(g)
			s+="\n"
			ret+=s
		}
	}
	ret = strings.ReplaceAll(ret,"\r","")
	return ret
}

func (this *ModelPackageObject) GetCsMessageRegString(g *Generator)string {
	ret := ""
	ids:=make([]*ModelId,0)
	for _,id:=range this.Ids {
		ids= append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Id<ids[j].Id
	})
	for _,id:=range ids {
		if id.Type!="" {
			s:=id.GetCsMessageRegisterString(g)
			if s!="" {
				s+="\n"
				ret+=s
			}
		}
	}
	return ret
}

func (this *ModelPackageObject) GetCsIdRegString(g *Generator)string {
	ret := ""
	ids:=make([]*ModelId,0)
	for _,v:=range this.Ids {
		ids = append(ids, v)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Id<ids[j].Id
	})
	for _,id:=range ids {
		s:=id.GetCsIdRegisterString(g)
		s+="\n"
		ret+=s
	}
	return ret
}

func (this *ModelPackageObject) GetGoFuncString(g *Generator)string {
	ret := ""
	idArr:=make([]*ModelId,0)
	for _,v:=range this.Ids {
		idArr = append(idArr, v)
	}
	sort.Slice(idArr, func(i, j int) bool {
		return idArr[i].Id<idArr[j].Id
	})
	for _,id:=range idArr {
		if id.Type!="" {
			s:=id.GetGoProtocolFuncString(g)
			s+="\n"
			s+="\n"
			ret+=s
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelPackageObject) GetGoIdAssignString(g *Generator)string {
	ret := ""
	idArr:=make([]*ModelId,0)
	for _,v:=range this.Ids {
		idArr = append(idArr, v)
	}
	sort.Slice(idArr, func(i, j int) bool {
		return idArr[i].Id<idArr[j].Id
	})
	for _,id:=range idArr {
		s:=id.GetGoIdAssignString(g)
		s+="\n"
		ret+=s
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelPackageObject) GetGoIdRegString(g *Generator)string {
	ret := ""
	idArr:=make([]*ModelId,0)
	for _,v:=range this.Ids {
		idArr = append(idArr, v)
	}
	sort.Slice(idArr, func(i, j int) bool {
		return idArr[i].Id<idArr[j].Id
	})
	for _,id:=range idArr {
		s:=id.GetGoIdRegisterString(g)
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
	log.Info(line.Text)
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
	this.file.(*ModelFile).Package = this.PackageName
	this.file.(*ModelFile).GoPackage = this.GoPackageName
	this.file.(*ModelFile).TsPackage = this.TsPackageName
	this.file.(*ModelFile).CsPackage = this.CsPackageName

	return false
}

type ModelError struct {
	ErrorName string
	ErrorId int
	ErrorText string
	PackageName string
}


func (this *ModelError) TsString(g *Generator)string{
	ret := "\texport const "+"ERR_"+stringutil.SplitCamelCaseUpperSnake(this.ErrorName)+` = new ErrMsg(`+strconv.Itoa(this.ErrorId)+`,"`+this.ErrorText+`")`
	return ret
}


func (this *ModelError) GoString(g *Generator)string{
	ret := `{ErrorName}     = protocol.CreateError({Id}, "{Text}")`
	ret  = strings.Replace(ret,"{ErrorName}","ERR_"+stringutil.SplitCamelCaseUpperSnake(this.ErrorName),-1)
	ret  = strings.Replace(ret,"{Id}",strconv.Itoa(this.ErrorId),-1)
	ret  = strings.Replace(ret,"{Text}",this.ErrorText,-1)
	return ret
}

func (this *ModelError) CsString(g *Generator)string{
	ret := `public static readonly FunnelServerError {ErrorName} = NewServerErrorMsg({Id}, "{Text}");`
	ret  = strings.Replace(ret,"{ErrorName}","ERR_"+stringutil.SplitCamelCaseUpperSnake(this.ErrorName),-1)
	ret  = strings.Replace(ret,"{Id}",strconv.Itoa(this.ErrorId),-1)
	ret  = strings.Replace(ret,"{Text}",this.ErrorText,-1)
	ret = strings.ReplaceAll(ret,"\r","")
	ret = strings.ReplaceAll(ret,"\n","")
	return ret
}

type ModelErrorsObject struct {
	DefaultGeneratorObj
	PackageName   string
	GoPackageName string
	CsPackageName string
	TsPackageName string
	Errors map[int]*ModelError
}

func NewModelErrorsObject(file GeneratorFile) *ModelErrorsObject {
	ret := &ModelErrorsObject{DefaultGeneratorObj: DefaultGeneratorObj{},Errors: map[int]*ModelError{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_ERRORS, file)
	return ret
}

func (this *ModelErrorsObject) CheckLine(line *LineText) bool {
	if this.state == 0 {
		if this.TryAddLine(line, LINE_MODEL_ERRORS_HEADER) {
			this.state = 1
			return true
		}
		log.Infof("parse ModelErrorsObject Error", this.state,line.Text,line.LineNum)
		return false
	} else if this.state == 1 {
		if this.TryAddLine(line, LINE_MODEL_ERROR) {
			p:=&ModelError{
				ErrorName:   line.GetName(),
				ErrorId:     line.GetValue(),
				ErrorText:   strings.TrimRight(strings.TrimRight(line.GetTagName()," "),"\n"),
				PackageName: "",
			}
			p.ErrorText = strings.ReplaceAll(p.ErrorText,"\n","")
			this.Errors[line.GetValue()] = p
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
	log.Panicf("parse ModelErrorsObject Error", this.state)
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
	Comment string
	PackageName string
	ClassObj *ModelClassObject
	RespClassObj *ModelClassObject
}

func (this *ModelId) Deps(g *Generator)[]string{
	ret:= []string{}
	if this.ClassObj!=nil&&this.ClassObj.Package!=this.PackageName {
		ret = append(ret, this.ClassObj.Package)
	}
	if this.RespClassObj!=nil&&this.RespClassObj.Package!=this.PackageName {
		ret = append(ret, this.RespClassObj.Package)
	}
	return ret
}

func (this *ModelId) GetGoProtocolFuncString(g *Generator)string{
	ret := ""
	switch this.Type {
	case "REQ":
		ret+="//"
		if this.Comment!= "" {
			ret+="-----"
			ret+=strings.ReplaceAll(this.Comment,"//","")
			ret+="-----"
		}
		ret+="     [{ClassB}]<---->[{ClassA}]"
		ret+="\n"
		ret+=`func Request{ClassB}(c lokas.INetClient,req *{ClassB})(*{ClassA},error){
	resp,err:=c.Request(req).Await()
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	return resp.(*{ClassA}),nil
}

type OnRequest{ClassB}Func func(data *DataMap,avatar lokas.IActor, actorId util.ID, transId uint32,req *{ClassB}) (*{ClassA},error)

func Register{ClassB}(f OnRequest{ClassB}Func,r func(protocol.BINARY_TAG,func(data *DataMap,avatar lokas.IActor,actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error))){
	r({TAG}, func(data *DataMap, avatar lokas.IActor, actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error) {
		return f(data,avatar,actorId,transId,msg.(*{ClassB}))
	})
}`
		ret = strings.ReplaceAll(ret,"{TAG}","TAG_"+stringutil.SplitCamelCaseUpperSnake(this.Name))
		ret = strings.ReplaceAll(ret,"{ClassB}",this.Name)
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Resp)
		return ret
	case "NTF":
		ret+="//"
		if this.Comment!= "" {
			ret+="-----"
			ret+=strings.ReplaceAll(this.Comment,"//","")
			ret+="-----"
		}
		ret+="//     [{ClassB}]----->"
		ret+="\n"
		ret+=`type OnNotify{ClassB}Func func(data *DataMap,avatar lokas.IActor, actorId util.ID,notify *{ClassB})

func Register{ClassB}(f OnNotify{ClassB}Func,r func(protocol.BINARY_TAG,func(data *DataMap,avatar lokas.IActor,actorId util.ID, msg protocol.ISerializable))){
	r({TAG}, func(data *DataMap, avatar lokas.IActor, actorId util.ID, msg protocol.ISerializable) {
		f(data,avatar,actorId,msg.(*{ClassB}))
	})
}`
		ret = strings.ReplaceAll(ret,"{TAG}","TAG_"+stringutil.SplitCamelCaseUpperSnake(this.Name))
		ret = strings.ReplaceAll(ret,"{ClassB}",this.Name)
		return ret
	case "EVT":
		ret+="//"
		if this.Comment!= "" {
			ret+="-----"
			ret+=strings.ReplaceAll(this.Comment,"//","")
			ret+="(客户端)"
			ret+="-----"
		}
		ret+="\n"
		ret+=`type OnEvent{ClassB} func(c lokas.IEntityNetClient,event *{ClassB})error

func RegisterOnEvent{ClassB}(f func(c lokas.IEntityNetClient,event *{ClassB})error,r  func(protocol.BINARY_TAG,func(lokas.IEntityNetClient,protocol.ISerializable))){
	r({TAG}, func(client lokas.IEntityNetClient, serializable protocol.ISerializable) {
		f(client,serializable.(*{ClassB}))
	})
}`
		ret = strings.ReplaceAll(ret,"{TAG}","TAG_"+stringutil.SplitCamelCaseUpperSnake(this.Name))
		ret = strings.ReplaceAll(ret,"{ClassB}",this.Name)
		return ret
	default:
		return ""
	}
}

func (this *ModelId) GetCsProtocolFuncString(g *Generator)string{
	switch this.Type {
	case "REQ":
		comment:=strings.ReplaceAll(this.Comment,`//`,"")
		errReg:=regexp.MustCompile(`\{\s*([0-9|A-z]+[\s|\,]*)+\}`)
		errcodesString:=errReg.FindString(this.Comment)
		errCodes:=regexp.MustCompile(`[0-9|A-z]+`).FindAllString(errcodesString,-1)
		comment = errReg.ReplaceAllString(comment,"")
		ret:=`		///<summary>`+comment+"</summary>\n"
		for _,v:=range errCodes {
			if regexp.MustCompile(`[0-9]+`).FindString(v) == v {
				code,err:=strconv.Atoi(v)
				if err != nil {
					log.Panic(err.Error())
					return ""
				}
				codeName:=g.GetErrorName(code)
				if codeName!="" {
					ret+=`		/// <exception cref="Exception">`+codeName+`</exception>`+"\n"
				}
			} else {
				log.Infof("IsErrorName",v,g.IsErrorName(v))
				if g.IsErrorName(v) {
					ret+=`		/// <exception cref="Exception">`+v+`</exception>`+"\n"
				}
			}

		}
		ret+=`		public static async Task<{ClassA}> Request{ClassB}({ClassB} msg)
		{
			return await _client.Request<{ClassA}, {ClassB}>(msg);
		}`
		ret = strings.ReplaceAll(ret,"{ClassB}",this.Name)
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Resp)
		return ret
	case "NTF":
		comment:=strings.ReplaceAll(this.Comment,`//`,"")
		ret:=`		///<summary>`+comment+"</summary>\n"
		ret+=`		public static FunnelError Send{ClassA}({ClassA} msg)
        {
            return _client.SendMessage(msg);
        }`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		return ret
	case "EVT":
		comment:=strings.ReplaceAll(this.Comment,`//`,"")
		ret:=`		///<summary>`+comment+"</summary>\n"
		ret+=`		public static Action<{ClassA}> OnEvent{ClassA} { get; set; }`
		ret = strings.ReplaceAll(ret,"{ClassA}",this.Name)
		return ret
	default:
		return ""
	}
}

func (this *ModelId) GetCsMessageRegisterString(g *Generator)string{
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

func (this *ModelId) GetCsIdRegisterString(g *Generator)string {
	ret := `			FunnelSerializable.RegisterType(new {ClassA}(), {MsgId});`
	ret = strings.ReplaceAll(ret, "{ClassA}", this.Name)
	ret = strings.ReplaceAll(ret, "{MsgId}", strconv.Itoa(this.Id))
	return ret
}

func (this *ModelId) GetGoIdAssignString(g *Generator)string {
	ret:= "\t{TagName}  protocol.BINARY_TAG = {TagId}"
	ret = strings.ReplaceAll(ret,"{TagName}","TAG_"+stringutil.SplitCamelCaseUpperSnake(this.Name))
	ret = strings.ReplaceAll(ret,"{TagId}",strconv.Itoa(this.Id))
	return ret
}

func (this *ModelId) GetGoIdRegisterString(g *Generator)string {
	ret:= line_parse_map[LINE_GO_TAG_REGISTRY]
	ret = strings.ReplaceAll(ret,"{$type}",this.Name)
	ret = strings.ReplaceAll(ret,"{$name}","TAG_"+stringutil.SplitCamelCaseUpperSnake(this.Name))
	return ret
}

type ModelIdsObject struct {
	DefaultGeneratorObj
	state int
	PackageName string
	Ids   map[int]*ModelId
}

func (this *ModelIdsObject) Deps(g *Generator)[]string{
	ret:=[]string{}
	for _,v:=range this.Ids {
		for _,d:=range v.Deps(g) {
			ret = append(ret, d)
		}
	}
	return ret
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
		log.Infof("parse ModelIdsObject Error", this.state,line.Text,line.LineNum)
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
				Comment: line.Comment,
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
	Comment string
}

func NewModelEnumObject(file GeneratorFile) *ModelEnumObject {
	ret := &ModelEnumObject{DefaultGeneratorObj: DefaultGeneratorObj{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_ENUM, file)
	return ret
}

func (this *ModelEnumObject) CsString(g *Generator)string{
	ret:=`//this is a generate file,do not modify it!
using Funnel.Protocol.Abstractions;
{Comment}
namespace {CsPackageName}
{
    public enum {EnumName}
    {
{ClassBody}
    }
}
`
	if this.Comment!="" {
		comment:="\n"+this.Comment
		ret = strings.Replace(ret,`{Comment}`,comment,-1)
	} else {
		ret = strings.Replace(ret,`{Comment}`,"",-1)
	}
	ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackage,1)
	ret = strings.Replace(ret,`{EnumName}`,stringutil.SplitCamelCaseUpperSnake(this.EnumName),-1)
	ret = strings.Replace(ret,`{ClassBody}`,this.csFields(g),-1)
	return ret
}

func (this *ModelEnumObject) csFields(g *Generator)string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			ret+="\t\t"
			ret+=stringutil.SplitCamelCaseUpperSnake(l.Name)
			ret+= " = "
			ret+=strconv.Itoa(l.GetValue())
			ret+=","
			ret+=" "
			ret+=l.Comment
			ret+="\n"
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelEnumObject) TsString(g *Generator)string{
	ret:=`export enum {EnumName} {
{ClassBody}
}

`
	if this.Comment!="" {
		comment:="\n"+this.Comment
		ret = strings.Replace(ret,`{Comment}`,comment,-1)
	} else {
		ret = strings.Replace(ret,`{Comment}`,"",-1)
	}
	ret = strings.Replace(ret,`{EnumName}`,stringutil.SplitCamelCaseUpperSnake(this.EnumName),-1)
	ret = strings.Replace(ret,`{ClassBody}`,this.tsFields(g),-1)
	return ret
}

func (this *ModelEnumObject) tsFields(g *Generator)string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			ret+="\t"
			ret+=stringutil.SplitCamelCaseUpperSnake(l.Name)
			ret+= " = "
			ret+=strconv.Itoa(l.GetValue())
			ret+=","
			ret+=" "
			ret+=l.Comment
			ret+="\n"
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelEnumObject) GoString(g *Generator)string{
	ret:=`//this is a generate file,do not modify it!
package {PackageName}

import (
	"github.com/nomos/go-lokas/protocol"
)

type {EnumName} protocol.Enum {Comment}

const (
{ClassBody}
)

var ALL_{EnumName} protocol.IEnumCollection = []protocol.IEnum{{EnumList}}

func TO_{EnumName}(s string){EnumName}{
	switch s {
{StringToEnum}
	}
	return -1
}

func (this {EnumName}) Enum()protocol.Enum{
	return protocol.Enum(this)
}


func (this {EnumName}) ToString()string{
	switch this {
{EnumToString}
	}
	return ""
}
`
	if this.Comment!="" {
		comment:=" "+this.Comment
		ret = strings.Replace(ret,`{Comment}`,comment,-1)
	} else {
		ret = strings.Replace(ret,`{Comment}`,"",-1)
	}
	ret = strings.Replace(ret,`{PackageName}`,this.Package,-1)
	ret = strings.Replace(ret,`{EnumName}`,stringutil.SplitCamelCaseUpperSnake(this.EnumName),-1)
	ret = strings.Replace(ret,`{ClassBody}`,this.goFields(g),-1)
	ret = strings.Replace(ret,`{StringToEnum}`,this.gsString2EnumFields(g),-1)
	ret = strings.Replace(ret,`{EnumToString}`,this.gsEnum2StringFields(g),-1)
	ret = strings.Replace(ret,`{EnumList}`,this.gsEnumListFields(g),-1)
	return ret
}

func (this *ModelEnumObject) gsEnumListFields(g *Generator)string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			ret+=stringutil.SplitCamelCaseUpperSnake(this.EnumName)
			ret+="_"
			ret+=stringutil.SplitCamelCaseUpperSnake(l.Name)
			ret+=","
		}
	}
	ret = strings.TrimRight(ret,",")
	return ret
}

func (this *ModelEnumObject) gsString2EnumFields(g *Generator)string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			comment:=strings.TrimSpace(l.Comment)
			if comment=="" {
				break
			}
			comment=strings.ReplaceAll(l.Comment,"//","")
			comment = strings.TrimSpace(comment)
			comment = `"`+comment+`"`
			ret+="\tcase "
			ret+=comment
			ret+=":\n"
			ret+="\t\t"
			ret+= "return "
			ret+=stringutil.SplitCamelCaseUpperSnake(this.EnumName)
			ret+="_"
			ret+=stringutil.SplitCamelCaseUpperSnake(l.Name)
			ret+="\n"
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelEnumObject) gsEnum2StringFields(g *Generator)string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			comment:=strings.TrimSpace(l.Comment)
			if comment=="" {
				break
			}
			ret+="\tcase "
			ret+=stringutil.SplitCamelCaseUpperSnake(this.EnumName)
			ret+="_"
			ret+=stringutil.SplitCamelCaseUpperSnake(l.Name)
			ret+=":\n"
			ret+="\t\t"
			ret+= "return "
			comment=strings.ReplaceAll(l.Comment,"//","")
			comment = strings.TrimSpace(comment)
			comment = `"`+comment+`"`
			ret+=comment
			ret+="\n"
		}
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelEnumObject) goFields(g *Generator)string{
	ret:=""
	for _,l:=range this.lines {
		if l.LineType ==LINE_MODEL_ENUM_FIELD {
			ret+="\t"
			ret+=stringutil.SplitCamelCaseUpperSnake(this.EnumName)
			ret+="_"
			ret+=stringutil.SplitCamelCaseUpperSnake(l.Name)
			ret+=" "
			ret+=stringutil.SplitCamelCaseUpperSnake(this.EnumName)
			ret+=" "
			ret+= " = "
			ret+=strconv.Itoa(l.GetValue())
			ret+=" "
			ret+=l.Comment
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
			this.Comment = line.Comment
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
	Comment string
}

//TODO:dep
func (this *ModelClassFields) Deps(g *Generator)[]string{
	protos:=[]string{}
	t:= MatchModelProtoTag(this.Type)
	if t!=0 {
		t=GetModelProtoTag(this.Type)
		protos = append(protos, this.Type)
	} else if t,s1,s2:=MatchModelSystemTag(this.Type);t!=0 {
		if t==TAG_List {
			t = MatchModelProtoTag(s1)
			if t!=0 {
			} else {
				protos = append(protos, s1)
			}
		} else if t==TAG_Map {
			t1 := MatchModelProtoTag(s1)
			t2 := MatchModelProtoTag(s2)
			//type1 := "*"+s1
			if t1!=0 {
				//type1 = t1.GoTypeString()
			} else {
				if g.IsEnum(s1) {
					//type1 = stringutil.SplitCamelCaseUpperSnake(s1)
					protos = append(protos,s1)
				} else {
					protos = append(protos,s1)
				}
			}

			//type2 := "*"+s2
			if t2!=0 {
				//type2 = t2.GoTypeString()
			}else {
				if g.IsEnum(s2) {
					//type2 = stringutil.SplitCamelCaseUpperSnake(s2)
					protos = append(protos,s2)
				} else {
					protos = append(protos,s2)
				}
			}

		}
	} else {
		if g.IsEnum(this.Type) {
			protos = append(protos,this.Type)
			//ret = "\t"+this.Name+" "+stringutil.SplitCamelCaseUpperSnake(this.Type)
		} else {
			protos = append(protos,this.Type)
			//ret = "\t"+this.Name+" *"+this.Type
		}
	}
	ret:=[]string{}
	for _,p:=range protos {
		m:=g.GetModelByName(p)
		if m!=nil {
			ret = append(ret, m.Package)
		}
		e:=g.GetEnumByName(p)
		if e!=nil {
			ret = append(ret, e.Package)
		}
	}
	return ret
}

func (this *ModelClassFields) csString(g *Generator,lower bool)string {
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
				if t==TAG_Byte {
					ret = "byte[] "+name
				} else {
					ret = "List<"+t.CsTypeString()+"> "+name
				}
			} else if g.IsEnum(s1) {
				ret = "List<"+stringutil.SplitCamelCaseUpperSnake(s1)+"> "+name
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
			if g.IsEnum(s1) {
				type1 = stringutil.SplitCamelCaseUpperSnake(s1)
			}
			type2 := s2
			if t2!=0 {
				type2 = t2.CsTypeString()
			}
			if g.IsEnum(s2) {
				type1 = stringutil.SplitCamelCaseUpperSnake(s2)
			}
			ret =  "Dictionary<"+type1+","+type2+"> "+name

		}
	} else if g.IsEnum(this.Type) {
		ret = stringutil.SplitCamelCaseUpperSnake(this.Type)+" "+name
	} else {
		ret = this.Type+" "+name
	}
	return ret
}

func (this *ModelClassFields) CsString(g *Generator)string {
	return "\t\tpublic "+this.csString(g,false)+"{ get;set; }"+" "+this.Comment
}

func (this *ModelClassFields) TsDefineTags(g *Generator, tsClass *TsClassObject)string{
	str := "\t[" + `"`
	str += this.Name + `",`
	log.Infof("this.Type",this.Type,this.Name)
	t:= MatchModelProtoTag(this.Type)
	if t!=0 {
		t:=GetModelProtoTag(this.Type)
		str += t.TsTagString()
	} else if t,s1,s2:=MatchModelSystemTag(this.Type);t!=0 {
		if t==TAG_List {
			t = MatchModelProtoTag(s1)
			str += "Tag.List"
			if t!=0 {
				str+="," + t.TsTagString()
			} else {
				str+="," + s1
			}
		} else if t==TAG_Map {
			str += "Tag.Map"
			t1 := MatchModelProtoTag(s1)
			t2 := MatchModelProtoTag(s2)
			type1 := s1
			if t1!=0 {
				type1 = t1.TsTagString()
			}
			if g.IsEnum(s1) {
				type1 = "Tag.Int"
			}
			type2 := s2
			if t2!=0 {
				type2 = t2.TsTagString()
			}
			if g.IsEnum(s2) {
				type1 = stringutil.SplitCamelCaseUpperSnake(s2)
			}
			str+="," + type1+","+type2
		}
	} else if g.IsEnum(this.Type) {
		str+=TAG_Int.TsTagString()
	} else {
		str+=this.Type
	}
	if tsClass.CheckBuffer(this.Name) {
		str+=`,Tag.Buffer`
	}
	if tsClass.CheckLongString(this.Name) {
		str+=`,Tag.LongString`
	}
	str+="],"
	return str
}

func (this *ModelClassFields) ParamString(g *Generator)string {
	return this.csString(g,true)
}

func (this *ModelClassFields) ParamAssignString(g *Generator)string {
	return "\t\t\t"+this.Name+" = "+stringutil.FirstToLower(this.Name)+";"
}

func (this *ModelClassFields) GoString(g *Generator)string {
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
			} else if g.IsEnum(s1) {
				ret = "\t"+this.Name+" []"+stringutil.SplitCamelCaseUpperSnake(s1)
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
			if g.IsEnum(s1) {
				type1 = stringutil.SplitCamelCaseUpperSnake(s1)
			}
			type2 := "*"+s2
			if t2!=0 {
				type2 = t2.GoTypeString()
			}
			if g.IsEnum(s2) {
				type2 = stringutil.SplitCamelCaseUpperSnake(s2)
			}
			ret =  "\t"+this.Name+" map["+type1+"]"+type2

		}
	} else {
		if g.IsEnum(this.Type) {
			ret = "\t"+this.Name+" "+stringutil.SplitCamelCaseUpperSnake(this.Type)
		} else {
			ret = "\t"+this.Name+" *"+this.Type
		}
	}
	ret+=" "
	ret+=this.Comment
	return ret
}

func (this *ModelClassFields) TsPublicString(g *Generator)string{
	ret:="\tpublic "
	ret+= this.Name+":"
	ret += this.TsPublicType(g)
	return ret
}

func (this *ModelClassFields) TsPublicType(g *Generator)string{
	ret:=""
	t:= MatchModelProtoTag(this.Type)
	if t!=0 {
		t:=GetModelProtoTag(this.Type)
		ret = t.TsTypeString()
	} else if t,s1,s2:=MatchModelSystemTag(this.Type);t!=0 {
		if t==TAG_List {
			t = MatchModelProtoTag(s1)
			if t!=0 {
				ret = t.TsTypeString()+"[]"
			} else if g.IsEnum(s1) {
				ret = stringutil.SplitCamelCaseUpperSnake(s1)+"[]"
			} else {
				ret = s1+"[]"
			}
		} else if t==TAG_Map {
			t1 := MatchModelProtoTag(s1)
			t2 := MatchModelProtoTag(s2)
			type1 := s1
			if t1!=0 {
				type1 = t1.TsTypeString()
			}
			if g.IsEnum(s1) {
				type1 = stringutil.SplitCamelCaseUpperSnake(s1)
			}
			type2 := s2
			if t2!=0 {
				type2 = t2.TsTypeString()
			}
			if g.IsEnum(s2) {
				type1 = stringutil.SplitCamelCaseUpperSnake(s2)
			}
			ret =  "Map<"+type1+","+type2+"> "

		}
	} else if g.IsEnum(this.Type) {
		ret = stringutil.SplitCamelCaseUpperSnake(this.Type)
	} else {
		ret = this.Type
	}
	return ret
}

type ModelClassObject struct {
	DefaultGeneratorObj
	TagId     BINARY_TAG
	state     int
	Fields    []*ModelClassFields
	Component bool
	Package   string
	CsPackage string
	GoPackage string
	TsPackage string
	ClassName string
	Comment string
	Depends []string
}

func (this *ModelClassObject) Deps(g *Generator)[]string{
	ret:=[]string{}
	for _,f:=range this.Fields {
		deps:=f.Deps(g)
		for _,d:=range deps {
			if d == this.Package {
				continue
			}
			ret = append(ret, d)
		}
	}
	return ret
}

func NewModelClassObject(file GeneratorFile) *ModelClassObject {
	ret := &ModelClassObject{DefaultGeneratorObj: DefaultGeneratorObj{},Component: true,Fields: []*ModelClassFields{},Depends: []string{}}
	ret.DefaultGeneratorObj.init(OBJ_MODEL_CLASS, file)
	return ret
}

func (this *ModelClassObject) TsDefineLines(g *Generator, tsClass *TsClassObject) []string {
	ret := make([]string, 0)
	for _, field := range this.Fields {
		str:=field.TsDefineTags(g,tsClass)
		ret = append(ret,str)
	}
	return ret
}

func (this *ModelClassObject) TsDefineEnd(g *Generator) string {
	ret := "]"
	for _, depend := range this.Depends {
		ret += `,"` + depend + `"`
	}
	ret += `)`
	return ret
}

func (this *ModelClassObject) TsDefineStart(g *Generator) string {
	return `@define("` + this.ClassName + `"` + `,[`
}

func (this *ModelClassObject) TsDefineSingleLine(g *Generator) string {
	ret := `@define("` + this.ClassName + `"`
	if len(this.Fields) > 0 {
		log.Panicf("single line must not have members")
	}
	if len(this.Depends) > 0 {
		ret += ",[]"
		for _, depend := range this.Depends {
			ret += `,"` + depend + `"`
		}
	}
	ret += `)`
	return ret
}

func (this *ModelClassObject) ToTsClassHeader(g *Generator,object *TsClassObject)string{
	compStr:="ISerializable"
	if object.IsComponent {
		compStr = "BaseComponent"
	}
	if object.IsRenderComponent {
		compStr = "RenderComponent"
	}
	ret:="export class "+this.ClassName+" extends "+compStr+" {"
	return ret
}

func (this *ModelClassObject) CsString(g *Generator)string{
	ret:=`//this is a generate file,do not modify it!
using System;
using System.Collections.Generic;
using Funnel.Protocol.Abstractions;
{Comment}
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
{Comment}
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
		ret = strings.Replace(ret,`{ClassBody}`,this.csFields(g),-1)
	} else {
		ret = strings.Replace(ret,`{CsPackageName}`,this.CsPackage,-1)
		ret = strings.Replace(ret,`{ClassName}`,this.ClassName,-1)
		ret = strings.Replace(ret,`{ClassBody}`,this.csFields(g),-1)
		ret = strings.Replace(ret,`{CsParams}`,this.csParams(g),-1)
		ret = strings.Replace(ret,`{ParamAssign}`,this.csParamAssign(g),-1)
	}
	if this.Comment!="" {
		comment:="\n"+this.Comment
		ret = strings.Replace(ret,`{Comment}`,comment,-1)
	} else {
		ret = strings.Replace(ret,`{Comment}`,"",-1)
	}
	return ret
}

func (this *ModelClassObject) csFields(g *Generator)string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.CsString(g)
		ret+="\n"
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelClassObject) csParams(g *Generator)string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.ParamString(g)
		ret+=","
	}
	ret = strings.TrimRight(ret,",")
	return ret
}

func (this *ModelClassObject) csParamAssign(g *Generator)string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.ParamAssignString(g)
		ret+="\n"
	}
	ret = strings.TrimRight(ret,"\n")
	return ret
}

func (this *ModelClassObject) GoImplString(g *Generator)string{
	ret:=`//this is a generate file,edit implement on this file only!
package {PackageName}

import (
	"github.com/nomos/go-lokas"
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
	ret = strings.Replace(ret,`{PackageName}`,this.Package,-1)
	ret = strings.Replace(ret,`{ClassName}`,this.ClassName,-1)
	return ret
}

func (this *ModelClassObject) GoString(g *Generator)string{
	ret:=`//this is a generate file,do not modify it!
package {PackageName}

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/protocol"
	"reflect"{OtherImport}{Import}
)

var _ lokas.IComponent = (*{EnumName})(nil)
{Comment}
type {EnumName} struct {
	ecs.Component `+"`json:\"-\" bson:\"-\"`"+`
{ClassBody}
}

func (this *{EnumName}) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *{EnumName}) Serializable()protocol.ISerializable {
	return this
}
`
	if this.hasTime()||this.hasColor() {
		str:="\n\t"
		if this.hasColor() {
			str+=`"github.com/nomos/go-lokas/util/colors"`+"\n"
		}
		if this.hasTime() {
			str+=`"time"`+"\n"
		}
		str = strings.TrimRight(str,"\n")
		ret = strings.Replace(ret,`{OtherImport}`,str,-1)
	} else {
		ret = strings.Replace(ret,`{OtherImport}`,"",-1)
	}
	if len(this.Deps(g))>0 {
		ret = strings.Replace(ret,`{Import}`,g.getGoImportsString(this.Deps(g)),-1)
	} else {
		ret = strings.Replace(ret,`{Import}`,"",-1)
	}
	if this.Comment!="" {
		comment:="\n"+this.Comment
		ret = strings.Replace(ret,`{Comment}`,comment,-1)
	} else {
		ret = strings.Replace(ret,`{Comment}`,"",-1)
	}
	ret = strings.Replace(ret,`{PackageName}`,this.Package,-1)
	ret = strings.Replace(ret,`{EnumName}`,this.ClassName,-1)
	ret = strings.Replace(ret,`{ClassBody}`,this.goFields(g),-1)
	return ret
}

func (this *ModelClassObject) hasTime()bool {
	for _,f:=range this.Fields {
		if f.Type=="time" {
			return true
		}
	}
	return false
}

func (this *ModelClassObject) hasColor()bool {
	for _,f:=range this.Fields {
		if f.Type=="color" {
			return true
		}
	}
	return false
}

func (this *ModelClassObject) goFields(g *Generator)string{
	ret:=""
	for _,f:=range this.Fields {
		ret+=f.GoString(g)
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
			this.Comment = line.Comment
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
					Comment: line.Comment,
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
		offset, success = this.parse(offset, OBJ_MODEL_IDS,OBJ_MODEL_ERRORS,OBJ_MODEL_CLASS, OBJ_MODEL_ENUM)
		log.Infof("isFinish", len(this.Lines), offset)
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

func (this *ModelFile) ProcessImports() []string {
	ret := make([]string, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_IMPORTS {
			for _,v:=range obj.(*ModelImportObject).imports {
				ret = append(ret, v)
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i]<ret[j]
	})
	return ret
}

func (this *ModelFile) ProcessIds() []*ModelId {
	ret := make([]*ModelId, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_IDS {
			for _,v:=range obj.(*ModelIdsObject).Ids {
				v.PackageName = this.Package
				ret = append(ret, v)
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Id<ret[j].Id
	})
	return ret
}

func (this *ModelFile) ProcessErrors() []*ModelError {
	ret := make([]*ModelError, 0)
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_ERRORS {
			for _,v:=range obj.(*ModelErrorsObject).Errors {
				ret = append(ret, v)
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ErrorId<ret[j].ErrorId
	})
	return ret
}

func (this *ModelFile) ProcessPackages() *ModelPackageObject {
	var ret *ModelPackageObject = nil
	defer func() {
		r:=recover()
		if r!=nil {
			this.GetLogger().Errorf(r)
		}
	}()
	for _, obj := range this.Objects {
		if obj.ObjectType() == OBJ_MODEL_PACKAGE {
			this.GetLogger().Infof("this.Package",this.Package)
			obj.(*ModelPackageObject).PackageName = this.Package
			ret =  obj.(*ModelPackageObject)
		}
		if obj.ObjectType() ==  OBJ_MODEL_IDS{
			for _,id:=range obj.(*ModelIdsObject).Ids {
				id.PackageName = this.Package
			}
		}
		if obj.ObjectType() ==  OBJ_MODEL_ERRORS{
			for _,id:=range obj.(*ModelErrorsObject).Errors {
				id.PackageName = this.Package
			}
		}
	}
	this.GetLogger().Info("ProcessPackages")
	return ret
}
