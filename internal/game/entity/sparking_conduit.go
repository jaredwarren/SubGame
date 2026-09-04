package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// SparkingConduit is a severed electrical cable hazard inside shipwrecks that periodically arcs high voltage.
type SparkingConduit struct {
	BaseEntity
	Timer             int
	CyclePeriod       int
	DischargeDuration int
	PreSparkDuration  int
	IsActiveArc       bool
	IsWarning         bool
	ArcLength         float32
	ShockCooldown     int
}

// NewSparkingConduit creates a conduit hazard at (x, y).
func NewSparkingConduit(x, y float64, seed int64) *SparkingConduit {
	r := rand.New(rand.NewSource(seed))
	cycle := 150 + r.Intn(60) // ~2.5 - 3.5 seconds
	return &SparkingConduit{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 16, Y: 24},
			Active:     true,
		},
		Timer:             r.Intn(cycle), // Desynchronize different conduits
		CyclePeriod:       cycle,
		DischargeDuration: 22,
		PreSparkDuration:  20,
		ArcLength:         18.0,
	}
}

func (c *SparkingConduit) Update(gr Runtime) {
	c.Timer++
	frame := c.Timer % c.CyclePeriod

	startPre := c.CyclePeriod - c.DischargeDuration - c.PreSparkDuration
	startDischarge := c.CyclePeriod - c.DischargeDuration

	c.IsWarning = (frame >= startPre && frame < startDischarge)
	wasActive := c.IsActiveArc
	c.IsActiveArc = (frame >= startDischarge)

	if c.IsActiveArc && !wasActive {
		audio.Get().PlaySFXVaried("sfx/conduit_spark.wav", 0.45, 0.05)
	}

	if c.ShockCooldown > 0 {
		c.ShockCooldown--
	}

	if c.IsActiveArc && c.ShockCooldown <= 0 {
		tipX := c.Pos.X + c.Dimensions.X/2
		tipY := c.Pos.Y + c.Dimensions.Y + 4.0
		pPos := gr.PlayerPos()
		pDims := gr.PlayerDims()
		px := pPos.X + pDims.X/2
		py := pPos.Y + pDims.Y/2

		dist := math.Hypot(px-tipX, py-tipY)
		if dist <= 24.0 {
			c.ShockCooldown = 45
			gr.Emit(&DamagePlayerCmd{Amount: 6.0, Kind: DamageElectric})
			var kx, ky float64
			if dist > 0.001 {
				kx = (px - tipX) / dist * 3.5
				ky = (py - tipY) / dist * 3.5
			} else {
				ky = 3.5
			}
			gr.Emit(&KnockbackPlayerCmd{Force: gvec.Vec2{X: kx, Y: ky}})
		}
	}
}

func (c *SparkingConduit) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	w := float32(c.Dimensions.X)
	h := float32(c.Dimensions.Y)

	// 1. Junction Box mounted to ceiling/wall
	boxColor := color.RGBA{48, 52, 60, 255}
	boxBorder := color.RGBA{75, 82, 95, 255}
	vector.FillRect(screen, sx+2.0, sy, w-4.0, 6.0, boxColor, false)
	vector.StrokeRect(screen, sx+2.0, sy, w-4.0, 6.0, 1.0, boxBorder, false)

	// Warning amber hazard indicator on box
	vector.FillRect(screen, sx+w/2.0-1.5, sy+1.5, 3.0, 3.0, color.RGBA{225, 160, 30, 255}, false)

	// 2. Severed hanging rubber cables
	wireColor := color.RGBA{28, 30, 36, 255}
	copperColor := color.RGBA{220, 125, 60, 255}

	tipX := sx + w/2.0
	tipY := sy + h

	// Cable 1: curved left
	vector.StrokeLine(screen, sx+4.0, sy+6.0, sx+3.0, sy+14.0, 1.8, wireColor, false)
	vector.StrokeLine(screen, sx+3.0, sy+14.0, tipX-2.0, tipY-2.0, 1.8, wireColor, false)

	// Cable 2: curved right
	vector.StrokeLine(screen, sx+w-4.0, sy+6.0, sx+w-2.0, sy+12.0, 1.6, wireColor, false)
	vector.StrokeLine(screen, sx+w-2.0, sy+12.0, tipX+2.0, tipY-1.0, 1.6, wireColor, false)

	// Exposed frayed copper tips
	vector.FillRect(screen, tipX-3.0, tipY-2.0, 2.0, 2.5, copperColor, false)
	vector.FillRect(screen, tipX+1.0, tipY-1.0, 2.0, 2.5, copperColor, false)

	// 3. Warning sizzle effect
	if c.IsWarning {
		sizzleAlpha := uint8(160 + rand.Intn(95))
		sparkColor := color.RGBA{140, 220, 255, sizzleAlpha}
		vector.FillCircle(screen, tipX, tipY, 1.5, sparkColor, false)
	}

	// 4. Full high-voltage electric discharge
	if c.IsActiveArc {
		glowColor := color.RGBA{60, 180, 255, 65}
		coreColor := color.RGBA{255, 255, 255, 255}
		arcColor := color.RGBA{130, 235, 255, 255}

		// Ambient halo glow
		vector.FillCircle(screen, tipX, tipY+4.0, 14.0, glowColor, false)
		vector.FillCircle(screen, tipX, tipY+4.0, 5.0, color.RGBA{100, 210, 255, 140}, false)

		// Draw 3-4 jagged lightning branches
		for b := 0; b < 3; b++ {
			ang := (float64(b) - 1.0) * 0.7 + (rand.Float64()-0.5)*0.5 + math.Pi/2.0
			curX := tipX
			curY := tipY

			for seg := 0; seg < 3; seg++ {
				segLen := float32(4.0 + rand.Float64()*4.0)
				jitter := float32((rand.Float64() - 0.5) * 5.0)
				nextX := curX + float32(math.Cos(ang))*segLen + jitter
				nextY := curY + float32(math.Sin(ang))*segLen
				vector.StrokeLine(screen, curX, curY, nextX, nextY, 1.5, arcColor, false)
				vector.StrokeLine(screen, curX, curY, nextX, nextY, 0.8, coreColor, false)
				curX = nextX
				curY = nextY
			}
		}
	}
}
