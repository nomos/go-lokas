package lokas

import (
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
