package logger

import (
	"context"
	"fmt"
	"github.com/LyricTian/queue"
	"github.com/nomos/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

type ExecCloser interface {
	Exec(entry *zapcore.Entry,m map[string]interface{}) error
	Close() error
}

type defaultExec struct {
	sess     qmgo.Database
	dbName   string
	cName    string
	canClose bool
}

func (this *defaultExec) Exec(entry *zapcore.Entry,m map[string]interface{}) error {
	item := make(bson.M)
	item["level"] = entry.Level
	item["time"] = entry.Time
	item["caller"] = entry.Caller
	item["msg"] = entry.Message
	for k,b:=range m {
		item[k] = b
	}

	_,err:=this.sess.Collection(this.cName).InsertOne(context.TODO(),item)
	return err
}

func (this *defaultExec) Close() error {
	if this.canClose {
		//this.sess.Close()
	}
	return nil
}

func NewExec(sess qmgo.Database, dbName, cName string) ExecCloser {
	return &defaultExec{
		sess:   sess,
		dbName: dbName,
		cName:  cName,
	}
}

var defaultOptions = mongoHookOpt{
	maxQueues:  512,
	maxWorkers: 2,
	level:zap.PanicLevel,
}

type MongoHookOption func(*mongoHookOpt)

type mongoHookOpt struct {
	maxQueues  int
	maxWorkers int
	extra      map[string]interface{}
	level      zapcore.Level
	filter FilterHandle
	exec ExecCloser
	out io.Writer
}

type MongoHook struct {
	opts mongoHookOpt
	q    *queue.Queue
}

//设置队列数量
func SetMaxQueues(maxQueues int) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.maxQueues = maxQueues
	}
}

//设置Worker数量
func SetMaxWorkers(maxWorkers int) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.maxWorkers = maxWorkers
	}
}

//设置额外的Field
func SetExtra(extra map[string]interface{}) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.extra = extra
	}
}

func SetExec(exec ExecCloser) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.exec = exec
	}
}

type FilterHandle func(m map[string]interface{}) map[string]interface{}

//设置过滤器
func SetFilter(filter FilterHandle) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.filter = filter
	}
}

//设置日志级别
func SetLevels(level zapcore.Level) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.level = level
	}
}

//设置额外输出流
func SetOut(out io.Writer) MongoHookOption {
	return func(o *mongoHookOpt) {
		o.out = out
	}
}

func NewMongoHook(opt ...MongoHookOption) *MongoHook {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	q := queue.NewQueue(opts.maxQueues, opts.maxWorkers)
	q.Run()

	return &MongoHook{
		opts: opts,
		q:    q,
	}
}

func (this *MongoHook) Level() zapcore.Level {
	return zapcore.WarnLevel
}

func (this *MongoHook) Fire(entry *zapcore.Entry,m map[string]interface{}) error {
	this.q.Push(queue.NewJob(entry, func(v interface{}) {
		this.exec(v.(*zapcore.Entry),m)
	}))
	return nil
}

func (this *MongoHook) copyEntry(e *zapcore.Entry) *zapcore.Entry {
	entry := &zapcore.Entry{
		Level:      e.Level,
		Time:       e.Time,
		LoggerName: e.LoggerName,
		Message:    e.Message,
		Caller:     e.Caller,
		Stack:      e.Stack,
	}
	return entry
}

func (this *MongoHook) exec(entry *zapcore.Entry,m map[string]interface{}) {
	for k,v:=range this.opts.extra {
		m[k] = v
	}
	if filter := this.opts.filter; filter != nil {
		m = filter(m)
	}

	err := this.opts.exec.Exec(entry,m)
	if err != nil && this.opts.out != nil {
		fmt.Fprintf(this.opts.out, "[Mongo-MongoHook] Execution error: %s\n", err.Error())
	}
}

//等待直到队列为空
func (this *MongoHook) Flush() {
	this.q.Terminate()
	this.opts.exec.Close()
}

func (this *MongoHook) WriteConsole(ent zapcore.Entry,p []byte) error {
	return nil
}

func (this *MongoHook) WriteJson(ent zapcore.Entry,p []byte) error {
	return nil
}

func (this *MongoHook) WriteObject(ent zapcore.Entry,m map[string]interface{}) error {
	return this.Fire(&ent,m)
}