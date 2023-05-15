package colors

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"reflect"
	"strconv"
	"strings"
)

type Color struct {
	R byte
	G byte
	B byte
	A byte
}

func NewColor(r, g, b byte) Color {
	return NewColorRGBA(r, g, b, 255)
}

func NewColorAnyArg(arg ...interface{}) Color {
	if len(arg) == 1 {
		if reflect.TypeOf(arg[0]).Kind() == reflect.String {
			return NewColorHexString(arg[0].(string))
		}
		return arg[0].(Color)
	} else if len(arg) == 3 {
		return NewColor(byte(arg[0].(int)), byte(arg[1].(int)), byte(arg[2].(int)))
	} else if len(arg) == 4 {
		return NewColorRGBA(byte(arg[0].(int)), byte(arg[1].(int)), byte(arg[2].(int)), byte(arg[3].(int)))
	}
	return Color{}
}

func (this Color) EqualAny(arg ...interface{}) bool {
	if len(arg) == 1 {
		switch arg[0].(type) {
		case Color:
			c := arg[0].(Color)
			return this.R == c.R && this.G == c.G && this.B == c.B && this.A == c.A
		case string:
			c := NewColorHexString(arg[0].(string))
			return this.R == c.R && this.G == c.G && this.B == c.B && this.A == c.A
		}
		return false
	} else if len(arg) == 3 {
		c := NewColor(byte(arg[0].(int)), byte(arg[1].(int)), byte(arg[2].(int)))
		return this.R == c.R && this.G == c.G && this.B == c.B && this.A == c.A
	} else if len(arg) == 4 {
		c := NewColorRGBA(byte(arg[0].(int)), byte(arg[1].(int)), byte(arg[2].(int)), byte(arg[3].(int)))
		return this.R == c.R && this.G == c.G && this.B == c.B && this.A == c.A
	}
	return false
}

func (this Color) Equal(c Color) bool {
	return this.R == c.R && this.G == c.G && this.B == c.B && this.A == c.A
}

func (this Color) EqualTolerantAny(t byte, arg ...interface{}) bool {
	c := NewColorAnyArg(arg...)
	max := maxbyte(byteDiff(this.R, c.R), byteDiff(this.G, c.G), byteDiff(this.B, c.B))
	return max <= t
}

func (this Color) EqualTolerant(c Color, t byte) bool {
	max := maxbyte(byteDiff(this.R, c.R), byteDiff(this.G, c.G), byteDiff(this.B, c.B))
	return max <= t
}

func byteDiff(a, b byte) byte {
	if a > b {
		return a - b
	}
	return b - a
}

func maxbyte(b ...byte) byte {
	var max byte = 0
	for _, v := range b {
		if v > max {
			max = v
		}
	}
	return max
}

func NewColorRGBA(r, g, b, a byte) Color {
	ret := Color{R: r, G: g, B: b, A: a}
	return ret
}

func NewColorHSV(h, s, v float64) Color {
	r, g, b := Hsv2Rgb(h, s, v)
	return NewColor(r, g, b)
}

func NewColorHSL(h, s, l float64) Color {
	r, g, b := Hsv2Rgb(h, s, l)
	return NewColor(r, g, b)
}

func NewColorHexString(hex string) Color {
	r, g, b := Hex2Rgb(hex)
	return NewColor(r, g, b)
}

func NewColorUint32(v uint32) Color {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, v)
	arr := bytebuf.Bytes()
	return NewColorRGBA(arr[0], arr[1], arr[2], arr[3])
}

func (this Color) RGB() (byte, byte, byte) {
	return this.R, this.G, this.B
}

func (this Color) RGBA() (byte, byte, byte, byte) {
	return this.R, this.G, this.B, this.A
}

func (this Color) ToRGBA() color.RGBA {
	return color.RGBA{this.R, this.G, this.B, this.A}
}

func (this Color) SetRGBA(r, g, b, a byte) {
	this.R = r
	this.G = g
	this.B = b
	this.A = a
}

func (this Color) SetHSL(h, s, l float64) {
	r, g, b := Hsl2Rgb(h, s, l)
	this.R = r
	this.G = g
	this.B = b
}

func (this Color) H() float64 {
	h, _, _ := Rgb2Hsv(this.R, this.G, this.B)
	return h
}

func (this Color) Hue() float64 {
	h, _, _ := Rgb2Hsv(this.R, this.G, this.B)
	return h
}

func (this Color) S() float64 {
	_, s, _ := Rgb2Hsv(this.R, this.G, this.B)
	return s
}

func (this Color) SV() float64 {
	_, s, _ := Rgb2Hsv(this.R, this.G, this.B)
	return s
}

func (this Color) SaturationV() float64 {
	_, s, _ := Rgb2Hsv(this.R, this.G, this.B)
	return s
}

func (this Color) SL() float64 {
	_, s, _ := Rgb2Hsl(this.R, this.G, this.B)
	return s
}

func (this Color) SaturationL() float64 {
	_, s, _ := Rgb2Hsl(this.R, this.G, this.B)
	return s
}

func (this Color) V() float64 {
	_, _, v := Rgb2Hsv(this.R, this.G, this.B)
	return v
}

func (this Color) Value() float64 {
	_, _, v := Rgb2Hsv(this.R, this.G, this.B)
	return v
}

func (this Color) L() float64 {
	_, _, l := Rgb2Hsl(this.R, this.G, this.B)
	return l
}

func (this Color) Lightness() float64 {
	_, _, l := Rgb2Hsl(this.R, this.G, this.B)
	return l
}

func (this Color) SetH(h float64) {
	_, s, v := Rgb2Hsv(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsv2Rgb(h, s, v)
}

func (this Color) SetS(s float64) {
	h, _, v := Rgb2Hsv(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsv2Rgb(h, s, v)
}

func (this Color) SetSV(s float64) {
	h, _, v := Rgb2Hsv(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsv2Rgb(h, s, v)
}

func (this Color) SetSaturationV(s float64) {
	h, _, v := Rgb2Hsv(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsv2Rgb(h, s, v)
}

func (this Color) SetSL(s float64) {
	h, _, l := Rgb2Hsl(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsl2Rgb(h, s, l)
}

func (this Color) SetSaturationL(s float64) {
	h, _, l := Rgb2Hsl(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsl2Rgb(h, s, l)
}

func (this Color) SetV(v float64) {
	h, s, _ := Rgb2Hsv(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsv2Rgb(h, s, v)
}

func (this Color) SetValue(v float64) {
	h, s, _ := Rgb2Hsv(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsv2Rgb(h, s, v)
}

func (this Color) SetL(l float64) {
	h, s, _ := Rgb2Hsl(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsl2Rgb(h, s, l)
}

func (this Color) SetLightness(l float64) {
	h, s, _ := Rgb2Hsl(this.R, this.G, this.B)
	this.R, this.G, this.B = Hsl2Rgb(h, s, l)
}

func (this Color) HSL() (float64, float64, float64) {
	return Rgb2Hsl(this.R, this.G, this.B)
}

func (this Color) HSV() (float64, float64, float64) {
	return Rgb2Hsv(this.R, this.G, this.B)
}

func (this Color) Hex() string {
	return Rgb2Hex(this.R, this.G, this.B)
}

func (this Color) HexWithAlpha() string {
	alphaStr := ""
	if this.A < 16 {
		alphaStr += "0"
	}
	alphaStr += strconv.FormatInt(int64(this.A), 16)
	return Rgb2Hex(this.R, this.G, this.B) + strings.ToUpper(alphaStr)
}

func (this Color) SetHex(s string) {
	r, g, b := Hex2Rgb(s)
	this.R = r
	this.G = g
	this.B = b
}

func (this Color) SetHSV(h, s, v float64) {
	r, g, b := Hsv2Rgb(h, s, v)
	this.R = r
	this.G = g
	this.B = b
}

func (this Color) Uint32() uint32 {
	return uint32(this.R)<<24 + uint32(this.G)<<16 + uint32(this.B)<<8 + uint32(this.A)
}
