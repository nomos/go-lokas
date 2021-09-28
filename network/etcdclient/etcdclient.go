package etcdclient

import (
	"github.com/nomos/go-lokas/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"time"
)

type ETCD_ACTION int32

const (
	ETCD_DELETE = iota
	ETCD_PUT
)

func (this ETCD_ACTION) ToString()string{
	switch this {
	case ETCD_DELETE:
		return "delete"
	case ETCD_PUT:
		return "put"
	}
	return ""
}

func LogAction(action ETCD_ACTION) zap.Field {
	return zap.String("action", action.ToString())
}

func LogKey(key string) zap.Field {
	return zap.String("key", key)
}
func LogOk(v bool) zap.Field {
	if v {
		return zap.String("result", "ok")
	} else {
		return zap.String("result", "failed")
	}
}

func LogResp(resp interface{}) zap.Field{
	return zap.String("resp",log.PrettyStruct(resp))
}


type Option func(client *Client)

func New(opts... Option)*Client{
	ret:=&Client{}
	ret.config.Endpoints = []string{}
	for _,o:=range opts {
		o(ret)
	}
	ret.Client,_ = clientv3.New(ret.config)
	ret.KV = clientv3.NewKV(ret.Client)
	ret.Watcher = clientv3.NewWatcher(ret.Client)
	return ret
}

func WithEndPoints(endPoints... string)func(client *Client){
	return func(client *Client) {
		client.config.Endpoints = append(client.config.Endpoints,endPoints...)
	}
}

func WithAuth(user string,pass string)func(client *Client){
	return func(client *Client) {
		client.config.Username = user
		client.config.Password = pass
	}
}

func WithTimeout(timeout time.Duration)func(client *Client){
	return func(client *Client) {
		client.config.DialKeepAliveTimeout = timeout
	}
}

func (this *Client)NewMutex(key string, ttl int)(*Mutex, error){
	return CreateMutex(this.Client,key,ttl)
}

type Client struct {
	clientv3.KV
	clientv3.Watcher
	*clientv3.Client
	config clientv3.Config
}


