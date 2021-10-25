package lox

import (
	"bytes"
	"context"
	"errors"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/etcdclient"
	"github.com/nomos/go-lokas/util"
	"github.com/spf13/viper"
	"os"
	"path"
	"sync"
	"time"
)

var _ lokas.IConfig = (*AppConfig)(nil)

type AppConfig struct {
	*viper.Viper
	name     string
	root     string
	folder   string
	fromFile bool
	etcdPath string
	etcdAddr string
	parent   *AppConfig
	mu       sync.Mutex
}

type ConfigOption func(*AppConfig) *AppConfig

func ConfigFile(folder string) ConfigOption {
	return func(config *AppConfig) *AppConfig {
		config.folder = folder
		config.fromFile = true
		return config
	}
}

func MongoFile(folder string) ConfigOption {
	return func(config *AppConfig) *AppConfig {
		config.folder = folder
		config.fromFile = false
		return config
	}
}

func EtcdFile(etcdPath string, addr string) ConfigOption {
	return func(config *AppConfig) *AppConfig {
		config.SetRemoteConfig(etcdPath, addr)
		return config
	}
}

func NewAppConfig(name string, opts ...ConfigOption) *AppConfig {
	ret := &AppConfig{}
	ret.name = name
	ret.root = ""
	for _, opt := range opts {
		ret = opt(ret)
	}
	if ret.fromFile && ret.folder == "" {
		ret.folder, _ = util.ExecPath()
	}
	ret.Viper = viper.New()
	ret.Viper.AddConfigPath(".")
	ret.Viper.SetConfigName(name)
	ret.Viper.SetConfigType("toml")
	return ret
}

func NewSubAppConfig(name string, parent *AppConfig, conf *viper.Viper) *AppConfig {
	ret := &AppConfig{}
	ret.name = name
	ret.root = name
	ret.parent = parent
	if conf!=nil {
		ret.Viper = conf.Sub(name)
	}
	return ret
}
func (this *AppConfig) GetFolder() string {
	return this.folder
}

func (this *AppConfig) SetFolder(f string) {
	this.folder = f
}

func (this *AppConfig) SetRemoteConfig(p string, etcd string) {
	this.etcdPath = p
	this.etcdAddr = etcd
}

func (this *AppConfig) LoadFromRemote() error {
	if this.parent != nil {
		log.Warnf("parent", this.parent)
		err := this.parent.LoadFromRemote()
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return nil
	} else {
		client := etcdclient.New(etcdclient.WithEndPoints(this.etcdAddr))
		defer client.Client.Close()
		resp, err := client.Get(context.TODO(), this.etcdPath)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		if len(resp.Kvs) == 1 {
			err := this.ReadConfig(bytes.NewBuffer(resp.Kvs[0].Value))
			if err != nil {
				log.Error(err.Error())
				return err
			}
		} else {
			return errors.New("wrong etcd return")
		}
		return nil
	}
}

func (this *AppConfig) Save() error {
	if this.parent != nil {
		log.Warnf("parent", this.parent)
		err := this.parent.Save()
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return nil
	}
	//err:=this.Viper.WatchRemoteConfig()
	//if err != nil {
	//	log.Error(err.Error())
	//	return err
	//}
	err := this.Viper.WriteConfigAs(path.Join(this.folder, this.name+".toml"))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *AppConfig) Load() error {
	this.mu.Lock()
	defer this.mu.Unlock()
	file, err := os.OpenFile(path.Join(this.folder, this.name+".toml"), os.O_CREATE, 0664)
	if err != nil {
		log.Warn(err.Error())
		return err
	}
	if file != nil {
		return this.Viper.ReadConfig(file)
	} else {
		log.Warn("no file")
	}
	return nil
}

func (this *AppConfig) Set(key string, value interface{}) {
	if this.parent != nil {
		key = this.name + "." + key
		this.parent.Set(key, value)
		return
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Viper.Set(key, value)
	err := this.Save()
	if err != nil {
		log.Error(err.Error())
	}
}

func (this *AppConfig) Sub(key string) lokas.IConfig {
	return NewSubAppConfig(key, this, this.Viper)
}

func (this *AppConfig) Get(key string) interface{} {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.Get(key)
	}
	return this.Viper.Get(key)
}
func (this *AppConfig) GetBool(key string) bool {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetBool(key)
	}
	return this.Viper.GetBool(key)
}
func (this *AppConfig) GetFloat64(key string) float64 {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetFloat64(key)
	}
	return this.Viper.GetFloat64(key)
}
func (this *AppConfig) GetInt(key string) int {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetInt(key)
	}
	return this.Viper.GetInt(key)
}
func (this *AppConfig) GetString(key string) string {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetString(key)
	}
	return this.Viper.GetString(key)
}
func (this *AppConfig) GetStringMap(key string) map[string]interface{} {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetStringMap(key)
	}
	return this.Viper.GetStringMap(key)
}
func (this *AppConfig) GetStringMapString(key string) map[string]string {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetStringMapString(key)
	}
	return this.Viper.GetStringMapString(key)
}
func (this *AppConfig) GetStringMapStringSlice(key string) map[string][]string {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetStringMapStringSlice(key)
	}
	return this.Viper.GetStringMapStringSlice(key)
}
func (this *AppConfig) GetIntSlice(key string) []int {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetIntSlice(key)
	}
	return this.Viper.GetIntSlice(key)
}
func (this *AppConfig) GetSizeInBytes(key string) uint {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetSizeInBytes(key)
	}
	return this.Viper.GetSizeInBytes(key)
}
func (this *AppConfig) GetStringSlice(key string) []string {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetStringSlice(key)
	}
	return this.Viper.GetStringSlice(key)
}
func (this *AppConfig) GetTime(key string) time.Time {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetTime(key)
	}
	return this.Viper.GetTime(key)
}
func (this *AppConfig) GetDuration(key string) time.Duration {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.GetDuration(key)
	}
	return this.Viper.GetDuration(key)
}
func (this *AppConfig) IsSet(key string) bool {
	if this.parent != nil {
		key = this.name + "." + key
		return this.parent.IsSet(key)
	}
	return this.Viper.IsSet(key)
}
func (this *AppConfig) AllSettings() map[string]interface{} {
	if this.parent != nil {
		return this.parent.AllSettings()
	}
	return this.Viper.AllSettings()
}

var _ lokas.IConfig = (*DefaultConfig)(nil)

type DefaultConfig struct {
	*AppConfig
	ProcessId util.ProcessId  `mapstructure:"pid"`
	ServerId  int32           `mapstructure:"sid"`
	GameId    string          `mapstructure:"gid"`
	Host      string          `mapstructure:"host"`
	Port      string          `mapstructure:"port"`
	Version   string          `mapstructure:"version"`
	SName     string          `mapstructure:"serverName"`
	Name      string          `mapstructure:"name"`
	Etcd      EtcdConfig      `mapstructure:"-"`
	Mongo     MongoConfig     `mapstructure:"-"`
	Mysql     MysqlConfig     `mapstructure:"-"`
	Redis     RedisConfig     `mapstructure:"-"`
	Mods      []lokas.IConfig `mapstructure:"-"`
	Modules   []string        `mapstructure:"modules"`
}

func (this *DefaultConfig) GetDb(t string) interface{} {
	switch t {
	case "mongo":
		return this.Mongo
	case "mysql":
		return this.Mysql
	case "redis":
		return this.Redis
	case "etcd":
		return this.Etcd
	default:
		log.Panicf("unrecognized db type", t)
	}
	return nil
}

type MongoConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type EtcdConfig struct {
	EndPoints []string `mapstructure:"endpoints"`
}

type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

var _ lokas.IConfig = (*DefaultConfig)(nil)

func NewDefaultConfig() *DefaultConfig {
	ret := &DefaultConfig{
		AppConfig: NewAppConfig("service"),
		Mods:      []lokas.IConfig{},
	}
	return ret
}

func (this *DefaultConfig) Load() error {
	err := this.AppConfig.Load()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.loadInner()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *DefaultConfig) LoadFromRemote() error {
	err := this.AppConfig.LoadFromRemote()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.loadInner()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *DefaultConfig) loadInner() error {
	err := this.Viper.Unmarshal(this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.Viper.UnmarshalKey("db.mongo", &this.Mongo)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.Viper.UnmarshalKey("db.redis", &this.Redis)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.Viper.UnmarshalKey("db.etcd", &this.Etcd)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	this.loadModules()
	return nil
}

func (this *DefaultConfig) loadModules() {
	for _, v := range this.Modules {
		modConf := this.Sub(v)
		this.Mods = append(this.Mods, modConf)
	}
}

func (this *DefaultConfig) LoadFromString(s string) error {
	err := this.Viper.ReadConfig(bytes.NewBuffer([]byte(s)))
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = this.loadInner()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *DefaultConfig) GetName() string {
	return this.Name
}

func (this *DefaultConfig) ServerName() string {
	return this.SName
}

func (this *DefaultConfig) GetProcessId() util.ProcessId {
	return this.ProcessId
}

func (this *DefaultConfig) GetGameId() string {
	return this.GameId
}

func (this *DefaultConfig) GetServerId() int32 {
	return this.ServerId
}

func (this *DefaultConfig) GetVersion() string {
	return this.Version
}

func (this *DefaultConfig) GetIdType(key string) util.ID {
	return this.Get(key).(util.ID)
}

func (this *DefaultConfig) GetProcessIdType(key string) util.ProcessId {
	return this.Get(key).(util.ProcessId)
}

func (this *DefaultConfig) GetAllSub() []lokas.IConfig {
	return this.Mods
}
