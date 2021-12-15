package keys

var _scancode2key map[SCAN_CODE]KEY
var _key2scancode map[KEY]SCAN_CODE
func init(){
	_scancode2key = map[SCAN_CODE]KEY{}
	_key2scancode = map[KEY]SCAN_CODE{}
	for _,s:=range ALL_SCAN_CODE {
		for _,k:=range ALL_KEY {
			if s.ToString()==k.ToString() {
				_scancode2key[SCAN_CODE(s.Enum())] = KEY(k.Enum())
				_key2scancode[KEY(k.Enum())] = SCAN_CODE(s.Enum())
			}
		}
	}
	_key2scancode[KEY_RALT] = SCAN_CODE_LALT
	_key2scancode[KEY_RCTRL] = SCAN_CODE_LCTRL
}

func (this SCAN_CODE) Key()KEY {
	return _scancode2key[this]
}

func (this SCAN_CODE) Is(key KEY)bool{
	return this.ToString()==key.ToString()
}

func (this KEY) ScanCode()SCAN_CODE {
	return _key2scancode[this]
}

func (this KEY) Is(key SCAN_CODE)bool{
	return this.ToString()==key.ToString()
}