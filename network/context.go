package network

import (
	"context"
	"errors"
	"github.com/nomos/go-lokas"
	"github.com/nomos/go-lokas/util"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type stringer interface {
	String() string
}


func contextName(this context.Context) string {
	if s, ok := this.(stringer); ok {
		return s.String()
	}
	return reflect.TypeOf(this).String()
}

type reasonTimerCtx struct {
	reasonCancelCtx
	timer *time.Timer // Under cancelCtx.mu.

	deadline time.Time
}

func (this *reasonTimerCtx) Deadline() (deadline time.Time, ok bool) {
	return this.deadline, true
}

func (this *reasonTimerCtx) String() string {
	return contextName(this.reasonCancelCtx.Context) + ".WithReasonDeadline(" +
		this.deadline.String() + " [" +
		time.Until(this.deadline).String() + "])"
}

func (this *reasonTimerCtx) cancel(removeFromParent bool, err error) {
	this.reasonCancelCtx.cancel(false, err)
	if removeFromParent {
		// Remove this reasonTimerCtx from its parent cancelCtx's children.
		removeChild(this.reasonCancelCtx.Context, this)
	}
	this.mu.Lock()
	if this.timer != nil {
		this.timer.Stop()
		this.timer = nil
	}
	this.mu.Unlock()
}

var cancelCtxKey int
// closedchan is a reusable closed channel.
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

func parentCancelCtx(parent context.Context) (*reasonCancelCtx, bool) {
	done := parent.Done()
	if done == closedchan || done == nil {
		return nil, false
	}
	p, ok := parent.Value(&cancelCtxKey).(*reasonCancelCtx)
	if !ok {
		return nil, false
	}
	p.mu.Lock()
	ok = p.done == done
	p.mu.Unlock()
	if !ok {
		return nil, false
	}
	return p, true
}

// removeChild removes a context from its parent.
func removeChild(parent context.Context, child reasonCanceler) {
	p, ok := parentCancelCtx(parent)
	if !ok {
		return
	}
	p.mu.Lock()
	if p.children != nil {
		delete(p.children, child)
	}
	p.mu.Unlock()
}

// A reasonCanceler is a context type that can be canceled directly. The
// implementations are *cancelCtx and *reasonTimerCtx.
type reasonCanceler interface {
	cancel(removeFromParent bool, err error)
	Cancel(err error)
	Done() <-chan struct{}
}

type reasonCancelCtx struct {
	context.Context

	mu       sync.Mutex                  // protects following fields
	done     chan struct{}               // created lazily, closed by first cancel call
	children map[reasonCanceler]struct{} // set to nil by the first cancel call
	err      error                       // set to non-nil by the first cancel call
}

func (this *reasonCancelCtx) Value(key interface{}) interface{} {
	if key == &cancelCtxKey {
		return this
	}
	return this.Context.Value(key)
}

func (this *reasonCancelCtx) Done() <-chan struct{} {
	this.mu.Lock()
	if this.done == nil {
		this.done = make(chan struct{})
	}
	d := this.done
	this.mu.Unlock()
	return d
}

// newCancelCtx returns an initialized cancelCtx.
func newCancelCtx(parent context.Context) reasonCancelCtx {
	return reasonCancelCtx{Context: parent}
}

func (this *reasonCancelCtx) Cancel(err error) {
	this.mu.Lock()
	this.err = err
	this.mu.Unlock()
	this.cancel(true,err)
}

var goroutines int32

func propagateCancel(parent context.Context, child reasonCanceler) {
	done := parent.Done()
	if done == nil {
		return // parent is never canceled
	}

	select {
	case <-done:
		// parent is already canceled
		child.cancel(false, parent.Err())
		return
	default:
	}

	if p, ok := parentCancelCtx(parent); ok {
		p.mu.Lock()
		if p.err != nil {
			// parent has already been canceled
			child.cancel(false, p.err)
		} else {
			if p.children == nil {
				p.children = make(map[reasonCanceler]struct{})
			}
			p.children[child] = struct{}{}
		}
		p.mu.Unlock()
	} else {
		atomic.AddInt32(&goroutines, +1)
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}

func WithReasonTimeout(parent context.Context, timeout time.Duration) (context.Context, CancelFunc) {
	return WithReasonDeadline(parent, time.Now().Add(timeout))
}

func WithReasonDeadline(parent context.Context, d time.Time) (context.Context, CancelFunc) {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		// The current deadline is already sooner than the new one.
		return WithReasonCancel(parent)
	}
	c := &reasonTimerCtx{
		reasonCancelCtx: newCancelCtx(parent),
		deadline:  d,
	}
	propagateCancel(parent, c)
	dur := time.Until(d)
	if dur <= 0 {
		c.cancel(true, context.DeadlineExceeded) // deadline has already passed
		return c, func(err error) { c.cancel(false, err) }
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err == nil {
		c.timer = time.AfterFunc(dur, func() {
			c.cancel(true, context.DeadlineExceeded)
		})
	}
	return c, func(err error) { c.cancel(true,err) }
}

func WithReasonCancel(parent context.Context) (ctx context.Context, cancel CancelFunc) {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	c := newCancelCtx(parent)
	propagateCancel(parent, &c)
	return &c, func(err error) { c.cancel(true, err) }
}

func (this *reasonCancelCtx) cancel(removeFromParent bool, err error) {
	this.mu.Lock()
	if this.err != nil {
		this.mu.Unlock()
		return // already canceled
	}
	this.err = err
	if this.done == nil {
		this.done = closedchan
	} else {
		close(this.done)
	}
	for child := range this.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err)
	}
	this.children = nil
	this.mu.Unlock()

	if removeFromParent {
		removeChild(this.Context, this)
	}
}

func (this *reasonCancelCtx) Err() error {
	this.mu.Lock()
	err := this.err
	this.mu.Unlock()
	return err
}

type CancelFunc func(error)

type DefaultContext struct {
	context.Context
	transId  uint32
	resp     interface{}
	cancelFunc CancelFunc
	callback lokas.AsyncCallBack
	pending  bool
}

func (this *DefaultContext) GetInt(key string) int {
	return this.Get(key).(int)
}

func (this *DefaultContext) GetBool(key string) bool {
	return this.Get(key).(bool)
}

func (this *DefaultContext) GetInt32(key string) int32 {
	return this.Get(key).(int32)
}

func (this *DefaultContext) GetInt64(key string) int64 {
	return this.Get(key).(int64)
}

func (this *DefaultContext) GetFloat32(key string) float32 {
	return this.Get(key).(float32)
}

func (this *DefaultContext) GetFloat64(key string) float64 {
	return this.Get(key).(float64)
}

var _ lokas.IReqContext = (*DefaultContext)(nil)

var ContextFinish = errors.New("context finished")


func NewDefaultContext(ctx context.Context)*DefaultContext{
	ctx1,cancelFunc := WithReasonCancel(ctx)
	ret:=&DefaultContext{
		Context:  ctx1,
		cancelFunc:   cancelFunc,
		callback: nil,
		pending:  false,
	}
	return ret
}

func NewDefaultContextWithTimeout(ctx context.Context,transId uint32,timeout time.Duration)*DefaultContext{
	ctx1,cancelFunc := WithReasonTimeout(ctx,timeout)
	ret:=&DefaultContext{
		Context:  ctx1,
		transId: transId,
		cancelFunc:   cancelFunc,
		callback: nil,
		pending:  false,
	}
	return ret
}

func (this *DefaultContext) GetResp()interface{} {
	return this.resp
}

func (this *DefaultContext) SetResp(data interface{}) {
	this.resp = data
}

func (this *DefaultContext) Cancel(err error) {
	this.cancelFunc(err)
}

func (this *DefaultContext) Finish(){
	if this.callback!= nil {
		this.callback(this)
	}
	this.cancelFunc(nil)
}

func (this *DefaultContext) GetTransId() uint32 {
	return this.transId
}

func (this *DefaultContext) Get(name string)interface{} {
	return this.Context.Value(name)
}

func (this *DefaultContext) Set(name string,data interface{}) {
	this.Context = context.WithValue(this.Context,name,data)
}

func (this *DefaultContext) GetProcessIdType(key string)util.ProcessId {
	value:= this.Context.Value(key)
	if value!=nil {
		return value.(util.ProcessId)
	}
	return 0
}

func (this *DefaultContext) GetString(key string)string {
	value:= this.Context.Value(key)
	if value!=nil {
		return value.(string)
	}
	return ""
}

func (this *DefaultContext) GetIdType(key string)util.ID {
	value:= this.Context.Value(key)
	if value!=nil {
		return value.(util.ID)
	}
	return 0
}

func (this *DefaultContext) SetCallback(cb lokas.AsyncCallBack)  {
	this.callback = cb
}

func (this *DefaultContext) GetCallback() lokas.AsyncCallBack {
	return this.callback
}