package promise

import (
	"errors"
	"sync"
	"time"
)

type Promise struct {
	pending bool

	executor func(resolve func(interface{}), reject func(interface{}))

	result interface{}

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

func Await(p *Promise) (interface{}, error) {
	return p.Await()

}

func (this *Promise) CalTime() *Promise {
	this.calTime = true
	return this
}

func (this *Promise) Elapse() time.Duration {
	return this.elapseTime
}

func Async(executor func(resolve func(interface{}), reject func(interface{}))) *Promise {
	var promise = &Promise{
		pending:  true,
		executor: executor,
		result:   nil,
		err:      nil,
		mutex:    sync.Mutex{},
		wg:       sync.WaitGroup{},
	}

	promise.wg.Add(1)

	go func() {
		defer promise.handlePanic()
		promise.executor(promise.Resolve, promise.Reject)
	}()

	return promise
}

func (this *Promise) Resolve(resolution interface{}) {
	this.mutex.Lock()

	if !this.pending {
		this.mutex.Unlock()
		return
	}

	switch result := resolution.(type) {
	case *Promise:
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

func (this *Promise) Reject(err interface{}) {
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

func (this *Promise) handlePanic() {
	var r = recover()
	if r != nil {
		if err, ok := r.(error); ok {
			this.Reject(errors.New(err.Error()))
		} else {
			this.Reject(errors.New(r.(string)))
		}
	}
}

func (this *Promise) Then(fulfillment func(data interface{}) interface{}) *Promise {
	return Async(func(resolve func(interface{}), reject func(interface{})) {
		result, err := this.Await()
		if err != nil {
			reject(err)
			return
		}
		resolve(fulfillment(result))
	})
}

func (this *Promise) Catch(rejection func(err error) interface{}) *Promise {
	return Async(func(resolve func(interface{}), reject func(interface{})) {
		result, err := this.Await()
		if err != nil {
			reject(rejection(err))
			return
		}
		resolve(result)
	})
}

func (this *Promise) Await() (interface{}, error) {
	if this.calTime {
		start := time.Now()
		this.wg.Wait()
		this.elapseTime = time.Now().Sub(start)
		return this.result, this.err
	}
	this.wg.Wait()
	return this.result, this.err
}

func (this *Promise) AsCallback(f func(interface{}, error)) {
	go func() {
		this.wg.Wait()
		f(this.result, this.err)
	}()
}

type resolutionHelper struct {
	index int
	data  interface{}
}

func Each(promises ...*Promise) *Promise {
	return Async(func(resolve func(interface{}), reject func(interface{})) {
		resolutions := make([]interface{}, 0)
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

func All(promises ...*Promise) *Promise {
	psLen := len(promises)
	if psLen == 0 {
		return Resolve(make([]interface{}, 0))
	}

	return Async(func(resolve func(interface{}), reject func(interface{})) {
		resolutionsChan := make(chan resolutionHelper, psLen)
		errorChan := make(chan error, psLen)

		for index, promise := range promises {
			func(i int) {
				promise.Then(func(data interface{}) interface{} {
					resolutionsChan <- resolutionHelper{i, data}
					return data
				}).Catch(func(err error) interface{} {
					errorChan <- err
					return err
				})
			}(index)
		}

		resolutions := make([]interface{}, psLen)
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

func Race(promises ...*Promise) *Promise {
	psLen := len(promises)
	if psLen == 0 {
		return Resolve(nil)
	}

	return Async(func(resolve func(interface{}), reject func(interface{})) {
		resolutionsChan := make(chan interface{}, psLen)
		errorChan := make(chan error, psLen)

		for _, promise := range promises {
			promise.Then(func(data interface{}) interface{} {
				resolutionsChan <- data
				return data
			}).Catch(func(err error) interface{} {
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

func AllSettled(promises ...*Promise) *Promise {
	psLen := len(promises)
	if psLen == 0 {
		return Resolve(nil)
	}

	return Async(func(resolve func(interface{}), reject func(interface{})) {
		resolutionsChan := make(chan resolutionHelper, psLen)

		for index, promise := range promises {
			func(i int) {
				promise.Then(func(data interface{}) interface{} {
					resolutionsChan <- resolutionHelper{i, data}
					return data
				}).Catch(func(err error) interface{} {
					resolutionsChan <- resolutionHelper{i, err}
					return err
				})
			}(index)
		}

		resolutions := make([]interface{}, psLen)
		for x := 0; x < psLen; x++ {
			resolution := <-resolutionsChan
			resolutions[resolution.index] = resolution.data
		}
		resolve(resolutions)
	})
}

func Resolve(resolution interface{}) *Promise {
	return Async(func(resolve func(interface{}), reject func(interface{})) {
		resolve(resolution)
	})
}

func Reject(err error) *Promise {
	return Async(func(resolve func(interface{}), reject func(interface{})) {
		reject(err)
	})
}
