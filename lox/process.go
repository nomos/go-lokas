package lox

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nomos/go-lokas/log/flog"
	"strconv"
	"sync"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/dockerclient"
	"github.com/nomos/go-lokas/network/etcdclient"
	"github.com/nomos/go-lokas/network/ossclient"
	"github.com/nomos/go-lokas/network/redisclient"
	"github.com/nomos/go-lokas/util"
	"github.com/nomos/go-lokas/util/slice"
	"github.com/nomos/qmgo"
	"go.uber.org/zap"
)

var _ lokas.IProcess = &Process{}
var _ lokas.IRegistry = &Process{}

var _pOnce sync.Once
var _processInstance *Process

func Instance() *Process {
	_pOnce.Do(func() {
		if _processInstance == nil {
			_processInstance = &Process{
				modulesMap:     map[string]util.ID{},
				modules:        []lokas.IModule{},
				modulesCreator: []lokas.IModuleCtor{},
			}
		}
	})
	return _processInstance
}

func CreateProcess() *Process {
	return Instance()
}

type Process struct {
	lokas.IActorContainer
	lokas.IRegistry
	lokas.IRouter
	lokas.IProxy
	modulesMap     map[string]util.ID
	modules        []lokas.IModule
	modulesCreator []lokas.IModuleCtor
	id             util.ProcessId
	idNode         *util.Snowflake
	mongo          *qmgo.Database
	etcd           *etcdclient.Client
	oss            *ossclient.Client
	redis          *redisclient.Client
	docker         *dockerclient.Client
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
	res, err := this.etcd.Get(context.TODO(), "/process/"+this.PId().ToString()+"/modules")
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
	_, err := this.etcd.Put(context.TODO(), "/process/"+this.PId().ToString()+"/modules", string(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) RegisterModule(creator lokas.IModuleCtor) {
	this.modulesCreator = append(this.modulesCreator, creator)
}

func (this *Process) Add(mod lokas.IModule) lokas.IModule {
	this.modules = append(this.modules, mod)
	mod.SetProcess(this)
	return mod
}

func (this *Process) Config() lokas.IConfig {
	return this.config
}

func (this *Process) getModuleByType(name string) lokas.IModule {
	for _, v := range this.modules {
		if v.Type() == name {
			return v
		}
	}
	return nil
}

func (this *Process) LoadMod(name string, conf lokas.IConfig) error {
	// log.Info("loading ", zap.String("module", name))
	mod := this.getModuleByType(name)
	if mod == nil {
		mod = this.createMod(name)
		this.Add(mod)
	}
	err := mod.Load(conf)
	if err != nil {
		return err
	}
	if _, ok := mod.(lokas.IActor); ok {
		this.AddActor(mod.(lokas.IActor))
		err = this.StartActor(mod.(lokas.IActor))
		if err != nil {
			return err
		}
	}
	log.Info("load success", zap.String("module", name))
	return nil
}

func (this *Process) UnloadMod(name string) error {
	mod := this.getModuleByType(name)
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
	this.modules, _ = slice.RemoveWithCondition(this.modules, func(i int, t lokas.IModule) bool {
		if t.Type() == name {
			return true
		}
		return false
	})
	return nil
}

// func (this *Process) GetServerRegisterMgr() *ServiceRegisterMgr {
// 	return this.Get
// }

func (this *Process) getModuleCreatorByType(name string) lokas.IModuleCtor {
	for _, v := range this.modulesCreator {
		if v.Type() == name {
			return v
		}
	}
	return nil
}

func (this *Process) createMod(name string) lokas.IModule {
	creator := this.getModuleCreatorByType(name)
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

	proxy := this.Get("FnProxy")
	if proxy == nil {
		proxy = this.Get("FnProxyClient")
	}
	// TODO temp do
	if proxy != nil {
		this.IProxy = proxy.(lokas.IProxy)
	}

	return nil
}

func (this *Process) StartAllModule() error {
	for _, mod := range this.modules {
		log.Info("starting", flog.FuncInfo(this, "StartAllModule").Append(lokas.LogModule(mod))...)
		err := mod.Start()
		if err != nil {
			return err
		}
		log.Info("success", flog.FuncInfo(this, "StartAllModule").Append(lokas.LogModule(mod))...)
	}
	return nil
}

func (this *Process) StopAllModule() error {
	log.Warn("StopAllModule", zap.Any("modules", this.modules))
	for _, mod := range this.modules {
		log.Info("stop", flog.FuncInfo(this, "StopAllModule").Append(lokas.LogModule(mod))...)
		err := mod.Stop()
		if err != nil {
			return err
		}
		log.Info("success",
			flog.FuncInfo(this, "StopAllModule").
				Append(lokas.LogModule(mod))...,
		)
	}
	return nil
}

func (this *Process) Get(name string) lokas.IModule {
	return this.getModuleByType(name)
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
		log.Error("pid is not set", flog.FuncInfo(this, "Load")...)
		return errors.New("pid is not set")
	}
	log.Info("process config",
		flog.FuncInfo(this, "Load").
			Append(zap.Uint16("id", uint16(this.PId()))).
			Append(zap.String("game", this.GameId())).
			Append(zap.String("version", this.Version())).
			Append(zap.Int32("server", this.ServerId()))...,
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
	err = this.loadOss(config.GetDb("oss").(OssConfig))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.loadDockerCLI(config.GetDockerCLI().(DockerConfig))

	this.idNode, _ = util.NewSnowflake(int64(config.GetProcessId()))
	err = this.LoadModuleRegistry()

	err = this.LoadAllModule(config)
	if err != nil {
		log.Error(err.Error())
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
			Concat(lokas.LogActorInfo(this)).
			Append(flog.DataBase(config.Database)).
			Append(flog.Address(url))...,
	)
	return nil
}

func (this *Process) loadRedis(config RedisConfig) error {
	var err error
	if config.Host == "" {
		return nil
	}
	this.redis, err = redisclient.NewClient(config.Host+":"+config.Port, config.Password)
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

func (this *Process) loadOss(config OssConfig) error {
	if config.EndPoint == "" {
		return nil
	}
	t := ossclient.MINIO
	if config.OssType == "aliyun" {
		t = ossclient.ALIYUN_OSS
	}
	var err error
	this.oss, err = ossclient.NewClient(t, config.EndPoint, config.AccessId, config.AccessSecret)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Process) loadDockerCLI(config DockerConfig) error {
	if config.Endpoint == "" || config.CertPath == "" {
		return nil
	}
	var err error
	this.docker, err = dockerclient.NewClient()
	if err != nil {
		log.Error(err.Error())
		return err
	}
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
			Concat(lokas.LogActorInfo(this))...,
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

func (this *Process) GetOss() *ossclient.Client {
	return this.oss
}

func (this *Process) GetDocker() (*dockerclient.Client, error) {
	if this.docker == nil {
		return nil, errors.New("invalid endpoint, certPath")
	}
	return this.docker, nil
}
