package lokas

import (
	"time"

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
	// GameId      string
	Host     string
	Port     uint16
	Version  string
	Cnt      int
	CreateAt time.Time
}
