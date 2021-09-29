package lox

import (
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

var _ lokas.IModel = (*User)(nil)

func UserId(user *User)zap.Field{
	return zap.Int64("user_id",user.Id.Int64())
}
func UserRefreshToken(user *User)zap.Field{
	return zap.String("refresh_token",user.RefreshToken)
}

func UserToken(user *User)zap.Field{
	return zap.String("token",user.Token)
}

func LogUserInfo(user *User)log.ZapFields{
	ret:=log.ZapFields{}
	ret = ret.Append(UserId(user))
	ret = ret.Append(flog.UserName(user.UserName))
	ret = ret.Append(UserToken(user))
	ret = ret.Append(UserRefreshToken(user))
	return ret
}

type User struct {
	Id      util.ID `bson:"_id"`
	Role    uint32
	Avatars map[string]util.ID
	Token string
	RefreshToken string
	UserName string
	Password string
}

func (this *User) GetId()util.ID{
	return this.Id
}

func (this *User) Deserialize(a lokas.IProcess) error {
	panic("implement me")
}

func (this *User) Serialize(a lokas.IProcess) error {
	panic("implement me")
}

func (this *User) SimpleUser()*User{
	return &User{
		Id:        this.Id,
		Role:      this.Role,
	}
}

func (this *User) GetServerInfo(s string)(string,int32){
	sarr:=strings.Split(s,"_")
	gameId:=sarr[0]
	serverId,_:=strconv.Atoi(sarr[1])
	return gameId,int32(serverId)
}

func (this *User) HasAvatarByServer(gameId string,serverId int32)(bool,util.ID){
	for k,v:=range this.Avatars {
		gid,sid:=this.GetServerInfo(k)
		if gameId==gid&&serverId==sid {
			return true,v
		}
	}
	return false,0
}

func (this *User) HasAvatarById(id util.ID)bool{
	for _,v:=range this.Avatars {
		if v==id {
			return true
		}
	}
	return false
}

type ClaimUser struct {
	Create time.Time
	Expire int64
	User *User
}

func (this* ClaimUser) Valid() error {
	expires:= this.Create.Add(time.Duration(this.Expire)).Before(time.Now())
	if expires {
		return protocol.ERR_TOKEN_EXPIRED
	}
	return nil
}

func (this* ClaimUser) Marshal()([]byte,error) {
	return protocol.MarshalBinary(this)
}

func (this* ClaimUser) Unmarshal(v []byte)error {
	return protocol.Unmarshal(v,this)
}

func (this *ClaimUser) SetUser(user interface{}){
	this.User = user.(*User)
}

func (this *ClaimUser) GetUser()interface{} {
	return this.User
}

func CreateUser()interface{} {
	return &User{}
}
