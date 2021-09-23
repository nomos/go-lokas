package lox

import (
	"context"
	"errors"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"sync"
	"time"
)

type avatarManagerCtor struct {
	option lokas.IGameHandler
}

func NewAvatarManagerCtor(option lokas.IGameHandler)avatarManagerCtor{
	return avatarManagerCtor{option: option}
}

func (this avatarManagerCtor) Type()string{
	return "AvatarManager"
}

func (this avatarManagerCtor) Create()lokas.IModule {
	ret:=&AvatarManager{
		Actor:NewActor(),
		Avatars: map[util.ID]*Avatar{},
	}
	ret.SetType(this.Type())
	ret.OnUpdateFunc = ret.OnUpdate
	ret.MsgHandler = ret.HandleMsg
	ret.option = this.option
	return ret
}

var _ lokas.IActor = (*AvatarManager)(nil)

type AvatarManager struct {
	*Actor
	Avatars map[util.ID]*Avatar
	option lokas.IGameHandler
	mu sync.Mutex
}

func (this *AvatarManager) HandleMsg(actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error) {
	id,err:=msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	if id== TAG_CREATE_AVATAR {
		createMsg:=msg.(*CreateAvatar)
		//多线程处理创建玩家Actor
		go func() {
			err:=this.CreateAvatar(util.ID(createMsg.Id))
			if err != nil {
				log.Error(err.Error())
				this.SendMessage(actorId,transId,NewResponse(false))
				return
			}
			this.SendMessage(actorId,transId,NewResponse(true))
		}()
	}
	return nil,nil
}

func (this *AvatarManager) GetAvatar(id util.ID)*Avatar{
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.Avatars[id]
}

func (this *AvatarManager) CreateAvatar(id util.ID)error{
	this.mu.Lock()
	defer this.mu.Unlock()
	avatar:=this.Avatars[id]
	if avatar !=nil {
		this.logRetention(avatar)
		return nil
	}
	log.Warnf("CreateAvatar",id)
	a:=this.GetProcess()
	var am AvatarMap
	err := a.GetMongo().Collection("avatarmap").Find(context.TODO(),bson.M{"_id": id}).One(&am)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	avatar= NewAvatar(id,this.option)
	avatar.UserName = am.UserName
	avatar.GameId = am.GameId
	avatar.ServerId = am.ServerId
	avatar.SetProcess(this.GetProcess())
	err =avatar.Deserialize(this.GetProcess())
	avatar.SetId(id)
	this.GetProcess().AddActor(avatar)
	err = this.GetProcess().StartActor(avatar)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.Avatars[id] = avatar
	this.logRetention(avatar)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *AvatarManager) logRetention(avatar lokas.IAvatar){
	regtime:=avatar.GetId().Time()
	timeDelta:= time.Now().Sub(regtime)/(time.Hour*24)
	log.Warn("retention day",flog.AvatarInfo(avatar).Append(zap.String("register_day",util.FormatTimeToISOString(regtime))).Append(zap.Int64("register_avatar",avatar.GetId().Int64())).Append(flog.RegElapsedDay(int(timeDelta)))...)
}

func (this *AvatarManager) RemoveAvatar(id util.ID){
	this.GetProcess().RemoveActorById(id)
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.Avatars,id)
}

func (this *AvatarManager) Load(conf lokas.IConfig) error {
	return nil
}

func (this *AvatarManager) Unload() error {
	return nil
}

func (this *AvatarManager) Start() error {
	this.StartMessagePump()
	return nil
}

func (this *AvatarManager) Stop() error {
	var err error
	for _,a:=range this.Avatars {
		err=a.Stop()
		if err != nil {
			log.Error(err.Error())
		}
	}
	if err != nil {
		log.Error(err.Error())
		return errors.New("actors stop failed")
	}
	return nil
}

func (this *AvatarManager) OnUpdate(){
	this.GetProcess().RegisterActorRemote(this)
}

func (this *AvatarManager) OnStart() error {
	log.Warnf("AvatarManager:OnStart")
	err:=this.GetProcess().RegisterActorLocal(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err=this.GetProcess().RegisterActorRemote(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *AvatarManager) OnStop() error {
	err:=this.GetProcess().UnregisterActorLocal(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err=this.GetProcess().UnregisterActorRemote(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}