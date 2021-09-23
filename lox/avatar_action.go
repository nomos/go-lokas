package lox

import (
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"reflect"
)

//创建角色
type CreateAvatar struct {
	Id int64
}

func (this *CreateAvatar) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *CreateAvatar) Serializable()protocol.ISerializable {
	return this
}

func NewCreateAvatar(id util.ID)*CreateAvatar{
	ret:=&CreateAvatar{
		Id: int64(id),
	}
	return ret
}

type KickAvatar struct {
	Id int64
}

func (this *KickAvatar) GetId()(protocol.BINARY_TAG,error){
	return protocol.GetTypeRegistry().GetTagByType(reflect.TypeOf(this).Elem())
}

func (this *KickAvatar) Serializable()protocol.ISerializable {
	return this
}

func NewKickAvatar(id util.ID)*KickAvatar{
	ret:=&KickAvatar{
		Id: id.Int64(),
	}
	return ret
}