package entity

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
)

// NerveMat is a floor carpet that slows the player on contact.
type NerveMat struct {
	BaseEntity
}

func (ent *NerveMat) Update(gr Runtime) {
	vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
	targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
	if gr.HasActiveVehicle() {
		vPos := gr.ActiveVehiclePos()
		targetX, targetY = vPos.X, vPos.Y
		vDims := gr.ActiveVehicleDims()
		vWidth, vHeight = vDims.X, vDims.Y
	}
	if rectsOverlap(ent.Pos.X, ent.Pos.Y, ent.Dimensions.X, ent.Dimensions.Y, targetX, targetY, vWidth, vHeight) {
		gr.Emit(SetPlayerSlowedCmd{Slowed: true})
	}
}

func (ent *NerveMat) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)

	vector.FillRect(screen, sx, sy+sh-4, sw, 4, color.RGBA{80, 25, 120, 255}, false)
	for o := float32(4); o < sw; o += 12 {
		vector.StrokeLine(screen, sx+o, sy+sh, sx+o, sy+sh-8, 1.5, color.RGBA{130, 40, 180, 255}, false)
		vector.FillCircle(screen, sx+o, sy+sh-8, 2.0, color.RGBA{180, 60, 220, 255}, false)
	}
}
