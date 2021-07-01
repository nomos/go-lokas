package log

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"time"
)

type ComposeEncoder struct {
	zapcore.ObjectEncoder
	Fields       map[string]interface{}
	cur          map[string]interface{}
	cfg *zapcore.EncoderConfig
	jsonEncoder  zapcore.Encoder
	enableJson   bool
	enableObject bool
}

func NewComposeEncoder(cfg *zapcore.EncoderConfig,enableJson ,enableObj bool) *ComposeEncoder {
	m := make(map[string]interface{})
	ret := &ComposeEncoder{
		jsonEncoder: zapcore.NewJSONEncoder(*cfg),
		Fields:      m,
		cur:         m,
		cfg: cfg,
		enableJson: enableJson,
		enableObject: enableObj,
	}
	return ret
}

func (this ComposeEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	return this.jsonEncoder.EncodeEntry(entry,fields)
}


func (this ComposeEncoder) Clone() *ComposeEncoder{
	m := make(map[string]interface{})
	ret := &ComposeEncoder{
		jsonEncoder: zapcore.NewJSONEncoder(*this.cfg),
		Fields:      m,
		cur:         m,
		cfg: this.cfg,
	}
	for k,v:=range this.Fields {
		ret.Fields[k] = v
	}
	return ret
}

func (this ComposeEncoder) AddArray(k string, v zapcore.ArrayMarshaler) error {
	if !this.enableObject {
		if this.enableJson {
			err := this.jsonEncoder.AddArray(k, v)
			return err
		}
		return nil
	}
	arr := &sliceArrayEncoder{elems: make([]interface{}, 0)}
	err := v.MarshalLogArray(arr)
	this.cur[k] = arr.elems
	return err
}

// AddObject implements composeEnc.
func (this ComposeEncoder) AddObject(k string, v zapcore.ObjectMarshaler) error {
	if !this.enableObject {
		if this.enableJson {
			err := this.jsonEncoder.AddObject(k, v)
			return err
		}
		return nil
	}
	newMap := zapcore.NewMapObjectEncoder()
	this.cur[k] = newMap.Fields
	return v.MarshalLogObject(newMap)
}

// AddBinary implements composeEnc.
func (this ComposeEncoder) AddBinary(k string, v []byte) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddBinary(k, v)
			return
		}
		return
	}

	if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddBinary(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddByteString implements composeEnc.
func (this ComposeEncoder) AddByteString(k string, v []byte) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddByteString(k, v)
			return
		}
		return
	}

	if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddBinary(k, v)
			return
		}
		return
	}
	this.cur[k] = string(v)
}

// AddBool implements composeEnc.
func (this ComposeEncoder) AddBool(k string, v bool) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddBool(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddDuration implements composeEnc.
func (this ComposeEncoder) AddDuration(k string, v time.Duration) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddDuration(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddComplex128 implements composeEnc.
func (this ComposeEncoder) AddComplex128(k string, v complex128) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddComplex128(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddComplex64 implements composeEnc.
func (this ComposeEncoder) AddComplex64(k string, v complex64) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddComplex64(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddFloat64 implements composeEnc.
func (this ComposeEncoder) AddFloat64(k string, v float64) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddFloat64(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddFloat32 implements composeEnc.
func (this ComposeEncoder) AddFloat32(k string, v float32) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddFloat32(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddInt implements composeEnc.
func (this ComposeEncoder) AddInt(k string, v int) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddInt(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddInt64 implements composeEnc.
func (this ComposeEncoder) AddInt64(k string, v int64) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddInt64(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddInt32 implements composeEnc.
func (this ComposeEncoder) AddInt32(k string, v int32) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddInt32(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddInt16 implements composeEnc.
func (this ComposeEncoder) AddInt16(k string, v int16) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddInt16(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddInt8 implements composeEnc.
func (this ComposeEncoder) AddInt8(k string, v int8) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddInt8(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddString implements composeEnc.
func (this ComposeEncoder) AddString(k string, v string) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddString(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddTime implements composeEnc.
func (this ComposeEncoder) AddTime(k string, v time.Time) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddTime(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddUint implements composeEnc.
func (this ComposeEncoder) AddUint(k string, v uint) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddUint(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddUint64 implements composeEnc.
func (this ComposeEncoder) AddUint64(k string, v uint64) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddUint64(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddUint32 implements composeEnc.
func (this ComposeEncoder) AddUint32(k string, v uint32) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddUint32(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddUint16 implements composeEnc.
func (this ComposeEncoder) AddUint16(k string, v uint16) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddUint16(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddUint8 implements composeEnc.
func (this ComposeEncoder) AddUint8(k string, v uint8) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddUint8(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddUintptr implements composeEnc.
func (this ComposeEncoder) AddUintptr(k string, v uintptr) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.AddUintptr(k, v)
			return
		}
		return
	}
	this.cur[k] = v
}

// AddReflected implements composeEnc.
func (this ComposeEncoder) AddReflected(k string, v interface{}) error {
	this.cur[k] = v
	return nil
}

// OpenNamespace implements composeEnc.
func (this ComposeEncoder) OpenNamespace(k string) {
if !this.enableObject {
		if this.enableJson {
			this.jsonEncoder.OpenNamespace(k)
			return
		}
		return
	}

	ns := make(map[string]interface{})
	this.cur[k] = ns
	this.cur = ns
}
