package log

import (
	"go.uber.org/zap/zapcore"
	"io"
)

type Hook interface {
	WriteConsole(zapcore.Entry, []byte) error
	WriteJson(zapcore.Entry, []byte) error
	WriteObject(zapcore.Entry, map[string]interface{}) error
}

type defaultHook struct {
	consoleWriter io.Writer
	jsonWriter io.Writer
	objWriter ObjectWriter
}

func NewConsoleWriter(writer io.Writer)Hook{
	return &defaultHook{
		consoleWriter: writer,
		jsonWriter:    nil,
		objWriter:     nil,
	}
}

func NewJsonWriter(writer io.Writer)Hook{
	return &defaultHook{
		consoleWriter: nil,
		jsonWriter:    writer,
		objWriter:     nil,
	}
}

func NewObjectWriter(writer ObjectWriter)Hook{
	return &defaultHook{
		consoleWriter: nil,
		jsonWriter:    nil,
		objWriter:     writer,
	}
}

func (this *defaultHook) WriteConsole(ent zapcore.Entry,p []byte) error {
	if this.consoleWriter!= nil {
		_,err:=this.consoleWriter.Write(p)
		return err
	}
	return nil
}

func (this *defaultHook) WriteJson(ent zapcore.Entry,p []byte) error {
	if this.jsonWriter!= nil {
		_,err:=this.jsonWriter.Write(p)
		return err
	}
	return nil
}

func (this *defaultHook) WriteObject(ent zapcore.Entry,p map[string]interface{}) error{
	if this.objWriter!= nil {
		return this.objWriter.Write(p)
	}
	return nil
}