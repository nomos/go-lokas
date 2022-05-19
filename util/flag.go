package util

import (
	"flag"
	"github.com/nomos/go-lokas/log"
	"go.uber.org/zap"
)

var FlagUtil = &flagUtil{
	items: map[string]*flagValue{},
}

type flagValue struct {
	k       string
	vInt    int
	vBool   bool
	vString string
	vFloat  float64
}

type flagUtil struct {
	items map[string]*flagValue
}

func (this *flagUtil) Items()map[string]*flagValue {
	return this.items
}

func (this *flagUtil) Parse() {
	flag.Parse()
}

func (this *flagUtil) IntVar(k string, v int, desc string) {
	if this.items[k] != nil {
		log.Warn("flag value exist", zap.String("flag", k))
		return
	}
	value := &flagValue{
		k:    k,
		vInt: v,
	}
	flag.IntVar(&value.vInt, k, v, desc)
	this.items[k] = value
}

func (this *flagUtil) BoolVar(k string, v bool, desc string) {
	if this.items[k] != nil {
		log.Warn("flag value exist", zap.String("flag", k))
		return
	}
	value := &flagValue{
		k:     k,
		vBool: v,
	}
	flag.BoolVar(&value.vBool, k, v, desc)
	this.items[k] = value
}

func (this *flagUtil) StringVar(k string, v string, desc string) {
	if this.items[k] != nil {
		log.Warn("flag value exist", zap.String("flag", k))
		return
	}
	value := &flagValue{
		k:       k,
		vString: v,
	}
	flag.StringVar(&value.vString, k, v, desc)
	this.items[k] = value
}

func (this *flagUtil) FloatVar(k string, v float64, desc string) {
	if this.items[k] != nil {
		log.Warn("flag value exist", zap.String("flag", k))
		return
	}
	value := &flagValue{
		k:      k,
		vFloat: v,
	}
	flag.Float64Var(&value.vFloat, k, v, desc)
	this.items[k] = value
}

func (this *flagUtil) Int(k string)int{
	value:=this.items[k]
	if value==nil {
		log.Panic("flag key not exist:"+k)
		return 0
	}
	return value.vInt
}

func (this *flagUtil) Bool(k string)bool{
	value:=this.items[k]
	if value==nil {
		log.Panic("flag key not exist:"+k)
		return false
	}
	return value.vBool
}

func (this *flagUtil) String(k string)string{
	value:=this.items[k]
	if value==nil {
		log.Panic("flag key not exist:"+k)
		return ""
	}
	return value.vString
}

func (this *flagUtil) Float(k string)float64{
	value:=this.items[k]
	if value==nil {
		log.Panic("flag key not exist:"+k)
		return 0
	}
	return value.vFloat
}

