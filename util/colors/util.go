package colors

import (
	"math"
	"strconv"
	"strings"
)

func maxFloat64(i ...float64) float64 {
	var max float64 = -math.MaxFloat64
	for _, v := range i {
		if max < v {
			max = v
		}
	}
	return max
}

func minFloat64(i ...float64) float64 {
	var min float64 = math.MaxFloat64
	for _, v := range i {
		if min > v {
			min = v
		}
	}
	return min
}

func maxByte(i ...byte) byte {
	var max byte = 0
	for _, v := range i {
		if max < v {
			max = v
		}
	}
	return max
}
func minByte(i ...byte) byte {
	var min byte = math.MaxUint8
	for _, v := range i {
		if min > v {
			min = v
		}
	}
	return min
}

func Rgb2Hsv(r, g, b byte) (float64, float64, float64) {
	var r1 = float64(r) / 255
	var g1 = float64(g) / 255
	var b1 = float64(b) / 255
	max := maxFloat64(r1, g1, b1)
	min := minFloat64(r1, g1, b1)
	diff := max - min
	var h float64 = 0
	v := max
	var s float64 = 0
	if max != 0 {
		s = diff / max
	}

	if max == min {
		h = 0
	} else if max == r1 && g1 >= b1 {
		h = 60 * ((g1 - b1) / diff)
	} else if max == r1 && g1 < b1 {
		h = 60*((g1-b1)/diff) + 360
	} else if max == g1 {
		h = 60*((b1-r1)/diff) + 120
	} else if max == b1 {
		h = 60*((r1-g1)/diff) + 240
	}
	return math.Round(h), math.Round(s * 100), math.Round(v * 100)
}

func Hsv2Rgb(h, s, v float64) (byte, byte, byte) {
	h /= 1
	s /= 100
	v /= 100
	var r float64 = 0
	var g float64 = 0
	var b float64 = 0

	if s == 0 {
		r = v
		g = v
		b = v
	} else {
		_h := h / 60
		i := math.Floor(_h)
		f := _h - i
		p := v * (1 - s)
		q := v * (1 - f*s)
		t := v * (1 - (1-f)*s)
		switch i {
		case 0:
			r = v
			g = t
			b = p
			break

		case 1:
			r = q
			g = v
			b = p
			break

		case 2:
			r = p
			g = v
			b = t
			break

		case 3:
			r = p
			g = q
			b = v
			break

		case 4:
			r = t
			g = p
			b = v
			break

		case 5:
			r = v
			g = p
			b = q
			break
		}
	}
	return byte(math.Round(r * 255)), byte(math.Round(g * 255)), byte(math.Round(b * 255))
}

func Rgb2Hsl(r, g, b byte) (float64, float64, float64) {
	var r1 = float64(r) / 255
	var g1 = float64(g) / 255
	var b1 = float64(b) / 255
	max := maxFloat64(r1, g1, b1)
	min := minFloat64(r1, g1, b1)
	h := (max + min) / 2
	s := h
	l := h
	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		s = d / (max + min)
		if l > 0.5 {
			s = (2 - max - min)
		}
		switch max {
		case r1:
			h = (g1 - b1) / d
			if g1 < b1 {
				h += 6
			}
		case g1:
			h = (b1-r1)/d + 2
		case b1:
			h = (r1-g1)/d + 4
		}
		h /= 6
	}
	return h * 360, s * 100, l * 100
}

func Hsl2Rgb(h, s, l float64) (byte, byte, byte) {
	h = h / 360.0 * 6.0
	s /= 100.0
	l /= 100.0
	v := (l + s) - (s * l)
	if l < 0.5 {
		v = l * (1 + s)
	}
	m := l + l - v
	var sv float64 = 0
	if v != 0 {
		sv = (v - m) / v
	}
	sextant := math.Floor(h)
	fract := h - sextant
	vsf := v * sv * fract
	t := m + vsf
	q := v - vsf
	mod := int(sextant) % 6

	return byte(math.Floor([]float64{v, q, m, m, t, v}[mod] * 255)),
		byte(math.Floor([]float64{t, v, v, q, m, m}[mod] * 255)),
		byte(math.Floor([]float64{m, m, t, v, v, q}[mod] * 255))
}

func Rgb2Hex(r, g, b byte) string {
	r1 := int64(r)
	g1 := int64(g)
	b1 := int64(b)

	str := ""
	if r1 < 16 {
		str += "0"
	}
	str += strconv.FormatInt(r1, 16)
	if g1 < 16 {
		str += "0"
	}
	str += strconv.FormatInt(g1, 16)
	if b1 < 16 {
		str += "0"
	}
	str += strconv.FormatInt(b1, 16)
	return strings.ToUpper(str)
}

func Hex2Rgb(hex string) (byte, byte, byte) {
	hex = strings.TrimLeft(hex, "#")
	hexArr := strings.Split(hex, "")
	rStr := ""
	gStr := ""
	bStr := ""
	if len(hexArr) == 3 {
		rStr += hexArr[0]
		rStr += hexArr[0]
		gStr += hexArr[1]
		gStr += hexArr[1]
		bStr += hexArr[2]
		bStr += hexArr[2]
	} else {
		rStr += hexArr[0]
		rStr += hexArr[1]
		gStr += hexArr[2]
		gStr += hexArr[3]
		bStr += hexArr[4]
		bStr += hexArr[5]
	}
	r, _ := strconv.ParseInt(rStr, 16, 64)
	g, _ := strconv.ParseInt(gStr, 16, 64)
	b, _ := strconv.ParseInt(bStr, 16, 64)
	return byte(r), byte(g), byte(b)
}

func luminanceUtil(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

const (
	LUMINANCE_R = 0.299
	LUMINANCE_G = 0.587
	LUMINANCE_B = 0.114
)

func GetLuminance(r, g, b byte, normalized bool) float64 {
	div := 1.0
	if !normalized {
		div = 255
	}
	r1 := float64(r) / div
	g1 := float64(g) / div
	b1 := float64(b) / div
	r1 = luminanceUtil(r1)
	g1 = luminanceUtil(g1)
	b1 = luminanceUtil(b1)
	return LUMINANCE_R*r1 + LUMINANCE_G*g1 + LUMINANCE_B*b1
}
