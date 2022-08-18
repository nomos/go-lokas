package lox

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"sync"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/lox/flog"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ServiceDiscoverMgr struct {
	process lokas.IProcess

	serviceMap map[string]map[uint16]*lokas.ServiceInfo

	mutex sync.RWMutex

	closeChan chan struct{}
}

func NewServiceDiscoverMgr(process lokas.IProcess) *ServiceDiscoverMgr {
	return &ServiceDiscoverMgr{
		process:    process,
		serviceMap: make(map[string]map[uint16]*lokas.ServiceInfo),
	}
}

func (mgr *ServiceDiscoverMgr) FindServiceInfo(serviceType string, serviceId uint16) (*lokas.ServiceInfo, bool) {

	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if _, ok := mgr.serviceMap[serviceType]; !ok {
		return nil, ok
	} else {
		serviceInfo, ok2 := mgr.serviceMap[serviceType][serviceId]
		return serviceInfo, ok2
	}

}

func (mgr *ServiceDiscoverMgr) StartDiscover() error {
	log.Info("start", flog.FuncInfo(mgr, "startDiscover")...)
	etcdClient := mgr.process.GetEtcd()
	resp, err := etcdClient.Get(context.TODO(), "/service/", clientv3.WithPrefix())
	if err != nil {
		log.Error(err.Error())
		return err
	}

	for _, v := range resp.Kvs {
		mgr.addServiceFromEtcd(v)
	}

	mgr.closeChan = make(chan struct{})

	watchChan := etcdClient.Watch(context.TODO(), "/service/", clientv3.WithPrefix(), clientv3.WithRev(resp.Header.Revision))
	go func() {
	LOOP:
		for {
			select {
			case resp := <-watchChan:
				for _, v := range resp.Events {
					switch v.Type {
					case mvccpb.PUT:
						mgr.addServiceFromEtcd(v.Kv)
					case mvccpb.DELETE:
						mgr.delServiceFromEtcd(v.Kv)
					}
				}
			case <-mgr.closeChan:
				break LOOP
			}
		}
		close(mgr.closeChan)

	}()

	return nil
}

func (mgr *ServiceDiscoverMgr) Stop() {
	mgr.closeChan <- struct{}{}
}

func (mgr *ServiceDiscoverMgr) addServiceFromEtcd(kv *mvccpb.KeyValue) error {
	serviceInfo := &lokas.ServiceInfo{}
	err := json.Unmarshal(kv.Value, serviceInfo)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if _, ok := mgr.serviceMap[serviceInfo.ServiceType]; !ok {
		mgr.serviceMap[serviceInfo.ServiceType] = make(map[uint16]*lokas.ServiceInfo)
	}

	mgr.serviceMap[serviceInfo.ServiceType][serviceInfo.ServiceId] = serviceInfo

	log.Info("update service", zap.Any("serviceInfo", serviceInfo))
	return nil
}

func (mgr *ServiceDiscoverMgr) delServiceFromEtcd(kv *mvccpb.KeyValue) error {

	reg := regexp.MustCompile("/service/(?P<type>[a-zA-Z]+)/(?P<id>[0-9]+)")
	matchs := reg.FindStringSubmatch(string(kv.Key))
	typIdx := reg.SubexpIndex("type")
	idIdx := reg.SubexpIndex("id")

	if len(matchs) < idIdx+1 || len(matchs) < typIdx+1 || typIdx < 0 || idIdx < 0 {
		err := errors.New("etc data invalid")
		log.Error(err.Error())
		return err
	}
	serviceType := matchs[typIdx]
	serviceId, err := strconv.ParseUint(matchs[idIdx], 10, 16)
	if err != nil {
		log.Error(err.Error(), zap.String("etcd.kv.Key", string(kv.Key)))
		return err
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if _, ok := mgr.serviceMap[serviceType]; ok {
		delete(mgr.serviceMap[serviceType], uint16(serviceId))
	}
	return nil
}
