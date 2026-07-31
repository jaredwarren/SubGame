package entity

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// BrimstoneSiphon is a volcanic vent that fires damaging thermal jets.
type BrimstoneSiphon struct {
	BaseEntity
	def       *BrimstoneSiphonDef
	Timer     int
	Direction string // "up", "down", "left", "right"
}

func (ent *BrimstoneSiphon) stats() *BrimstoneSiphonDef {
	if ent.def != nil {
		return ent.def
	}
	return BrimstoneSiphonArchetype
}

// NewBrimstoneSiphon creates a BrimstoneSiphon at the given position facing direction.
func NewBrimstoneSiphon(x, y float64, direction string) *BrimstoneSiphon {
	d := BrimstoneSiphonArchetype
	return &BrimstoneSiphon{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 32, Y: 32},
			Active:     true,
		},
		def:       d,
		Direction: direction,
	}
}

func (ent *BrimstoneSiphon) Update(gr Runtime) {
	d := ent.stats()
	ent.Timer = (ent.Timer + 1) % d.CycleFrames
	if ent.Timer >= d.ActiveStartFrame {
		var jx, jy, jw, jh float64
		jetRange := d.JetRange

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
				gr.Emit(DamageActiveVehicleCmd{Amount: d.VehicleDPS})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: d.PlayerDPS})
			}
		}
	}
}

func (ent *BrimstoneSiphon) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	d := ent.stats()
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
	if ent.Timer >= d.ActiveStartFrame {
		ventColor = color.RGBA{185, 85, 45, 255}
	} else {
		ventColor = color.RGBA{65, 55, 50, 255}
	}
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(ventColor)
	vector.FillPath(screen, entityPath, nil, &opts)

	if ent.Timer >= d.ActiveStartFrame {
		jetLen := float32(d.JetDrawLen)
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
