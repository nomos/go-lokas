package lox

import (
	"context"
	"sync"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
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
	// this.MsgChan = make(chan *protocol.RouteMessage, 100)
	// this.ReplyChan = make(chan *protocol.RouteMessage, 100)
	this.DoneChan = make(chan struct{})

	// this.DataChan = make(chan *protocol.RouteRecv, 100)
	// this.ReplyDataChan = make(chan *protocol.RouteRecv, 100)

	go func() {
	MSG_LOOP:
		for {
			select {
			case <-this.Timer.C:
				// log.Debug("timer", zap.String("type", this.Type()), zap.Uint64("actorId", uint64(this.GetId())))
				this.Update(0, time.Now())
			case rMsg := <-this.MsgChan:
				this.OnMessage(rMsg)
			case recv := <-this.DataChan:
				this.OnRecvData(recv)
			case msg := <-this.TimeHandler.EventChan():
				out := msg.(*timer.TimeEventMsg)
				out.Callback(out.TimeNoder)
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

		log.Debug("actor stop ", zap.String("type", this.Type()), zap.Uint64("actorId", uint64(this.GetId())))
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

	// log.Debug("receive data", zap.Uint64("actorId", uint64(this.GetId())), zap.Uint16("cmd", recv.GetCmd()))

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
			flog.ActorInfo(this).
				Concat(flog.ErrorInfo(msg.(*protocol.ErrMsg))).
				Concat(flog.MsgInfo(msg)).
				Append(flog.FromActorId(actorId)).
				Append(flog.TransId(transId))...,
		)
	} else {
		log.Info("Actor:handleMsg", flog.ActorReceiveMsgInfo(this, msg, transId, actorId)...)
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
				flog.ActorReceiveMsgInfo(this, msg.Body, msg.TransId, msg.FromActor).
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
		log.Error("route msg unmarsh err", zap.String("actorType", this.Type()), zap.Uint64("actorId", uint64(this.GetId())), zap.Uint16("cmd", uint16(dataMsg.Cmd)), zap.String("err", err.Error()))
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
	if err != nil {
		log.Error("handle msg err", zap.String("actorType", this.Type()), zap.Uint64("actorId", uint64(this.GetId())), zap.Uint16("cmd", uint16(dataMsg.Cmd)), zap.Any("body", body), zap.String("err", err.Error()))
	}

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
			log.Warn("DeadlineExceeded", flog.ActorSendMsgInfo(this, req, transId, actorId)...)
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
