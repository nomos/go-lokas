package mq

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

const (
	KEY_ACTOR_BASE = "actor.one.%d"

	// KEY_ACTOR_PID = "actor.pid.%d"

	KEY_SERVICE_ID = "service.%s.%d"
)

var once sync.Once
var ins *MsgQueue

func Instance() *MsgQueue {
	once.Do(func() {
		if ins == nil {
			ins = &MsgQueue{
				actorSubscriberMap: make(map[util.ID]*ActorSubscriber),
			}
		}
	})
	return ins
}

func Init(config lokas.IConfig) error {

	if config == nil {
		log.Warn("mq not find config")
		return protocol.ERR_CONFIG_ERROR
	}

	url := config.GetString("url")
	if url == "" {
		log.Error("mq config not find url")
		return protocol.ERR_CONFIG_ERROR
	}

	protocolType := protocol.JSON
	protocolName := config.GetString("protocolType")
	if protocolName != "" {
		protocolType = protocol.String2Type(protocolName)
	}

	conn, err := nats.Connect(url,
		// nats.RetryOnFailedConnect(true),
		// nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			log.Debug("nats disconnected ", zap.String("err", err.Error()))
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			log.Debug("nats reconnect to ", zap.String("url", c.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			log.Debug("nats connection closed ", zap.String("err", c.LastError().Error()))
		}),
		nats.ConnectHandler(func(c *nats.Conn) {
			log.Debug("nats connect succ ", zap.String("url", c.ConnectedUrl()))
		}),
	)

	if err != nil {
		log.Error("init mq connect err", zap.String("url", url), zap.String("err", err.Error()))
		return err
	}

	Instance()

	ins.nc = conn
	ins.url = url
	ins.protocolType = protocolType

	return nil
}

func MarshalMsg(msg protocol.ISerializable) ([]byte, error) {
	var out bytes.Buffer

	cmdId, _ := msg.GetId()

	w(&out, uint16(0))
	w(&out, uint16(cmdId))

	data, _ := json.Marshal(msg)
	w(&out, data)

	ret := out.Bytes()
	binary.LittleEndian.PutUint16(ret[0:2], uint16(out.Len()))

	return ret, nil
}

func UnmarshalHead(data []byte) (*MQHead, error) {
	head := &MQHead{
		Len:   binary.LittleEndian.Uint16(data[0:2]),
		CmdId: protocol.BINARY_TAG(binary.LittleEndian.Uint16(data[2:4])),
		Body:  data[4:],
	}

	return head, nil
}

func UnmarshalMsg(data []byte) (protocol.ISerializable, error) {
	head, _ := UnmarshalHead(data)

	return protocol.UnmarshJsonBody(head.CmdId, head.Body)
}

func GetProtocolType() (protocol.TYPE, error) {
	if ins == nil {
		log.Debug("mq not connect")
		return protocol.JSON, protocol.ERR_MQ_NOT_CONNECT
	}

	return ins.protocolType, nil
}

func CreateActorSubscriber(actor lokas.IActor, ch chan *nats.Msg) (*ActorSubscriber, error) {
	if ins == nil {
		// log.Debug("mq not connect")
		return nil, protocol.ERR_MQ_NOT_CONNECT
	}

	sub, err := ins.addActorSubscriber(actor, ch)
	if err != nil {
		return nil, err
	}

	sub.SubscribeForActor()

	log.Debug("add mq subscribe", zap.String("type", actor.Type()), zap.Uint64("actorId", uint64(actor.GetId())))

	return sub, nil
}

func TryReplyMessage(mqMsg *nats.Msg, msg protocol.ISerializable) error {
	if mqMsg.Reply == "" {
		return nil
	}

	out, err := ins.MarshalMsg(msg)
	if err != nil {
		cmdId, _ := msg.GetId()
		log.Error("mq marshal msg err", zap.Uint16("cmdId", uint16(cmdId)))
		return protocol.ERR_MQ_MARSHAL_ERROR
	}

	return ins.nc.Publish(mqMsg.Reply, out)
}

func RequestToActorOne(actorId util.ID, msg protocol.ISerializable) (protocol.ISerializable, protocol.ErrCode) {
	if ins == nil {
		log.Warn("mq not connect")
		return nil, protocol.ERR_MQ_NOT_CONNECT
	}
	data, err := ins.MarshalMsg(msg)
	if err != nil {
		return nil, protocol.ERR_MQ_MARSHAL_ERROR
	}
	mqMsg, err := ins.nc.Request(fmt.Sprintf(KEY_ACTOR_BASE, actorId), data, 5*time.Second)
	if err != nil {
		log.Warn("nats request err", zap.Uint64("actorId", uint64(actorId)), zap.String("err", err.Error()))
		return nil, protocol.ERR_MQ_REQUEST_ERROR
	}

	recvMsg, err := ins.UnmarshalMsg(mqMsg.Data)
	if err != nil {
		return nil, protocol.ERR_MQ_UNMARSHAL_ERROR
	}

	if v, ok := recvMsg.(*protocol.ErrMsg); ok {
		return nil, protocol.ErrCode(v.Code)
	} else {
		return recvMsg, protocol.ERR_SUCC
	}

}

func Publsih(key string, msg protocol.ISerializable) error {
	if ins == nil {
		log.Warn("mq not connect")
		return protocol.ERR_MQ_NOT_CONNECT
	}

	data, err := ins.MarshalMsg(msg)
	if err != nil {
		return err
	}

	err = ins.nc.Publish(key, data)
	if err != nil {
		log.Error("mq publish err", zap.String("key", key), zap.Any("msg", msg))
		return protocol.ERR_MQ_ERROR
	}

	return nil
}

func Flush() error {
	if ins == nil {
		log.Warn("mq not connect")
		return protocol.ERR_MQ_NOT_CONNECT
	}

	return ins.nc.Flush()
}

func PublishToActorOne(actorId util.ID, msg protocol.ISerializable) error {
	if ins == nil {
		log.Warn("mq not connect")
		return protocol.ERR_MQ_NOT_CONNECT
	}

	data, err := ins.MarshalMsg(msg)
	if err != nil {
		return protocol.ERR_MQ_MARSHAL_ERROR
	}
	return ins.nc.Publish(fmt.Sprintf(KEY_ACTOR_BASE, actorId), data)
}

func PublishToService(serviceType string, serviceId uint16, msg protocol.ISerializable) error {
	if ins == nil {
		log.Warn("mq not connect")
		return protocol.ERR_MQ_NOT_CONNECT
	}

	data, err := ins.MarshalMsg(msg)
	if err != nil {
		return protocol.ERR_MQ_MARSHAL_ERROR
	}
	return ins.nc.Publish(fmt.Sprintf(KEY_SERVICE_ID, serviceType, serviceId), data)
}
