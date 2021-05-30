package protocol

import (
	"bytes"
	"github.com/nomos/go-log/log"
	"github.com/nomos/go-lokas/util/xmath"
	"go.uber.org/zap"
	"reflect"
	"testing"
	"time"
)

type TestStruct struct {
	Name     string
	Weight   int32
	Height   float32
	Time     time.Time
	Sex      bool
	SexArray []bool
	Map1     map[string]bool
	Friends  []*TestStruct
}

type ByteTest struct {
	Bytes *bytes.Buffer
}

func TestBenchMark(t *testing.T) {
	//for i:=0;i<100000;i++ {
	//	marshalToMessage(false)
	//}
	//marshalToMessage(true)
}

func TestEncode(t *testing.T) {
	GetTypeRegistry().RegistryType(44, reflect.TypeOf((*TestStruct)(nil)).Elem())
	GetTypeRegistry().RegistryType(45, reflect.TypeOf((*ByteTest)(nil)).Elem())
	a := 0
	for i := 0; i < 1000; i++ {
		if xmath.XInY(3, 5) {
			a++
		}
	}
	log.Warnf(float64(a) / float64(1000))
	marshalTest(true)
}

func marshalTest(log1 bool) {
	test1 := &TestStruct{
		Name:     "小明aaaaaaaadsadasd",
		Weight:   33344,
		Height:   123.3123,
		Time:     time.Now(),
		Sex:      true,
		Map1:     map[string]bool{"aaa": true, "bbb": false},
		SexArray: []bool{true, true, false, false, true},
		Friends: []*TestStruct{&TestStruct{
			Name:     "小白",
			Weight:   1134551,
			Height:   555.3123,
			Sex:      false,
			SexArray: []bool{true, true, false, false, false, false, false},
			Friends:  nil,
		}},
	}
	data, err := MarshalBinaryMessage(0, test1)
	if err != nil {
		log.Error(err.Error())
		return
	}
	var test2 TestStruct
	err = UnmarshalFromBytes(data, &test2)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if log1 {
		log.Warn("", zap.Any("test2", test2))
	}
	test3inc := &ByteTest{Bytes: bytes.NewBuffer([]byte{8,12,123,53})}
	data, err = MarshalBinaryMessage(0, test3inc)
	if err != nil {
		log.Error(err.Error())
		return
	}

	var test3dec ByteTest
	err = UnmarshalFromBytes(data, &test3dec)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Warnf("bytes3 %v",test3dec.Bytes.Bytes())
}
