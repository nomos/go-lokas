package oss

import (
	aliyun_oss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/minio/minio-go/v7"
	"github.com/nomos/go-lokas/protocol"
)

type Type protocol.Enum

const (
	ALIYUN_OSS = 1
	MINIO      = 2
)

type Client struct {
	typ          Type
	endPoint     string
	accessId     string
	accessSecret string
	minio        *minio.Client
	aliyun       *aliyun_oss.Client
}

func NewClient(typ Type, endPoint string, accessId string, accessSecret string) *Client {
	ret := &Client{
		typ:          typ,
		endPoint:     endPoint,
		accessId:     accessId,
		accessSecret: accessSecret,
	}
	return ret
}

func (this *Client) Connect() error {

	return nil
}
