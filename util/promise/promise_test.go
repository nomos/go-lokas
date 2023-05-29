package promise

import (
	"github.com/nomos/go-lokas/log"
	"testing"
	"time"
)

type ClassA struct {
	Name string
}

type ClassB struct {
	Name string
}

func TestPromise(t *testing.T) {
	p1 := Async(func(resolve func(a *ClassA), reject func(any)) {
		timeout2 := SetTimeout(time.Millisecond*2000, func(timeout *Timeout) {
			reject(log.Error("timeout"))
		})
		SetTimeout(time.Millisecond*1000, func(timeout *Timeout) {
			a := &ClassA{
				Name: "AAA",
			}
			timeout2.Close()
			resolve(a)
		})
	})
	p2 := Async(func(resolve func(*ClassB), reject func(any)) {
		timeout2 := SetTimeout(time.Millisecond*500, func(timeout *Timeout) {
			reject(log.Error("timeout"))
		})
		SetTimeout(time.Millisecond*1000, func(timeout *Timeout) {
			a := &ClassB{
				Name: "BBBB",
			}
			timeout2.Close()
			resolve(a)
		})
	})

	d, err := Each(p1.Any(), p2.Any()).Await()
	if err != nil {
		log.Error(err.Error())
	}
	log.Infof("Each", d)
	//log.Infof("Each", d.([]interface{})[0].(*ClassA), d.([]interface{})[1].(*ClassB))
	d, err = All(p1.Any(), p2.Any()).Await()
	if err != nil {
		log.Error(err.Error())
	}
	log.Infof("All", d)
	d, err = Race(p1.Any(), p2.Any()).Await()
	if err != nil {
		log.Error(err.Error())
	}
	log.Infof("Race", d)
	d, err = AllSettled(p1.Any(), p2.Any()).Await()
	if err != nil {
		log.Error(err.Error())
	}
	log.Infof("AllSettled", d)
	p1.Then(func(data any) any {
		log.Infof("Then", data)
		return data
	}).Catch(func(err error) any {
		log.Error(err.Error())
		return nil
	}).Await()
}
