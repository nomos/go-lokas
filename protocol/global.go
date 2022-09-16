package protocol

var global_protocol_protobuf bool = false

func IsProtobuf() bool {
	return global_protocol_protobuf
}

func SetProtobuf() {
	global_protocol_protobuf = true
}

const (
	REQ_TYPE_REPLAY  = 0
	REQ_TYPE_MAIN    = 1
	REQ_TYPE_OUTSIDE = 2
)
