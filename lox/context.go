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
	if _,ok:=this.Content[key];!ok {
		return ""
	}
	return this.Content[key].(string)
}

func (this *Context) GetInt(key string)int {
	if _,ok:=this.Content[key];!ok {
		return 0
	}
	return this.Content[key].(int)
}

func (this *Context) GetBool(key string)bool {
	if _,ok:=this.Content[key];!ok {
		return false
	}
	return this.Content[key].(bool)
}

func (this *Context) GetInt32(key string)int32 {
	if _,ok:=this.Content[key];!ok {
		return 0
	}
	return this.Content[key].(int32)
}

func (this *Context) GetInt64(key string)int64 {
	if _,ok:=this.Content[key];!ok {
		return 0
	}
	return this.Content[key].(int64)
}

func (this *Context) GetFloat32(key string)float32 {
	if _,ok:=this.Content[key];!ok {
		return 0
	}
	return this.Content[key].(float32)
}

func (this *Context) GetFloat64(key string)float64 {
	if _,ok:=this.Content[key];!ok {
		return 0
	}
	return this.Content[key].(float64)
}

func (this *Context) GetIdType(key string)util.ID {
	return this.Content[key].(util.ID)
}

func (this *Context) Set(key string,value interface{}) {
	this.Content[key] = value
}