package world

import (
	"math"
	"math/rand"
)

// Noise2D implements a self-contained 2D Value Noise generator with FBM.
type Noise2D struct {
	grid [256][256]float64
}

// NewNoise2D creates a seeded noise generator.
func NewNoise2D(seed int64) *Noise2D {
	r := rand.New(rand.NewSource(seed))
	n := &Noise2D{}
	for x := 0; x < 256; x++ {
		for y := 0; y < 256; y++ {
			n.grid[x][y] = r.Float64()
		}
	}
	return n
}

// Noise returns the 2D value noise at coordinate (x, y).
func (n *Noise2D) Noise(x, y float64) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)

	// Smoothstep curve for cubic interpolation S-curve
	u := xf * xf * (3.0 - 2.0*xf)
	v := yf * yf * (3.0 - 2.0*yf)

	xnext := (xi + 1) & 255
	ynext := (yi + 1) & 255

	g00 := n.grid[xi][yi]
	g10 := n.grid[xnext][yi]
	g01 := n.grid[xi][ynext]
	g11 := n.grid[xnext][ynext]

	// Bilinear interpolation using smoothstep weightings
	return (1-v)*((1-u)*g00+u*g10) + v*((1-u)*g01+u*g11)
}

// FBM generates Fractal Brownian Motion summing multiple octaves of noise.
func (n *Noise2D) FBM(x, y float64, octaves int) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxVal := 0.0

	for i := 0; i < octaves; i++ {
		value += amplitude * n.Noise(x*frequency, y*frequency)
		maxVal += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	return value / maxVal
}
