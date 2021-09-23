package lox

import (
	"context"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/ecs"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/network"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

const (
	TimeOut            = time.Second * 15
	UpdateTime = time.Second*15
	LeaseDuration int64 = 23
	LeaseRenewDuration int64 = 15
)

//Actor是基础的游戏物体,一个物体只有一个可寻址实例

var _ lokas.IActor = (*Actor)(nil)


func NewActor() *Actor {
	ret := &Actor{
		IEntity:ecs.CreateEntity(),
		reqContexts: make(map[uint32]lokas.IReqContext),
		timer:time.NewTicker(UpdateTime),
		timeout:     TimeOut,
	}
	ret.SetType("Actor")
	return ret
}

type Actor struct {
	lokas.IEntity
	typeString string
	process     lokas.IProcess
	msgChan     chan *protocol.RouteMessage
	idGen       uint32
	reqContexts map[uint32]lokas.IReqContext
	timeout     time.Duration
	ctxMutex    sync.Mutex
	timer 		*time.Ticker
	done        chan struct{}
	leaseId     clientv3.LeaseID
	OnUpdateFunc func()
	MsgHandler  func(actorId util.ID, transId uint32, msg protocol.ISerializable) (protocol.ISerializable, error)
}

func (this *Actor) Send(id util.ProcessId, msg *protocol.RouteMessage) error {
	return this.process.Send(id,msg)
}

func (this *Actor) Load(conf lokas.IConfig) error {
	panic("implement me")
}

func (this *Actor) Unload() error {
	panic("implement me")
}

func (this *Actor) Start() error {
	panic("implement me")
}

func (this *Actor) Stop() error {
	panic("implement me")
}

func (this *Actor) OnStart() error {
	panic("implement me")
}

func (this *Actor) OnStop() error {
	panic("implement me")
}

func (this *Actor) Type()string{
	return this.typeString
}

func (this *Actor) SetType(s string){
	this.typeString = s
}

//return leaseId,(bool)is registered,error
func (this *Actor) GetLeaseId() (clientv3.LeaseID,bool, error) {
	c := this.process.GetEtcd()
	if this.leaseId!= 0 {
		resToLive,err:= c.Lease.TimeToLive(context.Background(),this.leaseId)
		if err != nil {
			log.Error(err.Error())
			return 0,false,err
		}
		//if lease id is expired,create a new lease id
		if resToLive.TTL<=0 {
			res, err := c.Lease.Grant(context.Background(), LeaseDuration)
			if err != nil {
				log.Error(err.Error())
				return 0,false, err
			}
			this.leaseId = res.ID
			return this.leaseId,false, nil
		}
		if resToLive.TTL<LeaseRenewDuration {
			_,err := c.Lease.KeepAliveOnce(context.Background(),this.leaseId)
			if err != nil {
				log.Error(err.Error())
				return 0,false,err
			}
		}
		return this.leaseId,true,nil
	}

	res, err := c.Lease.Grant(context.Background(), LeaseDuration)
	if err != nil {
		log.Error(err.Error())
		return 0,false, err
	}
	this.leaseId = res.ID
	return this.leaseId,false, nil
}

func (this *Actor) Update(dt time.Duration, now time.Time) {
	if this.OnUpdateFunc!=nil {
		this.OnUpdateFunc()
	}
}

func (this *Actor) GetProcess() lokas.IProcess {
	return this.process
}

func (this *Actor) SetProcess(process lokas.IProcess) {
	this.process = process
	log.SetDefaultLogger(process.GetLogger())
}

func (this *Actor) StartMessagePump() {
	this.msgChan = make(chan *protocol.RouteMessage, 100)
	this.done = make(chan struct{})
	go func() {
		LOOP:
		for {
			select {
			case <-this.timer.C:
				this.Update(0,time.Now())
			case rMsg := <-this.msgChan:
				this.OnMessage(rMsg)
			case <-this.done:
				break LOOP
			}
		}
		close(this.msgChan)
		close(this.done)
	}()
}

func (this *Actor) ReceiveMessage(msg *protocol.RouteMessage) {
	this.msgChan <- msg
}

func (this *Actor) clearContext(err error) {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	for _, v := range this.reqContexts {
		v.Cancel(err)
	}
	this.reqContexts = map[uint32]lokas.IReqContext{}
}

func (this *Actor) addContext(transId uint32, ctx lokas.IReqContext) {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	this.reqContexts[transId] = ctx
}

func (this *Actor) removeContext(transId uint32) {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	delete(this.reqContexts, transId)
}

func (this *Actor) PId()util.ProcessId{
	if this.process==nil {
		return 0
	}
	return this.process.PId()
}

func (this *Actor) getContext(transId uint32) lokas.IReqContext {
	this.ctxMutex.Lock()
	defer this.ctxMutex.Unlock()
	return this.reqContexts[transId]
}

func (this *Actor) genId() uint32 {
	this.idGen++
	return this.idGen
}

func (this *Actor) hookReceive(msg *protocol.RouteMessage) *protocol.RouteMessage {
	ctx := this.getContext(msg.TransId)
	if ctx != nil {
		ctx.SetResp(msg.Body)
		ctx.Finish()
		return nil
	}
	return msg
}

func (this *Actor) handleMsg(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	id,err:=msg.GetId()
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
			Append(flog.TransId(transId))...
		)
	} else {
		log.Info("Actor:handleMsg",flog.ActorReceiveMsgInfo(this,msg,transId,actorId)...)
	}
	if this.MsgHandler != nil {
		resp, err := this.MsgHandler(actorId, transId, msg)
		if err != nil {
			if protocol.ERR_ACTOR_NOT_FOUND.Is(err) {
				log.Error("ERR_ACTOR_NOT_FOUND")
				return protocol.ERR_ACTOR_NOT_FOUND
			}
			if uerr, ok := err.(protocol.IError); ok {
				if transId!=0 {
					this.SendReply(actorId, transId, protocol.NewErrorMsg(int32(uerr.ErrCode()), uerr.Error()))
				}
			} else {
				if transId!=0 {
					this.SendReply(actorId, transId, protocol.ERR_INTERNAL_ERROR.NewErrMsg())
				}
			}
			return err
		}
		if resp != nil {
			if transId!=0 {
				err:=this.SendReply(actorId, transId, resp)
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
	msg = this.hookReceive(msg)
	if msg != nil {
		err:=this.handleMsg(msg.FromActor, msg.TransId, msg.Body)
		if err != nil {
			log.Error("Actor:OnMessage:Error",
				flog.ActorReceiveMsgInfo(this,msg.Body,msg.TransId,msg.FromActor).
				Append(flog.Error(err))...
			)
		}
	}
}

func (this *Actor) SendReply(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	_,err:=msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("Actor:SendReply",flog.ActorSendMsgInfo(this,msg,transId,actorId)...)
	routeMsg := protocol.NewRouteMessage(this.GetId(), actorId, transId, msg, false)
	this.process.RouteMsg(routeMsg)
	return nil
}

func (this *Actor) SendMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error {
	_,err:=msg.GetId()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("Actor:SendMessage",flog.ActorSendMsgInfo(this,msg,transId,actorId)...)
	routeMsg := protocol.NewRouteMessage(this.GetId(), actorId, transId, msg, true)
	this.process.RouteMsg(routeMsg)
	return nil
}

func (this *Actor) Call(actorId util.ID, req protocol.ISerializable) (protocol.ISerializable, error) {
	ctx := network.NewDefaultContextWithTimeout(context.TODO(), this.genId(), this.timeout)
	transId := ctx.GetTransId()
	this.addContext(transId, ctx)
	err:=this.SendMessage(actorId, transId, req)
	if err != nil {
		log.Error(err.Error())
		return nil,err
	}
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			log.Warn("DeadlineExceeded",flog.ActorSendMsgInfo(this,req,transId,actorId)...)
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
