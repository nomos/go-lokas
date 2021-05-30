package netstat

import (
	"sync/atomic"
	"time"
)

type StatData struct {
	messagesRecv  int64 // message count received
	messagesSend  int64 // message count sent
	bytesRecv     int64 // byte count received
	bytesSend     int64 // byte count sent
	sendChanItems int64 // buffered item count of send chan
	recvChanItems int64 // buffered item count of recv chan
}

func (this *StatData) GetMessagesRecv() int {
	return int(this.messagesRecv)
}

func (this *StatData) GetMessagesSend() int {
	return int(this.messagesSend)
}

func (this *StatData) GetBytesRecv() int {
	return int(this.bytesRecv)
}

func (this *StatData) GetBytesSend() int {
	return int(this.bytesSend)
}

func (this *StatData) GetSendChanItemCount() int {
	return int(this.sendChanItems)
}

func (this *StatData) GetRecvChanItemCount() int {
	return int(this.recvChanItems)
}

type NetStat struct {
	current   *StatData
	max       *StatData
	total     *StatData
	lastTotal *StatData
}

func NewNetStat() *NetStat {
	s := &NetStat{
		current:   new(StatData),
		max:       new(StatData),
		total:     new(StatData),
		lastTotal: new(StatData),
	}
	if atomic.CompareAndSwapInt32(&isRunning, 0, 1) {
		manager = newStatManager()
		manager.run()
	}
	<-manager.start
	return s
}

func (this *NetStat) Start() {
	manager.startChan <- this
}

func (this *NetStat) Stop() {
	manager.stopChan <- this
}

func (this *NetStat) GetCurrent() *StatData {
	return this.current
}

func (this *NetStat) GetMax() *StatData {
	return this.max
}

func (this *NetStat) GetTotal() *StatData {
	return this.total
}

func (this *NetStat) doCalc() {
	total, current, max, lastTotal := this.total, this.current, this.max, this.lastTotal
	messagesSend := atomic.LoadInt64(&total.messagesSend)
	messagesRecv := atomic.LoadInt64(&total.messagesRecv)
	bytesSend := atomic.LoadInt64(&total.bytesSend)
	bytesRecv := atomic.LoadInt64(&total.bytesRecv)

	current.messagesSend = messagesSend - lastTotal.messagesSend
	current.messagesRecv = messagesRecv - lastTotal.messagesRecv
	current.bytesSend = bytesSend - lastTotal.bytesSend
	current.bytesRecv = bytesRecv - lastTotal.bytesRecv

	if current.messagesSend > max.messagesSend {
		max.messagesSend = current.messagesSend
	}
	if current.messagesRecv > max.messagesRecv {
		max.messagesRecv = current.messagesRecv
	}
	if current.bytesSend > max.bytesSend {
		max.bytesSend = current.bytesSend
	}
	if current.bytesRecv > max.bytesRecv {
		max.bytesSend = current.bytesRecv
	}

	lastTotal.messagesSend = messagesSend
	lastTotal.messagesRecv = messagesRecv
	lastTotal.bytesSend = bytesSend
	lastTotal.bytesRecv = bytesRecv
}

func (this *NetStat) AddSendStat(msgLen int, msgCount int) {
	data := this.total
	atomic.AddInt64(&data.bytesSend, int64(msgLen))
	atomic.AddInt64(&data.messagesSend, int64(msgCount))
}

func (this *NetStat) AddRecvStat(msgLen int, msgCount int) {
	data := this.total
	atomic.AddInt64(&data.bytesRecv, int64(msgLen))
	atomic.AddInt64(&data.messagesRecv, int64(msgCount))
}

func (this *NetStat) SetSendChanItemCount(chanItemCount int) {
	data := this.total
	count := atomic.LoadInt64(&data.sendChanItems)
	if int64(chanItemCount) > count {
		atomic.StoreInt64(&data.sendChanItems, int64(chanItemCount))
	}
}

func (this *NetStat) SetRecvChanItemCount(chanItemCount int) {
	data := this.total
	count := atomic.LoadInt64(&data.recvChanItems)
	if int64(chanItemCount) > count {
		atomic.StoreInt64(&data.recvChanItems, int64(chanItemCount))
	}
}

type statManager struct {
	startChan chan *NetStat
	stopChan  chan *NetStat
	stats     map[*NetStat]bool
	start     chan struct{}
}

var (
	isRunning int32
	manager   *statManager
)

func newStatManager() *statManager {
	return &statManager{
		startChan: make(chan *NetStat, 1000),
		stopChan:  make(chan *NetStat, 1000),
		stats:     make(map[*NetStat]bool),
		start:     make(chan struct{}),
	}
}

func (this *statManager) run() {
	close(this.start)
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case s := <-this.startChan:
			this.stats[s] = true
		case s := <-this.stopChan:
			delete(this.stats, s)
		case <-ticker.C:
			for stat := range this.stats {
				stat.doCalc()
			}
		}
	}
}
