package scene

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/base"
	"github.com/jaredwarren/SubGame/internal/game/config"
	oe "github.com/jaredwarren/SubGame/internal/game/entity/overworld"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Update handles input, movement physics, and checks state transition triggers.
func (o *OverworldScene) Update(g GameContext) error {
	if o.whirlpool == nil {
		o.whirlpool = oe.NewWhirlpool(g.GetWorld().Seed)
		rng := rand.New(rand.NewSource(g.GetWorld().Seed + 997))
		pos := o.FindSafeWhirlpoolSpawnPos(g.GetBaseStation().Pos, rng)
		o.whirlpool.Relocate(pos)
	}

	o.whirlpool.Update(whirlpoolContext{scene: o, g: g})
	o.UpdateExtras(g)

	// If piloting a vehicle, apply forces to the vehicle and handle death.
	if v := g.GetActiveVehicle(); v != nil {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		vCenter := gvec.Vec2{
			X: vPos.X + vDims.X/2.0,
			Y: vPos.Y + vDims.Y/2.0,
		}

		// Check death center distance (within 15 pixels of whirlpool eye)
		wpDx := o.whirlpool.Pos.X - vCenter.X
		wpDy := o.whirlpool.Pos.Y - vCenter.Y
		wpDist := math.Hypot(wpDx, wpDy)

		if wpDist < 15.0 {
			g.SetDeathReason("Your vehicle was dragged into the abyss by a violent whirlpool!")
			g.DestroyOverworldVehicle(v)
			g.GetPlayer().CurrentHealth = 0
			return nil
		}

		// Apply pulling force
		force := o.whirlpool.PullForce(vCenter)
		if force.X != 0 || force.Y != 0 {
			newPos := gvec.Vec2{
				X: vPos.X + force.X,
				Y: vPos.Y + force.Y,
			}
			// Collision checks before updating position
			if !o.IsSolid(newPos.X, newPos.Y, vDims.X, vDims.Y) {
				v.SetPos(newPos)
				// Immediately sync player position to center of vehicle to avoid lag
				p := g.GetPlayer()
				p.Pos.X = newPos.X + (vDims.X-p.Width)/2.0
				p.Pos.Y = newPos.Y + (vDims.Y-p.Height)/2.0
			}
		}
		return nil
	}

	p := g.GetPlayer()
	inp := g.GetInput()

	// Calculate and apply whirlpool force to player velocity
	pCenter := gvec.Vec2{
		X: p.Pos.X + p.Width/2.0,
		Y: p.Pos.Y + p.Height/2.0,
	}

	// Check death center distance (within 15 pixels of whirlpool eye)
	wpDx := o.whirlpool.Pos.X - pCenter.X
	wpDy := o.whirlpool.Pos.Y - pCenter.Y
	wpDist := math.Hypot(wpDx, wpDy)

	if wpDist < 15.0 {
		g.SetDeathReason("You were sucked into the center of a swirling vortex!")
		g.GetPlayer().CurrentHealth = 0
		return nil
	}

	// Apply pulling force to player velocity
	force := o.whirlpool.PullForce(pCenter)
	p.Vel.X += force.X
	p.Vel.Y += force.Y

	speedProp := p.Speed["overworld"]
	accel := speedProp.Acceleration
	maxSpeed := speedProp.TopSpeed
	isSprinting := inp.IsKeyPressed(ebiten.KeyShift)

	if isSprinting && p.CurrentStamina > 0 {
		accel *= 1.5
		maxSpeed *= 1.5
	}

	moving := false
	var dx, dy float64
	if inp.IsKeyPressed(ebiten.KeyW) || inp.IsKeyPressed(ebiten.KeyArrowUp) {
		dy -= 1.0
		moving = true
	}
	if inp.IsKeyPressed(ebiten.KeyS) || inp.IsKeyPressed(ebiten.KeyArrowDown) {
		dy += 1.0
		moving = true
	}
	if inp.IsKeyPressed(ebiten.KeyA) || inp.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1.0
		moving = true
	}
	if inp.IsKeyPressed(ebiten.KeyD) || inp.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1.0
		moving = true
	}

	if moving {
		angle := math.Atan2(dy, dx)
		p.Facing = angle
		p.Vel.X += math.Cos(angle) * accel
		p.Vel.Y += math.Sin(angle) * accel
	}

	const drag = 0.88
	p.Vel = p.Vel.Scale(drag)

	speed := p.Vel.Length()
	if speed > maxSpeed {
		p.Vel = p.Vel.Scale(maxSpeed / speed)
	}

	o.CheckCollisions(p, g.GetBaseStation())

	isMoving := speed > 0.1
	p.UpdateStats(false, isSprinting && isMoving && moving)

	tx := tileAt(p.Pos.X+p.Width/2.0, config.TileSize)
	ty := tileAt(p.Pos.Y+p.Height/2.0, config.TileSize)
	if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
		if inp.IsKeyJustPressed(ebiten.KeyE) {
			g.EnterCave(tx, ty)
			return nil
		}
	} else {
		tile := o.World.OverworldMap[tx][ty]
		if tile == world.TileTrench || tile == world.TileWater || tile == world.TileWreckage {
			if inp.IsKeyJustPressed(ebiten.KeyE) && g.GetBaseStation().DistanceToPlayer(p) >= 100.0 {
				g.EnterCave(tx, ty)
				return nil
			}
		}
	}

	return nil
}

// CheckCollisions resolves player collisions against solid land tiles and the base station.
func (o *OverworldScene) CheckCollisions(p *player.Player, baseStation *base.BaseStation) {
	hasBase := baseStation != nil && baseStation.Size.X > 0 && baseStation.Size.Y > 0

	newX := p.Pos.X + p.Vel.X
	collidesX := o.IsSolid(newX, p.Pos.Y, p.Width, p.Height)
	if !collidesX && hasBase {
		bPos, bSize := baseStation.Pos, baseStation.Size
		collidesX = newX < bPos.X+bSize.X && newX+p.Width > bPos.X &&
			p.Pos.Y < bPos.Y+bSize.Y && p.Pos.Y+p.Height > bPos.Y
	}

	if collidesX {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}

	newY := p.Pos.Y + p.Vel.Y
	collidesY := o.IsSolid(p.Pos.X, newY, p.Width, p.Height)
	if !collidesY && hasBase {
		bPos, bSize := baseStation.Pos, baseStation.Size
		collidesY = p.Pos.X < bPos.X+bSize.X && p.Pos.X+p.Width > bPos.X &&
			newY < bPos.Y+bSize.Y && newY+p.Height > bPos.Y
	}

	if collidesY {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}
