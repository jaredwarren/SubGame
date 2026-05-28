package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type ParticleType int

const (
	ParticleBubble ParticleType = iota
	ParticleDebris
	ParticlePlankton
)

type Particle struct {
	Pos    gvec.Vec2
	Vel    gvec.Vec2
	Color  color.RGBA
	Type   ParticleType
	Size   float32
	Life   float64 // 1.0 down to 0
	Decay  float64
	Wobble float64
}

// SpawnBubble spawns a floating bubble particle at the specified position.
func (g *Game) SpawnBubble(x, y float64) {
	vx := (rand.Float64() - 0.5) * 0.4
	vy := -rand.Float64()*0.8 - 0.4 // floating upwards

	g.Particles = append(g.Particles, Particle{
		Pos:    gvec.Vec2{X: x, Y: y},
		Vel:    gvec.Vec2{X: vx, Y: vy},
		Color:  color.RGBA{200, 230, 255, 160},
		Type:   ParticleBubble,
		Size:   float32(rand.Float64()*3.0 + 1.5),
		Life:   1.0,
		Decay:  rand.Float64()*0.01 + 0.008,
		Wobble: rand.Float64() * 10.0,
	})
}

// SpawnDebris spawns a shower of colored mineral debris particles.
func (g *Game) SpawnDebris(x, y float64, c color.RGBA) {
	for i := 0; i < 6; i++ {
		angle := rand.Float64() * 2.0 * math.Pi
		speed := rand.Float64()*1.6 + 0.6
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle)*speed - 0.4

		g.Particles = append(g.Particles, Particle{
			Pos:   gvec.Vec2{X: x, Y: y},
			Vel:   gvec.Vec2{X: vx, Y: vy},
			Color: c,
			Type:  ParticleDebris,
			Size:  float32(rand.Float64()*3.0 + 1.5),
			Life:  1.0,
			Decay: rand.Float64()*0.025 + 0.015,
		})
	}
}

// SpawnPlankton spawns a slow-drifting plankton / marine snow particle.
func (g *Game) SpawnPlankton(x, y float64) {
	vx := (rand.Float64() - 0.5) * 0.15
	vy := rand.Float64()*0.2 + 0.1 // floating downwards slowly

	g.Particles = append(g.Particles, Particle{
		Pos:    gvec.Vec2{X: x, Y: y},
		Vel:    gvec.Vec2{X: vx, Y: vy},
		Color:  color.RGBA{220, 240, 255, 120}, // soft translucent white/cyan
		Type:   ParticlePlankton,
		Size:   float32(rand.Float64()*1.6 + 0.8),
		Life:   1.0,
		Decay:  rand.Float64()*0.003 + 0.002, // long-lived
		Wobble: rand.Float64() * 100.0,
	})
}

// UpdateParticles updates all active particles and decays/reaps them.
func (g *Game) UpdateParticles() {
	var active []Particle
	for _, p := range g.Particles {
		p.Life -= p.Decay
		if p.Life <= 0 {
			continue
		}

		switch p.Type {
		case ParticleBubble:
			p.Wobble += 0.08
			p.Vel.X = math.Sin(p.Wobble) * 0.3
			p.Pos = p.Pos.Add(p.Vel)
		case ParticleDebris:
			p.Vel.Y += 0.06 // Gravity in water
			p.Vel = p.Vel.Scale(0.93)
			p.Pos = p.Pos.Add(p.Vel)
		case ParticlePlankton:
			p.Wobble += 0.03
			p.Vel.X = math.Sin(p.Wobble) * 0.15
			p.Pos = p.Pos.Add(p.Vel)
		}
		active = append(active, p)
	}
	g.Particles = active
}

// DrawParticles renders the active particles.
func (g *Game) DrawParticles(screen *ebiten.Image) {
	camX := g.camera.Pos.X
	camY := g.camera.Pos.Y

	for _, p := range g.Particles {
		if p.Type == ParticlePlankton {
			// Skip plankton, drawn in the cave background layer
			continue
		}

		sx := float32(p.Pos.X - camX)
		sy := float32(p.Pos.Y - camY)

		c := p.Color
		c.A = uint8(float64(c.A) * p.Life)

		if p.Type == ParticleBubble {
			vector.StrokeCircle(screen, sx, sy, p.Size, 0.8, c, false)
		} else {
			vector.FillRect(screen, sx-p.Size/2.0, sy-p.Size/2.0, p.Size, p.Size, c, false)
		}
	}
}
