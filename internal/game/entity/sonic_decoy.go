package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// SonicDecoy attracts hostile creatures by emitting sound waves.
type SonicDecoy struct {
	BaseEntity
	LifeTimer int
}

// NewSonicDecoy creates a new SonicDecoy entity.
func NewSonicDecoy(x, y float64, vel gvec.Vec2) *SonicDecoy {
	return &SonicDecoy{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Vel:        vel,
			Dimensions: gvec.Vec2{X: 16, Y: 16},
			Active:     true,
		},
		LifeTimer: 360,
	}
}

// Update updates the decoy's position, physics, and emits sound waves.
func (d *SonicDecoy) Update(gr Runtime) {
	d.LifeTimer--
	if d.LifeTimer <= 0 {
		d.Active = false
		return
	}

	// Apply drag to velocity
	d.Vel = d.Vel.Scale(0.95)

	// Move X and Y independently with simple wall bounce
	nextX := d.Pos.X + d.Vel.X
	nextY := d.Pos.Y + d.Vel.Y

	if gr.IsSolid(nextX, d.Pos.Y, d.Dimensions.X, d.Dimensions.Y) {
		d.Vel.X = -d.Vel.X * 0.6
	} else {
		d.Pos.X = nextX
	}

	if gr.IsSolid(d.Pos.X, nextY, d.Dimensions.X, d.Dimensions.Y) {
		d.Vel.Y = -d.Vel.Y * 0.6
	} else {
		d.Pos.Y = nextY
	}

	// Emit TriggerSoundWaveCmd every 60 frames (1 second)
	if d.LifeTimer%60 == 0 {
		gr.Emit(TriggerSoundWaveCmd{
			Pos: gvec.Vec2{
				X: d.Pos.X + d.Dimensions.X/2.0,
				Y: d.Pos.Y + d.Dimensions.Y/2.0,
			},
		})
	}
}

// Draw renders the SonicDecoy with a wiggling cylinder and concentric ring pulses.
func (d *SonicDecoy) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(d.Pos.X - camera.Pos.X)
	sy := float32(d.Pos.Y - camera.Pos.Y)
	sw := float32(d.Dimensions.X)
	sh := float32(d.Dimensions.Y)
	cx := sx + sw/2.0
	cy := sy + sh/2.0

	// Draw wiggling cylinder
	wiggle := float32(math.Sin(timeOfDay*0.2 + d.Pos.X + d.Pos.Y)) * 2.0
	cylColor := color.RGBA{180, 210, 50, 255}
	vector.FillRect(screen, cx-4+wiggle, cy-8, 8, 16, cylColor, false)
	vector.StrokeRect(screen, cx-4+wiggle, cy-8, 8, 16, 1.0, color.RGBA{255, 255, 255, 100}, false)

	// Draw concentric pulse circles
	waveProgress := float32((360 - d.LifeTimer) % 60)
	r1 := waveProgress * 1.5
	alpha1 := uint8(max(0, 255-waveProgress*4.25))
	vector.StrokeCircle(screen, cx, cy, r1, 1.0, color.RGBA{180, 210, 50, alpha1}, false)

	waveProgress2 := float32((360 - d.LifeTimer + 30) % 60)
	r2 := waveProgress2 * 1.5
	alpha2 := uint8(max(0, 255-waveProgress2*4.25))
	vector.StrokeCircle(screen, cx, cy, r2, 1.0, color.RGBA{180, 210, 50, alpha2}, false)

	// Spark details
	if math.Sin(timeOfDay*0.5+d.Pos.X) > 0.7 {
		vector.StrokeLine(screen, cx+wiggle-5, cy-5, cx+wiggle+3, cy-8, 1.0, color.RGBA{0, 240, 255, 200}, false)
	}
	if math.Cos(timeOfDay*0.3+d.Pos.Y) > 0.7 {
		vector.StrokeLine(screen, cx+wiggle+5, cy+5, cx+wiggle-3, cy+8, 1.0, color.RGBA{0, 240, 255, 200}, false)
	}
}
