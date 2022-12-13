package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/nomos/go-lokas/protocol/encoding/number_json"

	"reflect"
	"time"

	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/util"
	"go.uber.org/zap"
)

type TYPE int

const (
	BINARY TYPE = 0
	JSON   TYPE = 1
)

const (
	AUTH_STATE_CONNECTED uint8 = 0
	AUTH_STATE_AUTHING   uint8 = 1
	AUTH_STATE_AUTHED    uint8 = 2
)

const (
	ROUTE_MSG_HEAD_SIZE  int = 25
	BINARY_MSG_HEAD_SIZE int = 8
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
	MAX_PACKET_SIZE    = 20 * 1024
	DEFAULT_PACKET_LEN = 2048 * 4
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
	FromPid   util.ProcessId // TODO  add route
	ToActor   util.ID
	ToPid     util.ProcessId
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
	if ret.Req {
		ret.ReqType = REQ_TYPE_MAIN
	}
	return ret
}

// new ctor
func NewRouteMsg(fromActor util.ID, toActor util.ID, transId uint32, msg ISerializable, reqType uint8) *RouteMessage {
	cmd, _ := msg.GetId()
	ret := &RouteMessage{
		TransId:   transId,
		Len:       0,
		ReqType:   reqType,
		CmdId:     TAG_RouteMessage,
		InnerId:   cmd,
		FromActor: fromActor,
		ToActor:   toActor,
		Body:      msg,
	}
	if ret.ReqType == REQ_TYPE_REPLAY {
		ret.Req = false
	} else {
		ret.Req = true
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

type RouteHead struct {
	Len       uint16
	Cmd       uint16
	TransId   uint32
	ToActor   uint64
	FromActor uint64
}

type RouteRecv struct {
	Protocol TYPE
	// FromActor util.ID
	FromPid util.ProcessId
	Head    *RouteHead
	Data    []byte
}

func NewRouteRecv(data []byte, protocolType TYPE, fromPid util.ProcessId) *RouteRecv {
	return &RouteRecv{
		Protocol: protocolType,
		FromPid:  fromPid,
		Data:     data,
	}
}

func (recv *RouteRecv) UnmarshalHead() {

}

func (recv *RouteRecv) GetReq() bool {

	// reqType := uint8(recv.Data[16])

	transId := recv.GetTransId()

	if transId == 0 {
		return false
	} else {
		return true
	}

}

func (recv *RouteRecv) GetCmd() uint16 {
	cmd := binary.LittleEndian.Uint16(recv.Data[2:4])

	return cmd
}

func (recv *RouteRecv) GetTransId() uint32 {
	transId := binary.LittleEndian.Uint32(recv.Data[4:8])
	return transId
}

type RouteDataMsg struct {
	Protocol  TYPE
	Len       uint16
	Cmd       BINARY_TAG
	TransId   uint32
	ReqType   uint8
	FromActor util.ID
	FromPid   util.ProcessId
	ToActor   util.ID
	ToPid     util.ProcessId
	BodyData  []byte
}

func NewRouteDataMsg(fromActorId util.ID, toActorId util.ID, transId uint32, cmd BINARY_TAG, reqType uint8, body []byte, protocolType TYPE) *RouteDataMsg {
	return &RouteDataMsg{
		Protocol:  protocolType,
		TransId:   transId,
		Cmd:       cmd,
		ReqType:   reqType,
		FromActor: fromActorId,
		ToActor:   toActorId,
		BodyData:  body,
	}
}

func UnmarshalRouteDataMsg(data []byte, protocolType TYPE, fromPid util.ProcessId) (*RouteDataMsg, error) {

	if len(data) < ROUTE_MSG_HEAD_SIZE {
		return nil, ERR_MSG_FORMAT
	}

	routeMsg := &RouteDataMsg{
		Protocol: protocolType,
		FromPid:  fromPid,

		BodyData: data[ROUTE_MSG_HEAD_SIZE:],
	}

	routeMsg.Len = binary.LittleEndian.Uint16(data[0:2])
	routeMsg.Cmd = BINARY_TAG(binary.LittleEndian.Uint16(data[2:4]))
	routeMsg.TransId = binary.LittleEndian.Uint32(data[4:8])
	routeMsg.ToActor = util.ID(binary.LittleEndian.Uint64(data[8:16]))
	routeMsg.FromActor = util.ID(binary.LittleEndian.Uint64(data[16:24]))
	routeMsg.ReqType = uint8(data[24])

	return routeMsg, nil
}

func (msg *RouteDataMsg) MarshalData() ([]byte, error) {
	var buff bytes.Buffer
	binary.Write(&buff, binary.LittleEndian, uint16(0))
	binary.Write(&buff, binary.LittleEndian, msg.Cmd)
	binary.Write(&buff, binary.LittleEndian, msg.TransId)
	binary.Write(&buff, binary.LittleEndian, uint64(msg.ToActor))
	binary.Write(&buff, binary.LittleEndian, uint64(msg.FromActor))
	binary.Write(&buff, binary.LittleEndian, uint8(msg.ReqType))
	binary.Write(&buff, binary.LittleEndian, msg.BodyData)

	if buff.Len() > 65535 {
		log.Error("error data len", zap.Int("len", buff.Len()))
		return nil, ERR_MSG_LEN_INVALID
	}
	out := buff.Bytes()
	binary.LittleEndian.PutUint16(out[0:2], uint16(buff.Len()))

	return out, nil
}

func (msg *RouteDataMsg) UnmarshalData() (ISerializable, error) {
	body, err := GetTypeRegistry().GetInterfaceByTag(msg.Cmd)
	if err != nil {
		log.Error("not find cmd", zap.Uint16("cmdId", uint16(msg.Cmd)), zap.String("err", err.Error()))
		return nil, err
	}
	dec := number_json.NewDecoder(bytes.NewBuffer(msg.BodyData))
	dec.UseNumber()
	err = dec.Decode(body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return body, nil
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

func GetTransId(data []byte) uint32 {
	transId := binary.LittleEndian.Uint32(data[2:6])
	return transId
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
	ret, _ := MarshalBinaryMessage(0, ack)
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
