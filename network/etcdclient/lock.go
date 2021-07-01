package etcdclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/nomos/go-lokas/log"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"go.etcd.io/etcd/client/v3"
	"io"
	"os"
	"sync"
	"time"
)

const (
	defaultTTL   = 15
	defaultTry   = 3
	prefix = "/mutex"
)

type Mutex struct {
	key    string
	id     string // The identity of the caller
	client *clientv3.Client
	kapi   clientv3.KV
	ctx    context.Context
	ttl    int64
	mutex  *sync.Mutex
	logger io.Writer
	lease clientv3.Lease
	watcher clientv3.Watcher
	leaseId clientv3.LeaseID
}

func CreateMutex(c *clientv3.Client,key string, ttl int) (*Mutex, error) {

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, errors.New("wrong lock key")
	}

	if key[0] != '/' {
		key = "/" + key
	}

	if ttl < 1 {
		ttl = defaultTTL
	}

	return &Mutex{
		key:    key,
		id:     fmt.Sprintf("%v-%v-%v", hostname, os.Getpid(), time.Now().Format("20060102-15:04:05.999999999")),
		client: c,
		mutex:  new(sync.Mutex),
		kapi:   clientv3.NewKV(c),
		ctx:    context.TODO(),
		ttl:    int64(ttl),

	}, nil
}

func (this *Mutex) Lock() (err error) {
	this.lease = clientv3.NewLease(this.client)
	this.watcher = clientv3.NewWatcher(this.client)
	this.mutex.Lock()
	for try := 1; try <= defaultTry; try++ {
		err = this.lock()
		if err == nil {
			return nil
		}

		log.Errorf("Lock node %v ERROR %v", this.key, err)
		if try < defaultTry {
			log.Infof("Try to lock node %v again", this.key, err)
		}
	}
	return err
}

func (this *Mutex) lock() (err error) {
	log.Infof("Trying to create a node : key=%v", prefix+this.key)

	for {
		leaseResp,err:=this.lease.Grant(this.ctx,this.ttl)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		this.leaseId = leaseResp.ID
		resp, err := this.kapi.Put(this.ctx, prefix+this.key, this.id, clientv3.WithPrevKV(),clientv3.WithLease(leaseResp.ID))
		if err == nil&&resp.PrevKv==nil {
			log.Infof("Create node %v OK [%q]", this.key, log.PrettyStruct(resp))
			return nil
		}
		log.Errorf("Create node %v failed [%v]", this.key, err)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		gResp, err := this.kapi.Get(this.ctx, this.key)
		if err != nil {
			return err
		}
		log.Infof("Get node %v OK", this.key)
		watcher := this.watcher.Watch(this.ctx,prefix+this.key, clientv3.WithRev(gResp.Header.Revision))
		log.Infof("Watching %v ...", this.key)
		for {
			select {
			case wResp:=<-watcher:
				if wResp.Err()!=nil {
					return err
				}
				for _,e:=range wResp.Events {
					if e.Type  == clientv3.EventTypeDelete {
						goto LoopEnd
					}
					log.Infof("Received an event : %q", resp)
					log.Infof("Received an event : %q", e)
				}
			}
		}
		LoopEnd:
	}
	return err
}

func (this *Mutex) Unlock() (err error) {
	defer this.watcher.Close()
	defer this.lease.Close()
	defer this.mutex.Unlock()
	for i := 1; i <= defaultTry; i++ {
		var resp *clientv3.DeleteResponse
		resp, err = this.kapi.Delete(this.ctx, prefix+this.key)
		if err == nil {
			log.Infof("Delete %v OK", this.key)
			return nil
		}
		log.Errorf("Delete %v falied: %q", this.key, resp)

		if err == rpctypes.ErrEmptyKey {
			return nil
		}
		log.Error(err.Error())
	}
	return err
}

func (this *Mutex) RefreshLockTTL(ttl time.Duration) (err error) {
	resp,err:=this.lease.KeepAliveOnce(context.TODO(),this.leaseId)
	if err != nil {
		log.Errorf("Refresh ttl of %v failed [%q]", this.key, resp)
	} else {
		log.Infof("Refresh ttl of %v OK", this.key)
	}
	return err
}
