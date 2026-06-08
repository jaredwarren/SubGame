package light

import "image/color"

func ApplyLight(c color.RGBA, mult float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * mult),
		G: uint8(float64(c.G) * mult),
		B: uint8(float64(c.B) * mult),
		A: c.A,
	}
}
