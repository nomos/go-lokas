package protocol

import (
	"encoding/binary"
	"encoding/json"
	"reflect"
	"time"

	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
)

type TYPE int

const (
	BINARY TYPE = 0
	JSON   TYPE = 1
)

func String2Type(s string) TYPE {
	switch s {
	case "binary":
		return BINARY
	case "json":
		return JSON
	default:
		panic("not a valid type")
	}
}

func (this TYPE) String() string {
	switch this {
	case BINARY:
		return "binary"
	case JSON:
		return "json"
	default:
		panic("not a valid type")
	}
}

const (
	// len 2 transId 4
	HEADER_SIZE        = 2 + 4
	MAX_PACKET_SIZE    = 4 * 1024 * 1024
	DEFAULT_PACKET_LEN = 4 * 1024 * 1024
	ROUTE_MSG_ADDITION = 2 + 8 + 8
)

type ISerializable interface {
	GetId() (BINARY_TAG, error)
	Serializable() ISerializable
}

var _ ISerializable = &ComposeData{}

type ComposeData struct {
	Idx  uint32
	Data []byte
}

func (this *ComposeData) Unmarshal(from []byte) error {
	return Unmarshal(from, this)
}

func (this *ComposeData) Marshal() ([]byte, error) {
	return MarshalBinary(this)
}

func (this *ComposeData) GetId() (BINARY_TAG, error) {
	return TAG_Compose, nil
}

func (this *ComposeData) Serializable() ISerializable {
	return this
}

var _ ISerializable = &RouteMessage{}

//RouteMessage rpc message across server
type RouteMessage struct {
	TransId   uint32
	Len       uint16
	Req       bool
	ReqType   uint8
	CmdId     BINARY_TAG
	InnerId   BINARY_TAG
	FromActor util.ID
	ToActor   util.ID
	Body      ISerializable
}

func NewRouteMessage(fromActor util.ID, toActor util.ID, transId uint32, msg ISerializable, isReq bool) *RouteMessage {
	id, _ := msg.GetId()
	ret := &RouteMessage{
		TransId:   transId,
		Len:       0,
		Req:       isReq,
		CmdId:     TAG_RouteMessage,
		InnerId:   id,
		FromActor: fromActor,
		ToActor:   toActor,
		Body:      msg,
	}
	return ret
}

func (this *RouteMessage) GetId() (BINARY_TAG, error) {
	return TAG_BinaryMessage, nil
}

func (this *RouteMessage) Serializable() ISerializable {
	return this
}

func (this *RouteMessage) Unmarshal(from []byte) error {
	header, err := unmarshalHeader(from)
	if err != nil {
		return err
	}
	header.Body, err = GetTypeRegistry().GetInterfaceByTag(header.CmdId)
	if err != nil {
		return err
	}
	err = Unmarshal(from[HEADER_SIZE:], header.Body)
	if err != nil {
		return err
	}
	return nil
}

func (this *RouteMessage) Marshal() ([]byte, error) {
	ret, err := MarshalBinaryMessage(this.TransId, this.Body)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (this *RouteMessage) BinaryMessage() *BinaryMessage {
	ret := &BinaryMessage{
		TransId: this.TransId,
		Len:     this.Len,
		CmdId:   this.InnerId,
		Body:    this.Body,
	}
	return ret
}

var _ ISerializable = &BinaryMessage{}

//BinaryMessage base protocol
type BinaryMessage struct {
	ISerializable
	TransId uint32
	Len     uint16
	CmdId   BINARY_TAG
	Body    ISerializable
}

func (this *BinaryMessage) Unmarshal(from []byte) error {
	header, err := unmarshalHeader(from)
	if err != nil {
		return err
	}
	header.Body, err = GetTypeRegistry().GetInterfaceByTag(header.CmdId)
	if err != nil {
		return err
	}
	err = Unmarshal(from[HEADER_SIZE:], header.Body)
	if err != nil {
		return err
	}
	return nil
}

func (this *BinaryMessage) Marshal() ([]byte, error) {
	return nil, nil
}

func (this *BinaryMessage) GetId() (BINARY_TAG, error) {
	return TAG_BinaryMessage, nil
}

var _ ISerializable = &Ping{}

type Ping struct {
	Time time.Time
}

func (this *Ping) Unmarshal(from []byte) error {
	return Unmarshal(from, this)
}

func (this *Ping) Marshal() ([]byte, error) {
	return MarshalBinary(this)
}

func (this *Ping) GetId() (BINARY_TAG, error) {
	return TAG_Ping, nil
}

func (this *Ping) Serializable() ISerializable {
	return this
}

var _ ISerializable = &Pong{}

type Pong struct {
	Time time.Time
}

func (this *Pong) Unmarshal(from []byte) error {
	return Unmarshal(from, this)
}

func (this *Pong) Marshal() ([]byte, error) {
	return MarshalBinary(this)
}

func (this *Pong) GetId() (BINARY_TAG, error) {
	return TAG_Pong, nil
}

func (this *Pong) Serializable() ISerializable {
	return this
}

var _ ISerializable = &HandShake{}

type HandShake struct {
	Data []byte
}

func (this *HandShake) Unmarshal(from []byte) error {
	this.Data = from
	return nil
}

func (this *HandShake) Marshal() ([]byte, error) {
	return this.Data, nil
}

func (this *HandShake) GetId() (BINARY_TAG, error) {
	return TAG_HandShake, nil
}

func (this *HandShake) Serializable() ISerializable {
	return this
}

func GetCmdId16(data []byte) BINARY_TAG {
	cmdId := BINARY_TAG(binary.LittleEndian.Uint16(data[6:8]))
	return cmdId
}

func GetCmdId(data []byte) BINARY_TAG {
	temp := data[6:7][0]
	var cmdId BINARY_TAG
	if temp < 128 {
		cmdId = BINARY_TAG(temp)
	} else {
		cmdId = BINARY_TAG(binary.LittleEndian.Uint16(data[6:8]))
	}
	return cmdId
}

func GetCmdIdFromType(data interface{}) (BINARY_TAG, error) {
	return GetTypeRegistry().GetTagByType(reflect.TypeOf(data))
}

func PickLongPacket(protocol TYPE) func(data []byte) (bool, int, []byte) {
	if protocol == JSON {
		return PickJsonLongPacket
	} else if protocol == BINARY {
		return PickBinaryLongPacket
	} else {
		panic("unidentified protocol")
	}
}

func PickBinaryLongPacket(data []byte) (bool, int, []byte) {
	cmdId := GetCmdId16(data)
	if cmdId != TAG_Compose {
		return false, 0, nil
	}
	body, err := unmarshalBodyByTag(cmdId, data)
	if err != nil {
		return false, 0, nil
	}
	ack, ok := body.(*ComposeData)
	if !ok {
		return false, 0, nil
	}
	log.Warnf("PickLongPacket", ack.Idx)
	return true, int(ack.Idx), ack.Data
}

func PickJsonLongPacket(data []byte) (bool, int, []byte) {
	cmdId := GetCmdId16(data)
	if cmdId != TAG_Compose {
		return false, 0, nil
	}
	bodyData := data[HEADER_SIZE+2:]
	var ack = &ComposeData{}
	err := json.Unmarshal(bodyData, ack)
	if err != nil {
		log.Error(err.Error())
		return false, 0, nil
	}
	//log.Infof("PickLongPacket", ack.Idx)
	return true, int(ack.Idx), ack.Data
}

func CreateLongPacket(protocol TYPE) func(data []byte, idx int) ([]byte, error) {
	if protocol == JSON {
		return CreateJsonLongPacket
	} else if protocol == BINARY {
		return CreateBinaryLongPacket
	} else {
		panic("unidentified protocol")
	}
}

func CreateBinaryLongPacket(data []byte, idx int) ([]byte, error) {
	ack := &ComposeData{Idx: uint32(idx), Data: data}
	return MarshalBinaryMessage(0, ack)
}

func CreateJsonLongPacket(data []byte, idx int) ([]byte, error) {
	ack := &ComposeData{Idx: uint32(idx), Data: data}
	ret, _ := MarshalJsonMessage(0, ack)
	return ret, nil
}

func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	dataLen := len(data)
	if dataLen < HEADER_SIZE {
		return 0, nil, nil
	}
	n := int(binary.LittleEndian.Uint16(data[0:2]))
	if n == 0 || n > MAX_PACKET_SIZE {
		return 0, nil, ERR_PACKAGE_FORMAT
	}
	if dataLen < n {
		return 0, nil, nil
	}
	return n, data[0:n], nil
}
