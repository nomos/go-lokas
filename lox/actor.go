package lox

import (
	"context"
	"fmt"
	"github.com/nomos/go-lokas/log/flog"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/mq"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/timer"
	"github.com/nomos/go-lokas/util"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	TimeOut                  = time.Second * 120
	UpdateTime               = time.Second * 15
	LeaseDuration      int64 = 23
	LeaseRenewDuration int64 = 15
)

//Actor是基础的游戏物体,一个物体只有一个可寻址实例

var _ lokas.IActor = (*Actor)(nil)

func NewActor() *Actor {
	ctx, cancel := context.WithCancel(context.Background())

	ret := &Actor{
		IEntity:     ecs.CreateEntity(),
		ReqContexts: make(map[uint32]lokas.IReqContext),
		Timer:       time.NewTicker(UpdateTime),
		Timeout:     TimeOut,
		TimeHandler: timer.NewHandler(),

		MsgChan:   make(chan *protocol.RouteMessage, 100),
		ReplyChan: make(chan *protocol.RouteMessage, 100),

		DataChan:      make(chan *protocol.RouteDataMsg, 100),
		ReplyDataChan: make(chan *protocol.RouteDataMsg, 100),

		Ctx:    ctx,
		Cancel: cancel,
	}
	ret.SetType("Actor")
	return ret
}

type Actor struct {
	lokas.IEntity
	timer.TimeHandler
	typeString   string
	process      lokas.IProcess
	idGen        uint32
	leaseId      clientv3.LeaseID
	MsgChan      chan *protocol.RouteMessage
	ReplyChan    chan *protocol.RouteMessage
	ReqContexts  map[uint32]lokas.IReqContext
	Timeout      time.Duration
	CtxMutex     sync.Mutex
	Timer        *time.Ticker
	DoneChan     chan struct{}
	OnUpdateFunc func()
	MsgHandler   func(actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error)

	DataChan      chan *protocol.RouteDataMsg
	ReplyDataChan chan *protocol.RouteDataMsg

	Ctx    context.Context
	Cancel context.CancelFunc

	MQChan chan *nats.Msg
	Sub    *mq.ActorSubscriber

	isStarted bool
}

func (this *Actor) LogInfo() log.ZapFields {
	return lokas.LogActorInfo(this)
}

func (this *Actor) Send(id util.ProcessId, msg *protocol.RouteMessage) error {
	return this.process.Send(id, msg)
}

func (this *Actor) Load(conf lokas.IConfig) error {
	panic("implement me")
}

func (this *Actor) Unload() error {
	panic("implement me")
}

func (this *Actor) Start() error {
	return nil
}

func (this *Actor) Stop() error {
	// this.DoneChan <- struct{}{}

	this.Cancel()
	return nil
}

func (this *Actor) OnStart() error {
	return nil
}

func (this *Actor) OnStop() error {
	return nil
}

func (this *Actor) Type() string {
	return this.typeString
}

func (this *Actor) SetType(s string) {
	this.typeString = s
}

//return leaseId,(bool)is registered,error
func (this *Actor) GetLeaseId() (clientv3.LeaseID, bool, error) {
	c := this.process.GetEtcd()
	if this.leaseId != 0 {
		resToLive, err := c.Lease.TimeToLive(context.Background(), this.leaseId)
		if err != nil {
			log.Error(err.Error())
			return 0, false, err
		}
		//if lease id is expired,create a new lease id
		if resToLive.TTL <= 0 {
			res, err := c.Lease.Grant(context.Background(), LeaseDuration)
			if err != nil {
				log.Error(err.Error())
				return 0, false, err
			}
			this.leaseId = res.ID
			return this.leaseId, false, nil
		}
		if resToLive.TTL < LeaseRenewDuration {
			_, err := c.Lease.KeepAliveOnce(context.Background(), this.leaseId)
			if err != nil {
				log.Error(err.Error())
				return 0, false, err
			}
		}
		return this.leaseId, true, nil
	}

	res, err := c.Lease.Grant(context.Background(), LeaseDuration)
	if err != nil {
		log.Error(err.Error())
		return 0, false, err
	}
	this.leaseId = res.ID
	return this.leaseId, false, nil
}

func (this *Actor) Update(dt time.Duration, now time.Time) {
	if this.OnUpdateFunc != nil {
		this.OnUpdateFunc()
	}
}

// func (this *Actor) GetTimeHandler() timer.TimeHandler {
// 	return this.TimeHandler
// }

func (this *Actor) GetProcess() lokas.IProcess {
	return this.process
}

func (this *Actor) SetProcess(process lokas.IProcess) {
	this.process = process
	log.SetDefaultLogger(process.GetLogger())
}

func (this *Actor) StartMessagePump() {

	if this.isStarted {
		log.Warn("actor has started message pump", this.LogInfo()...)
		return
	}

	this.isStarted = true

	// this.MsgChan = make(chan *protocol.RouteMessage, 100)
	// this.ReplyChan = make(chan *protocol.RouteMessage, 100)
	this.DoneChan = make(chan struct{})

	// this.DataChan = make(chan *protocol.RouteRecv, 100)
	// this.ReplyDataChan = make(chan *protocol.RouteRecv, 100)

	this.MQChan = make(chan *nats.Msg, 128)

	this.CreateSubscriber()

	go func() {
	MSG_LOOP:
		for {
			select {
			case <-this.Timer.C:
				this.Update(0, time.Now())
			case rMsg := <-this.MsgChan:
				this.OnMessage(rMsg)
			case recv := <-this.DataChan:
				this.OnRecvData(recv)
			case msg := <-this.TimeHandler.EventChan():
				out := msg.(*timer.TimeEventMsg)
				out.Callback(out.TimeNoder)
			case mqMsg := <-this.MQChan:
				this.OnRecvMQ(mqMsg)
			case <-this.DoneChan:
				break MSG_LOOP
			case <-this.Ctx.Done():
				break MSG_LOOP
			}
		}
		close(this.MsgChan)
		this.MsgChan = nil
		close(this.DoneChan)
		this.DoneChan = nil

		this.TimeHandler.DelTimer()
		this.TimeHandler = nil

		if this.Sub != nil {
			this.Sub.Drain()
			this.Sub = nil
		}

		close(this.MQChan)
		this.MQChan = nil

		log.Debug("actor stop ", this.LogInfo()...)
	}()

	go func() {
	REP_LOOP:
		for {
			select {
			case rMsg := <-this.ReplyChan:
				this.OnMessage(rMsg)
			case recv := <-this.ReplyDataChan:
				this.OnRecvData(recv)
			case <-this.DoneChan:
				break REP_LOOP
			case <-this.Ctx.Done():
				break REP_LOOP
			}
		}
		close(this.ReplyChan)
		this.ReplyChan = nil
	}()

}

func (this *Actor) ReceiveMessage(msg *protocol.RouteMessage) {
	if msg.Req {
		this.MsgChan <- msg
	} else {
		this.ReplyChan <- msg
	}
}

func (this *Actor) ReceiveData(msg *protocol.RouteDataMsg) error {

	if msg.ReqType == protocol.REQ_TYPE_REPLAY {
		this.ReplyDataChan <- msg
	} else {
		this.DataChan <- msg
	}

	return nil
}

func (this *Actor) clearContext(err error) {
	this.CtxMutex.Lock()
	defer this.CtxMutex.Unlock()
	for _, v := range this.ReqContexts {
		v.Cancel(err)
	}
	this.ReqContexts = map[uint32]lokas.IReqContext{}
}

func (this *Actor) addContext(transId uint32, ctx lokas.IReqContext) {
	this.CtxMutex.Lock()
	defer this.CtxMutex.Unlock()
	this.ReqContexts[transId] = ctx
}

func (this *Actor) removeContext(transId uint32) {
	this.CtxMutex.Lock()
	defer this.CtxMutex.Unlock()
	delete(this.ReqContexts, transId)
}

func (this *Actor) PId() util.ProcessId {
	if this.process == nil {
		return 0
	}
	return this.process.PId()
}

func (this *Actor) getContext(transId uint32) lokas.IReqContext {
	this.CtxMutex.Lock()
	defer this.CtxMutex.Unlock()
	return this.ReqContexts[transId]
}

func (this *Actor) genId() uint32 {
	this.idGen++
	return this.idGen
}

func (this *Actor) HookReceive(msg *protocol.RouteMessage) *protocol.RouteMessage {
	ctx := this.getContext(msg.TransId)
	if ctx != nil {
		ctx.SetResp(msg.Body)
		ctx.Finish()
		return nil
	}
	return msg
}

func (this *Actor) HandleMsg(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	id, err := msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if id == protocol.TAG_Error {
		log.Warn("Actor:handleMsg:errmsg",
			this.LogInfo().
				Concat(protocol.LogErrorInfo(msg.(*protocol.ErrMsg))).
				Concat(protocol.LogMsgInfo(msg)).
				Append(flog.FromActorId(actorId)).
				Append(flog.TransId(transId))...,
		)
	} else {
		log.Debug("Actor:handleMsg", lokas.LogActorReceiveMsgInfo(this, msg, transId, actorId)...)
	}
	if this.MsgHandler != nil {
		resp, err := this.MsgHandler(actorId, transId, msg)
		if err != nil {
			if protocol.ERR_ACTOR_NOT_FOUND.Is(err) {
				log.Error("ERR_ACTOR_NOT_FOUND")
				return protocol.ERR_ACTOR_NOT_FOUND
			}
			if uerr, ok := err.(protocol.IError); ok {
				if transId != 0 {
					this.SendReply(actorId, transId, protocol.NewErrorMsg(int32(uerr.ErrCode()), uerr.Error()))
				}
			} else {
				if transId != 0 {
					this.SendReply(actorId, transId, protocol.ERR_INTERNAL_ERROR.NewErrMsg())
				}
			}
			return err
		}
		if resp != nil {
			if transId != 0 {
				err := this.SendReply(actorId, transId, resp)
				if err != nil {
					log.Error(err.Error())
					return err
				}
			}
		}
	}
	return nil
}

func (this *Actor) OnMessage(msg *protocol.RouteMessage) {
	if !msg.Req {
		this.HookReceive(msg)
		return
	}
	if msg != nil && msg.Req {
		err := this.HandleMsg(msg.FromActor, msg.TransId, msg.Body)
		if err != nil {
			log.Error("Actor:OnMessage:Error",
				msg.LogInfo().
					Append(flog.Error(err))...,
			)
		}
	}
}

func (this *Actor) OnRecvData(dataMsg *protocol.RouteDataMsg) {
	// rMsg, err := protocol.UnmarshalRouteMsg(recv.Data, recv.Protocol)
	// if err != nil {
	// 	log.Error("route msg unmarsh err", zap.Int("len", len(recv.Data)))
	// }
	// rMsg.FromActor = recv.FromActor
	// this.OnMessage(rMsg)

	body, err := dataMsg.UnmarshalData()
	if err != nil {
		log.Error("route msg unmarsh err", this.LogInfo().Append(protocol.LogCmdId(dataMsg.Cmd)).Append(flog.Error(err))...)
		return
	}

	if dataMsg.ReqType == protocol.REQ_TYPE_REPLAY {
		ctx := this.getContext(dataMsg.TransId)
		if ctx != nil {
			ctx.SetResp(body)
			ctx.Finish()
		}
		return
	}

	err = this.HandleMsg(dataMsg.FromActor, dataMsg.TransId, body)

}

func (this *Actor) OnRecvMQ(mqMsg *nats.Msg) (errRet error) {
	// var retMsg protocol.ISerializable
	defer func() {
		if errRet == nil {
			return
		}

		var errMsg *protocol.ErrMsg
		errCode, ok := errRet.(protocol.ErrCode)
		if !ok {
			desc := fmt.Sprintf("%s(%s)", errCode.Error(), errRet.Error())
			errMsg = protocol.NewErrorMsg(int32(protocol.ERR_MQ_ERROR), desc)
		} else {
			errMsg = errCode.NewErrMsg()
		}
		mq.TryReplyMessage(mqMsg, errMsg)
	}()

	if this.MsgHandler == nil {
		return protocol.ERR_MSG_HANDLER_NOT_FOUND
	}
	recvMsg, err := mq.UnmarshalMsg(mqMsg.Data)
	if err != nil {
		return protocol.ERR_MQ_UNMARSHAL_ERROR
	}

	retMsg, err := this.MsgHandler(0, 0, recvMsg)

	if mqMsg.Reply == "" {
		return nil
	}

	if err != nil {
		if v, ok := err.(protocol.ErrCode); ok {
			retMsg = v.NewErrMsg()
		} else {
			retMsg = protocol.ERR_INTERNAL_ERROR.NewErrMsg()
		}
	}

	if retMsg == nil {
		return protocol.ERR_MQ_NOT_RELPY
	}
	mq.TryReplyMessage(mqMsg, retMsg)

	return nil
}

func (this *Actor) SendReply(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	_, err := msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// log.Info("Actor:SendReply", flog.ActorSendMsgInfo(this, msg, transId, actorId)...)
	isReq := false
	if transId == 0 {
		isReq = true
	}
	routeMsg := protocol.NewRouteMessage(this.GetId(), actorId, transId, msg, isReq)
	this.process.RouteMsg(routeMsg)
	return nil
}

func (this *Actor) SendMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	_, err := msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// log.Info("Actor:SendMessage", flog.ActorSendMsgInfo(this, msg, transId, actorId)...)
	routeMsg := protocol.NewRouteMessage(this.GetId(), actorId, transId, msg, true)
	this.process.RouteMsg(routeMsg)
	return nil
}

func (this *Actor) Call(actorId util.ID, req protocol.ISerializable) (protocol.ISerializable, error) {
	ctx := network.NewDefaultContextWithTimeout(context.TODO(), this.genId(), this.Timeout)
	transId := ctx.GetTransId()
	this.addContext(transId, ctx)
	err := this.SendMessage(actorId, transId, req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			log.Warn("DeadlineExceeded", lokas.LogActorSendMsgInfo(this, req, transId, actorId)...)
			this.removeContext(transId)
			return nil, protocol.ERR_RPC_TIMEOUT
		default:
			this.removeContext(transId)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			resp := ctx.GetResp().(protocol.ISerializable)
			if err, ok := resp.(*protocol.ErrMsg); ok {
				return nil, err
			}
			return ctx.GetResp().(protocol.ISerializable), nil
		}
	}
}

func (this *Actor) CreateSubscriber() (err error) {
	this.Sub, err = mq.CreateActorSubscriber(this, this.MQChan)
	return err
}

func (this *Actor) Subscribe(key string) error {
	if strings.HasPrefix(key, "actor") || strings.HasPrefix(key, "service") {

		log.Error("mq subscribe err, key not begin with 'actor' or 'service'", this.LogInfo().Append(zap.String("key", key))...)
		return protocol.ERR_MQ_SUBJ_PREFIX_ERR
	}

	if this.Sub == nil {
		log.Warn("actor is not create subscriber", this.LogInfo().Append(zap.String("key", key))...)
		return nil
	}

	return this.Sub.Subscribe(key)
}

func (this *Actor) Unsubscribe(key string) error {
	if strings.HasPrefix(key, "actor") || strings.HasPrefix(key, "service") {

		log.Error("mq unsubscribe err, key not begin with 'actor' or 'service'", this.LogInfo().Append(zap.String("key", key))...)
		return protocol.ERR_MQ_SUBJ_PREFIX_ERR
	}

	if this.Sub == nil {
		log.Warn("actor is not create subscriber", this.LogInfo().Append(zap.String("key", key))...)
		return nil
	}

	return this.Sub.Unsubscribe(key)
}

func (this *Actor) Publish(key string, msg protocol.ISerializable) error {
	return mq.Publsih(key, msg)
}

func (this *Actor) Request(key string, msg protocol.ISerializable) (protocol.ISerializable, error) {
	return mq.Request(key, msg)
}
