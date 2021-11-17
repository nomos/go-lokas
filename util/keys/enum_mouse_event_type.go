//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas/protocol"
)

type MOUSE_EVENT_TYPE protocol.Enum 

const (
	MOUSE_EVENT_TYPE_NULL MOUSE_EVENT_TYPE  = 0 //0
	MOUSE_EVENT_TYPE_DOWN MOUSE_EVENT_TYPE  = 1 //1
	MOUSE_EVENT_TYPE_UP MOUSE_EVENT_TYPE  = 2 //2
	MOUSE_EVENT_TYPE_PRESS MOUSE_EVENT_TYPE  = 3 //3
	MOUSE_EVENT_TYPE_SCROLL MOUSE_EVENT_TYPE  = 4 //4
	MOUSE_EVENT_TYPE_MOVE MOUSE_EVENT_TYPE  = 5 //5
	MOUSE_EVENT_TYPE_ENTER MOUSE_EVENT_TYPE  = 6 //6
	MOUSE_EVENT_TYPE_LEAVE MOUSE_EVENT_TYPE  = 7 //7
	MOUSE_EVENT_TYPE_CANCEL MOUSE_EVENT_TYPE  = 8 //8
)

var ALL_MOUSE_EVENT_TYPE protocol.IEnumCollection = []protocol.IEnum{MOUSE_EVENT_TYPE_NULL,MOUSE_EVENT_TYPE_DOWN,MOUSE_EVENT_TYPE_UP,MOUSE_EVENT_TYPE_PRESS,MOUSE_EVENT_TYPE_SCROLL,MOUSE_EVENT_TYPE_MOVE,MOUSE_EVENT_TYPE_ENTER,MOUSE_EVENT_TYPE_LEAVE,MOUSE_EVENT_TYPE_CANCEL}

func TO_MOUSE_EVENT_TYPE(s string)MOUSE_EVENT_TYPE{
	switch s {
	case "0":
		return MOUSE_EVENT_TYPE_NULL
	case "1":
		return MOUSE_EVENT_TYPE_DOWN
	case "2":
		return MOUSE_EVENT_TYPE_UP
	case "3":
		return MOUSE_EVENT_TYPE_PRESS
	case "4":
		return MOUSE_EVENT_TYPE_SCROLL
	case "5":
		return MOUSE_EVENT_TYPE_MOVE
	case "6":
		return MOUSE_EVENT_TYPE_ENTER
	case "7":
		return MOUSE_EVENT_TYPE_LEAVE
	case "8":
		return MOUSE_EVENT_TYPE_CANCEL
	}
	return -1
}

func (this MOUSE_EVENT_TYPE) Enum()protocol.Enum{
	return protocol.Enum(this)
}


func (this MOUSE_EVENT_TYPE) ToString()string{
	switch this {
	case MOUSE_EVENT_TYPE_NULL:
		return "0"
	case MOUSE_EVENT_TYPE_DOWN:
		return "1"
	case MOUSE_EVENT_TYPE_UP:
		return "2"
	case MOUSE_EVENT_TYPE_PRESS:
		return "3"
	case MOUSE_EVENT_TYPE_SCROLL:
		return "4"
	case MOUSE_EVENT_TYPE_MOVE:
		return "5"
	case MOUSE_EVENT_TYPE_ENTER:
		return "6"
	case MOUSE_EVENT_TYPE_LEAVE:
		return "7"
	case MOUSE_EVENT_TYPE_CANCEL:
		return "8"
	}
	return ""
}
