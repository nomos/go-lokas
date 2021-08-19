package convert

import (
	"github.com/nomos/go-lokas/log"
	"github.com/nomos/go-lokas/protocol"
)

type TYPE protocol.Enum

const (
	STRING TYPE = iota
	UNICODE
	RUNE
	BINARY
	DECIMAL
	OCTAL
	HEX
)

var _ protocol.IEnum = (*TYPE)(nil)

var ALL_ENC_TYPES protocol.IEnumCollection = []protocol.IEnum{STRING, UNICODE, RUNE, BINARY,DECIMAL,OCTAL,HEX}

func (this TYPE) ToString() string {
	switch this {
	case STRING:
		return "string"
	case UNICODE:
		return "unicode"
	case RUNE:
		return "rune"
	case BINARY:
		return "0x2"
	case DECIMAL:
		return "0x10"
	case OCTAL:
		return "0x8"
	case HEX:
		return "0x16"
	default:
		log.Panic("type not supported")
	}
	return ""
}

func (this TYPE) Enum() protocol.Enum {
	return protocol.Enum(this)
}
