package vehicle

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// emptyImage is a 3x3 white image reused by drawFilledTriangle.
var emptyImage = ebiten.NewImage(3, 3)

// Pre-allocated buffers for drawing triangles — avoids per-frame heap allocation.
var (
	triangleVertices = make([]ebiten.Vertex, 3)
	triangleIndices  = []uint16{0, 1, 2}
)

func init() {
	emptyImage.Fill(color.White)
}

// drawFilledTriangle fills a 2D triangle using Ebitengine DrawTriangles.
func drawFilledTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, clr color.Color) {
	r, g, b, a := clr.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff

	triangleVertices[0] = ebiten.Vertex{DstX: x1, DstY: y1, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af}
	triangleVertices[1] = ebiten.Vertex{DstX: x2, DstY: y2, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af}
	triangleVertices[2] = ebiten.Vertex{DstX: x3, DstY: y3, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af}

	screen.DrawTriangles(triangleVertices, triangleIndices, emptyImage, nil)
}

// drawArc draws a curved arc vector on the screen.
func drawArc(screen *ebiten.Image, cx, cy, radius float64, centerAngle, halfSweep float64, thickness float32, clr color.Color) {
	const segments = 8
	step := (halfSweep * 2.0) / float64(segments)
	startAngle := centerAngle - halfSweep
	var lastX, lastY float32
	for i := 0; i <= segments; i++ {
		angle := startAngle + float64(i)*step
		px := float32(cx + math.Cos(angle)*radius)
		py := float32(cy + math.Sin(angle)*radius)
		if i > 0 {
			vector.StrokeLine(screen, lastX, lastY, px, py, thickness, clr, true)
		}
		lastX, lastY = px, py
	}
}
