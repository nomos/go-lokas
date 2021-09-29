package lox

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"github.com/nomos/go-lokas/network/etcdclient"
	"github.com/nomos/go-lokas/network/redisclient"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/qmgo"
	"go.uber.org/zap"
	"strconv"
)

var _ lokas.IProcess = &Process{}
var _ lokas.IRegistry = &Process{}

func CreateProcess() *Process {
	ret := &Process{
		modulesMap:     map[string]util.ID{},
		modules:        map[string]lokas.IModule{},
		modulesCreator: map[string]lokas.IModuleCtor{},
	}
	return ret
}

type Process struct {
	lokas.IActorContainer
	lokas.IRegistry
	lokas.IRouter
	lokas.IProxy
	modulesMap     map[string]util.ID
	modules        map[string]lokas.IModule
	modulesCreator map[string]lokas.IModuleCtor
	id             util.ProcessId
	idNode         *util.Snowflake
	mongo          *qmgo.Database
	etcd           *etcdclient.Client
	redis          *redisclient.Client
	config         lokas.IConfig
	gameId         string
	serverId       int32
	version        string
}

func (this *Process) GameId() string {
	return this.gameId
}

func (this *Process) ServerId() int32 {
	return this.serverId
}

func (this *Process) GameServerId() string {
	return this.gameId + "_" + strconv.Itoa(int(this.serverId))
}

func (this *Process) Version() string {
	return this.version
}

func (this *Process) PId() util.ProcessId {
	return this.id
}

func (this *Process) GetId() util.ID {
	return this.id.Snowflake()
}

func (this *Process) Type() string {
	return "Process"
}

func (this *Process) LoadModuleRegistry() error {
	res, err := this.etcd.Get(context.TODO(), "/process/"+this.PId().String()+"/modules")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if len(res.Kvs) == 0 {
		return nil
	}
	if len(res.Kvs) != 1 {
		log.Error("incorrect etcd result")
		return errors.New("incorrect etcd result")
	}
	err = json.Unmarshal(res.Kvs[0].Value, &this.modulesMap)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) SaveModuleRegistry() error {
	s, _ := json.Marshal(this.modulesMap)
	_, err := this.etcd.Put(context.TODO(), "/process/"+this.PId().String()+"/modules", string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) RegisterModule(creator lokas.IModuleCtor) {
	this.modulesCreator[creator.Type()] = creator
}

func (this *Process) Add(mod lokas.IModule) lokas.IModule {
	this.modules[mod.Type()] = mod
	mod.SetProcess(this)
	return mod
}

func (this *Process) Config() lokas.IConfig {
	return this.config
}

func (this *Process) LoadMod(name string, conf lokas.IConfig) error {
	log.Info("loading ", zap.String("module", name))
	mod := this.modules[name]
	if mod == nil {
		mod = this.createMod(name)
		this.Add(mod)
	}
	err := mod.Load(conf)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if _, ok := mod.(lokas.IActor); ok {
		this.AddActor(mod.(lokas.IActor))
		this.StartActor(mod.(lokas.IActor))
	}
	log.Info("load success", zap.String("module", name))
	return nil
}

func (this *Process) UnloadMod(name string) error {
	mod := this.modules[name]
	if mod == nil {
		log.Error("module is not exist", zap.String("mod name", name))
		return nil
	}
	if _, ok := mod.(lokas.IActor); ok {
		this.RemoveActor(mod.(lokas.IActor))
	}
	err := mod.Unload()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	delete(this.modules, name)
	return nil
}

func (this *Process) createMod(name string) lokas.IModule {
	creator := this.modulesCreator[name]
	if creator == nil {
		log.Panic("module not exist", zap.String("mod name", name))
	}
	ret := creator.Create()
	if ret == nil {
		log.Panic("module create failed", zap.String("mod name", name))
	}

	if _, ok := ret.(lokas.IActor); ok {
		if id, ok := this.modulesMap[name]; ok {
			ret.(lokas.IActor).SetId(id)
		} else {
			id := this.GenId()
			ret.(lokas.IActor).SetId(id)
			this.modulesMap[name] = id
		}
	}
	return ret
}

func (this *Process) LoadAllModule(conf lokas.IProcessConfig) error {
	modConfigs := conf.GetAllSub()
	for _, modConf := range modConfigs {
		mod := this.createMod(modConf.GetString("name"))
		this.Add(mod)
	}

	for _, mod := range this.modules {
		err := this.LoadMod(mod.Type(), conf.Sub(mod.Type()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Process) StartAllModule() error {
	for _, mod := range this.modules {
			log.Info("starting",flog.FuncInfo(this,"StartAllModule").Append(flog.Module(mod))...)
			err := mod.Start()
			if err != nil {
				return err
			}
			log.Info("success",flog.FuncInfo(this,"StartAllModule").Append(flog.Module(mod))...)
	}
	return nil
}

func (this *Process) StopAllModule() error {
	log.Warn("StopAllModule",zap.Any("modules",this.modules))
	for _, mod := range this.modules {
			log.Info("stop", flog.FuncInfo(this,"StopAllModule").Append(flog.Module(mod))...)
			err := mod.Stop()
			if err != nil {
				return err
			}
			log.Info("success",
				flog.FuncInfo(this, "StopAllModule").
					Append(flog.Module(mod))...
			)
	}
	return nil
}

func (this *Process) Get(name string) lokas.IModule {
	return this.modules[name]
}

func (this *Process) GlobalMutex(key string, ttl int) (*etcdclient.Mutex, error) {
	return this.etcd.NewMutex(key, ttl)
}

func (this *Process) Load(config lokas.IProcessConfig) error {
	this.config = config
	this.serverId = config.GetServerId()
	this.gameId = config.GetGameId()
	this.version = config.GetVersion()
	this.id = config.GetProcessId()

	this.IProxy = this.Add(NewProxy(this)).(lokas.IProxy)
	this.IProxy.(*Proxy).SetPort(config.GetString("port"))
	this.IRegistry = this.Add(NewRegistry(this)).(lokas.IRegistry)
	this.IActorContainer = this.Add(NewActorContainer(this)).(lokas.IActorContainer)
	this.IRouter = this.Add(NewRouter(this)).(lokas.IRouter)
	if this.id == 0 {
		log.Error("pid is not set",flog.FuncInfo(this,"Load")...)
		return errors.New("pid is not set")
	}
	log.Info("process config",
		flog.FuncInfo(this, "Load").
			Append(zap.Uint16("id", uint16(this.PId()))).
			Append(zap.String("game", this.GameId())).
			Append(zap.String("version", this.Version())).
			Append(zap.Int32("server", this.ServerId()))...
	)
	err := this.loadMongo(config.GetDb("mongo").(MongoConfig))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.loadRedis(config.GetDb("redis").(RedisConfig))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.loadEtcd(config.GetDb("etcd").(EtcdConfig))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.idNode, _ = util.NewSnowflake(int64(config.GetProcessId()))
	err = this.LoadModuleRegistry()

	err = this.LoadAllModule(config)
	if err != nil {
		return err
	}
	err = this.SaveModuleRegistry()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) loadMongo(config MongoConfig) error {
	url := "mongodb://" + config.User + ":" + config.Password + "@" + config.Host + ":" + config.Port
	client, err := qmgo.NewClient(context.TODO(), &qmgo.Config{
		Uri:      url,
		Database: config.Database,
	})
	if err != nil {
		log.Error("Process:loadMongo:Error",
			flog.Error(err),
		)

		return err
	}
	this.mongo = client.Database(config.Database)
	log.Info("success",
		flog.FuncInfo(this, "loadMongo").
			Concat(flog.ActorInfo(this)).
			Append(flog.DataBase(config.Database)).
			Append(flog.Address(url))...
	)
	return nil
}

func (this *Process) loadRedis(config RedisConfig) error {
	var err error
	this.redis, err = redisclient.NewClient(config.Host + ":" + config.Port,config.Password)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) loadEtcd(config EtcdConfig) error {
	this.etcd = etcdclient.New(etcdclient.WithEndPoints(config.EndPoints...))
	return nil
}

func (this *Process) Start() error {
	err := this.StartAllModule()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) Stop() error {
	err := this.StopAllModule()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Warn("success",
		flog.FuncInfo(this, "Stop").
			Concat(flog.ActorInfo(this))...
	)
	return nil
}

//GenId generate snowflake id
func (this *Process) GenId() util.ID {
	if this.idNode == nil {
		return 0
	}
	return this.idNode.Generate()
}

func (this *Process) GetLogger() *log.ComposeLogger {
	return log.DefaultLogger()
}

func (this *Process) GetMongo() *qmgo.Database {
	return this.mongo
}

func (this *Process) GetRedis() *redisclient.Client {
	return this.redis
}

func (this *Process) GetEtcd() *etcdclient.Client {
	return this.etcd
}
