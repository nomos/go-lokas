package lokas

import "github.com/nomos/go-lokas/protocol"

type KEY protocol.Enum

const (
	KEY_LBUTTON    KEY = 1  //   1
	KEY_RBUTTON    KEY = 2  //   2
	KEY_CANCEL     KEY = 3  //   3
	KEY_MBUTTON    KEY = 4  //   4
	KEY_XBUTTON1   KEY = 5  //   5
	KEY_XBUTTON2   KEY = 6  //   6
	KEY_BACK       KEY = 8  //   8
	KEY_TAB        KEY = 9  //   9
	KEY_LINEFEED   KEY = 10 //  10
	KEY_CLEAR      KEY = 12 //  12
	KEY_RETURN     KEY = 13 //  13
	KEY_SHIFT      KEY = 16 //  16
	KEY_CONTROL    KEY = 17 //  17
	KEY_MENU       KEY = 18 //  18
	KEY_PAUSE      KEY = 19 //  19
	KEY_CAPITAL    KEY = 20 //  20
	KEY_KANA       KEY = 21 //  21
	KEY_HANGUL     KEY = 21 //  21
	KEY_JUNJA      KEY = 23 //  23
	KEY_FINAL      KEY = 24 //  24
	KEY_HANJA      KEY = 25 //   25
	KEY_KANJI      KEY = 25 //   25
	KEY_CONVERT    KEY = 28 //   28
	KEY_NONCONVERT KEY = 29 //   29
	KEY_ACCEPT     KEY = 30 //   30
	KEY_MODECHANGE KEY = 31 //   31
	KEY_ESCAPE     KEY = 27 //   27
	KEY_SPACE      KEY = 32 //   32
	KEY_PGUP      KEY = 33 //   33
	KEY_PGDN       KEY = 34 //   34
	KEY_END        KEY = 35 //   35
	KEY_HOME       KEY = 36 //   36
	KEY_LEFT       KEY = 37 //   37
	KEY_UP         KEY = 38 //   38
	KEY_RIGHT      KEY = 39 //   39
	KEY_DOWN       KEY = 40 //   40
	KEY_SELECT     KEY = 41 //   41
	KEY_PRINT      KEY = 42 //   42
	KEY_EXECUTE    KEY = 43 //   43
	KEY_SNAPSHOT   KEY = 44 //   44
	KEY_INSERT     KEY = 45 //   45
	KEY_DELETE     KEY = 46 //   46
	KEY_HELP       KEY = 47 //   47
	// VK0 THRU VK9 ARE THE SAME AS ASCII '0' THRU '9' (0X30 - 0X39)
	KEY_0 KEY = 48 //   48
	KEY_1 KEY = 49 //   49
	KEY_2 KEY = 50 //   50
	KEY_3 KEY = 51 //   51
	KEY_4 KEY = 52 //   52
	KEY_5 KEY = 53 //   53
	KEY_6 KEY = 54 //   54
	KEY_7 KEY = 55 //   55
	KEY_8 KEY = 56 //   56
	KEY_9 KEY = 57 //   57
	// VKA THRU VKZ ARE THE SAME AS ASCII 'A' THRU 'Z' (0X41 - 0X5A)
	KEY_A         KEY = 65  //   65
	KEY_B         KEY = 66  //   66
	KEY_C         KEY = 67  //   67
	KEY_D         KEY = 68  //   68
	KEY_E         KEY = 69  //   69
	KEY_F         KEY = 70  //   70
	KEY_G         KEY = 71  //   71
	KEY_H         KEY = 72  //   72
	KEY_I         KEY = 73  //   73
	KEY_J         KEY = 74  //   74
	KEY_K         KEY = 75  //   75
	KEY_L         KEY = 76  //   76
	KEY_M         KEY = 77  //   77
	KEY_N         KEY = 78  //   78
	KEY_O         KEY = 79  //   79
	KEY_P         KEY = 80  //   80
	KEY_Q         KEY = 81  //   81
	KEY_R         KEY = 82  //   82
	KEY_S         KEY = 83  //   83
	KEY_T         KEY = 84  //   84
	KEY_U         KEY = 85  //   85
	KEY_V         KEY = 86  //   86
	KEY_W         KEY = 87  //   87
	KEY_X         KEY = 88  //   88
	KEY_Y         KEY = 89  //   89
	KEY_Z         KEY = 90  //   90
	KEY_LWIN      KEY = 91  //   91
	KEY_RWIN      KEY = 92  //   92
	KEY_APPS      KEY = 93  //   93
	KEY_SLEEP     KEY = 95  //   95
	KEY_NUMPAD0   KEY = 96  //   96
	KEY_NUMPAD1   KEY = 97  //   97
	KEY_NUMPAD2   KEY = 98  //   98
	KEY_NUMPAD3   KEY = 99  //   99
	KEY_NUMPAD4   KEY = 100 //   100
	KEY_NUMPAD5   KEY = 101 //   101
	KEY_NUMPAD6   KEY = 102 //   102
	KEY_NUMPAD7   KEY = 103 //   103
	KEY_NUMPAD8   KEY = 104 //   104
	KEY_NUMPAD9   KEY = 105 //   105
	KEY_NUMPADENTER KEY = -1
	KEY_MULTIPLY  KEY = 106 //   106
	KEY_ADD       KEY = 107 //   107
	KEY_SEPARATOR KEY = 108 //   108
	KEY_SUBTRACT  KEY = 109 //   109
	KEY_DECIMAL   KEY = 110 //   110
	KEY_DIVIDEKEY KEY = 111 //   111
	KEY_F1        KEY = 112 //   112
	KEY_F2        KEY = 113 //   113
	KEY_F3        KEY = 114 //   114
	KEY_F4        KEY = 115 //   115
	KEY_F5        KEY = 116 //   116
	KEY_F6        KEY = 117 //   117
	KEY_F7        KEY = 118 //   118
	KEY_F8        KEY = 119 //   119
	KEY_F9        KEY = 120 //   120
	KEY_F10       KEY = 121 //   121
	KEY_F11       KEY = 122 //   122
	KEY_F12       KEY = 123 //   123
	KEY_F13       KEY = 124 //   124
	KEY_F14       KEY = 125 //   125
	KEY_F15       KEY = 126 //   126
	KEY_F16       KEY = 127 //   127
	KEY_F17       KEY = 128 //   128
	KEY_F18       KEY = 129 //   129
	KEY_F19       KEY = 130 //   130
	KEY_F20       KEY = 131 //   131
	KEY_F21       KEY = 132 //   132
	KEY_F22       KEY = 133 //   133
	KEY_F23       KEY = 134 //   134
	KEY_F24       KEY = 135 //   135

	KEY_CAMERA       KEY = 136 //   136
	KEY_HARDWAREBACK KEY = 137 //   137

	KEY_NUMLOCK  KEY = 144 //   144
	KEY_SCROLL   KEY = 145 //   145
	KEY_LSHIFT   KEY = 160 //   160
	KEY_RSHIFT   KEY = 161 //   161
	KEY_LCONTROL KEY = 162 //   162
	KEY_RCONTROL KEY = 163 //   163
	KEY_LALT KEY = 164 //   164
	KEY_RALT KEY = 165 //   165

	KEY_BROWSERBACK       KEY = 166 //   166
	KEY_BROWSERFORWARDKEY KEY = 167 //   167
	KEY_BROWSERREFRESHKEY KEY = 168 //   168
	KEY_BROWSERSTOP       KEY = 169 //   169
	KEY_BROWSERSEARCH     KEY = 170 //   170
	KEY_BROWSERFAVORITES  KEY = 171 //   171
	KEY_BROWSERHOME       KEY = 172 //   172
	KEY_VOLUMEMUTE        KEY = 173 //   173
	KEY_VOLUMEDOWN        KEY = 174 //   174
	KEY_VOLUMEUP          KEY = 175 //   175
	KEY_MEDIANEXTTRACKKEY KEY = 176 //   176
	KEY_MEDIAPREVTRACKKEY KEY = 177 //   177
	KEY_MEDIASTOP         KEY = 178 //   178
	KEY_MEDIAPLAYPAUSEKEY KEY = 179 //   179
	KEY_LAUNCHMAIL        KEY = 180 //   180
	KEY_LAUNCHMEDIASELECT KEY = 181 //   181
	KEY_LAUNCHAPP1        KEY = 182 //   182
	KEY_LAUNCHAPP2        KEY = 183 //   183

	KEY_SEMICOLONKEY KEY = 186 //   186
	KEY_EQUAL        KEY = 187 //   187
	KEY_COMMA        KEY = 188 //   188
	KEY_MINUS        KEY = 189 //   189
	KEY_PERIOD       KEY = 190 //   190
	KEY_SLASH        KEY = 191 //   191
	KEY_TILDE        KEY = 192 //   192
	KEY_LEFTBRACKET  KEY = 219 //   219
	KEY_BACKSLASHKEY KEY = 220 //   220
	KEY_RIGHTBRACKET KEY = 221 //   221
	KEY_QUOTE        KEY = 222 //   222
	KEY_PARA         KEY = 223 //   223

	KEY_OEM102     KEY = 226 //   226
	KEY_ICOHELPKEY KEY = 227 //   227
	KEY_ICO00      KEY = 228 //   228
	KEY_PROCESSKEY KEY = 229 //   229
	KEY_ICOCLEAR   KEY = 230 //   230
	KEY_PACKET     KEY = 231 //   231
	KEY_ATTN       KEY = 246 //   246
	KEY_CRSEL      KEY = 247 //   247
	KEY_EXSEL      KEY = 248 //   248
	KEY_EREOF      KEY = 249 //   249
	KEY_PLAY       KEY = 250 //   250
	KEY_ZOOM       KEY = 251 //   251
	KEY_NONAME     KEY = 252 //   252
	KEY_PA1        KEY = 253 //   253
	KEY_OEMCLEAR   KEY = 254 //   254
	KEY_NONE       KEY = 255 //   255
)

type INPUT_EVENT protocol.Enum

func (this INPUT_EVENT) IsKeyEvent()bool{
	return this<=KEY_EVENT_CLICK
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