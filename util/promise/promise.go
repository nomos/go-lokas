package promise

import (
	"errors"
	"github.com/nomos/go-lokas/util"
	"sync"
	"time"
)

type Promise[T any] struct {
	pending bool

	executor func(resolve func(any), reject func(any))

	result any

	err error

	mutex sync.Mutex

	elapseTime time.Duration

	calTime bool

	wg sync.WaitGroup
}

type Timeout struct {
	isClose   bool
	closeChan chan struct{}
}

func (this *Timeout) IsClose() bool {
	return this.closeChan == nil
}

func (this *Timeout) Close() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				util.Recover(r, false)
			}
		}()
		if this.closeChan != nil {
			this.closeChan <- struct{}{}
		}
	}()
}

type Interval struct {
	interval  time.Duration
	ticker    *time.Ticker
	closeChan chan struct{}
	f         func(interval *Interval)
}

func (this *Interval) IsClose() bool {
	return this.closeChan == nil
}

func (this *Interval) Close() {
	go func() {
		this.closeChan <- struct{}{}
	}()
}

func (this *Timeout) execute(duration time.Duration, f func(*Timeout)) {
	this.closeChan = make(chan struct{})
	go func() {
		for {
			select {
			case <-this.closeChan:
				return
			case <-time.After(duration):
				f(this)
				this.isClose = true
				close(this.closeChan)
				this.closeChan = nil
				return
			}
		}
	}()
}

func (this *Promise[T]) Any() *Promise[any] {
	return (*Promise[any])(this)
}
func (this *Promise[T]) CalTime() *Promise[T] {
	this.calTime = true
	return this
}

func (this *Promise[T]) Elapse() time.Duration {
	return this.elapseTime
}

func (this *Promise[T]) Resolve(resolution any) {
	this.mutex.Lock()

	if !this.pending {
		this.mutex.Unlock()
		return
	}

	switch result := resolution.(type) {
	case *Promise[T]:
		flattenedResult, err := result.Await()
		if err != nil {
			this.mutex.Unlock()
			this.Reject(err)
			return
		}
		this.result = flattenedResult
	default:
		this.result = result
	}
	this.pending = false

	this.wg.Done()
	this.mutex.Unlock()
}

func (this *Promise[T]) Reject(err any) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if !this.pending {
		return
	}
	if err1, ok := err.(error); ok {
		this.err = err1
	} else {
		this.err = errors.New(err.(string))
	}
	this.pending = false

	this.wg.Done()
}

func (this *Promise[T]) handlePanic() {
	var r = recover()
	if r != nil {
		if err, ok := r.(error); ok {
			this.Reject(errors.New(err.Error()))
		} else {
			this.Reject(errors.New(r.(string)))
		}
	}
}

func (this *Promise[T]) Then(fulfillment func(data any) any) *Promise[T] {
	return Async[T](func(resolve func(T), reject func(any)) {
		result, err := this.Await()
		if err != nil {
			reject(err)
			return
		}
		resolve(fulfillment(result).(T))
	})
}

func (this *Promise[T]) Catch(rejection func(err error) any) *Promise[T] {
	return Async[T](func(resolve func(T), reject func(any)) {
		result, err := this.Await()
		if err != nil {
			reject(rejection(err))
			return
		}
		resolve(result.(T))
	})
}

func (this *Promise[T]) Await() (any, error) {
	if this.calTime {
		start := time.Now()
		this.wg.Wait()
		this.elapseTime = time.Now().Sub(start)
		return this.result, this.err
	}
	this.wg.Wait()
	return this.result, this.err
}

func (this *Promise[T]) AsCallback(f func(any, error)) {
	go func() {
		this.wg.Wait()
		f(this.result, this.err)
	}()
}

type resolutionHelper struct {
	index int
	data  any
}

func SetTimeout(duration time.Duration, f func(*Timeout)) *Timeout {
	ret := &Timeout{
		isClose: true,
	}
	ret.execute(duration, f)
	return ret
}

func SetInterval(duration time.Duration, f func(*Interval)) *Interval {
	ret := &Interval{
		interval:  duration,
		ticker:    time.NewTicker(duration),
		closeChan: nil,
		f:         f,
	}
	go func() {
		for {
			select {
			case <-ret.ticker.C:
				f(ret)
			case <-ret.closeChan:
				close(ret.closeChan)
				return
			}
		}
	}()
	return ret
}

func Async[T any](executor func(resolve func(T), reject func(any))) *Promise[T] {
	var promise = &Promise[T]{
		pending: true,
		executor: func(resolve func(any), reject func(any)) {
			executor(func(t T) {
				resolve(t)
			}, reject)
		},
		result: any(nil),
		err:    nil,
		mutex:  sync.Mutex{},
		wg:     sync.WaitGroup{},
	}

	promise.wg.Add(1)

	go func() {
		defer promise.handlePanic()
		promise.executor(promise.Resolve, promise.Reject)
	}()

	return promise
}

func Await[T any](p *Promise[T]) (any, error) {
	return p.Await()
}

func Each(promises ...*Promise[any]) *Promise[any] {
	return Async[any](func(resolve func(any), reject func(any)) {
		resolutions := make([]any, 0)
		for _, promise := range promises {
			result, err := promise.Await()
			if err != nil {
				reject(err)
				return
			}
			resolutions = append(resolutions, result)
		}
		resolve(resolutions)
	})
}

func All(promises ...*Promise[any]) *Promise[any] {
	psLen := len(promises)
	if psLen == 0 {
		return Resolve[any](make([]any, 0))
	}

	return Async[any](func(resolve func(any), reject func(any)) {
		resolutionsChan := make(chan resolutionHelper, psLen)
		errorChan := make(chan error, psLen)

		for index, promise := range promises {
			func(i int) {
				promise.Then(func(data any) any {
					resolutionsChan <- resolutionHelper{i, data}
					return data
				}).Catch(func(err error) any {
					errorChan <- err
					return err
				})
			}(index)
		}

		resolutions := make([]any, psLen)
		for x := 0; x < psLen; x++ {
			select {
			case resolution := <-resolutionsChan:
				resolutions[resolution.index] = resolution.data

			case err := <-errorChan:
				reject(err)
				return
			}
		}
		resolve(resolutions)
	})
}

func Race(promises ...*Promise[any]) *Promise[any] {
	psLen := len(promises)
	if psLen == 0 {
		return Resolve[any](nil)
	}

	return Async[any](func(resolve func(any), reject func(any)) {
		resolutionsChan := make(chan any, psLen)
		errorChan := make(chan error, psLen)

		for _, promise := range promises {
			promise.Then(func(data any) any {
				resolutionsChan <- data
				return data
			}).Catch(func(err error) any {
				errorChan <- err
				return err
			})
		}

		select {
		case resolution := <-resolutionsChan:
			resolve(resolution)

		case err := <-errorChan:
			reject(err)
		}
	})
}

func AllSettled(promises ...*Promise[any]) *Promise[any] {
	psLen := len(promises)
	if psLen == 0 {
		return Resolve[any](nil)
	}

	return Async[any](func(resolve func(any), reject func(any)) {
		resolutionsChan := make(chan resolutionHelper, psLen)

		for index, promise := range promises {
			func(i int) {
				promise.Then(func(data any) any {
					resolutionsChan <- resolutionHelper{i, data}
					return data
				}).Catch(func(err error) any {
					resolutionsChan <- resolutionHelper{i, err}
					return err
				})
			}(index)
		}

		resolutions := make([]any, psLen)
		for x := 0; x < psLen; x++ {
			resolution := <-resolutionsChan
			resolutions[resolution.index] = resolution.data
		}
		resolve(resolutions)
	})
}

func Resolve[T any](resolution any) *Promise[T] {
	return Async[T](func(resolve func(T), reject func(any)) {
		resolve(resolution.(T))
	})
}

func Reject[T any](err error) *Promise[T] {
	return Async[T](func(resolve func(T), reject func(any)) {
		reject(err)
	})
}
