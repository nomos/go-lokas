package lokas

import (
	"context"
	"errors"
	"fmt"
	"github.com/nomos/go-events"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/etcdclient"
	"github.com/nomos/go-lokas/network/redisclient"
	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/promise"
	"github.com/nomos/qmgo"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"reflect"
	"time"
)

const (
	DEFAULT_PACKET_LEN = 2048 * 4
	DEFAULT_ACTOR_TTL  = 15
)

type Key string

func (this Key) Assemble(args... interface{})string{
	return fmt.Sprintf(string(this),args...)
}

type ServiceType int

const (
	//rpc have a actorId<util.Id> to route message
	//SERVICE_RPC ServiceType = iota + 1
	//subscribe have a actorId<util.Id> to pub message
	SERVICE_SUB ServiceType = iota + 1
	//stateless  do not have local state
	SERVICE_STATELESS
	SERVICE_SERVER
)

type ActorState int //Actor health state

const (
	ACTOR_STARTING ActorState = iota + 1
	ACTOR_HEALTHY
	ACTOR_UNHEALTHY
	ACTOR_ERRORED
	ACTOR_STOPPED
)

var (
	ErrDataNil = errors.New("data nil")
)

//IProcess the interface for application entry
type IProcess interface {
	IRegistry
	IActorContainer
	IRouter
	Add(modules IModule) IModule             //add a module
	RegisterModule(ctor IModuleCtor)         //register a module ctor
	LoadAllModule(IProcessConfig) error      //load all module from config
	LoadMod(name string, conf IConfig) error //load a module with config
	UnloadMod(name string) error             //unload a module
	Get(name string) IModule                 //get a module
	Load(IProcessConfig) error               //load config
	Start() error                            //start process
	Stop() error                             //stop process
	PId() util.ProcessId                     //PId
	GetId() util.ID                          //PId to snowflake
	Type()string
	GenId() util.ID                                             //gen snowflake Id,goroutine safe
	GetLogger() *log.ComposeLogger                              //get process logger
	GetMongo() *qmgo.Database                                   //get mongo client
	GetRedis() *redisclient.Client                              //get redis client
	GetEtcd() *etcdclient.Client                                //get etcd client
	GlobalMutex(key string, ttl int) (*etcdclient.Mutex, error) //create a distributed global mutex based on etcd
	Config() IConfig                                            //get config
	GameId() string                                             //get game id
	ServerId() int32                                            //get server id
	GameServerId()string                                        //get game and server id
	Version() string                                            //get version
}

//IProxy universal module interface for connection
type IProxy interface {
}

type IActorContainer interface {
	AddActor(actor IActor)
	RemoveActor(actor IActor)
	RemoveActorById(id util.ID) IActor
	GetActorIds() []util.ID
	GetActor(id util.ID) IActor
}

type IActorInfo interface {
	GetId()util.ID
	PId()util.ProcessId
	Type() string
}

//IActor standard interface for actor
type IActor interface {
	IEntity
	IProxy
	IModule
	PId()util.ProcessId
	ReceiveMessage(msg *protocol.RouteMessage)
	OnMessage(msg *protocol.RouteMessage)
	SendMessage(actorId util.ID, transId uint32, msg protocol.ISerializable) error
	Call(actorId util.ID, req protocol.ISerializable) (protocol.ISerializable, error)
	GetLeaseId() (clientv3.LeaseID,bool, error)
	Update(dt time.Duration, now time.Time)
}

type IEntity interface {
	Dirty()bool
	SetDirty(bool)
	Get(t protocol.BINARY_TAG)IComponent
	GetOrCreate(t protocol.BINARY_TAG)IComponent
	Add(c IComponent)
	Remove(t protocol.BINARY_TAG)IComponent
	RemoveAll()
	SetId(id util.ID)
	GetId()util.ID
	Components()map[protocol.BINARY_TAG]IComponent
}


type IComponentPool interface {
	Recycle(comp IComponent)
	Get() IComponent
	Create(args...interface{}) IComponent
	Destroy()
}

type IComponent interface {
	protocol.ISerializable
	GetEntity()IEntity
	SetDirty(d bool)
	SetRuntime(engine IRuntime)
	GetRuntime()IRuntime
	SetEntity(e IEntity)
	OnAdd(e IEntity, r IRuntime)
	OnRemove(e IEntity, r IRuntime)
	OnCreate(r IRuntime)
	OnDestroy(r IRuntime)
	GetSibling(t protocol.BINARY_TAG) IComponent
}

type IRuntime interface {
	Init(updateTime int64,timeScale float32,server bool)
	GetContext(name string)interface{}
	SetContext(name string,value interface{})
	GetEntity(util.ID)IEntity
	CurrentTick()int64
	Start()
	Stop()
	RunningTime()int64
	GetTimeScale()float32
	SetTimeScale(scale float32)
	RegisterComponent(name string,c IComponent)
	RegisterSingleton(name string,c IComponent)
	GetComponentType(name string)reflect.Type
	IsSyncAble(compName string)bool
	CreateEntity()IEntity
	IsServer()bool
	//private
	MarkDirtyEntity(e IEntity)
}

//IModuleCtor module export interface
type IModuleCtor interface {
	Type() string
	Create() IModule
}

//IModule module interface
type IModule interface {
	Type() string
	Load(conf IConfig) error
	Unload() error
	GetProcess() IProcess
	SetProcess(IProcess)
	Start() *promise.Promise
	Stop() *promise.Promise
	OnStart() error
	OnStop() error
}

type IRegistry interface {
	RegisterProcessInfo() error
	GetProcessIdByActor(actorId util.ID)(util.ProcessId,error)
	RegisterActors() error
	RegisterActorRemote(actor IActor) error
	UnregisterActorRemote(actor IActor)error
	RegisterServiceRemote(service *Service)error
	UnregisterServiceRemote(service *Service)error
	RegisterActorLocal(actor IActor) error
	UnregisterActorLocal(actor IActor)error
	RegisterServiceLocal(service *Service)error
	UnregisterServiceLocal(service *Service)error
	GetActorIdsByTypeAndServerId(serverId int32,typ string)[]util.ID

}

//IRouter interface for router
type IRouter interface {
	RouteMsg(msg *protocol.RouteMessage)
}

//IContext context interface
type IContext interface {
	Get(key string) interface{}
	GetIdType(key string) util.ID
	GetProcessIdType(key string) util.ProcessId
	GetString(key string) string
	Set(key string, value interface{})
}

type IConfig interface {
	Save() error
	Load() error
	LoadFromRemote() error
	SetRemoteConfig(p string, etcd string)
	Set(key string, value interface{})
	Get(key string) interface{}
	Sub(key string) IConfig
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetIntSlice(key string) []int
	GetSizeInBytes(key string) uint
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	IsSet(key string) bool
	AllSettings() map[string]interface{}
}

type IProcessConfig interface {
	IConfig
	GetName() string    //config name
	ServerName() string //serverName
	GetProcessId() util.ProcessId
	GetGameId() string
	GetServerId() int32
	GetVersion() string
	GetDb(string) interface{}
	GetAllSub() []IConfig //get all config for sub modules
}

//INetClient interface for client
type INetClient interface {
	events.EventEmmiter
	Connect(addr string) *promise.Promise
	Disconnect(bool) *promise.Promise
	Request(req interface{}) *promise.Promise
	Connected() bool
	OnRecvCmd(cmdId protocol.BINARY_TAG, time time.Duration) *promise.Promise
	OnRecv(conn IConn, data []byte)
}

type IEntityNetClient interface {
	IEntity
	INetClient
}

type AsyncCallBack func(context IReqContext)

type IReqContext interface {
	context.Context
	IContext
	GetTransId() uint32
	GetResp() interface{}
	SetResp(interface{})
	SetCallback(cb AsyncCallBack)
	GetCallback() AsyncCallBack
	Cancel(err error)
	Finish()
}

type ITestCase interface {
	Test() *promise.Promise
	TestWithContext(ctx IContext) *promise.Promise
	Name() string
	SetName(string)
	GetTestCase(num int) ITestCase
	GetTestCaseByName(name string) ITestCase
	AddTestCase(name string, testCase ITestCase)
	GetLength() int
}

//ISession connection session interface
type ISession interface {
	GetId() util.ID
	GetConn() IConn                 // get corresponding IConn
	OnOpen(conn IConn)              // called when IConn is opened
	OnClose(conn IConn)             // called when IConn is closed
	OnRecv(conn IConn, data []byte) // called when IConn receives data, ATTENTION: data is a slice and must be consumed immediately
}

// SessionManager manage all the sessions for a server network service

type ISessionManager interface {
	AddSession(id util.ID, session ISession)
	RemoveSession(id util.ID)
	GetSession(id util.ID) ISession
	GetSessionCount() int
	Clear()
	Range(f func(id util.ID, session ISession) bool)
}

type IModel interface {
	GetId() util.ID
	Deserialize(a IProcess) error
	Serialize(a IProcess) error
}

// IConn interface for a connection, either a client connection or server accepted connection
type IConn interface {
	net.Conn
	SetUserData(userData interface{}) // set application layer reference data
	GetUserData() interface{}         // get application layer reference data
	GetSession() ISession             // get the bound session
	GetConnTime() time.Time           // get connection init time
	Activate()                        // activate the conn, used with idle check
	// GraceClose() error                // gracefully close the connection, after all the pending packets sent to the peer
	Wait() // wait for close
}

// Server the interface for wsserver and tcpserver
type Server interface {
	Start(addr string) error                     //start server
	Stop()                                       //stop server
	Broadcast(sessionIds []util.ID, data []byte) // broadcast data to all connected sessions
	GetActiveConnNum() int                       // get current count of connections
}

// pick long packet from binary
// args: data:binary
// return: 1.is long packet 2.long packet index, if index = 0 is the last long packet 3.if is long packet  return real packet binary else return nil
type LongPacketPicker func(data []byte) (bool, int, []byte)

// create long packet
// args: data:binary to create  idx: long packet index
// return: 1.long packet
type LongPacketCreator func(data []byte, idx int) ([]byte, error)
