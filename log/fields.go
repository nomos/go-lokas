package log

import "go.uber.org/zap"

type ZapFields []zap.Field

func (this ZapFields) Concat(fields ZapFields) ZapFields {
	ret := []zap.Field{}
	for _, f := range this {
		ret = append(ret, f)
	}
	for _, f := range fields {
		ret = append(ret, f)
	}
	return ret
}

func (this ZapFields) Append(field ...zap.Field) ZapFields {
	ret := []zap.Field{}
	for _, f := range this {
		ret = append(ret, f)
	}
	for _, v := range field {
		ret = append(ret, v)
	}
	return ret
}
