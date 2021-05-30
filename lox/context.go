package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
)

type Context struct {
	Content map[string]interface{} `json:"content"`
}

func NewContext()*Context{
	ret:=&Context{Content: map[string]interface{}{}}
	return ret
}

var _ lokas.IContext = (*Context)(nil)

func (this *Context) Get(key string)interface{} {
	return this.Content[key]
}

func (this *Context) GetProcessIdType(key string)util.ProcessId {
	return this.Content[key].(util.ProcessId)
}

func (this *Context) GetString(key string)string {
	return this.Content[key].(string)
}

func (this *Context) GetIdType(key string)util.ID {
	return this.Content[key].(util.ID)
}

func (this *Context) Set(key string,value interface{}) {
	this.Content[key] = value
}