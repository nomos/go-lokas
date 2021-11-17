//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas/protocol"
)

type KEY_EVENT_TYPE protocol.Enum 

const (
	KEY_EVENT_TYPE_NULL KEY_EVENT_TYPE  = 0 //0
	KEY_EVENT_TYPE_DOWN KEY_EVENT_TYPE  = 1 //1
	KEY_EVENT_TYPE_UP KEY_EVENT_TYPE  = 2 //2
	KEY_EVENT_TYPE_PRESS KEY_EVENT_TYPE  = 3 //3
)

var ALL_KEY_EVENT_TYPE protocol.IEnumCollection = []protocol.IEnum{KEY_EVENT_TYPE_NULL,KEY_EVENT_TYPE_DOWN,KEY_EVENT_TYPE_UP,KEY_EVENT_TYPE_PRESS}

func TO_KEY_EVENT_TYPE(s string)KEY_EVENT_TYPE{
	switch s {
	case "0":
		return KEY_EVENT_TYPE_NULL
	case "1":
		return KEY_EVENT_TYPE_DOWN
	case "2":
		return KEY_EVENT_TYPE_UP
	case "3":
		return KEY_EVENT_TYPE_PRESS
	}
	return -1
}

func (this KEY_EVENT_TYPE) Enum()protocol.Enum{
	return protocol.Enum(this)
}


func (this KEY_EVENT_TYPE) ToString()string{
	switch this {
	case KEY_EVENT_TYPE_NULL:
		return "0"
	case KEY_EVENT_TYPE_DOWN:
		return "1"
	case KEY_EVENT_TYPE_UP:
		return "2"
	case KEY_EVENT_TYPE_PRESS:
		return "3"
	}
	return ""
}
