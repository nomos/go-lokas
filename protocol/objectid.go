package protocol

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
)

var staticMachine = getMachineHash()
var staticIncrement = getRandomNumber()
var staticPid = int32(os.Getpid())

// A globally unique identifier for Objects.
type ObjectId struct {
	timestamp int64
	machine   int32
	pid       int32
	increment int32
}

// News generates new ObjectID with a unique value.
func NewObjectId() ObjectId {
	timestamp := time.Now().Unix()
	return ObjectId{timestamp, staticMachine, staticPid, atomic.AddInt32(&staticIncrement, 1) & 0xffffff}
}

// Parses a string and creates a new ObjectId.
func Parse(input string) (o ObjectId, err error) {
	if o, err = tryParse(input); err == nil {
		return
	}
	return o, fmt.Errorf("%s is not a valid 24 digit hex string", input)
}

func (o ObjectId) Timestamp() int64 {
	return o.timestamp
}

func (o ObjectId) Machine() int32 {
	return o.machine
}

func (o ObjectId) Pid() int32 {
	return o.pid
}

func (o ObjectId) Increment() int32 {
	return o.increment & 0xffffff
}

// String returns the ObjectID id as a 24 byte hex string representation.
func (o ObjectId) String() string {
	array := []byte{
		byte(o.timestamp >> 0x18),
		byte(o.timestamp >> 0x10),
		byte(o.timestamp >> 8),
		byte(o.timestamp),
		byte(o.machine >> 0x10),
		byte(o.machine >> 8),
		byte(o.machine),
		byte(o.pid >> 8),
		byte(o.pid),
		byte(o.increment >> 0x10),
		byte(o.increment >> 8),
		byte(o.increment),
	}
	return hex.EncodeToString(array)
}

func getMachineHash() int32 {
	machineName, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	buf := md5.Sum([]byte(machineName))
	return (int32(buf[0])<<0x10 + int32(buf[1])<<8) + int32(buf[2])
}

func getRandomNumber() int32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int31()
}

func tryParse(input string) (o ObjectId, err error) {
	if len(input) != 0x18 {
		return o, errors.New("invalid input length")
	}
	array, err := hex.DecodeString(input)
	if err != nil {
		return o, err
	}
	return ObjectId{
		timestamp: int64(array[0])<<0x18 + int64(array[1])<<0x10 + int64(array[2])<<8 + int64(array[3]),
		machine:   int32(array[4])<<0x10 + int32(array[5])<<8 + int32(array[6]),
		pid:       int32(array[7])<<8 + int32(array[8]),
		increment: int32(array[9])<<0x10 + (int32(array[10]) << 8) + int32(array[11]),
	}, nil
}
