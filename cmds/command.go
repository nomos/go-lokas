package cmds

import (
	"errors"
	"fmt"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util/promise"
	"go.uber.org/zap"
	"reflect"
	"strconv"
	"strings"
)

type CmdType string

const (
	Cmd_Shell = "shell"
	Cmd_Expect = "expect"
)

func (this CmdType) GetCmdPrefix()string {
	switch this {
	case Cmd_Shell:
		return "/bin/sh"
	case Cmd_Expect:
		return "/usr/bin/expect"
	}
	return ""
}

type CmdResult struct {
	Outputs []string
	Success bool
	Results map[string]interface{}
}

type CmdOutput []string

func (this CmdOutput) LastOutput()string{
	if len(this)==0 {
		return ""
	}
	return this[len(this)-1]
}

type CmdHandler func(CmdOutput)*CmdResult

type WrappedCmd struct {
	CmdString string
	Tips string
	ParamsNum int
	ParamsMap []string
	CmdType CmdType
	CmdHandler CmdHandler
}

func (this *WrappedCmd) FillParams(param... string)string{
	ret :=this.CmdString
	for i,s:=range param{
		ret = strings.Replace(ret,"$"+strconv.Itoa(i+1),s,-1)
	}
	return ret
}

var wrappedCmds = make(map[string]*WrappedCmd)

func GetAllCmds()map[string]*WrappedCmd {
	return wrappedCmds
}

func RegisterCmd(s string,cmd *WrappedCmd){
	if _,ok:=wrappedCmds[s];ok{
		log.Warnf("duplicated cmd:%s,overwrite",s)
	}
	wrappedCmds[s] = cmd
}

func GetCmdByName(s string)*WrappedCmd {
	return wrappedCmds[s]
}

func SplitCommandParams(cmd string)(string,[]string){
	splits := strings.Split(cmd," ")
	if len(splits)==0 {
		return "",[]string{}
	}
	return splits[0],splits[1:]
}

type Command struct {
	name string
	execFunc func(value *ParamsValue,console IConsole)*promise.Promise
	tips string
	console IConsole
}

func (this *Command) Name()string{
	return this.name
}

func (this *Command) Exec(params... interface{})*promise.Promise{
	param:=&ParamsValue{
		cmd:    "",
		value: params,
		offset: 0,
	}
	return this.ConsoleExec(param,this.console)
}

func (this *Command) ExecWithConsole(console IConsole,params... interface{})*promise.Promise{
	param:=&ParamsValue{
		cmd:    "",
		value: params,
		offset: 0,
	}
	return this.ConsoleExec(param,console)
}

func (this *Command) ConsoleExec(param *ParamsValue,console IConsole)*promise.Promise {
	if this.execFunc!=nil {
		if param.IsHelp() {
			log.Info(this.tips)
			if console!=nil {
				console.Write([]byte(this.tips))
			}
			return promise.Resolve(nil)
		}
		return promise.Async(func(resolve func(interface{}), reject func(interface{})) {
			defer func() {
				if r:=recover();r!=nil {
					err := r.(error)
					if cmdErr,ok := r.(*CmdError);ok{
						if cmdErr.errorType == CMD_ERROR_PARAM_LEN {
							errStr := cmdErr.cmd+" 命令长度必须大于"+strconv.Itoa(cmdErr.offset+1)
							log.Error(errStr)
							if console!=nil {
								console.Write([]byte(errStr))
							}
						}
						if cmdErr.errorType == CMD_ERROR_PARAM_TYPE {
							errStr := cmdErr.cmd+" 命令参数("+strconv.Itoa(cmdErr.offset+1)+")类型必须为"+cmdErr.paramType
							log.Error(errStr)
							if console!=nil {
								console.Write([]byte(errStr))
							}
						}
					} else {
						log.Error(err.Error())
						console.Write([]byte(this.name+" 执行命令时出现未知错误:"+err.Error()))
					}
					errStr:="type "+this.name+" help|?"
					log.Info(errStr)
					if console!=nil {
						console.Write([]byte(errStr))
					}
					reject(err)
				}
			}()
			res,err:=this.execFunc(param,console).Await()
			if err!=nil {
				reject(err)
				return
			}
			resolve(res)
		})
	}
	return promise.Reject(errors.New("cant found exec"))
}

func (this *Command) Tips()string{
	return this.tips
}

func (this *Command) SetConsole(console IConsole) {
	this.console = console
}

func NewCommand(name string,tips string,f func(value *ParamsValue,console IConsole)*promise.Promise,console IConsole)ICommand{
	ret:=&Command{
		name: name,
		execFunc: f,
		tips:tips,
		console:console,
	}
	return ret
}

func NewCommandNoConsole(name string,tips string,f func(value *ParamsValue,console IConsole)*promise.Promise)ICommand{
	ret:=&Command{
		name: name,
		execFunc: f,
		tips:tips,
		console:nil,
	}
	return ret
}

type ParamsValue struct {
	cmd string
	value []interface{}
	offset int
}

func NewParamsValue(cmd string,value ...interface{})*ParamsValue{
	return &ParamsValue{
		cmd:   cmd,
		value:  value,
		offset: 0,
	}
}

func (this *ParamsValue) Len()int {
	return len(this.value)
}

func (this *ParamsValue) IsHelp()bool{
	return len(this.value)==1&&(this.value[0] == "?"||this.value[0]=="help")
}

const (
	CMD_ERROR_PARAM_LEN = iota
	CMD_ERROR_PARAM_TYPE
)

type CmdError struct {
	errorType int
	cmd string
	offset int
	paramType string
}

func (this *CmdError) Error() string {
	switch this.errorType {
	case CMD_ERROR_PARAM_LEN:
		return fmt.Sprintf("cmd %s params length error, type %s help|? ",this.cmd,this.cmd)
	case CMD_ERROR_PARAM_TYPE:
		return fmt.Sprintf("cmd %s params[%d] type error,must be %s,type %s help|?",this.cmd,this.offset,this.paramType,this.cmd)
	default:
		return fmt.Sprintf("CmdError Type Error:%d",this.errorType)
	}
}

func NewCmdError(errType int,cmd string,offset int,typ string)*CmdError {
	return &CmdError{
		errorType: errType,
		cmd:cmd,
		offset:offset,
		paramType: typ,
	}
}

func (this *ParamsValue) LeftParams()[]interface{}{
	ret:=[]interface{}{}
	for i :=this.offset;i<len(this.value);i++{
		ret = append(ret, this.value[i])
	}
	return ret
}

func (this *ParamsValue) StringOpt()string {
	if len(this.value)-1<this.offset {
		return ""
	}
	ret:=this.value[this.offset]
	this.offset++
	return ret.(string)
}

func (this *ParamsValue) String()string {
	if len(this.value)-1<this.offset {
		panic(NewCmdError(CMD_ERROR_PARAM_LEN,this.cmd,this.offset,"string"))
	}
	ret:=this.value[this.offset]
	this.offset++
	return ret.(string)
}

func (this *ParamsValue) Int32()int32 {
	return int32(this.Int())
}

func (this *ParamsValue) Int64()int64 {
	return int64(this.Int())
}

func (this *ParamsValue) Int32Opt()int32 {
	return int32(this.IntOpt())
}

func (this *ParamsValue) Int64Opt()int64 {
	return int64(this.IntOpt())
}

func (this *ParamsValue) Int()int {
	if len(this.value)-1<this.offset {
		panic(NewCmdError(CMD_ERROR_PARAM_LEN,this.cmd,this.offset,"int"))
	}
	return this.int()
}

func (this *ParamsValue) int()int{
	raw:=this.value[this.offset]
	var ret int
	var err error
	switch reflect.TypeOf(raw).Kind() {
	case reflect.String:
		ret,err=strconv.Atoi(raw.(string))
		if err != nil {
			panic(NewCmdError(CMD_ERROR_PARAM_TYPE,this.cmd,this.offset,"int"))
		}
	case reflect.Uint:
		ret=int(raw.(uint))
	case reflect.Int:
		ret=int(raw.(int))
	case reflect.Uint8:
		ret=int(raw.(uint8))
	case reflect.Int8:
		ret=int(raw.(int8))
	case reflect.Uint16:
		ret=int(raw.(uint16))
	case reflect.Int16:
		ret=int(raw.(int16))
	case reflect.Uint32:
		ret=int(raw.(uint32))
	case reflect.Int32:
		ret=int(raw.(int32))
	case reflect.Uint64:
		ret=int(raw.(uint64))
	case reflect.Int64:
		ret=int(raw.(int64))
	default:
		log.Panic("type not supported",zap.Any("raw",raw))
	}
	this.offset++
	return ret
}

func (this *ParamsValue) float32()float32{
	raw:=this.value[this.offset]
	var ret float32
	switch reflect.TypeOf(raw).Kind() {
	case reflect.String:
		v,err:=strconv.ParseFloat(raw.(string),3)
		ret = float32(v)
		if err != nil {
			panic(NewCmdError(CMD_ERROR_PARAM_TYPE,this.cmd,this.offset,"float"))
		}
	case reflect.Float64:
		ret=float32(raw.(float64))
	case reflect.Float32:
		ret=raw.(float32)
	default:
		log.Panic("type not supported",zap.Any("raw",raw))
	}
	this.offset++
	return ret
}

func (this *ParamsValue) float64()float64{
	raw:=this.value[this.offset]
	var ret float64
	var err error
	switch reflect.TypeOf(raw).Kind() {
	case reflect.String:
		ret,err=strconv.ParseFloat(raw.(string),3)
		if err != nil {
			panic(NewCmdError(CMD_ERROR_PARAM_TYPE,this.cmd,this.offset,"float"))
		}
	case reflect.Float64:
		ret=raw.(float64)
	case reflect.Float32:
		ret=float64(raw.(float32))
	default:
		log.Panic("type not supported",zap.Any("raw",raw))
	}
	this.offset++
	return ret
}

func (this *ParamsValue) IntOpt()int {
	if len(this.value)-1<this.offset {
		return 0
	}
	return this.int()
}

func (this *ParamsValue) Bool()bool {
	if len(this.value)-1<this.offset {
		panic(NewCmdError(CMD_ERROR_PARAM_LEN,this.cmd,this.offset,"bool"))
	}
	ret:=false
	v:=this.value[this.offset]
	if v=="true"||v=="True"||v=="TRUE"||v=="1" {
		ret = true
	} else if v=="false"||v=="False"||v=="FALSE"||v=="0" {
		ret = false
	} else {
		panic(NewCmdError(CMD_ERROR_PARAM_TYPE,this.cmd,this.offset,"bool"))
	}
	this.offset++
	return ret
}

func (this *ParamsValue) BoolOpt()bool {
	if len(this.value)-1<this.offset {
		return false
	}
	ret:=false
	v:=this.value[this.offset]
	if v=="true"||v=="True"||v=="TRUE"||v=="1" {
		ret = true
	} else if v=="false"||v=="False"||v=="FALSE"||v=="0" {
		ret = false
	} else {
		panic(NewCmdError(CMD_ERROR_PARAM_TYPE,this.cmd,this.offset,"bool"))
	}
	this.offset++
	return ret
}

func (this *ParamsValue) Float()float64 {
	if len(this.value)-1<this.offset {
		panic(NewCmdError(CMD_ERROR_PARAM_LEN,this.cmd,this.offset,"float"))
	}
	return this.float64()
}

func (this *ParamsValue) FloatOpt()float64 {
	if len(this.value)-1<this.offset {
		return 0
	}
	return this.float64()
}
