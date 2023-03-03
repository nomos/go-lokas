package lox

import (
	"github.com/nomos/go-lokas/log/flog"
	"go.uber.org/zap"
	"runtime"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
)

var _ lokas.IModel = (*Avatar)(nil)

func NewAvatar(id util.ID, handler lokas.IGameHandler, manager *AvatarManager) *Avatar {
	ret := &Avatar{
		Actor:         NewActor(),
		AvatarSession: NewAvatarSession(id),
		manager:       manager,
	}
	ret.SetType("Avatar")
	ret.SetId(id)
	ret.MsgHandler = ret.handleMsg
	ret.OnUpdateFunc = ret.OnUpdate
	ret.Serializer = handler.GetSerializer()
	ret.Deserializer = handler.GetDeserializer()
	ret.Initializer = handler.GetInitializer()
	ret.Updater = handler.GetUpdater()
	ret.MsgDelegator = handler.GetMsgDelegator()

	return ret
}

var _ lokas.IActor = (*Avatar)(nil)

type Avatar struct {
	*Actor
	*AvatarSession
	manager      *AvatarManager
	Serializer   func(avatar lokas.IActor, process lokas.IProcess) error
	Initializer  func(avatar lokas.IActor, process lokas.IProcess) error
	Deserializer func(avatar lokas.IActor, process lokas.IProcess) error
	Updater      func(avatar lokas.IActor, process lokas.IProcess) error
	MsgDelegator func(avatar lokas.IActor, actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error)
	ClientHost   string
}

func (this *Avatar) SendEvent(msg protocol.ISerializable) error {
	return this.SendReply(this.AvatarSession.GateId, 0, msg)
}

func (this *Avatar) handleMsg(actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				buf := make([]byte, 1<<12)
				runtime.Stack(buf, true)
				log.Error("avatar routine panic detail", lokas.LogAvatarInfo(this).Append(flog.Error(err), zap.String("stack", string(buf)))...)
			}
		}
	}()

	id, err := msg.GetId()
	if transId == 0 && id == protocol.TAG_Error {
		log.Errorf("SendMessageError", actorId, this.Type(), this.GetId(), msg.(*protocol.ErrMsg).Message)
		return nil, nil
	}
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return this.MsgDelegator(this, actorId, transId, msg)
}

func (this *Avatar) Deserialize(a lokas.IProcess) error {
	return this.Deserializer(this, a)
}

func (this *Avatar) Initialize(a lokas.IProcess) error {
	return this.Initializer(this, a)
}

func (this *Avatar) Serialize(a lokas.IProcess) error {
	return this.Serializer(this, a)
}

func (this *Avatar) Load(conf lokas.IConfig) error {
	return nil
}

func (this *Avatar) Unload() error {
	return nil
}

func (this *Avatar) OnUpdate() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Error(err.Error())
				buf := make([]byte, 1<<14)
				runtime.Stack(buf, true)
				log.Error(string(buf))
				this.Stop()
			}
		}
	}()
	this.GetProcess().RegisterActorRemote(this)

	this.Updater(this, this.GetProcess())
	if this.Dirty() {
		this.Serialize(this.GetProcess())
	}
}

// func (this *Avatar) Start() error {
// 	err := this.AvatarSession.Deserialize(this.GetProcess())
// 	if err != nil {
// 		log.Error(err.Error())
// 		return err
// 	}
// 	if this.AvatarSession.UserName == "" || this.AvatarSession.GameId == "" {
// 		err = this.AvatarSession.FetchData(this.GetProcess())
// 		if err != nil {
// 			log.Error(err.Error())
// 			return err
// 		}
// 	}
// 	err = this.AvatarSession.Serialize(this.GetProcess())
// 	if err != nil {
// 		log.Error(err.Error())
// 		return err
// 	}
// 	this.AvatarSession.SetOnGateWayChanged(func(session *AvatarSession) {
// 		this.GateId = session.GateId
// 	})
// 	this.AvatarSession.SetOnSessionClosed(func() {
// 		this.GetProcess().RemoveActor(this)
// 	})
// 	this.AvatarSession.StartAvatarSession()
// 	this.GetProcess().RegisterActorLocal(this)
// 	this.GetProcess().RegisterActorRemote(this)
// 	this.StartMessagePump()
// 	return nil
// }

func (this *Avatar) Start() error {

	this.GetProcess().RegisterActorLocal(this)
	this.GetProcess().RegisterActorRemote(this)
	this.StartMessagePump()
	return nil
}

func (this *Avatar) Stop() error {
	this.Actor.Stop()
	this.Dirty()
	log.Warn("save player state", flog.AvatarId(this.GetId()))
	err := this.Serialize(this.GetProcess())
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.RemoveAll()
	this.GetProcess().UnregisterActorLocal(this)
	this.GetProcess().UnregisterActorRemote(this)
	this.manager.RemoveAvatar(this.GetId())
	return nil
}

func (this *Avatar) OnStart() error {
	return nil
}

func (this *Avatar) OnStop() error {
	this.AvatarSession.StopAvatarSession()
	return nil
}
