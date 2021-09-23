package lox

import (
	"github.com/nomos/go-lokas/protocol"
	"reflect"
)

//默认回复
type Response struct {
	OK bool
}

func (this *Response) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *Response) Serializable()protocol.ISerializable {
	return this
}

func NewResponse(v bool) *Response{
	return &Response{OK: v}
}
