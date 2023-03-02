package flog

import (
	"github.com/nomos/go-lokas/log"
	"go.uber.org/zap"
)

func ServiceId(serviceId uint16) zap.Field {
	return zap.Uint16("serviceid", serviceId)
}

func LineId(lineId uint16) zap.Field {
	return zap.Uint16("lineid", lineId)
}

func ServiceType(serviceType string) zap.Field {
	return zap.String("servicetype", serviceType)
}
func ServiceInfo(serviceType string, serviceId uint16, lineId uint16) log.ZapFields {
	ret := []zap.Field{}
	ret = append(ret, ServiceType(serviceType))
	ret = append(ret, LineId(lineId))
	ret = append(ret, ServiceId(serviceId))
	return ret
}
