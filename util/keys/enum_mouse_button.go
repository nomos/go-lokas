//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas/protocol"
)

type MOUSE_BUTTON protocol.Enum 

const (
	MOUSE_BUTTON_NULL MOUSE_BUTTON  = 0 //0
	MOUSE_BUTTON_LEFT MOUSE_BUTTON  = 1 //1
	MOUSE_BUTTON_RIGHT MOUSE_BUTTON  = 2 //2
	MOUSE_BUTTON_MIDDLE MOUSE_BUTTON  = 3 //3
	MOUSE_BUTTON_EXTRA_1 MOUSE_BUTTON  = 4 //4
	MOUSE_BUTTON_EXTRA_2 MOUSE_BUTTON  = 5 //5
)

var ALL_MOUSE_BUTTON protocol.IEnumCollection = []protocol.IEnum{MOUSE_BUTTON_NULL,MOUSE_BUTTON_LEFT,MOUSE_BUTTON_RIGHT,MOUSE_BUTTON_MIDDLE,MOUSE_BUTTON_EXTRA_1,MOUSE_BUTTON_EXTRA_2}

func TO_MOUSE_BUTTON(s string)MOUSE_BUTTON{
	switch s {
	case "0":
		return MOUSE_BUTTON_NULL
	case "1":
		return MOUSE_BUTTON_LEFT
	case "2":
		return MOUSE_BUTTON_RIGHT
	case "3":
		return MOUSE_BUTTON_MIDDLE
	case "4":
		return MOUSE_BUTTON_EXTRA_1
	case "5":
		return MOUSE_BUTTON_EXTRA_2
	}
	return -1
}

func (this MOUSE_BUTTON) Enum()protocol.Enum{
	return protocol.Enum(this)
}


func (this MOUSE_BUTTON) ToString()string{
	switch this {
	case MOUSE_BUTTON_NULL:
		return "0"
	case MOUSE_BUTTON_LEFT:
		return "1"
	case MOUSE_BUTTON_RIGHT:
		return "2"
	case MOUSE_BUTTON_MIDDLE:
		return "3"
	case MOUSE_BUTTON_EXTRA_1:
		return "4"
	case MOUSE_BUTTON_EXTRA_2:
		return "5"
	}
	return ""
}
