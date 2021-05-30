package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

func LoadRSAPrivateKeyPem(p string)(*rsa.PrivateKey,error){
	keyStr,err:=ioutil.ReadFile(p)
	if err != nil {
		return nil,err
	}
	block, _:=pem.Decode(keyStr)
	if block == nil {
		return nil,errors.New("decode pem error:"+p)
	}
	key,err:=x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil,err
	}
	return key,nil
}

func LoadRSAPublicKeyPem(p string)(*rsa.PublicKey,error){
	keyStr,err:=ioutil.ReadFile(p)
	if err != nil {
		return nil,err
	}
	block, _:=pem.Decode(keyStr)
	if block == nil {
		return nil,errors.New("decode pem error:"+p)
	}
	key,err:=x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil,err
	}
	return key,nil
}