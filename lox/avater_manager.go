package lox

import (
	"context"
	"errors"
	"github.com/nomos/go-lokas/log/flog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type avatarManagerCtor struct {
	option lokas.IGameHandler
}

func NewAvatarManagerCtor(option lokas.IGameHandler) avatarManagerCtor {
	return avatarManagerCtor{option: option}
}

func (this avatarManagerCtor) Type() string {
	return "AvatarManager"
}

func (this avatarManagerCtor) Create() lokas.IModule {
	ctx, cancel := context.WithCancel(context.Background())
	ret := &AvatarManager{
		Actor:   NewActor(),
		Avatars: map[util.ID]*Avatar{},
		Ctx:     ctx,
		Cancel:  cancel,
	}
	ret.SetType(this.Type())
	ret.OnUpdateFunc = ret.OnUpdate
	ret.MsgHandler = ret.HandleMsg
	ret.Option = this.option
	return ret
}

var _ lokas.IActor = (*AvatarManager)(nil)

type AvatarManager struct {
	*Actor
	Avatars   map[util.ID]*Avatar
	AvatarCnt int32
	Option    lokas.IGameHandler
	Mu        sync.Mutex
	Ctx       context.Context
	Cancel    context.CancelFunc
}

func (this *AvatarManager) HandleMsg(actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error) {
	id, err := msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if id == TAG_CREATE_AVATAR {
		createMsg := msg.(*CreateAvatar)
		//多线程处理创建玩家Actor
		go func() {
			err := this.CreateAvatar(util.ID(createMsg.Id))
			if err != nil {
				log.Error(err.Error())
				this.SendMessage(actorId, transId, NewResponse(false))
				return
			}
			this.SendMessage(actorId, transId, NewResponse(true))
		}()
	}
	return nil, nil
}

func (this *AvatarManager) GetAvatar(id util.ID) *Avatar {
	this.Mu.Lock()
	defer this.Mu.Unlock()
	return this.Avatars[id]
}

func (this *AvatarManager) CreateAvatar(id util.ID) error {
	this.Mu.Lock()
	defer this.Mu.Unlock()
	avatar := this.Avatars[id]
	if avatar != nil {
		this.logRetention(avatar)
		return nil
	}

	a := this.GetProcess()
	var am AvatarMap
	err := a.GetMongo().Collection("avatarmap").Find(context.TODO(), bson.M{"_id": id}).One(&am)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	avatar = NewAvatar(id, this.Option, this)
	avatar.UserName = am.UserName
	avatar.GameId = am.GameId
	avatar.ServerId = am.ServerId
	avatar.SetProcess(this.GetProcess())
	err = avatar.Deserialize(this.GetProcess())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	avatar.SetId(id)
	this.GetProcess().AddActor(avatar)
	err = this.GetProcess().StartActor(avatar)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = avatar.Initialize(this.GetProcess())
	if err != nil {
		log.Error(err.Error())
		avatar.Stop()
		return err
	}
	log.Info("CreateAvatar", flog.AvatarId(id), flog.UserName(avatar.UserName), flog.ServerId(avatar.ServerId))
	this.Avatars[id] = avatar
	atomic.AddInt32(&this.AvatarCnt, 1)
	this.logRetention(avatar)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *AvatarManager) logRetention(avatar lokas.IAvatar) {
	regtime := avatar.GetId().Time()
	timeDelta := time.Now().Sub(regtime) / (time.Hour * 24)
	log.Warn("retention day", lokas.LogAvatarInfo(avatar).Append(zap.String("register_day", util.FormatTimeToISOString(regtime))).Append(zap.Int64("register_avatar", avatar.GetId().Int64())).Append(flog.RegElapsedDay(int(timeDelta)))...)
}

func (this *AvatarManager) RemoveAvatar(id util.ID) {
	this.GetProcess().RemoveActorById(id)
	this.Mu.Lock()
	defer this.Mu.Unlock()
	delete(this.Avatars, id)
	atomic.AddInt32(&this.AvatarCnt, -1)
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
	for _, a := range this.Avatars {
		err = a.Stop()
		if err != nil {
			log.Error(err.Error())
		}
	}
	this.Cancel()
	if err != nil {
		log.Error(err.Error())
		return errors.New("actors stop failed")
	}
	return nil
}

func (this *AvatarManager) OnUpdate() {
	this.GetProcess().RegisterActorRemote(this)
}

func (this *AvatarManager) OnStart() error {
	// log.Warn("AvatarManager:OnStart")
	err := this.GetProcess().RegisterActorLocal(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.GetProcess().RegisterActorRemote(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *AvatarManager) OnStop() error {
	err := this.GetProcess().UnregisterActorLocal(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.GetProcess().UnregisterActorRemote(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
