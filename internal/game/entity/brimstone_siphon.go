package entity

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
)

// BrimstoneSiphon is a volcanic vent that fires damaging thermal jets.
type BrimstoneSiphon struct {
	BaseEntity
	Timer     int
	Direction string // "up", "down", "left", "right"
}

func (ent *BrimstoneSiphon) Update(gr Runtime) {
	ent.Timer = (ent.Timer + 1) % 120
	if ent.Timer >= 60 {
		var jx, jy, jw, jh float64
		const jetRange = 160.0

		switch ent.Direction {
		case "up":
			jx, jy, jw, jh = ent.Pos.X, ent.Pos.Y-jetRange, ent.Dimensions.X, jetRange
		case "down":
			jx, jy, jw, jh = ent.Pos.X, ent.Pos.Y+ent.Dimensions.Y, ent.Dimensions.X, jetRange
		case "left":
			jx, jy, jw, jh = ent.Pos.X-jetRange, ent.Pos.Y, jetRange, ent.Dimensions.Y
		default:
			jx, jy, jw, jh = ent.Pos.X+ent.Dimensions.X, ent.Pos.Y, jetRange, ent.Dimensions.Y
		}

		vWidth, vHeight := gr.PlayerDims().X, gr.PlayerDims().Y
		targetX, targetY := gr.PlayerPos().X, gr.PlayerPos().Y
		if gr.HasActiveVehicle() {
			vPos := gr.ActiveVehiclePos()
			targetX, targetY = vPos.X, vPos.Y
			vDims := gr.ActiveVehicleDims()
			vWidth, vHeight = vDims.X, vDims.Y
		}

		if rectsOverlap(jx, jy, jw, jh, targetX, targetY, vWidth, vHeight) {
			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: 0.4})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: 0.6})
			}
		}
	}
}

func (ent *BrimstoneSiphon) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(ent.Pos.X - camera.Pos.X)
	sy := float32(ent.Pos.Y - camera.Pos.Y)
	sw := float32(ent.Dimensions.X)
	sh := float32(ent.Dimensions.Y)
	cx := sx + sw/2.0

	entityPath.Reset()
	entityPath.MoveTo(cx-16, sy+32)
	entityPath.LineTo(cx+16, sy+32)
	entityPath.LineTo(cx+8, sy+12)
	entityPath.LineTo(cx-8, sy+12)
	entityPath.Close()

	var ventColor color.RGBA
	if ent.Timer >= 60 {
		ventColor = color.RGBA{185, 85, 45, 255}
	} else {
		ventColor = color.RGBA{65, 55, 50, 255}
	}
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(ventColor)
	vector.FillPath(screen, entityPath, nil, &opts)

	if ent.Timer >= 60 {
		jetLen := float32(120.0)
		switch ent.Direction {
		case "up":
			vector.FillRect(screen, cx-8, sy-jetLen+float32(sh)/2, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-3, sy-jetLen+float32(sh)/2, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
		case "down":
			vector.FillRect(screen, cx-8, sy+16, 16, jetLen, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-3, sy+16, 6, jetLen+10, color.RGBA{245, 220, 40, 160}, false)
		case "left":
			vector.FillRect(screen, cx-jetLen-16, sy-8+float32(sh)/2, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx-jetLen-26, sy-3+float32(sh)/2, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
		default:
			vector.FillRect(screen, cx+16, sy-8+float32(sh)/2, jetLen, 16, color.RGBA{245, 120, 20, 90}, false)
			vector.FillRect(screen, cx+16, sy-3+float32(sh)/2, jetLen+10, 6, color.RGBA{245, 220, 40, 160}, false)
		}
	}
}
