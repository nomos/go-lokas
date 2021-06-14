package lokas

import "github.com/nomos/go-lokas/protocol"

type KEY protocol.Enum

const (
	KEY_BACK       KEY = 51  //   8
	KEY_TAB        KEY = 48  //   9
	KEY_RETURN     KEY = 36 //  13
	KEY_MENU       KEY = 110 //  18
	KEY_PAUSE      KEY = 113 //  19
	KEY_CAPITAL    KEY = 57 //  20
	KEY_ESCAPE     KEY = 53  //   27
	KEY_SPACE      KEY = 49 //   32
	KEY_PGUP      KEY = 116 //   33
	KEY_PGDN       KEY = 121 //   34
	KEY_END        KEY = 119 //   35
	KEY_HOME       KEY = 115 //   36
	KEY_LEFT       KEY = 123 //   37
	KEY_UP         KEY = 126 //   38
	KEY_RIGHT      KEY = 124 //   39
	KEY_DOWN       KEY = 125 //   40
	KEY_SELECT     KEY = 10 //   41
	KEY_PRINT      KEY = 105 //   42
	KEY_INSERT     KEY = 114 //   45
	KEY_DELETE     KEY = 117 //   46
	KEY_HELP       KEY = -1 //   47
	// VK0 THRU VK9 ARE THE SAME AS ASCII '0' THRU '9' (0X30 - 0X39)
	KEY_0 KEY = 29 //   48
	KEY_1 KEY = 18 //   49
	KEY_2 KEY = 19 //   50
	KEY_3 KEY = 20 //   51
	KEY_4 KEY = 21 //   52
	KEY_5 KEY = 23 //   53
	KEY_6 KEY = 22 //   54
	KEY_7 KEY = 26 //   55
	KEY_8 KEY = 28 //   56
	KEY_9 KEY = 25 //   57
	// VKA THRU VKZ ARE THE SAME AS ASCII 'A' THRU 'Z' (0X41 - 0X5A)
	KEY_A         KEY = 0  //   65
	KEY_B         KEY = 11  //   66
	KEY_C         KEY = 8  //   67
	KEY_D         KEY = 2  //   68
	KEY_E         KEY = 14  //   69
	KEY_F         KEY = 3  //   70
	KEY_G         KEY = 5  //   71
	KEY_H         KEY = 4  //   72
	KEY_I         KEY = 34  //   73
	KEY_J         KEY = 38  //   74
	KEY_K         KEY = 40  //   75
	KEY_L         KEY = 37  //   76
	KEY_M         KEY = 46  //   77
	KEY_N         KEY = 45  //   78
	KEY_O         KEY = 31  //   79
	KEY_P         KEY = 35  //   80
	KEY_Q         KEY = 12  //   81
	KEY_R         KEY = 15  //   82
	KEY_S         KEY = 0x01  //   83
	KEY_T         KEY = 17  //   84
	KEY_U         KEY = 32  //   85
	KEY_V         KEY = 9  //   86
	KEY_W         KEY = 13  //   87
	KEY_X         KEY = 7  //   88
	KEY_Y         KEY = 16  //   89
	KEY_Z         KEY = 6  //   90
	KEY_LWIN      KEY = 59  //   91
	KEY_RWIN      KEY = 62  //   92
	KEY_NUMPAD0   KEY = 82  //   96
	KEY_NUMPAD1   KEY = 83  //   97
	KEY_NUMPAD2   KEY = 84  //   98
	KEY_NUMPAD3   KEY = 85  //   99
	KEY_NUMPAD4   KEY = 86 //   100
	KEY_NUMPAD5   KEY = 87 //   101
	KEY_NUMPAD6   KEY = 88 //   102
	KEY_NUMPAD7   KEY = 89 //   103
	KEY_NUMPAD8   KEY = 91 //   104
	KEY_NUMPAD9   KEY = 92 //   105
	KEY_MULTIPLY  KEY = 67 //   106
	KEY_ADD       KEY = 69 //   107
	KEY_NUMPADENTER KEY = 76
	KEY_SEPARATOR KEY = 108 //   108
	KEY_SUBTRACT  KEY = 78 //   109
	KEY_DECIMAL   KEY = 65 //   110
	KEY_DIVIDEKEY KEY = 75 //   111
	KEY_F1        KEY = 122 //   112
	KEY_F2        KEY = 120 //   113
	KEY_F3        KEY = 99 //   114
	KEY_F4        KEY = 118 //   115
	KEY_F5        KEY = 96 //   116
	KEY_F6        KEY = 97 //   117
	KEY_F7        KEY = 98 //   118
	KEY_F8        KEY = 100 //   119
	KEY_F9        KEY = 101 //   120
	KEY_F10       KEY = 109 //   121
	KEY_F11       KEY = 103 //   122
	KEY_F12       KEY = 111 //   123
	KEY_F13       KEY = -1 //   124
	KEY_F14       KEY = -1 //   125
	KEY_F15       KEY = -1 //   126
	KEY_F16       KEY = -1 //   127
	KEY_F17       KEY = -1 //   128
	KEY_F18       KEY = -1 //   129
	KEY_F19       KEY = -1 //   130
	KEY_F20       KEY = -1 //   131
	KEY_F21       KEY = -1 //   132
	KEY_F22       KEY = -1 //   133
	KEY_F23       KEY = -1 //   134
	KEY_F24       KEY = -1 //   135

	KEY_NUMLOCK  KEY = 71 //   144
	KEY_SCROLL   KEY = 107 //   145
	KEY_LSHIFT   KEY = 56 //   160
	KEY_RSHIFT   KEY = 60 //   161
	KEY_LCONTROL KEY = 54 //   162
	KEY_RCONTROL KEY = 55 //   163
	KEY_LALT KEY = 58 //   164
	KEY_RALT KEY = 61 //   165

	KEY_VOLUMEMUTE        KEY = 0x4A //   173
	KEY_VOLUMEDOWN        KEY = 0x49 //   174
	KEY_VOLUMEUP          KEY = 0x48 //   175

	KEY_SEMICOLONKEY KEY = 41 //   186
	KEY_EQUAL        KEY = 24 //   187
	KEY_COMMA        KEY = 43 //   188
	KEY_MINUS        KEY = 27 //   189
	KEY_PERIOD       KEY = 47 //   190
	KEY_SLASH        KEY = 44 //   191
	KEY_TILDE        KEY = 50 //   192
	KEY_LEFTBRACKET  KEY = 33 //   219
	KEY_BACKSLASHKEY KEY = 42 //   220
	KEY_RIGHTBRACKET KEY = 30 //   221
	KEY_QUOTE        KEY = 39 //   222
	KEY_NONE       KEY = 255 //   255
)

type INPUT_EVENT protocol.Enum

func (this INPUT_EVENT) IsKeyEvent()bool{
	return this<=KEY_EVENT_CLICK
}

func (this INPUT_EVENT) String()string{
	switch this {
		case KEY_EVENT_DOWN:
			return "KEY_EVENT_DOWN"
		case KEY_EVENT_UP:
			return "KEY_EVENT_UP"
		case KEY_EVENT_CLICK:
			return "KEY_EVENT_CLICK"
		case MOUSE_EVENT_BUTTON_DOWN:
			return "MOUSE_EVENT_BUTTON_DOWN"
		case MOUSE_EVENT_BUTTON_UP:
			return "MOUSE_EVENT_BUTTON_UP"
		case MOUSE_EVENT_BUTTON_CLICK:
			return "MOUSE_EVENT_BUTTON_CLICK"
		case MOUSE_EVENT_SCROLL:
			return "MOUSE_EVENT_SCROLL"
		case MOUSE_EVENT_MOVE:
			return "MOUSE_EVENT_MOVE"
		case MOUSE_EVENT_ENTER:
			return "MOUSE_EVENT_ENTER"
		case MOUSE_EVENT_LEAVE:
			return "MOUSE_EVENT_LEAVE"
		case MOUSE_EVENT_CANCEL:
			return "MOUSE_EVENT_CANCEL"
	default:
		return ""
	}
}

const (
	KEY_EVENT_DOWN INPUT_EVENT = iota + 1
	KEY_EVENT_UP
	KEY_EVENT_CLICK
	MOUSE_EVENT_BUTTON_DOWN
	MOUSE_EVENT_BUTTON_UP
	MOUSE_EVENT_BUTTON_CLICK
	MOUSE_EVENT_SCROLL
	MOUSE_EVENT_MOVE
	MOUSE_EVENT_ENTER
	MOUSE_EVENT_LEAVE
	MOUSE_EVENT_CANCEL
)

type MOUSE_BUTTON protocol.Enum

const (
	MOUSE_BUTTON_LEFT MOUSE_BUTTON = iota + 1
	MOUSE_BUTTON_RIGHT
	MOUSE_BUTTON_MIDDLE
	MOUSE_BUTTON_EXTRA1
	MOUSE_BUTTON_EXTRA2
)