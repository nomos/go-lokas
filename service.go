package lokas

import (
	"strings"

	"github.com/nomos/go-lokas/protocol"
	"github.com/nomos/go-lokas/util"
)

//Service the registry for service details
type Service struct {
	Id        protocol.BINARY_TAG //service Id
	ProcessId util.ProcessId      //server Id
	ActorId   util.ID             //actor Id
	Type      ServiceType         //rpc pub/sub stateless
	Host      string
	Port      int
}

type ServiceInfo struct {
	// Id          uint32
	ServiceType string
	ServiceId   uint16
	LineId      uint16
	ProcessId   util.ProcessId
	ActorId     util.ID
	// GameId      string
	Host    string
	Port    uint16
	Version string
	Cnt     int

	// CreateAt time.Time
}

type ServiceInfos []*ServiceInfo

func (infos ServiceInfos) Len() int { return len(infos) }

func (infos ServiceInfos) Swap(i, j int) { infos[i], infos[j] = infos[j], infos[i] }

func (infos ServiceInfos) Less(i, j int) bool {
	diff1 := strings.Compare(infos[i].ServiceType, infos[j].ServiceType)
	if diff1 > 0 {
		return false
	} else if diff1 < 0 {
		return true
	}

	if infos[i].ServiceId < infos[j].ServiceId {
		return true
	} else if infos[i].ServiceId > infos[j].ServiceId {
		return false
	}

	if infos[i].LineId <= infos[j].LineId {
		return true
	} else {
		return false
	}

}

type UserRouteInfo struct {
	UserId        util.ID
	GateServiceId uint16
	ServiceInfo   *ServiceInfo
}
