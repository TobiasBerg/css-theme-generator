package colors

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type (
	RGB struct{ R, G, B float64 }
	HSL struct{ H, S, L float64 }
)

func HexToRGB(hex string) (RGB, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return RGB{}, fmt.Errorf("invalid hex: #%s", hex)
	}
	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return RGB{}, fmt.Errorf("invalid hex: #%s", hex)
	}
	return RGB{float64((n >> 16) & 0xff), float64((n >> 8) & 0xff), float64(n & 0xff)}, nil
}

func rgbToHex(r RGB) string {
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(math.Round(v))
	}
	return fmt.Sprintf("#%02x%02x%02x", clamp(r.R), clamp(r.G), clamp(r.B))
}

func rgbToHSL(c RGB) HSL {
	r, g, b := c.R/255, c.G/255, c.B/255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l := (max + min) / 2
	if max == min {
		return HSL{0, 0, l * 100}
	}
	d := max - min
	s := d / (2 - max - min)
	if l <= 0.5 {
		s = d / (max + min)
	}
	var h float64
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	return HSL{h / 6 * 360, s * 100, l * 100}
}

func hslToRGB(c HSL) RGB {
	h, s, l := c.H/360, c.S/100, c.L/100
	if s == 0 {
		v := l * 255
		return RGB{v, v, v}
	}
	hue2rgb := func(p, q, t float64) float64 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		if t < 1.0/6 {
			return p + (q-p)*6*t
		}
		if t < 0.5 {
			return q
		}
		if t < 2.0/3 {
			return p + (q-p)*(2.0/3-t)*6
		}
		return p
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	return RGB{
		hue2rgb(p, q, h+1.0/3) * 255,
		hue2rgb(p, q, h) * 255,
		hue2rgb(p, q, h-1.0/3) * 255,
	}
}

func HslToHex(h, s, l float64) string {
	return rgbToHex(hslToRGB(HSL{h, s, l}))
}

func Lighten(hex string, amt float64) string {
	c, _ := HexToRGB(hex)
	hsl := rgbToHSL(c)
	hsl.L = math.Min(100, hsl.L+amt)
	return rgbToHex(hslToRGB(hsl))
}

func Darken(hex string, amt float64) string {
	c, _ := HexToRGB(hex)
	hsl := rgbToHSL(c)
	hsl.L = math.Max(0, hsl.L-amt)
	return rgbToHex(hslToRGB(hsl))
}

func Mix(hex1, hex2 string, t float64) string {
	a, _ := HexToRGB(hex1)
	b, _ := HexToRGB(hex2)
	return rgbToHex(RGB{
		a.R*(1-t) + b.R*t,
		a.G*(1-t) + b.G*t,
		a.B*(1-t) + b.B*t,
	})
}

func Luminance(hex string) float64 {
	c, _ := HexToRGB(hex)
	lin := func(v float64) float64 {
		v /= 255
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

func IsLight(hex string) bool { return Luminance(hex) > 0.3 }
