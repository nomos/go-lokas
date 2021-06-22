package lox

import (
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
	"sync"
)

var AvatarManagerCtor = avatarManagerCtor{}

type avatarManagerCtor struct {}

func (this avatarManagerCtor) Type()string{
	return "AvatarManager"
}

func (this avatarManagerCtor) Create()lokas.IModule {
	ret:=&AvatarManager{
		Actor:NewActor(),
		Avatars: map[util.ID]*Avatar{},
	}
	ret.OnUpdateFunc = ret.OnUpdate
	ret.SetType(this.Type())
	return ret
}

var _ lokas.IActor = (*AvatarManager)(nil)

type AvatarManager struct {
	*Actor
	Avatars map[util.ID]*Avatar
	mu sync.Mutex
}

func (this *AvatarManager) GetAvatar(id util.ID)*Avatar{
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.Avatars[id]
}

func (this *AvatarManager) CreateAvatar(id util.ID)(*Avatar,error){
	avatar:=this.GetAvatar(id)
	this.mu.Lock()
	defer this.mu.Unlock()
	if avatar !=nil {
		return avatar,nil
	}
	avatar=NewAvatar(id)
	this.GetProcess().AddActor(avatar)
	this.Avatars[id] = avatar
	err:=avatar.Deserialize(this.GetProcess())
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	return avatar,err
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

func (this *AvatarManager) Start() *promise.Promise {
	this.StartMessagePump()
	return promise.Resolve(nil)
}

func (this *AvatarManager) Stop() *promise.Promise {
	return promise.Resolve(nil)
}

func (this *AvatarManager) OnUpdate(){
	this.GetProcess().RegisterActorRemote(this)
}

func (this *AvatarManager) OnStart() error {
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