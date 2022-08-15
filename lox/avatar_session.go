package lox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/util"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

const (
	MUTEX_AVATAR_SESSION_KEY lokas.Key = "avatar_session_%d"
	AVATAR_SESSION_KEY       lokas.Key = "/avatar_session/%d"
)

var _ lokas.IModel = (*AvatarSession)(nil)

type AvatarSession struct {
	Id               util.ID
	UserId           util.ID
	GateId           util.ID
	UserName         string
	GameId           string
	ServerId         int32
	watchChan        clientv3.WatchChan `bson:"-",json:"-"`
	done             chan struct{}      `bson:"-",json:"-"`
	onGateWayChanged func(session *AvatarSession)
	onSessionClosed  func()
}

func (this *AvatarSession) Initialize(a lokas.IProcess) error {
	return nil
}

func NewAvatarSession(id util.ID) *AvatarSession {
	ret := &AvatarSession{}
	ret.Id = id
	return ret
}

func (this *AvatarSession) SetOnGateWayChanged(f func(session *AvatarSession)) {
	this.onGateWayChanged = f
}

func (this *AvatarSession) SetOnSessionClosed(f func()) {
	this.onSessionClosed = f
}

func (this *AvatarSession) ChangeUserName(a lokas.IProcess, name string) error {
	this.UserName = name
	err := this.Serialize(a)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *AvatarSession) StartAvatarSession() {
	go func() {
		if this.watchChan == nil {
			if this.onSessionClosed != nil {
				log.Errorf("AvatarSession:onSessionClosed")
				this.onSessionClosed()
			}
			return
		}
		this.done = make(chan struct{})
	Loop:
		for {
			select {
			case msg := <-this.watchChan:
				for _, e := range msg.Events {
					switch e.Type {
					case mvccpb.PUT:
						sess := NewAvatarSession(0)
						err := json.Unmarshal(e.Kv.Value, sess)
						if err != nil {
							log.Error(err.Error())
							break Loop
						}
						if sess.GateId != this.GateId {
							if this.onGateWayChanged != nil {
								this.onGateWayChanged(sess)
							}
						}
					case mvccpb.DELETE:
						break Loop
					}
				}
			case <-this.done:
				break Loop
			}
		}
		if this.onSessionClosed != nil {
			this.onSessionClosed()
		}
		if this.done != nil {
			close(this.done)
			this.done = nil
		}
		if this.watchChan != nil {
			this.watchChan = nil
		}
	}()
}

func (this *AvatarSession) StopAvatarSession() {
	log.Warn("AvatarSession StopAvatarSession")
	if this.done != nil {
		this.done <- struct{}{}
	}
}

func (this *AvatarSession) GetGameId() string {
	return this.GameId
}

func (this *AvatarSession) GetServerId() int32 {
	return this.ServerId
}

func (this *AvatarSession) GetGateId() util.ID {
	return this.GateId
}

func (this *AvatarSession) GetUserId() util.ID {
	return this.UserId
}

func (this *AvatarSession) GetUserName() string {
	return this.UserName
}

func (this *AvatarSession) GetId() util.ID {
	return this.Id
}

func (this *AvatarSession) fetchUserName(a lokas.IProcess) error {
	user := &AvatarMap{}
	err := a.GetMongo().Collection("avatarmap").Find(context.TODO(), bson.M{"_id": this.Id}).One(user)
	if err != nil || user.Id == 0 {
		return err
	}
	log.Warn("fetch user", flog.UserName(user.UserName), flog.GameId(user.GameId), flog.ServerId(user.ServerId))
	this.UserName = user.UserName
	return nil
}

func (this *AvatarSession) Deserialize(a lokas.IProcess) error {
	etcd := a.GetEtcd()
	key := AVATAR_SESSION_KEY.Assemble(this.Id)
	mutexKey := MUTEX_AVATAR_SESSION_KEY.Assemble(this.Id)
	mutex, err := a.GlobalMutex(mutexKey, 6)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	res, err := etcd.Get(context.TODO(), key)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if len(res.Kvs) == 0 {
		return errors.New("")
	}
	sess := NewAvatarSession(0)
	err = json.Unmarshal(res.Kvs[0].Value, sess)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.UserName = sess.UserName
	this.GameId = sess.GameId
	this.ServerId = sess.ServerId
	this.GateId = sess.GateId
	log.Warn("AvatarSession Deserialize", flog.AvatarSessionInfo(this).Append(zap.Any("res", res.Kvs))...)
	return nil
}

func (this *AvatarSession) FetchData(a lokas.IProcess) error {
	err := this.fetchUserName(a)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	avatar := &AvatarMap{}
	err = a.GetMongo().Collection("avatarmap").Find(context.TODO(), bson.M{"_id": this.Id}).One(avatar)
	if err != nil {
		// TODO create avatar
		log.Error(err.Error())
		return err
	}
	this.UserName = avatar.UserName
	this.GameId = avatar.GameId
	this.ServerId = avatar.ServerId
	return nil
}

func (this *AvatarSession) SetGateId(id util.ID) {
	this.GateId = id
}

func (this *AvatarSession) Serialize(a lokas.IProcess) error {
	log.Info("Serialize AvatarSession", flog.AvatarSessionInfo(this)...)
	etcd := a.GetEtcd()
	key := AVATAR_SESSION_KEY.Assemble(this.Id)
	mutexKey := MUTEX_AVATAR_SESSION_KEY.Assemble(this.Id)
	s, _ := json.Marshal(this)
	log.Warn("AvatarSession:Serialize", flog.AvatarSessionInfo(this).Append(zap.String("key", key))...)
	mutex, err := a.GlobalMutex(mutexKey, 6)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	res, err := etcd.Put(context.TODO(), key, string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.watchChan = etcd.Watch(context.TODO(), key, clientv3.WithRev(res.Header.Revision))
	return nil
}
