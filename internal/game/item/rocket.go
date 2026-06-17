package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// EscapeRocket represents the escape rocket end-game item.
type EscapeRocket struct{}

func (e *EscapeRocket) GetName() string       { return "Escape Rocket" }
func (e *EscapeRocket) GetMaxStack() int      { return 1 }
func (e *EscapeRocket) GetColor() color.Color { return color.RGBA{255, 100, 50, 255} }
func (e *EscapeRocket) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, e.GetName(), cx, cy, size) {
		return
	}
	topY := cy - size/2.0
	bottomY := cy + size/2.0
	leftX := cx - size/4.0
	rightX := cx + size/4.0
	midY := cy - size/6.0

	// Nose cone (triangle)
	var path vector.Path
	path.MoveTo(cx, topY)
	path.LineTo(leftX, midY)
	path.LineTo(rightX, midY)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(e.GetColor())
	vector.FillPath(screen, &path, nil, &opts)

	// Body (rectangle)
	vector.FillRect(screen, leftX, midY, size/2.0, bottomY-midY, color.RGBA{220, 220, 220, 255}, false)

	// Thruster flame (orange triangle at bottom)
	var flamePath vector.Path
	flamePath.MoveTo(cx, bottomY+size/4.0)
	flamePath.LineTo(cx-size/6.0, bottomY)
	flamePath.LineTo(cx+size/6.0, bottomY)
	flamePath.Close()
	var flameOpts vector.DrawPathOptions
	flameOpts.ColorScale.ScaleWithColor(color.RGBA{255, 165, 0, 255})
	vector.FillPath(screen, &flamePath, nil, &flameOpts)
}
