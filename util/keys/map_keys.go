package keys

var keycodes = map[string]KEY{
	"0":         KEY_N_0,
	"1":         KEY_N_1,
	"2":         KEY_N_2,
	"3":         KEY_N_3,
	"4":         KEY_N_4,
	"5":         KEY_N_5,
	"6":         KEY_N_6,
	"7":         KEY_N_7,
	"8":         KEY_N_8,
	"9":         KEY_N_9,
	"a":         KEY_A,
	"b":         KEY_B,
	"c":         KEY_C,
	"d":         KEY_D,
	"e":         KEY_E,
	"f":         KEY_F,
	"g":         KEY_G,
	"h":         KEY_H,
	"i":         KEY_I,
	"j":         KEY_J,
	"k":         KEY_K,
	"l":         KEY_L,
	"m":         KEY_M,
	"n":         KEY_N,
	"o":         KEY_O,
	"p":         KEY_P,
	"q":         KEY_Q,
	"r":         KEY_R,
	"s":         KEY_S,
	"t":         KEY_T,
	"u":         KEY_U,
	"v":         KEY_V,
	"w":         KEY_W,
	"x":         KEY_X,
	"y":         KEY_Y,
	"z":         KEY_Z,
	" ":         KEY_SPACE,
	"`":         KEY_TILDE,
	"-":         KEY_MINUS,
	"=":         KEY_EQUAL,
	"[":         KEY_LBRACE,
	"]":         KEY_RBRACE,
	"\\":        KEY_BACKSLASH,
	";":         KEY_SEMICOLON,
	"'":         KEY_QUOTE,
	",":         KEY_COMMA,
	".":         KEY_PERIOD,
	"/":         KEY_SLASH,
	"enter":     KEY_ENTER,
	"backspace": KEY_BACKSPACE,
	"tab":       KEY_TAB,
	"capslock":  KEY_CAPSLOCK,
	"esc":       KEY_ESCAPE,
	"insert":    KEY_INSERT,
	"delete":    KEY_DELETE,
	"home":      KEY_HOME,
	"end":       KEY_END,
	"pageup":    KEY_PGUP,
	"pagedown":  KEY_PGDN,
	"right":     KEY_RIGHT,
	"left":      KEY_LEFT,
	"down":      KEY_DOWN,
	"up":        KEY_UP,
}

func GetKeyCode(str string) KEY {
	return keycodes[str]
}
