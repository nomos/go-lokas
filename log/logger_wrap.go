package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sort"
	"sync"
)

type ObjectWriter interface {
	Write(map[string]interface{}) error
}

type ComposeLogger struct {
	cfg            *LogConfig
	fields         []zap.Field
	fieldLogger    []*ComposeLogger
	initFields     []zap.Field
	logger         *zap.Logger
	hooks          []Hook
	skip           int
	dev            bool
	consoleWriters []io.Writer
	jsonWriters    []io.Writer
	objWriters     []ObjectWriter
	fieldMu        sync.Mutex
	parent         *ComposeLogger
	stdOut         zapcore.WriteSyncer
	zapCompose 	*ZapCompose
}

func (this *ComposeLogger) Clear() {
}

func (this *ComposeLogger) recycle() {
	if this.parent != nil {
		this.parent.recycleFieldLogger(this)
	}
}

func (this *ComposeLogger) reset() {
	this.fields = make([]zap.Field, 0)
	this.fieldLogger = make([]*ComposeLogger, 0)
}

func (this *ComposeLogger) Clone() *ComposeLogger {
	return NewComposeLogger(this.dev, this.cfg, this.skip-1, this.hooks...)
}

func NewComposeLogger(dev bool, conf *LogConfig, skip int, hooks ...Hook) *ComposeLogger {
	ret := &ComposeLogger{
		cfg:            conf,
		fields:         []zap.Field{},
		fieldLogger:    []*ComposeLogger{},
		initFields:     []zap.Field{},
		hooks:          hooks,
		skip:           skip,
		dev:            dev,
		consoleWriters: []io.Writer{},
		jsonWriters:    []io.Writer{},
		objWriters:     []ObjectWriter{},
		parent:         nil,
		stdOut:         zapcore.Lock(os.Stderr),
	}
	ret.skip = skip
	ret.dev = dev
	ret.hooks = []Hook{ret}
	for _, h := range hooks {
		ret.hooks = append(ret.hooks, h)
	}
	ret.zapCompose = NewZapCompose(conf, ret.hooks...)
	logger := zap.New(ret.zapCompose, zap.AddCallerSkip(skip), zap.WithCaller(true))
	if dev {
		logger = logger.WithOptions(zap.Development())
	}
	ret.logger = logger
	return ret
}

func (this *ComposeLogger) AddHook(hook Hook) {
	this.hooks = append(this.hooks, hook)
	this.zapCompose = NewZapCompose(this.cfg, this.hooks...)
	logger := zap.New(this.zapCompose, zap.AddCallerSkip(this.skip), zap.WithCaller(true))
	if this.dev {
		logger = logger.WithOptions(zap.Development())
	}
	this.logger = logger
}

func (this *ComposeLogger) RemoveHook(hook Hook) {
	ret:=make([]Hook,0)
	for index,v:=range this.hooks {
		ret = this.hooks[:index]
		if v==hook {
			if index == len(this.hooks)-2 {
				break
			}
			for _,w:=range this.hooks[index+1:] {
				ret = append(ret, w)
			}
			break
		}
	}
	this.hooks = ret
	this.zapCompose = NewZapCompose(this.cfg, this.hooks...)
	logger := zap.New(this.zapCompose, zap.AddCallerSkip(this.skip), zap.WithCaller(true))
	if this.dev {
		logger = logger.WithOptions(zap.Development())
	}
	this.logger = logger
}

func (this *ComposeLogger) Write(p []byte) (int, error) {
	this.Info(string(p))
	return 0, nil
}

func (this *ComposeLogger) WriteString(s string) {
	this.Info(s)
}

func (this *ComposeLogger) WriteConsole(ent zapcore.Entry, p []byte) error {
	go func() {
		for _, writer := range this.consoleWriters {
			_, err := writer.Write(p)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()
	return nil
}

func (this *ComposeLogger) WriteJson(ent zapcore.Entry, p []byte) error {
	for _, writer := range this.jsonWriters {
		_, err := writer.Write(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *ComposeLogger) WriteObject(ent zapcore.Entry, data map[string]interface{}) error {
	for _, writer := range this.objWriters {
		err := writer.Write(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *ComposeLogger) getFieldLogger() *ComposeLogger {
	this.fieldMu.Lock()
	defer this.fieldMu.Unlock()
	if len(this.fieldLogger) == 0 {
		return this.Clone()
	}
	ret := this.fieldLogger[0]
	this.fieldLogger = this.fieldLogger[1:]
	return ret
}

func (this *ComposeLogger) recycleFieldLogger(logger *ComposeLogger) {
	this.fieldMu.Lock()
	defer this.fieldMu.Unlock()
	logger.reset()
	this.fieldLogger = append(this.fieldLogger, logger)
}

func (this *ComposeLogger) SetConsoleWriter(writer io.Writer) {
	this.consoleWriters = append(this.consoleWriters, writer)
}

func (this *ComposeLogger) RemoveConsoleWriter(writer io.Writer){
	ret:=make([]io.Writer,0)
	for index,v:=range this.consoleWriters {
		ret = this.consoleWriters[:index]
		if v==writer {
			if index == len(this.consoleWriters)-2 {
				break
			}
			for _,w:=range this.consoleWriters[index+1:] {
				ret = append(ret, w)
			}
			break
		}
	}
	this.consoleWriters = ret
}

func (this *ComposeLogger) SetJsonWriter(writer io.Writer) {
	this.jsonWriters = append(this.jsonWriters, writer)
}

func (this *ComposeLogger) SetObjectWriter(writer ObjectWriter) {
	this.objWriters = append(this.objWriters, writer)
}

func (this *ComposeLogger) WithFields(fields Fields) *ComposeLogger {
	logger := this.getFieldLogger()
	for k, v := range fields {
		logger.fields = append(logger.fields, zap.Any(k, v))
	}
	sort.Slice(logger.fields, func(i, j int) bool {
		return logger.fields[i].Key < logger.fields[j].Key
	})
	return logger
}

func (this *ComposeLogger) WithField(key string, value interface{}) *ComposeLogger {
	logger := this.getFieldLogger()
	logger.fields = append(logger.fields, zap.Any(key, value))
	sort.Slice(logger.fields, func(i, j int) bool {
		return logger.fields[i].Key < logger.fields[j].Key
	})
	return logger
}

func (this *ComposeLogger) cloneFields() []zap.Field {
	ret := make([]zap.Field, 0)
	for _, field := range this.initFields {
		ret = append(ret, field)
	}
	for _, field := range this.fields {
		ret = append(ret, field)
	}
	return ret
}

func (this *ComposeLogger) Info(msg string, fields ...zap.Field) {
	logFields := this.cloneFields()
	for _, field := range fields {
		logFields = append(logFields, field)
	}
	this.logger.Info(msg, logFields...)
	this.recycle()
}

func (this *ComposeLogger) Debug(msg string, fields ...zap.Field) {
	logFields := this.cloneFields()
	for _, field := range fields {
		logFields = append(logFields, field)
	}
	this.logger.Debug(msg, logFields...)
	this.recycle()
}

func (this *ComposeLogger) Warn(msg string, fields ...zap.Field) {
	logFields := this.cloneFields()
	for _, field := range fields {
		logFields = append(logFields, field)
	}
	this.logger.Warn(msg, logFields...)
	this.recycle()
}

func (this *ComposeLogger) Error(msg string, fields ...zap.Field) {
	logFields := this.cloneFields()
	for _, field := range fields {
		logFields = append(logFields, field)
	}
	this.logger.Error(msg, logFields...)
	this.recycle()
}

func (this *ComposeLogger) Panic(msg string, fields ...zap.Field) {
	logFields := this.cloneFields()
	for _, field := range fields {
		logFields = append(logFields, field)
	}
	this.logger.Panic(msg, logFields...)
	this.recycle()
}

func (this *ComposeLogger) Fatal(msg string, fields ...zap.Field) {
	logFields := this.cloneFields()
	for _, field := range fields {
		logFields = append(logFields, field)
	}
	this.logger.Fatal(msg, logFields...)
	this.recycle()
}

func (this *ComposeLogger) Infof(args ...interface{}) {
	this.logger.Sugar().Info(sprintf(args))
	this.recycle()
}

func (this *ComposeLogger) Warnf(args ...interface{}) {
	this.logger.Sugar().Warn(sprintf(args))
	this.recycle()
}

func (this *ComposeLogger) Debugf(args ...interface{}) {
	this.logger.Sugar().Debug(sprintf(args))
	this.recycle()
}

func (this *ComposeLogger) Errorf(args ...interface{}) {
	this.logger.Sugar().Error(sprintf(args))
	this.recycle()
}

func (this *ComposeLogger) Panicf(args ...interface{}) {
	this.logger.Sugar().Panic(sprintf(args))
	this.recycle()
}

func (this *ComposeLogger) Fatalf(args ...interface{}) {
	this.logger.Sugar().Fatal(sprintf(args))
	this.recycle()
}

func (this *ComposeLogger) Sync() error {
	return this.logger.Sync()
}
