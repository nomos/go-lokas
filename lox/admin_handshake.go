package lox

import (
	"encoding/json"
	"github.com/nomos/go-lokas/log"
)

type AdminHandShake struct {
	Account string
	Password string
}

func NewAdminHandShake(acc string,pass string)*AdminHandShake {
	ret:=&AdminHandShake{
		Account:  acc,
		Password: pass,
	}
	return ret
}


func (this *AdminHandShake) Marshal()([]byte,error) {
	return json.Marshal(this)
}

func (this *AdminHandShake) Unmarshal(from []byte)error {
	err:=json.Unmarshal(from,this)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return err
}