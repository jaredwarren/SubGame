package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// SonicDecoy represents the sonic decoy deployable item.
type SonicDecoy struct{}

func (d *SonicDecoy) GetName() string       { return "Sonic Decoy" }
func (d *SonicDecoy) GetMaxStack() int      { return 5 }
func (d *SonicDecoy) GetColor() color.Color { return color.RGBA{180, 210, 50, 255} }
func (d *SonicDecoy) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, d.GetName(), cx, cy, size) {
		return
	}
	// Draw neon yellow-green cylinder/concentric rings fallback
	vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, d.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/2.0, 1.5, color.RGBA{255, 255, 255, 180}, false)
	vector.FillCircle(screen, cx, cy, 3, color.White, false)
}
func (d *SonicDecoy) Use(ctx UsableContext) bool {
	playerCenter := gvec.Vec2{
		X: ctx.PlayerPos().X + ctx.PlayerDims().X/2.0,
		Y: ctx.PlayerPos().Y + ctx.PlayerDims().Y/2.0,
	}
	cursor := ctx.CursorWorldPos()
	dir := gvec.Vec2{X: cursor.X - playerCenter.X, Y: cursor.Y - playerCenter.Y}
	dist := dir.Length()
	if dist > 0 {
		dir = dir.Scale(1.0 / dist)
	} else {
		dir = gvec.Vec2{X: 1, Y: 0}
	}
	launchVel := dir.Scale(6.0)

	ctx.SpawnSonicDecoy(playerCenter, launchVel)
	ctx.SetMineWarning("Sonic Decoy Launched!", 90, 1)
	return true
}

// ChemicalDeterrent represents the chemical deterrent deployable item.
type ChemicalDeterrent struct{}

func (c *ChemicalDeterrent) GetName() string       { return "Chemical Deterrent" }
func (c *ChemicalDeterrent) GetMaxStack() int      { return 5 }
func (c *ChemicalDeterrent) GetColor() color.Color { return color.RGBA{40, 25, 60, 255} }
func (c *ChemicalDeterrent) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, c.GetName(), cx, cy, size) {
		return
	}
	// Draw dark purple capsule with warning stripes
	vector.FillCircle(screen, cx, cy, size/3.0, c.GetColor(), false)
	vector.FillRect(screen, cx-size/6.0, cy-size/2.0, size/3.0, size, c.GetColor(), false)
	// Orange hazard stripe
	vector.FillRect(screen, cx-size/6.0, cy-size/8.0, size/3.0, size/4.0, color.RGBA{240, 110, 40, 255}, false)
}
func (c *ChemicalDeterrent) Use(ctx UsableContext) bool {
	cursor := ctx.CursorWorldPos()
	ctx.SpawnDeterrentCloud(cursor)
	ctx.SetMineWarning("Chemical Deterrent Released!", 90, 1)
	return true
}
