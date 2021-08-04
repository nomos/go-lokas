package events

import (
	"github.com/nomos/go-lokas/log"
	"reflect"
	"sync"
)

const (
	DefaultMaxListeners = 0
	EnableWarning = false
)

type (
	EventName string
	Listener  func(...interface{})
	Events    map[EventName][]Listener

	EventEmmiter interface {
		AddListener(EventName, ...Listener)
		Emit(EventName, ...interface{})
		EventNames() []EventName
		GetMaxListeners() int
		ListenerCount(EventName) int
		Listeners(EventName) []Listener
		On(EventName, ...Listener)
		Once(EventName, ...Listener)
		RemoveAllListeners(EventName) bool
		RemoveListener(EventName, Listener) bool
		Clear()
		SetMaxListeners(int)
		Len() int
	}
)

type BaseEmitter struct {
	maxListeners int
	evtListeners Events
	mu           sync.Mutex
}

func (e Events) CopyTo(BaseEmitter EventEmmiter) {
	if e != nil && len(e) > 0 {
		for evt, listeners := range e {
			if len(listeners) > 0 {
				BaseEmitter.AddListener(evt, listeners...)
			}
		}
	}
}

func New() EventEmmiter {
	return &BaseEmitter{maxListeners: DefaultMaxListeners, evtListeners: Events{}}
}

var (
	_              EventEmmiter = &BaseEmitter{}
	defaultEmmiter              = New()
)

func AddListener(evt EventName, listener ...Listener) {
	defaultEmmiter.AddListener(evt, listener...)
}

func (e *BaseEmitter) AddListener(evt EventName, listener ...Listener) {
	if len(listener) == 0 {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.evtListeners == nil {
		e.evtListeners = Events{}
	}

	listeners := e.evtListeners[evt]

	if e.maxListeners > 0 && len(listeners) == e.maxListeners {
		if EnableWarning {
			log.Warnf(`(events) warning: possible EventEmitter memory '
                    leak detected. %d listeners added. '
                    Use emitter.SetMaxListeners(n int) to increase limit.`, len(listeners))
		}
		return
	}

	if listeners == nil {
		listeners = make([]Listener, e.maxListeners)
	}

	e.evtListeners[evt] = append(listeners, listener...)
}

func Emit(evt EventName, data ...interface{}) {
	defaultEmmiter.Emit(evt, data...)
}

func (e *BaseEmitter) Emit(evt EventName, data ...interface{}) {
	if e.evtListeners == nil {
		return
	}
	if listeners := e.evtListeners[evt]; listeners != nil && len(listeners) > 0 {
		for i := range listeners {
			l := listeners[i]
			if l != nil {
				l(data...)
			}
		}
	}
}

func EventNames() []EventName {
	return defaultEmmiter.EventNames()
}

func (e *BaseEmitter) EventNames() []EventName {
	if e.evtListeners == nil || e.Len() == 0 {
		return nil
	}

	names := make([]EventName, e.Len(), e.Len())
	i := 0
	for k := range e.evtListeners {
		names[i] = k
		i++
	}
	return names
}

func GetMaxListeners() int {
	return defaultEmmiter.GetMaxListeners()
}

func (e *BaseEmitter) GetMaxListeners() int {
	return e.maxListeners
}

func ListenerCount(evt EventName) int {
	return defaultEmmiter.ListenerCount(evt)
}

func (e *BaseEmitter) ListenerCount(evt EventName) int {
	if e.evtListeners == nil {
		return 0
	}
	len := 0

	if evtListeners := e.evtListeners[evt]; evtListeners != nil {
		for _, l := range evtListeners {
			if l == nil {
				continue
			}
			len++
		}
	}

	return len
}

func Listeners(evt EventName) []Listener {
	return defaultEmmiter.Listeners(evt)
}

func (e *BaseEmitter) Listeners(evt EventName) []Listener {
	if e.evtListeners == nil {
		return nil
	}
	var listeners []Listener
	if evtListeners := e.evtListeners[evt]; evtListeners != nil {
		for _, l := range evtListeners {
			if l == nil {
				continue
			}

			listeners = append(listeners, l)
		}

		if len(listeners) > 0 {
			return listeners
		}
	}

	return nil
}

func On(evt EventName, listener ...Listener) {
	defaultEmmiter.On(evt, listener...)
}

func (e *BaseEmitter) On(evt EventName, listener ...Listener) {
	e.AddListener(evt, listener...)
}

func Once(evt EventName, listener ...Listener) {
	defaultEmmiter.Once(evt, listener...)
}

func (e *BaseEmitter) Once(evt EventName, listener ...Listener) {
	if len(listener) == 0 {
		return
	}

	var modifiedListeners []Listener

	if e.evtListeners == nil {
		e.evtListeners = Events{}
	}

	for i, l := range listener {

		idx := len(e.evtListeners) + i
		func(listener Listener, index int) {
			fired := false
			modifiedListeners = append(modifiedListeners, func(data ...interface{}) {
				if e.evtListeners == nil {
					return
				}
				if !fired {
					if e.evtListeners[evt] != nil && (len(e.evtListeners[evt]) > index || index == 0) {

						e.mu.Lock()
						e.evtListeners[evt][index] = nil
						e.mu.Unlock()
					}
					fired = true
					listener(data...)
				}

			})
		}(l, idx)

	}
	e.AddListener(evt, modifiedListeners...)
}

func RemoveAllListeners(evt EventName) bool {
	return defaultEmmiter.RemoveAllListeners(evt)
}

func (e *BaseEmitter) RemoveAllListeners(evt EventName) bool {
	if e.evtListeners == nil {
		return false
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if listeners := e.evtListeners[evt]; listeners != nil {
		l := e.ListenerCount(evt)
		delete(e.evtListeners, evt)
		if l > 0 {
			return true
		}
	}

	return false
}

func (e *BaseEmitter) RemoveListener(evt EventName, listener Listener) bool {
	if e.evtListeners == nil {
		return false
	}

	if listener == nil {
		return false
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	listeners := e.evtListeners[evt]
	if listeners == nil {
		return false
	}

	idx := -1
	listenerPointer := reflect.ValueOf(listener).Pointer()

	for index, item := range listeners {
		itemPointer := reflect.ValueOf(item).Pointer()
		if itemPointer == listenerPointer {
			idx = index
			break
		}
	}

	if idx < 0 {
		return false
	}

	var modifiedListeners []Listener = nil

	if len(listeners) > 1 {
		modifiedListeners = append(listeners[:idx], listeners[idx+1:]...)
	}

	e.evtListeners[evt] = modifiedListeners

	return true
}

func Clear() {
	defaultEmmiter.Clear()
}

func (e *BaseEmitter) Clear() {
	e.evtListeners = Events{}
}

func SetMaxListeners(n int) {
	defaultEmmiter.SetMaxListeners(n)
}

func (e *BaseEmitter) SetMaxListeners(n int) {
	if n < 0 {
		if EnableWarning {
			log.Printf("(events) warning: MaxListeners must be positive number, tried to set: %d", n)
			return
		}
	}
	e.maxListeners = n
}

func Len() int {
	return defaultEmmiter.Len()
}

func (e *BaseEmitter) Len() int {
	if e.evtListeners == nil {
		return 0
	}
	return len(e.evtListeners)
}
