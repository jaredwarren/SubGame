package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// ShellVariant represents the visual and salvage junk type worn by the crab.
type ShellVariant int

const (
	ShellTinCan ShellVariant = iota
	ShellPipeElbow
	ShellCogGear
)

// ScrapHermitCrab is a derelict-adapted crustacean that wears ship scrap as its armor.
type ScrapHermitCrab struct {
	BaseEntity
	def         *ScrapHermitCrabDef
	FacingRight bool
	InShell     bool
	ShellTimer  int
	WalkTimer   int
	ShellType   ShellVariant
}

func (c *ScrapHermitCrab) stats() *ScrapHermitCrabDef {
	if c.def != nil {
		return c.def
	}
	return ScrapHermitCrabArchetype
}

// NewScrapHermitCrab creates a ScrapHermitCrab with a randomized scrap shell.
func NewScrapHermitCrab(x, y float64) *ScrapHermitCrab {
	return NewScrapHermitCrabWithShell(x, y, ShellVariant(rand.Intn(3)))
}

// NewScrapHermitCrabWithShell creates a ScrapHermitCrab with a specific shell variant.
func NewScrapHermitCrabWithShell(x, y float64, shell ShellVariant) *ScrapHermitCrab {
	d := ScrapHermitCrabArchetype
	return &ScrapHermitCrab{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: d.Dims,
			Active:     true,
		},
		def:         d,
		FacingRight: rand.Float64() < 0.5,
		ShellType:   shell,
	}
}

// GetHarvestedItem yields crab meat for survival food.
func (c *ScrapHermitCrab) GetHarvestedItem() item.Item {
	return &item.RawCrab{}
}

// GetBonusHarvestItem yields salvaged materials from the crab's worn shell.
func (c *ScrapHermitCrab) GetBonusHarvestItem() item.Item {
	if c.ShellType == ShellCogGear {
		return &item.ElectronicWaste{}
	}
	return &item.ScrapMetal{}
}

// CanCatch returns true when the player is within catch range.
func (c *ScrapHermitCrab) CanCatch(playerPos gvec.Vec2) bool {
	d := c.stats()
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	return math.Hypot(playerPos.X-cx, playerPos.Y-cy) <= d.CatchRange
}

// IsArmored reports whether the crab is completely tucked inside its metal shell, deflecting damage.
func (c *ScrapHermitCrab) IsArmored() bool {
	return c.InShell
}

func (c *ScrapHermitCrab) Update(gr Runtime) {
	c.update(gr)
}

func (c *ScrapHermitCrab) update(g CrabContext) {
	d := c.stats()
	px := g.PlayerPos().X + g.PlayerDims().X/2
	py := g.PlayerPos().Y + g.PlayerDims().Y/2
	cx := c.Pos.X + c.Dimensions.X/2
	cy := c.Pos.Y + c.Dimensions.Y/2
	dist := math.Hypot(px-cx, py-cy)

	isLit := false
	if g.FlashlightOn() && dist < d.LightRange {
		facingAngle := g.PlayerFacing()
		dx := cx - px
		dy := cy - py
		angleToEnt := math.Atan2(dy, dx)
		diff := angleToEnt - facingAngle
		for diff > math.Pi {
			diff -= 2 * math.Pi
		}
		for diff < -math.Pi {
			diff += 2 * math.Pi
		}
		if math.Abs(diff) < d.FlashlightConeHalfAngle {
			isLit = true
		}
	}

	threatened := (dist < d.ThreatRange || isLit)

	if threatened {
		if !c.InShell {
			// Play metallic clink as crab snaps inside
			audio.Get().PlaySFXVaried("sfx/crab_metal_clink.wav", 0.45, 0.05)
		}
		c.InShell = true
		c.ShellTimer = d.ShellFrames
		c.Vel.X = 0
	} else if c.ShellTimer > 0 {
		c.ShellTimer--
		if c.ShellTimer <= 0 {
			c.InShell = false
		}
	}

	if !c.InShell {
		c.WalkTimer++
		if c.WalkTimer%d.WalkTurnInterval == 0 {
			c.FacingRight = !c.FacingRight
		}
		speed := d.WalkSpeed
		if c.FacingRight {
			c.Vel.X = speed
		} else {
			c.Vel.X = -speed
		}
	}

	c.Vel.Y += d.Gravity
	if c.Vel.Y > d.MaxFallSpeed {
		c.Vel.Y = d.MaxFallSpeed
	}

	nextX := c.Pos.X + c.Vel.X
	if !g.IsSolid(nextX, c.Pos.Y, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.X = nextX
	} else {
		c.FacingRight = !c.FacingRight
		c.Vel.X = 0
	}

	nextY := c.Pos.Y + c.Vel.Y
	if !g.IsSolid(c.Pos.X, nextY, c.Dimensions.X, c.Dimensions.Y) {
		c.Pos.Y = nextY
	} else {
		c.Vel.Y = 0
	}
}

func (c *ScrapHermitCrab) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)
	ccx := sx + sw/2
	ccy := sy + sh/2

	crabBodyColor := color.RGBA{220, 90, 65, 255}
	crabLegColor := color.RGBA{185, 65, 45, 255}

	// 1. Draw organic parts if not fully tucked in
	if !c.InShell {
		legWiggle := float32(math.Sin(timeOfDay*0.15+float64(c.WalkTimer)*0.12)) * 2.0
		// Back/front crawling legs
		vector.StrokeLine(screen, ccx-5, ccy+3, ccx-9, ccy+6+legWiggle, 1.2, crabLegColor, false)
		vector.StrokeLine(screen, ccx-3, ccy+4, ccx-7, ccy+7-legWiggle, 1.2, crabLegColor, false)
		vector.StrokeLine(screen, ccx+5, ccy+3, ccx+9, ccy+6-legWiggle, 1.2, crabLegColor, false)
		vector.StrokeLine(screen, ccx+3, ccy+4, ccx+7, ccy+7+legWiggle, 1.2, crabLegColor, false)

		// Pincers poking out front
		clawColor := color.RGBA{240, 110, 80, 255}
		clawDir := float32(1.0)
		if !c.FacingRight {
			clawDir = -1.0
		}
		vector.FillCircle(screen, ccx+clawDir*8, ccy+2, 2.5, clawColor, false)
		vector.FillCircle(screen, ccx+clawDir*6, ccy+3, 1.8, clawColor, false)

		// Eyestalks peeking out
		eyeX := ccx + clawDir*4
		vector.StrokeLine(screen, eyeX-1.5, ccy-3, eyeX-1.5, ccy-6, 1.0, crabLegColor, false)
		vector.StrokeLine(screen, eyeX+1.5, ccy-3, eyeX+1.5, ccy-6, 1.0, crabLegColor, false)
		vector.FillCircle(screen, eyeX-1.5, ccy-6, 1.2, color.White, false)
		vector.FillCircle(screen, eyeX+1.5, ccy-6, 1.2, color.White, false)
		vector.FillCircle(screen, eyeX-1.5+clawDir*0.4, ccy-6, 0.6, color.Black, false)
		vector.FillCircle(screen, eyeX+1.5+clawDir*0.4, ccy-6, 0.6, color.Black, false)
	}

	// 2. Draw scrap shell armor on top of crab body
	shellCenterX := ccx
	shellCenterY := ccy - 1.0
	if !c.InShell {
		shellCenterX -= float32(1.0)
		if !c.FacingRight {
			shellCenterX += float32(2.0)
		}
	}

	switch c.ShellType {
	case ShellTinCan:
		// Corrugated rusted cylindrical tin can
		canW := float32(11.0)
		canH := float32(8.0)
		canX := shellCenterX - canW/2
		canY := shellCenterY - canH/2

		canBody := color.RGBA{140, 80, 60, 255}
		canRim := color.RGBA{185, 175, 165, 255}
		canRust := color.RGBA{170, 70, 45, 255}

		vector.FillRect(screen, canX, canY, canW, canH, canBody, false)
		vector.StrokeRect(screen, canX, canY, canW, canH, 1.0, canRim, false)
		// Corrugation ridges
		vector.StrokeLine(screen, canX+3, canY+1, canX+3, canY+canH-1, 0.8, canRust, false)
		vector.StrokeLine(screen, canX+6, canY+1, canX+6, canY+canH-1, 0.8, canRim, false)
		vector.StrokeLine(screen, canX+8, canY+1, canX+8, canY+canH-1, 0.8, canRust, false)

	case ShellPipeElbow:
		// Heavy industrial elbow pipe fitting
		pipeColor := color.RGBA{95, 105, 120, 255}
		pipeHighlight := color.RGBA{160, 175, 195, 255}
		pipeOpening := color.RGBA{30, 35, 45, 255}

		vector.FillCircle(screen, shellCenterX, shellCenterY, 5.5, pipeColor, false)
		vector.StrokeCircle(screen, shellCenterX, shellCenterY, 5.5, 1.0, pipeHighlight, false)
		// Threaded flange edge
		vector.FillRect(screen, shellCenterX-5.5, shellCenterY-2.0, 3.0, 4.0, pipeHighlight, false)
		// Dark hollow pipe mouth
		vector.FillCircle(screen, shellCenterX+1.5, shellCenterY, 2.5, pipeOpening, false)

	case ShellCogGear:
		// Rusted brass industrial gear with teeth
		gearColor := color.RGBA{175, 130, 65, 255}
		gearTeethColor := color.RGBA{210, 165, 90, 255}
		gearCenter := color.RGBA{45, 38, 30, 255}

		// Gear teeth points
		for i := 0; i < 6; i++ {
			ang := float64(i) * (math.Pi / 3.0)
			tx := shellCenterX + float32(math.Cos(ang))*5.5
			ty := shellCenterY + float32(math.Sin(ang))*5.5
			vector.FillRect(screen, tx-1.0, ty-1.0, 2.0, 2.0, gearTeethColor, false)
		}
		vector.FillCircle(screen, shellCenterX, shellCenterY, 4.8, gearColor, false)
		vector.StrokeCircle(screen, shellCenterX, shellCenterY, 4.8, 1.0, gearTeethColor, false)
		// Center axle hole
		vector.FillCircle(screen, shellCenterX, shellCenterY, 1.8, gearCenter, false)
	}

	// 3. Tucked shell defensive indicator (metallic glint)
	if c.InShell {
		vector.FillCircle(screen, shellCenterX-2.0, shellCenterY-2.5, 1.2, color.RGBA{255, 255, 255, 220}, false)
		// Tucked legs barely visible underneath
		vector.FillRect(screen, ccx-3, ccy+3, 6, 1.2, crabLegColor, false)
	} else {
		// Exposed crab belly underside
		vector.FillCircle(screen, ccx, ccy+1.5, 2.0, crabBodyColor, false)
	}
}
