package lox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/network/etcdclient"
	"github.com/nomos/go-lokas/protocol"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

type ServiceRegister struct {
	serverInfo *lokas.ServiceInfo

	etcdClient *etcdclient.Client

	leaseId clientv3.LeaseID

	mutex sync.Mutex

	closeChan chan struct{}
}

type ServiceRegisterMgr struct {
	process lokas.IProcess

	registerMap map[string]map[uint16]*ServiceRegister

	mutex sync.RWMutex
}

func NewServiceRegisterMgr(process lokas.IProcess) *ServiceRegisterMgr {
	return &ServiceRegisterMgr{
		process:     process,
		registerMap: make(map[string]map[uint16]*ServiceRegister),
	}
}

func (register *ServiceRegister) registerEtcd() error {
	etcd := register.etcdClient

	strServiceInfo, err := json.Marshal(register.serverInfo)
	if err != nil {
		log.Error(protocol.ERR_REGISTER_SERVICE_INFO_INVALID.Error(), zap.Any("serviceInfo", register.serverInfo))
		return protocol.ERR_REGISTER_SERVICE_INFO_INVALID
	}

	register.mutex.Lock()
	defer register.mutex.Unlock()

	leaseResp, err := etcd.Lease.Grant(context.TODO(), 5)
	if err != nil {
		log.Error(protocol.ERR_ETCD_ERROR.Error())
		return protocol.ERR_ETCD_ERROR
	}
	register.leaseId = leaseResp.ID

	_, err2 := concurrency.NewSTM(etcd.Client, func(s concurrency.STM) error {

		strKey := fmt.Sprintf("/service/%s/%d", register.serverInfo.ServiceType, register.serverInfo.ServiceId)
		remoteValue := s.Get(strKey)

		if remoteValue != "" {
			return protocol.ERR_REGISTER_SERVICE_DUPLICATED
		}

		s.Put(strKey, string(strServiceInfo), clientv3.WithLease(register.leaseId))
		return nil
	})

	if err2 != nil {
		log.Error(err2.Error(), zap.Any("serviceInfo", register.serverInfo))
		return err2
	}

	return nil
}

func (register *ServiceRegister) keepAliveEtcd() error {
	_, err := register.etcdClient.Lease.KeepAliveOnce(context.TODO(), register.leaseId)
	if err != nil {
		if err == rpctypes.ErrLeaseNotFound {
			log.Warn("lease not found, register again", zap.Any("serverInfo", register.serverInfo))
			register.registerEtcd()
			err = nil
		}
	}

	return err
}

func (register *ServiceRegister) updateEtcd() error {
	strServiceInfo, err := json.Marshal(register.serverInfo)
	if err != nil {
		log.Error(protocol.ERR_REGISTER_SERVICE_INFO_INVALID.Error(), zap.Any("serviceInfo", register.serverInfo))
		return protocol.ERR_REGISTER_SERVICE_INFO_INVALID
	}

	strKey := fmt.Sprintf("/service/%s/%d", register.serverInfo.ServiceType, register.serverInfo.ServiceId)
	_, err2 := register.etcdClient.KV.Put(context.TODO(), strKey, string(strServiceInfo), clientv3.WithLease(register.leaseId))

	if err2 != nil {
		log.Warn("etcd err", zap.Any("serviceInfo", register.serverInfo), zap.String("err", err2.Error()))
	}
	return err2

}

func (mgr *ServiceRegisterMgr) Register(info *lokas.ServiceInfo) error {

	if mgr.hasRegister(info.ServiceType, info.ServiceId) {
		log.Warn(protocol.ERR_REGISTER_SERVICE_DUPLICATED.Error(), zap.Any("serviceInfo", info))
		return protocol.ERR_REGISTER_SERVICE_DUPLICATED
	}

	register := &ServiceRegister{
		serverInfo: info,
		etcdClient: mgr.process.GetEtcd(),
	}

	// etcd register
	err := register.registerEtcd()
	if err != nil {
		return err
	}

	// etcd keep alive
	go func() {
		timer := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-timer.C:
				register.keepAliveEtcd()
			case <-register.closeChan:
				return
			}
		}
	}()

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if _, ok := mgr.registerMap[info.ServiceType]; !ok {
		mgr.registerMap[info.ServiceType] = make(map[uint16]*ServiceRegister)
	}
	mgr.registerMap[info.ServiceType][info.ServiceId] = register

	return nil
}

func (mgr *ServiceRegisterMgr) Unregister(serviceType string, serviceId uint16) error {
	// TODO
	return nil
}

func (mgr *ServiceRegisterMgr) UpdateServiceInfo(info *lokas.ServiceInfo) error {
	register, ok := mgr.findRegisterInfo(info.ServiceType, info.ServiceId)
	if !ok {
		log.Warn("not find service", zap.Any("serviceInfo", info))
		return protocol.ERR_REGISTER_SERVICE_NOT_FOUND
	}

	mgr.mutex.Lock()
	if register.serverInfo.Version == info.Version && register.serverInfo.Cnt == info.Cnt {
		mgr.mutex.Unlock()
		return nil
	}
	register.serverInfo.Version = info.Version
	register.serverInfo.Cnt = info.Cnt
	mgr.mutex.Unlock()

	err := register.updateEtcd()
	return err
}

func (mgr *ServiceRegisterMgr) hasRegister(serviceType string, serviceId uint16) bool {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	infos, ok := mgr.registerMap[serviceType]
	if !ok {
		return false
	}
	_, ok2 := infos[serviceId]
	return ok2
}

func (mgr *ServiceRegisterMgr) findRegisterInfo(serviceType string, serviceId uint16) (*ServiceRegister, bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	infos, ok := mgr.registerMap[serviceType]
	if !ok {
		return nil, false
	}
	info, ok2 := infos[serviceId]
	return info, ok2
}
