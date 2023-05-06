//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas/protocol"
)

type KEY_EVENT_TYPE protocol.Enum 

const (
	KEY_EVENT_TYPE_NULL KEY_EVENT_TYPE  = 0 //NULL
	KEY_EVENT_TYPE_DOWN KEY_EVENT_TYPE  = 1 //DOWN
	KEY_EVENT_TYPE_UP KEY_EVENT_TYPE  = 2 //UP
	KEY_EVENT_TYPE_PRESS KEY_EVENT_TYPE  = 3 //PRESS
)

var ALL_KEY_EVENT_TYPE protocol.IEnumCollection = []protocol.IEnum{KEY_EVENT_TYPE_NULL,KEY_EVENT_TYPE_DOWN,KEY_EVENT_TYPE_UP,KEY_EVENT_TYPE_PRESS}


var ENUM_KEY_EVENT_TYPE = protocol.NewEnumCollection([]KEY_EVENT_TYPE{KEY_EVENT_TYPE_NULL,KEY_EVENT_TYPE_DOWN,KEY_EVENT_TYPE_UP,KEY_EVENT_TYPE_PRESS})

func TO_KEY_EVENT_TYPE(s string)KEY_EVENT_TYPE{
	switch s {
	case "NULL":
		return KEY_EVENT_TYPE_NULL
	case "DOWN":
		return KEY_EVENT_TYPE_DOWN
	case "UP":
		return KEY_EVENT_TYPE_UP
	case "PRESS":
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
		return "NULL"
	case KEY_EVENT_TYPE_DOWN:
		return "DOWN"
	case KEY_EVENT_TYPE_UP:
		return "UP"
	case KEY_EVENT_TYPE_PRESS:
		return "PRESS"
	}
	return ""
}
