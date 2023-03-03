package mq

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

type ActorSubscriber struct {
	ActorId util.ID
	Pid     util.ProcessId
	MsgChan chan *nats.Msg

	mq     *MsgQueue
	subMap map[string]*nats.Subscription
}

func (this *ActorSubscriber) Subscribe(subj string) error {

	ns, err := this.mq.nc.ChanSubscribe(subj, this.MsgChan)
	if err != nil {
		log.Error("mq actor add subscribe err", zap.Int64("actorid", this.ActorId.Int64()), zap.String("subj", subj), zap.String("error", err.Error()))
		return protocol.ERR_MQ_ERROR
	}

	this.subMap[subj] = ns

	return nil
}

func (this *ActorSubscriber) Unsubscribe(subj string) error {

	v, ok := this.subMap[subj]
	if !ok {
		return protocol.ERR_MQ_SUBJ_NOT_FIND
	}

	delete(this.subMap, subj)

	v.Unsubscribe()
	v.Drain()

	return nil
}

func (this *ActorSubscriber) UnSubscribeAll() {

	for _, v := range this.subMap {
		v.Unsubscribe()
		v.Drain()
	}

	this.subMap = make(map[string]*nats.Subscription)
}

func (this *ActorSubscriber) SubscribeForActor() error {
	subj := fmt.Sprintf(KEY_ACTOR_BASE, this.ActorId)

	return this.Subscribe(subj)
}

// func (this *ActorSubscriber) SubscribeForAvatar() error {
// 	subj := fmt.Sprintf("actor.pid.%d", this.Pid)

// 	return this.Subscribe(subj)
// }

func (this *ActorSubscriber) SubscribeForService(serviceType string, serviceId uint16) error {
	subj := fmt.Sprintf("service.%s.%d", serviceType, serviceId)
	return this.Subscribe(subj)
}

func (this *ActorSubscriber) Drain() {
	this.mq.delActorSubscriber(this)
}

func (this *ActorSubscriber) PublishToActorOne(actorId util.ID, msg protocol.ISerializable) error {
	return PublishToActorOne(actorId, msg)
}

func (this *ActorSubscriber) PublishToService(serviceType string, serviceId uint16, msg protocol.ISerializable) error {

	return PublishToService(serviceType, serviceId, msg)

}
