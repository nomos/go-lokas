package ossclient

import (
	"bytes"
	"context"
	aliyun_oss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
	"io/ioutil"
)

type Type protocol.Enum

const (
	ALIYUN_OSS Type = 1
	MINIO      Type = 2
)

type Client struct {
	typ          Type
	endPoint     string
	accessId     string
	accessSecret string
	minio        *minio.Client
	aliyun       *aliyun_oss.Client
}

func NewClient(typ Type, endPoint string, accessId string, accessSecret string) (*Client, error) {
	ret := &Client{
		typ:          typ,
		endPoint:     endPoint,
		accessId:     accessId,
		accessSecret: accessSecret,
	}
	var err error
	if ret.typ == MINIO {
		ret.minio, err = minio.New(endPoint, &minio.Options{
			Creds: credentials.NewStaticV4(accessId, accessSecret, ""),
		})
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	} else if ret.typ == ALIYUN_OSS {
		ret.aliyun, err = aliyun_oss.New(endPoint, accessId, accessSecret)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}
	return ret, nil
}

func (this *Client) CreateBucket(name string) error {
	if this.typ == ALIYUN_OSS {
		return this.createBucketAliyun(name)
	} else if this.typ == MINIO {
		return this.createBucketMinio(name)
	}
	return nil
}

func (this *Client) RemoveBucket(name string) error {
	if this.typ == ALIYUN_OSS {
		return this.removeBucketAliyun(name)
	} else if this.typ == MINIO {
		return this.removeBucketMinio(name)
	}
	return nil
}

func (this *Client) PutObject(bucketName, name string, data []byte) error {
	if this.typ == ALIYUN_OSS {
		return this.putObjectAliyun(bucketName, name, data)
	} else if this.typ == MINIO {
		return this.putObjectMinio(bucketName, name, data)
	}
	return nil
}

func (this *Client) PutObjectFromFile(bucketName, name string, localFile string) error {
	if this.typ == ALIYUN_OSS {
		return this.putFileAliyun(bucketName, name, localFile)
	} else if this.typ == MINIO {
		return this.putFileMinio(bucketName, name, localFile)
	}
	return nil
}

func (this *Client) GetObject(bucketName, name string) ([]byte, error) {
	if this.typ == ALIYUN_OSS {
		return this.getObjectToBytesAliyun(bucketName, name)
	} else if this.typ == MINIO {
		return this.getObjectToBytesMinio(bucketName, name)
	}
	return nil, nil
}

func (this *Client) GetObjectToFile(bucketName, name string, localPath string) error {
	if this.typ == ALIYUN_OSS {
		return this.getObjectToFileAliyun(bucketName, name, localPath)
	} else if this.typ == MINIO {
		return this.getObjectToFileMinio(bucketName, name, localPath)
	}
	return nil
}

func (this *Client) DeleteObject(bucketName, name string) error {
	if this.typ == ALIYUN_OSS {
		return this.deleteObjectAliyun(bucketName, name)
	} else if this.typ == MINIO {
		return this.deleteObjectMinio(bucketName, name)
	}
	return nil
}

func (this *Client) createBucketAliyun(name string) error {
	return this.aliyun.CreateBucket(name)
}

func (this *Client) createBucketMinio(name string) error {
	return this.minio.MakeBucket(context.TODO(), name, minio.MakeBucketOptions{})
}

func (this *Client) bucketExistMinio(name string) (bool, error) {
	return this.minio.BucketExists(context.TODO(), name)
}

func (this *Client) bucketExistAliyun(name string) (bool, error) {
	bucket, err := this.aliyun.Bucket(name)
	if bucket != nil {
		return true, nil
	}
	return false, err
}

func (this *Client) removeBucketAliyun(name string) error {
	return this.aliyun.DeleteBucket(name)
}

func (this *Client) removeBucketMinio(name string) error {
	return this.minio.RemoveBucket(context.TODO(), name)
}

func (this *Client) putObjectAliyun(bucketName, name string, data []byte) error {
	bucket, err := this.aliyun.Bucket(bucketName)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	reader := bytes.NewReader(data)
	err = bucket.PutObject(name, reader)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) putFileAliyun(bucketName, name string, filePath string) error {
	bucket, err := this.aliyun.Bucket(bucketName)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = bucket.PutObjectFromFile(name, filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) getObjectToBytesAliyun(bucketName string, name string) ([]byte, error) {
	bucket, err := this.aliyun.Bucket(bucketName)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	rc, err := bucket.GetObject(name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(rc)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *Client) getObjectToFileAliyun(bucketName string, name, filePath string) error {
	bucket, err := this.aliyun.Bucket(bucketName)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = bucket.GetObjectToFile(name, filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) deleteObjectAliyun(bucketName string, name string) error {
	bucket, err := this.aliyun.Bucket(bucketName)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = bucket.DeleteObject(name)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) deletObjectsAliyun(bucketName string, names ...string) error {
	bucket, err := this.aliyun.Bucket(bucketName)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	_, err = bucket.DeleteObjects(names)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) putObjectMinio(bucketName, name string, data []byte) error {
	reader := bytes.NewReader(data)
	size := len(data)
	_, err := this.minio.PutObject(context.TODO(), bucketName, name, reader, int64(size), minio.PutObjectOptions{})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) putFileMinio(bucketName, name string, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	reader := bytes.NewReader(data)
	size := len(data)
	_, err = this.minio.PutObject(context.TODO(), bucketName, name, reader, int64(size), minio.PutObjectOptions{})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) getObjectToBytesMinio(bucketName string, name string) ([]byte, error) {
	object, err := this.minio.GetObject(context.TODO(), bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(object)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return buf.Bytes(), nil
}

func (this *Client) getObjectToFileMinio(bucketName string, name, filePath string) error {
	object, err := this.minio.GetObject(context.TODO(), bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(object)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	err = ioutil.WriteFile(filePath, buf.Bytes(), 0644)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func (this *Client) deleteObjectMinio(bucketName string, name string) error {
	err := this.minio.RemoveObject(context.TODO(), bucketName, name, minio.RemoveObjectOptions{})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
