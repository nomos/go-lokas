package lox

import (
	"encoding/json"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
)

type LoginHandShake struct {
	 GameId string
	Version string
	ServerId int32
	UserId util.ID
	AvatarId util.ID
	Token string
}

func NewLoginHandShake(gameId string,version string,serverId int32,userId util.ID,avatarId util.ID,token string)*LoginHandShake {
	ret:=&LoginHandShake{
		GameId:  gameId,
		Version:  version,
		ServerId: serverId,
		UserId: userId,
		AvatarId: avatarId,
		Token:    token,
	}
	return ret
}

func (this *LoginHandShake) Marshal()([]byte,error) {
	return json.Marshal(this)
}

func (this *LoginHandShake) Unmarshal(from []byte)error {
	err:=json.Unmarshal(from,this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return err
}
