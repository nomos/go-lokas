package protocol

var global_protocol_protobuf bool = false

func IsProtobuf()bool {
	return global_protocol_protobuf
}

func SetProtobuf() {
	global_protocol_protobuf = true
}