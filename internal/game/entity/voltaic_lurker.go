package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type LurkerState string

const (
	StateIdle       LurkerState = "idle"
	StateLunging    LurkerState = "lunging"
	StateRetracting LurkerState = "retracting"
	StateCooldown   LurkerState = "cooldown"
)

type VoltaicLurker struct {
	BaseEntity
	def           *VoltaicLurkerDef
	AnchorFace    string // "left", "right", "top", "bottom"
	State         LurkerState
	Extension     float64 // Current extension in pixels from the wall
	CooldownTimer int     // Cooldown duration in frames
	SwayPhase     float64
}

func (v *VoltaicLurker) stats() *VoltaicLurkerDef {
	if v.def != nil {
		return v.def
	}
	return VoltaicLurkerArchetype
}

// NewVoltaicLurker creates a new VoltaicLurker anchored to a specific tile face.
func NewVoltaicLurker(x, y float64, anchorFace string) *VoltaicLurker {
	d := VoltaicLurkerArchetype
	return &VoltaicLurker{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: d.Dims,
			Active:     true,
		},
		def:           d,
		AnchorFace:    anchorFace,
		State:         StateIdle,
		Extension:     0.0,
		CooldownTimer: 0,
		SwayPhase:     rand.Float64() * math.Pi * 2,
	}
}

func (v *VoltaicLurker) getAnchorPoint() gvec.Vec2 {
	d := v.stats()
	cx := v.Pos.X + d.Dims.X/2.0
	cy := v.Pos.Y + d.Dims.Y/2.0
	switch v.AnchorFace {
	case "left":
		return gvec.Vec2{X: v.Pos.X, Y: cy}
	case "right":
		return gvec.Vec2{X: v.Pos.X + d.Dims.X, Y: cy}
	case "top":
		return gvec.Vec2{X: cx, Y: v.Pos.Y}
	case "bottom":
		return gvec.Vec2{X: cx, Y: v.Pos.Y + d.Dims.Y}
	default:
		return gvec.Vec2{X: cx, Y: cy}
	}
}

func (v *VoltaicLurker) getTarget(gr Runtime) (gvec.Vec2, gvec.Vec2, bool) {
	if gr.HasActiveVehicle() {
		return gr.ActiveVehiclePos(), gr.ActiveVehicleDims(), true
	}
	return gr.PlayerPos(), gr.PlayerDims(), true
}

func (v *VoltaicLurker) isTargetInSight(gr Runtime) bool {
	tPos, tDims, ok := v.getTarget(gr)
	if !ok {
		return false
	}

	d := v.stats()
	cx := v.Pos.X + d.Dims.X/2.0
	cy := v.Pos.Y + d.Dims.Y/2.0
	halfW := d.SightHalfWidth
	sightH := halfW * 2.0

	var bx, by, bw, bh float64
	switch v.AnchorFace {
	case "left":
		bx, by, bw, bh = v.Pos.X, cy-halfW, d.SightRange, sightH
	case "right":
		bx, by, bw, bh = v.Pos.X+d.Dims.X-d.SightRange, cy-halfW, d.SightRange, sightH
	case "top":
		bx, by, bw, bh = cx-halfW, v.Pos.Y, sightH, d.SightRange
	case "bottom":
		bx, by, bw, bh = cx-halfW, v.Pos.Y+d.Dims.Y-d.SightRange, sightH, d.SightRange
	default:
		return false
	}

	return rectsOverlap(bx, by, bw, bh, tPos.X, tPos.Y, tDims.X, tDims.Y)
}

func (v *VoltaicLurker) Update(gr Runtime) {
	d := v.stats()
	v.SwayPhase += d.SwayPhaseSpeed

	switch v.State {
	case StateIdle:
		if v.isTargetInSight(gr) {
			v.State = StateLunging
		}
	case StateLunging:
		v.Extension += d.LungeSpeed
		if v.Extension >= d.MaxExtension {
			v.Extension = d.MaxExtension
			v.State = StateRetracting
		}

		anchor := v.getAnchorPoint()
		var headX, headY float64
		var dirX, dirY float64
		switch v.AnchorFace {
		case "left":
			dirX, dirY = 1.0, 0.0
		case "right":
			dirX, dirY = -1.0, 0.0
		case "top":
			dirX, dirY = 0.0, 1.0
		case "bottom":
			dirX, dirY = 0.0, -1.0
		}
		headX = anchor.X + dirX*v.Extension
		headY = anchor.Y + dirY*v.Extension

		hSize := d.HeadSize
		hx := headX - hSize/2.0
		hy := headY - hSize/2.0

		tPos, tDims, ok := v.getTarget(gr)
		if ok && rectsOverlap(hx, hy, hSize, hSize, tPos.X, tPos.Y, tDims.X, tDims.Y) {
			if gr.HasActiveVehicle() {
				gr.Emit(DamageActiveVehicleCmd{Amount: d.Damage})
				gr.Emit(SetMineWarningCmd{
					Message:  "VEHICLE GRABBED AND SHOCKED!",
					Duration: d.WarningDuration,
					Level:    2,
				})
			} else {
				gr.Emit(DamagePlayerCmd{Amount: d.Damage})
				gr.Emit(SetMineWarningCmd{
					Message:  "GRABBED AND SHOCKED BY VOLTAIC LURKER!",
					Duration: d.WarningDuration,
					Level:    2,
				})
			}
			gr.Emit(StunPlayerCmd{Duration: d.StunDuration})
			gr.Emit(TriggerShakeCmd{Duration: d.ShakeDuration, Intensity: d.ShakeIntensity})

			v.State = StateRetracting
		}

	case StateRetracting:
		v.Extension -= d.RetractSpeed
		if v.Extension <= 0.0 {
			v.Extension = 0.0
			v.State = StateCooldown
			v.CooldownTimer = d.CooldownFrames
		}

	case StateCooldown:
		v.CooldownTimer--
		if v.CooldownTimer <= 0 {
			v.State = StateIdle
		}
	}
}

func (v *VoltaicLurker) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	d := v.stats()
	anchor := v.getAnchorPoint()

	// Draw hole background
	holeRadius := 10.0
	vector.FillCircle(screen, float32(anchor.X-camera.Pos.X), float32(anchor.Y-camera.Pos.Y), float32(holeRadius), color.RGBA{10, 8, 15, 255}, false)
	vector.StrokeCircle(screen, float32(anchor.X-camera.Pos.X), float32(anchor.Y-camera.Pos.Y), float32(holeRadius), 1.5, color.RGBA{60, 25, 90, 255}, false)

	var dirX, dirY float64
	var swayX, swayY float64
	switch v.AnchorFace {
	case "left":
		dirX, dirY = 1.0, 0.0
		swayX, swayY = 0.0, 1.0
	case "right":
		dirX, dirY = -1.0, 0.0
		swayX, swayY = 0.0, 1.0
	case "top":
		dirX, dirY = 0.0, 1.0
		swayX, swayY = 1.0, 0.0
	case "bottom":
		dirX, dirY = 0.0, -1.0
		swayX, swayY = 1.0, 0.0
	}

	// Draw segments from base to head
	var headX, headY float64
	for i := 0; i < 7; i++ {
		t := float64(i) / 6.0
		dist := v.Extension * t
		swayVal := math.Sin(v.SwayPhase+float64(i)*0.8) * 3.0 * (v.Extension / d.MaxExtension) * t
		segX := anchor.X + dirX*dist + swayX*swayVal
		segY := anchor.Y + dirY*dist + swayY*swayVal

		if i == 6 {
			headX = segX
			headY = segY
		}

		var rVal float64
		var clr color.RGBA
		if i == 6 {
			rVal = 8.0
			clr = color.RGBA{110, 20, 160, 255}
		} else {
			rVal = 3.0 + float64(i)*0.6
			clr = color.RGBA{130, 30, 180, 255}
		}

		vector.FillCircle(screen, float32(segX-camera.Pos.X), float32(segY-camera.Pos.Y), float32(rVal), clr, false)

		if i == 6 {
			var eye1X, eye1Y, eye2X, eye2Y float64
			if v.AnchorFace == "left" || v.AnchorFace == "right" {
				eye1X = segX + dirX*3.0
				eye1Y = segY - 3.0
				eye2X = segX + dirX*3.0
				eye2Y = segY + 3.0
			} else {
				eye1X = segX - 3.0
				eye1Y = segY + dirY*3.0
				eye2X = segX + 3.0
				eye2Y = segY + dirY*3.0
			}
			vector.FillCircle(screen, float32(eye1X-camera.Pos.X), float32(eye1Y-camera.Pos.Y), 2.0, color.RGBA{0, 230, 255, 255}, false)
			vector.FillCircle(screen, float32(eye2X-camera.Pos.X), float32(eye2Y-camera.Pos.Y), 2.0, color.RGBA{0, 230, 255, 255}, false)
		} else {
			vector.FillCircle(screen, float32(segX-camera.Pos.X), float32(segY-camera.Pos.Y), float32(rVal*0.5), color.RGBA{190, 50, 240, 255}, false)
		}
	}

	// Cyan sparks
	if v.State == StateLunging || (v.State == StateRetracting && v.Extension > 40.0) {
		for s := 0; s < 3; s++ {
			rx := headX + float64(rand.Intn(20)-10)
			ry := headY + float64(rand.Intn(20)-10)
			vector.StrokeLine(screen, float32(headX-camera.Pos.X), float32(headY-camera.Pos.Y), float32(rx-camera.Pos.X), float32(ry-camera.Pos.Y), 1.0, color.RGBA{0, 230, 255, 255}, false)
		}
	}
}
