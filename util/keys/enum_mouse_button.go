//this is a generate file,do not modify it!
package keys

import (
	"github.com/nomos/go-lokas/protocol"
)

type MOUSE_BUTTON protocol.Enum 

const (
	MOUSE_BUTTON_NULL MOUSE_BUTTON  = 0 //NULL
	MOUSE_BUTTON_LEFT MOUSE_BUTTON  = 1 //LEFT
	MOUSE_BUTTON_RIGHT MOUSE_BUTTON  = 2 //RIGHT
	MOUSE_BUTTON_MIDDLE MOUSE_BUTTON  = 3 //MIDDLE
	MOUSE_BUTTON_EXTRA_1 MOUSE_BUTTON  = 4 //EXTRA1
	MOUSE_BUTTON_EXTRA_2 MOUSE_BUTTON  = 5 //5
)

var ALL_MOUSE_BUTTON protocol.IEnumCollection = []protocol.IEnum{MOUSE_BUTTON_NULL,MOUSE_BUTTON_LEFT,MOUSE_BUTTON_RIGHT,MOUSE_BUTTON_MIDDLE,MOUSE_BUTTON_EXTRA_1,MOUSE_BUTTON_EXTRA_2}

func TO_MOUSE_BUTTON(s string)MOUSE_BUTTON{
	switch s {
	case "NULL":
		return MOUSE_BUTTON_NULL
	case "LEFT":
		return MOUSE_BUTTON_LEFT
	case "RIGHT":
		return MOUSE_BUTTON_RIGHT
	case "MIDDLE":
		return MOUSE_BUTTON_MIDDLE
	case "EXTRA1":
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
		return "NULL"
	case MOUSE_BUTTON_LEFT:
		return "LEFT"
	case MOUSE_BUTTON_RIGHT:
		return "RIGHT"
	case MOUSE_BUTTON_MIDDLE:
		return "MIDDLE"
	case MOUSE_BUTTON_EXTRA_1:
		return "EXTRA1"
	case MOUSE_BUTTON_EXTRA_2:
		return "5"
	}
	return ""
}
