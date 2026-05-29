package particle

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

func NewBubbleParticle(x, y float64) *Particle {
	vx := (rand.Float64() - 0.5) * 0.4
	vy := -rand.Float64()*0.8 - 0.4 // floating upwards

	return &Particle{
		Pos:    gvec.Vec2{X: x, Y: y},
		Vel:    gvec.Vec2{X: vx, Y: vy},
		Color:  color.RGBA{200, 230, 255, 160},
		Type:   ParticleBubble,
		Size:   float32(rand.Float64()*3.0 + 1.5),
		Life:   1.0,
		Decay:  rand.Float64()*0.01 + 0.008,
		Wobble: rand.Float64() * 10.0,
	}
}

func NewDebrisParticles(x, y float64, c color.RGBA) []*Particle {
	particles := make([]*Particle, 6)
	for i := 0; i < 6; i++ {
		angle := rand.Float64() * 2.0 * math.Pi
		speed := rand.Float64()*1.6 + 0.6
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle)*speed - 0.4

		particles[i] = &Particle{
			Pos:   gvec.Vec2{X: x, Y: y},
			Vel:   gvec.Vec2{X: vx, Y: vy},
			Color: c,
			Type:  ParticleDebris,
			Size:  float32(rand.Float64()*3.0 + 1.5),
			Life:  1.0,
			Decay: rand.Float64()*0.025 + 0.015,
		}
	}
	return particles
}

func NewPlanktonParticle(x, y float64) *Particle {
	vx := (rand.Float64() - 0.5) * 0.15
	vy := rand.Float64()*0.2 + 0.1 // floating downwards slowly

	return &Particle{
		Pos:    gvec.Vec2{X: x, Y: y},
		Vel:    gvec.Vec2{X: vx, Y: vy},
		Color:  color.RGBA{220, 240, 255, 120}, // soft translucent white/cyan
		Type:   ParticlePlankton,
		Size:   float32(rand.Float64()*1.6 + 0.8),
		Life:   1.0,
		Decay:  rand.Float64()*0.003 + 0.002, // long-lived
		Wobble: rand.Float64() * 100.0,
	}
}

// UpdateParticles updates all active particles and decays/reaps them.
func UpdateParticles(particles []*Particle) []*Particle {
	var active []*Particle
	for _, p := range particles {
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
	return active
}

// DrawParticles renders the active particles.
func DrawParticles(screen *ebiten.Image, particles []*Particle, camX, camY float64) {
	for _, p := range particles {
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
