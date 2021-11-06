package lox

import (
	"github.com/nomos/go-lokas/protocol"
)

type ConsoleEvent struct {
	Text string
}

func NewConsoleEvent(text string)*ConsoleEvent{
	ret:=&ConsoleEvent{
		Text: text,
	}
	return ret
}

func (this *ConsoleEvent) GetId()(protocol.BINARY_TAG,error){
	return TAG_CONSOLE_EVENT,nil
}

func (this *ConsoleEvent) Serializable()protocol.ISerializable {
	return this
}
