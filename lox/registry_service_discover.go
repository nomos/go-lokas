package lox

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	ETCD_SERVICE_PREFIX_KEY = "/service/"
)

type ServiceDiscoverMgr struct {
	process lokas.IProcess

	serviceMap map[string]map[uint16]map[uint16]*lokas.ServiceInfo

	mutex sync.RWMutex

	closeChan chan struct{}
}

func NewServiceDiscoverMgr(process lokas.IProcess) *ServiceDiscoverMgr {
	return &ServiceDiscoverMgr{
		process:    process,
		serviceMap: make(map[string]map[uint16]map[uint16]*lokas.ServiceInfo),
	}
}

func (mgr *ServiceDiscoverMgr) FindServiceInfo(serviceType string, serviceId uint16, lineId uint16) (*lokas.ServiceInfo, bool) {

	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if _, ok := mgr.serviceMap[serviceType]; !ok {
		return nil, ok
	}
	if _, ok := mgr.serviceMap[serviceType][serviceId]; !ok {
		return nil, ok
	}

	serviceInfo, ok := mgr.serviceMap[serviceType][serviceId][lineId]
	return serviceInfo, ok
}

// if serviceId is zero,get random serviceId; if lineId is zero,get randmo lineId
func (mgr *ServiceDiscoverMgr) FindRandServiceInfo(serviceType string, serviceId uint16, lineId uint16) (*lokas.ServiceInfo, bool) {

	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	if _, ok := mgr.serviceMap[serviceType]; !ok {
		return nil, ok
	}

	if serviceId == 0 {
		// get a random serviceId
		infos := lokas.ServiceInfos{}
		for _, v1 := range mgr.serviceMap[serviceType] {
			for _, v2 := range v1 {
				infos = append(infos, v2)
			}
		}
		sort.Stable(infos)
		randIdx := rand.Intn(len(infos))

		return infos[randIdx], true
	}

	if _, ok := mgr.serviceMap[serviceType][serviceId]; !ok {
		return nil, ok
	}

	if lineId == 0 {
		// get a random lineId
		infos := lokas.ServiceInfos{}
		for _, v := range mgr.serviceMap[serviceType][serviceId] {
			infos = append(infos, v)
		}
		sort.Stable(infos)

		randIdx := rand.Intn(len(infos))

		return infos[randIdx], true
	}

	serviceInfo, ok := mgr.serviceMap[serviceType][serviceId][lineId]
	return serviceInfo, ok

}

func (mgr *ServiceDiscoverMgr) StartDiscover() error {

	log.Info("start discover service", zap.String("path", ETCD_SERVICE_PREFIX_KEY))
	etcdClient := mgr.process.GetEtcd()
	resp, err := etcdClient.Get(context.TODO(), ETCD_SERVICE_PREFIX_KEY, clientv3.WithPrefix())
	if err != nil {
		log.Error(err.Error())
		return err
	}

	for _, v := range resp.Kvs {
		mgr.addServiceFromEtcd(v)
	}

	mgr.closeChan = make(chan struct{})

	watchChan := etcdClient.Watch(context.TODO(), ETCD_SERVICE_PREFIX_KEY, clientv3.WithPrefix(), clientv3.WithRev(resp.Header.Revision))
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
		mgr.serviceMap[serviceInfo.ServiceType] = make(map[uint16]map[uint16]*lokas.ServiceInfo)
	}

	if _, ok := mgr.serviceMap[serviceInfo.ServiceType][serviceInfo.ServiceId]; !ok {
		mgr.serviceMap[serviceInfo.ServiceType][serviceInfo.ServiceId] = make(map[uint16]*lokas.ServiceInfo)
	}

	mgr.serviceMap[serviceInfo.ServiceType][serviceInfo.ServiceId][serviceInfo.LineId] = serviceInfo

	log.Info("update service", zap.Any("serviceInfo", serviceInfo))
	return nil
}

func (mgr *ServiceDiscoverMgr) delServiceFromEtcd(kv *mvccpb.KeyValue) error {

	reg := regexp.MustCompile(ETCD_SERVICE_PREFIX_KEY + "(?P<type>[a-zA-Z]+)/(?P<id>[0-9]+)/(?P<line>[0-9])")
	matchs := reg.FindStringSubmatch(string(kv.Key))
	typIdx := reg.SubexpIndex("type")
	idIdx := reg.SubexpIndex("id")
	lineIdx := reg.SubexpIndex("line")

	if len(matchs) < idIdx+1 || len(matchs) < typIdx+1 || typIdx < 0 || idIdx < 0 || lineIdx < 0 {
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

	lineId, err2 := strconv.ParseUint(matchs[lineIdx], 10, 16)
	if err2 != nil {
		log.Error(err.Error(), zap.String("etcd.kv.Key", string(kv.Key)))
		return err
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if _, ok := mgr.serviceMap[serviceType]; ok {
		if _, ok := mgr.serviceMap[serviceType][uint16(serviceId)]; ok {
			delete(mgr.serviceMap[serviceType][uint16(serviceId)], uint16(lineId))
		}
	}

	log.Info("del service from etcd", zap.String("serviceType", serviceType), zap.String("serviceId", matchs[idIdx]), zap.String("lineId", matchs[lineIdx]))
	return nil
}
