package scene

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

type WhirlpoolState int

const (
	WpFadeIn WhirlpoolState = iota
	WpActive
	WpFadeOut
)

type whirlpoolParticle struct {
	angle float64
	dist  float64
	speed float64
	size  float32
	color color.RGBA
}

type Whirlpool struct {
	Pos        gvec.Vec2
	Radius     float64
	Alpha      float64 // 0.0 to 1.0
	Rotation   float64 // in radians, increments over time
	State      WhirlpoolState
	StateTimer int // in ticks
	particles  []whirlpoolParticle
	rng        *rand.Rand
}

// NewWhirlpool creates an unpositioned whirlpool with an internal RNG.
func NewWhirlpool(seed int64) *Whirlpool {
	wp := &Whirlpool{
		Radius:    220.0, // approx 3.4 tiles radius
		rng:       rand.New(rand.NewSource(seed + 997)),
		particles: make([]whirlpoolParticle, 40),
	}
	return wp
}

// Relocate places the whirlpool at a random ocean water tile far from land and player spawn.
func (wp *Whirlpool) Relocate(w *world.World, baseStationPos gvec.Vec2) {
	var candidates []gvec.Vec2
	for tx := 0; tx < w.Width; tx++ {
		for ty := 0; ty < w.Height; ty++ {
			if w.OverworldMap[tx][ty] == world.TileWater {
				// Land distance check: must be far from land
				if w.LandDist[tx][ty] >= 3 {
					tileX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
					tileY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0

					// Distance check to player spawn / base station (at least 15 tiles = 960 pixels)
					dist := math.Hypot(tileX-baseStationPos.X, tileY-baseStationPos.Y)
					if dist >= 960.0 {
						candidates = append(candidates, gvec.Vec2{X: tileX, Y: tileY})
					}
				}
			}
		}
	}

	if len(candidates) > 0 {
		idx := wp.rng.Intn(len(candidates))
		wp.Pos = candidates[idx]
	} else {
		// Fallback: search for any water tile if no deep/far water tile is available
		for tx := 0; tx < w.Width; tx++ {
			for ty := 0; ty < w.Height; ty++ {
				if w.OverworldMap[tx][ty] == world.TileWater {
					tileX := float64(tx*config.TileSize) + float64(config.TileSize)/2.0
					tileY := float64(ty*config.TileSize) + float64(config.TileSize)/2.0
					candidates = append(candidates, gvec.Vec2{X: tileX, Y: tileY})
				}
			}
		}
		if len(candidates) > 0 {
			idx := wp.rng.Intn(len(candidates))
			wp.Pos = candidates[idx]
		} else {
			// Absolute fallback to map center
			wp.Pos = gvec.Vec2{
				X: float64(w.Width*config.TileSize) / 2.0,
				Y: float64(w.Height*config.TileSize) / 2.0,
			}
		}
	}

	wp.State = WpFadeIn
	wp.StateTimer = 300 // 5 seconds at 60 FPS
	wp.Alpha = 0.0

	// Initialize particles
	for i := 0; i < len(wp.particles); i++ {
		wp.respawnParticle(&wp.particles[i])
		wp.particles[i].dist = wp.rng.Float64() * wp.Radius // start at random distance
	}
}

func (wp *Whirlpool) respawnParticle(p *whirlpoolParticle) {
	p.dist = wp.Radius
	p.angle = wp.rng.Float64() * 2 * math.Pi
	p.speed = wp.rng.Float64()*1.2 + 0.8 // inward speed
	p.size = float32(wp.rng.Float64()*3.0 + 1.0)

	r := wp.rng.Float64()
	if r < 0.3 {
		p.color = color.RGBA{220, 240, 255, 180} // foam
	} else if r < 0.7 {
		p.color = color.RGBA{60, 180, 210, 140}  // teal
	} else {
		p.color = color.RGBA{20, 80, 160, 100}   // soft deep blue
	}
}

// Update ticks the state machine, rotates the vortex, and updates the foam particles.
func (wp *Whirlpool) Update(w *world.World, baseStationPos gvec.Vec2) {
	// Spin speed
	wp.Rotation += 0.04
	if wp.Rotation > 2*math.Pi {
		wp.Rotation -= 2 * math.Pi
	}

	if wp.StateTimer > 0 {
		wp.StateTimer--
	}

	switch wp.State {
	case WpFadeIn:
		wp.Alpha = 1.0 - float64(wp.StateTimer)/300.0
		if wp.StateTimer <= 0 {
			wp.State = WpActive
			wp.StateTimer = 7200 // 2 minutes active
			wp.Alpha = 1.0
		}
	case WpActive:
		wp.Alpha = 1.0
		if wp.StateTimer <= 0 {
			wp.State = WpFadeOut
			wp.StateTimer = 300 // 5 seconds fade out
		}
	case WpFadeOut:
		wp.Alpha = float64(wp.StateTimer)/300.0
		if wp.StateTimer <= 0 {
			wp.Relocate(w, baseStationPos)
		}
	}

	// Update foam particles
	for i := 0; i < len(wp.particles); i++ {
		p := &wp.particles[i]
		p.dist -= p.speed
		// Speed up angular rotation near center
		orbitSpeed := 0.01 + (1.0-p.dist/wp.Radius)*0.08
		p.angle += orbitSpeed

		if p.dist <= 10.0 {
			wp.respawnParticle(p)
		}
	}
}

// PullForce calculates the spiral suction vector at targetPos.
func (wp *Whirlpool) PullForce(targetPos gvec.Vec2) gvec.Vec2 {
	if wp.Alpha <= 0.0 {
		return gvec.Vec2{}
	}

	dx := wp.Pos.X - targetPos.X
	dy := wp.Pos.Y - targetPos.Y
	dist := math.Hypot(dx, dy)

	currentRadius := wp.Radius * wp.Alpha
	if dist > currentRadius || dist < 1.0 {
		return gvec.Vec2{}
	}

	nx := dx / dist
	ny := dy / dist

	// Closeness ratio (0.0 at outer edge, 1.0 at center)
	ratio := 1.0 - (dist / currentRadius)

	// Quadratic scaling for pull strength
	maxPull := 6.5
	radialPullStrength := maxPull * (ratio * ratio) * wp.Alpha

	// Tangential swirl force (clockwise orbiting)
	tangentialStrength := maxPull * 0.7 * ratio * wp.Alpha

	// Perpendicular direction (-ny, nx)
	fx := nx*radialPullStrength - ny*tangentialStrength
	fy := ny*radialPullStrength + nx*tangentialStrength

	return gvec.Vec2{X: fx, Y: fy}
}

// Draw renders the spiral vortex and its particles.
func (wp *Whirlpool) Draw(screen *ebiten.Image, camX, camY float64) {
	if wp.Alpha <= 0.0 {
		return
	}

	cx := float32(wp.Pos.X - camX)
	cy := float32(wp.Pos.Y - camY)

	currentRadius := wp.Radius * wp.Alpha

	// Draw dark abyss center (shrinks and fades)
	vector.FillCircle(screen, cx, cy, float32(18.0*wp.Alpha), color.RGBA{4, 8, 20, uint8(225 * wp.Alpha)}, false)

	// Draw spiral arms
	numArms := 6
	pointsPerArm := 25
	maxAngle := 3.2 * math.Pi

	for arm := 0; arm < numArms; arm++ {
		startAngle := float64(arm)*(2.0*math.Pi/float64(numArms)) + wp.Rotation

		var lastX, lastY float32
		hasLast := false

		for i := 0; i < pointsPerArm; i++ {
			t := float64(i) / float64(pointsPerArm-1)
			theta := startAngle + t*maxAngle
			r := t * currentRadius

			px := cx + float32(r*math.Cos(theta))
			py := cy + float32(r*math.Sin(theta))

			opacity := uint8(210 * (1.0 - t) * wp.Alpha)
			var clr color.RGBA
			if t < 0.25 {
				clr = color.RGBA{220, 240, 255, opacity} // White foam near center
			} else if t < 0.65 {
				clr = color.RGBA{45, 175, 205, opacity}  // Teal middle
			} else {
				clr = color.RGBA{15, 65, 130, opacity}   // Deep blue outer
			}

			if hasLast {
				thickness := float32((2.8*(1.0-t) + 0.6) * wp.Alpha)
				vector.StrokeLine(screen, lastX, lastY, px, py, thickness, clr, false)
			}

			lastX, lastY = px, py
			hasLast = true
		}
	}

	// Draw swirling foam particles (clamped to currentRadius)
	for _, p := range wp.particles {
		renderDist := p.dist
		if renderDist > currentRadius {
			renderDist = currentRadius
		}
		px := cx + float32(renderDist*math.Cos(p.angle))
		py := cy + float32(renderDist*math.Sin(p.angle))

		c := p.color
		var alphaMult float64
		if currentRadius > 1.0 {
			alphaMult = 1.0 - renderDist/currentRadius
		}
		c.A = uint8(float64(c.A) * wp.Alpha * alphaMult)

		vector.FillCircle(screen, px, py, p.size, c, false)
	}
}
