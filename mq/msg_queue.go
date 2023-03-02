package mq

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

type MsgQueue struct {
	nc *nats.Conn

	url          string
	protocolType protocol.TYPE

	actorSubscriberMap map[util.ID]*ActorSubscriber
	muActor            sync.RWMutex
}

type MQHead struct {
	Len   uint16
	CmdId protocol.BINARY_TAG
	Body  []byte
}

func (this *MsgQueue) Flush() error {
	return this.nc.Flush()
}

func (this *MsgQueue) UnmarshalHead(data []byte) (*MQHead, error) {
	// head := &MQHead{
	// 	Len:   binary.LittleEndian.Uint16(data[0:2]),
	// 	CmdId: protocol.BINARY_TAG(binary.LittleEndian.Uint16(data[2:4])),
	// 	Body:  data[4:],
	// }

	// return head, nil

	return UnmarshalHead(data)
}

func (this *MsgQueue) UnmarshalMsg(data []byte) (protocol.ISerializable, error) {
	// _ := binary.LittleEndian.Uint16(data[0:2])
	// cmdId := protocol.BINARY_TAG(binary.LittleEndian.Uint16(data[2:4]))

	// body := data[4:]

	// head, _ := this.UnmarshalHead(data)

	// return protocol.UnmarshJsonBody(head.CmdId, head.Body)

	return UnmarshalMsg(data)
}

func (this *MsgQueue) MarshalMsg(msg protocol.ISerializable) ([]byte, error) {
	// var out bytes.Buffer

	// cmdId, _ := msg.GetId()

	// w(&out, uint16(0))
	// w(&out, uint16(cmdId))

	// data, _ := json.Marshal(msg)
	// w(&out, data)

	// ret := out.Bytes()
	// binary.LittleEndian.PutUint16(ret[0:2], uint16(out.Len()))

	// return ret, nil

	return MarshalMsg(msg)
}

func w(out io.Writer, v interface{}) {
	err := binary.Write(out, binary.LittleEndian, v)
	if err != nil {
		log.Panic(err.Error())
	}
}

func (this *MsgQueue) addActorSubscriber(actor lokas.IActor, ch chan *nats.Msg) (*ActorSubscriber, error) {
	if actor.GetId() <= 0 {
		return nil, protocol.ERR_ACTOR_ID_INVALID
	}

	this.muActor.Lock()
	defer this.muActor.Unlock()

	if _, ok := this.actorSubscriberMap[actor.GetId()]; ok {
		log.Error("mq actor add duplicate", zap.Int64("actorid", actor.GetId().Int64()))
		return nil, protocol.ERR_MQ_ADD_ACTOR_DUPLICATE
	}

	v := &ActorSubscriber{
		ActorId: actor.GetId(),
		Pid:     actor.PId(),
		MsgChan: ch,

		mq:     this,
		subMap: make(map[string]*nats.Subscription),
	}

	this.actorSubscriberMap[v.ActorId] = v

	return v, nil
}

func (this *MsgQueue) delActorSubscriber(sub *ActorSubscriber) {

	ins.muActor.Lock()
	defer ins.muActor.Unlock()

	delete(ins.actorSubscriberMap, sub.ActorId)

	for _, v := range sub.subMap {
		v.Unsubscribe()
		v.Drain()
	}

	sub.MsgChan = nil
	sub.subMap = nil
}
