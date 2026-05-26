package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/world"
)

// Vehicle defines the interface that all player-piloted vehicles must implement.
type Vehicle interface {
	Update(g *Game)
	Draw(screen *ebiten.Image, camera *Camera)
	GetPos() Vec2
	SetPos(pos Vec2)
	GetDimensions() Vec2
	GetHealth() float64
	GetMaxHealth() float64
	TakeDamage(amount float64)
	GetOxygen() float64
	GetDepthLimit() float64
	GetCargo() *Inventory
	GetPerspective() string // "overworld" or "cave"
	GetName() string
	GetBattery() float64
	GetMaxBattery() float64
	GetFacing() float64
}

// ---------------------------------------------------------
// 1. THE SKIFF (Surface Boat)
// ---------------------------------------------------------

type Skiff struct {
	Pos        Vec2
	Vel        Vec2
	Dimensions Vec2
	Facing     float64
	Health     float64
	MaxHealth  float64
	Battery    float64
	MaxBattery float64
	Cargo      *Inventory
}

func NewSkiff(x, y float64) *Skiff {
	return &Skiff{
		Pos:        Vec2{X: x, Y: y},
		Dimensions: Vec2{X: 56, Y: 24},
		Facing:     0.0,
		Health:     150.0,
		MaxHealth:  150.0,
		Battery:    100.0,
		MaxBattery: 100.0,
		Cargo:      NewInventory(24),
	}
}

func (s *Skiff) GetPos() Vec2             { return s.Pos }
func (s *Skiff) SetPos(pos Vec2)          { s.Pos = pos }
func (s *Skiff) GetDimensions() Vec2      { return s.Dimensions }
func (s *Skiff) GetHealth() float64       { return s.Health }
func (s *Skiff) GetMaxHealth() float64    { return s.MaxHealth }
func (s *Skiff) TakeDamage(amount float64) {
	s.Health -= amount
	if s.Health < 0 {
		s.Health = 0
	}
}
func (s *Skiff) GetOxygen() float64     { return 100.0 } // Surface breathing
func (s *Skiff) GetDepthLimit() float64 { return 0.0 }   // Surface only
func (s *Skiff) GetCargo() *Inventory   { return s.Cargo }
func (s *Skiff) GetPerspective() string { return "overworld" }
func (s *Skiff) GetName() string        { return "The Skiff" }
func (s *Skiff) GetBattery() float64    { return s.Battery }
func (s *Skiff) GetMaxBattery() float64 { return s.MaxBattery }
func (s *Skiff) GetFacing() float64     { return s.Facing }

func (s *Skiff) Update(g *Game) {
	// Day/night solar charging (TimeOfDay < 7200 is daytime)
	isDaytime := g.TimeOfDay < 7200
	if isDaytime {
		s.Battery += 0.05
		if s.Battery > s.MaxBattery {
			s.Battery = s.MaxBattery
		}
	}

	if g.ActiveVehicle != s {
		s.Vel = Vec2{}
		return
	}

	// Steering inputs
	const turnSpeed = 0.04
	if g.Input.IsKeyPressed(ebiten.KeyA) || g.Input.IsKeyPressed(ebiten.KeyArrowLeft) {
		s.Facing -= turnSpeed
	}
	if g.Input.IsKeyPressed(ebiten.KeyD) || g.Input.IsKeyPressed(ebiten.KeyArrowRight) {
		s.Facing += turnSpeed
	}

	// Drive physics
	var accel = 0.15
	var maxSpeed = 5.0

	// Consumes battery to move. If battery is dead, move at a crawl (manual rowing).
	hasPower := s.Battery > 0
	if !hasPower {
		accel = 0.03
		maxSpeed = 1.2
	}

	moving := false
	if g.Input.IsKeyPressed(ebiten.KeyW) || g.Input.IsKeyPressed(ebiten.KeyArrowUp) {
		s.Vel.X += math.Cos(s.Facing) * accel
		s.Vel.Y += math.Sin(s.Facing) * accel
		moving = true
	} else if g.Input.IsKeyPressed(ebiten.KeyS) || g.Input.IsKeyPressed(ebiten.KeyArrowDown) {
		s.Vel.X -= math.Cos(s.Facing) * (accel * 0.4)
		s.Vel.Y -= math.Sin(s.Facing) * (accel * 0.4)
		moving = true
	}

	if moving && hasPower {
		s.Battery -= 0.02
		if s.Battery < 0 {
			s.Battery = 0
		}
	}

	// Fluid drag
	const drag = 0.94
	s.Vel = s.Vel.Scale(drag)

	speed := s.Vel.Length()
	if speed > maxSpeed {
		s.Vel = s.Vel.Scale(maxSpeed / speed)
	}

	s.checkCollisions(g)
}

func (s *Skiff) checkCollisions(g *Game) {
	newX := s.Pos.X + s.Vel.X
	if s.isSolid(g, Vec2{X: newX, Y: s.Pos.Y}) {
		s.Vel.X = 0
	} else {
		s.Pos.X = newX
	}

	newY := s.Pos.Y + s.Vel.Y
	if s.isSolid(g, Vec2{X: s.Pos.X, Y: newY}) {
		s.Vel.Y = 0
	} else {
		s.Pos.Y = newY
	}
}

func (s *Skiff) isSolid(g *Game, pos Vec2) bool {
	x1 := int(math.Floor(pos.X)) / TileSize
	x2 := int(math.Floor(pos.X+s.Dimensions.X)) / TileSize
	y1 := int(math.Floor(pos.Y)) / TileSize
	y2 := int(math.Floor(pos.Y+s.Dimensions.Y)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= g.world.Width || ty < 0 || ty >= g.world.Height {
				return true
			}
			if g.world.OverworldMap[tx][ty] == world.TileLand {
				return true
			}
		}
	}
	return false
}

func (s *Skiff) Draw(screen *ebiten.Image, camera *Camera) {
	// Rotate boat relative to its center
	cosF := math.Cos(s.Facing)
	sinF := math.Sin(s.Facing)

	rotatePoint := func(px, py float64) (float32, float32) {
		rx := px*cosF - py*sinF
		ry := px*sinF + py*cosF
		return float32(s.Pos.X + s.Dimensions.X/2.0 + rx - camera.Pos.X), float32(s.Pos.Y + s.Dimensions.Y/2.0 + ry - camera.Pos.Y)
	}

	// 1. Draw hull (rotated hexagon)
	x1, y1 := rotatePoint(28, 0)
	x2, y2 := rotatePoint(14, 12)
	x3, y3 := rotatePoint(-28, 12)
	x4, y4 := rotatePoint(-28, -12)
	x5, y5 := rotatePoint(14, -12)

	hullColor := color.RGBA{220, 230, 240, 255}
	stripeColor := color.RGBA{235, 100, 30, 255} // Orange stripe

	// Draw filled triangles for hull
	drawFilledTriangle(screen, x1, y1, x2, y2, x3, y3, hullColor)
	drawFilledTriangle(screen, x1, y1, x3, y3, x4, y4, hullColor)
	drawFilledTriangle(screen, x1, y1, x4, y4, x5, y5, hullColor)

	// Draw border outline
	vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x2, y2, x3, y3, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x3, y3, x4, y4, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x4, y4, x5, y5, 1.5, stripeColor, false)
	vector.StrokeLine(screen, x5, y5, x1, y1, 1.5, stripeColor, false)

	// 2. Draw Cabin glass windshield
	cx1, cy1 := rotatePoint(10, 0)
	cx2, cy2 := rotatePoint(-4, 7)
	cx3, cy3 := rotatePoint(-16, 7)
	cx4, cy4 := rotatePoint(-16, -7)
	cx5, cy5 := rotatePoint(-4, -7)

	cabinColor := color.RGBA{40, 80, 110, 255}
	drawFilledTriangle(screen, cx1, cy1, cx2, cy2, cx3, cy3, cabinColor)
	drawFilledTriangle(screen, cx1, cy1, cx3, cy3, cx4, cy4, cabinColor)
	drawFilledTriangle(screen, cx1, cy1, cx4, cy4, cx5, cy5, cabinColor)

	// 3. Draw Solar Panel on the deck (backside)
	sp1x, sp1y := rotatePoint(-10, -5)
	sp2x, sp2y := rotatePoint(-24, -5)
	sp3x, sp3y := rotatePoint(-24, 5)
	sp4x, sp4y := rotatePoint(-10, 5)

	solarColor := color.RGBA{15, 120, 215, 255}
	drawFilledTriangle(screen, sp1x, sp1y, sp2x, sp2y, sp3x, sp3y, solarColor)
	drawFilledTriangle(screen, sp1x, sp1y, sp3x, sp3y, sp4x, sp4y, solarColor)
	vector.StrokeLine(screen, sp1x, sp1y, sp2x, sp2y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp2x, sp2y, sp3x, sp3y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp3x, sp3y, sp4x, sp4y, 0.8, color.RGBA{220, 240, 255, 180}, false)
	vector.StrokeLine(screen, sp4x, sp4y, sp1x, sp1y, 0.8, color.RGBA{220, 240, 255, 180}, false)
}

// ---------------------------------------------------------
// 2. THE SCOUT SUB (Cave Mini-Sub)
// ---------------------------------------------------------

type ScoutSub struct {
	Pos        Vec2
	Vel        Vec2
	Dimensions Vec2
	Facing     float64
	Health     float64
	MaxHealth  float64
	Battery    float64
	MaxBattery float64
	Cargo      *Inventory
}

func NewScoutSub(x, y float64) *ScoutSub {
	return &ScoutSub{
		Pos:        Vec2{X: x, Y: y},
		Dimensions: Vec2{X: 48, Y: 32},
		Facing:     0.0,
		Health:     100.0,
		MaxHealth:  100.0,
		Battery:    100.0,
		MaxBattery: 100.0,
		Cargo:      NewInventory(12),
	}
}

func (sub *ScoutSub) GetPos() Vec2             { return sub.Pos }
func (sub *ScoutSub) SetPos(pos Vec2)          { sub.Pos = pos }
func (sub *ScoutSub) GetDimensions() Vec2      { return sub.Dimensions }
func (sub *ScoutSub) GetHealth() float64       { return sub.Health }
func (sub *ScoutSub) GetMaxHealth() float64    { return sub.MaxHealth }
func (sub *ScoutSub) TakeDamage(amount float64) {
	sub.Health -= amount
	if sub.Health < 0 {
		sub.Health = 0
	}
}
func (sub *ScoutSub) GetOxygen() float64     { return 100.0 } // Replenishes diver oxygen
func (sub *ScoutSub) GetDepthLimit() float64 { return 60.0 }  // Mid-depth limit (60 tiles)
func (sub *ScoutSub) GetCargo() *Inventory   { return sub.Cargo }
func (sub *ScoutSub) GetPerspective() string { return "cave" }
func (sub *ScoutSub) GetName() string        { return "Scout Sub" }
func (sub *ScoutSub) GetBattery() float64    { return sub.Battery }
func (sub *ScoutSub) GetMaxBattery() float64 { return sub.MaxBattery }
func (sub *ScoutSub) GetFacing() float64     { return sub.Facing }

func (sub *ScoutSub) Update(g *Game) {
	if g.ActiveVehicle != sub {
		sub.Vel = Vec2{}
		return
	}

	// Flashlight pointing direction follows mouse cursor
	cursor := g.Input.Cursor()
	dx := cursor.X - pCenterX(g.player)
	dy := cursor.Y - pCenterY(g.player)
	sub.Facing = math.Atan2(dy, dx)

	// Physics forces
	var force = 0.20
	var maxSpeed = 4.5
	const drag = 0.94

	hasPower := sub.Battery > 0
	if !hasPower {
		force = 0.04
		maxSpeed = 1.0
	}

	if g.playerSlowed {
		force *= 0.5
		maxSpeed *= 0.5
	}

	moving := false
	if g.Input.IsKeyPressed(ebiten.KeyW) || g.Input.IsKeyPressed(ebiten.KeyArrowUp) {
		sub.Vel.Y -= force
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyS) || g.Input.IsKeyPressed(ebiten.KeyArrowDown) {
		sub.Vel.Y += force
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyA) || g.Input.IsKeyPressed(ebiten.KeyArrowLeft) {
		sub.Vel.X -= force
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyD) || g.Input.IsKeyPressed(ebiten.KeyArrowRight) {
		sub.Vel.X += force
		moving = true
	}

	// Consume battery on move
	if moving && hasPower {
		sub.Battery -= 0.03
		if sub.Battery < 0 {
			sub.Battery = 0
		}
	}

	// Water friction
	sub.Vel = sub.Vel.Scale(drag)

	speed := sub.Vel.Length()
	if speed > maxSpeed {
		sub.Vel = sub.Vel.Scale(maxSpeed / speed)
	}

	// Collisions
	sub.checkCollisions(g)

	// Sonar Ping activation (Key Q)
	if hasPower && g.Input.IsKeyJustPressed(ebiten.KeyQ) {
		if g.SonarTimer <= 0 {
			sub.Battery -= 10.0
			if sub.Battery < 0 {
				sub.Battery = 0
			}
			g.SonarTimer = 180 // 3 seconds reveal duration
			g.SonarRadius = 0
			g.SonarSourceX = sub.Pos.X + sub.Dimensions.X/2.0
			g.SonarSourceY = sub.Pos.Y + sub.Dimensions.Y/2.0
		}
	}
}

func (sub *ScoutSub) checkCollisions(g *Game) {
	newX := sub.Pos.X + sub.Vel.X
	if sub.isSolid(g, Vec2{X: newX, Y: sub.Pos.Y}) {
		sub.Vel.X = -sub.Vel.X * 0.3 // Bounce back slightly
		// High speed collision damages vehicle hull
		speed := math.Abs(sub.Vel.X)
		if speed > 2.0 {
			sub.TakeDamage(speed * 4.0)
		}
		sub.Vel.X = 0
	} else {
		sub.Pos.X = newX
	}

	newY := sub.Pos.Y + sub.Vel.Y
	if sub.isSolid(g, Vec2{X: sub.Pos.X, Y: newY}) {
		// High speed collision damages vehicle hull
		speed := math.Abs(sub.Vel.Y)
		if speed > 2.0 {
			sub.TakeDamage(speed * 4.0)
		}
		sub.Vel.Y = 0
	} else {
		sub.Pos.Y = newY
	}
}

func (sub *ScoutSub) isSolid(g *Game, pos Vec2) bool {
	grid := g.caveState.CaveGrid
	if grid == nil {
		return false
	}
	gridW := len(grid)
	gridH := len(grid[0])

	x1 := int(math.Floor(pos.X)) / TileSize
	x2 := int(math.Floor(pos.X+sub.Dimensions.X)) / TileSize
	y1 := int(math.Floor(pos.Y)) / TileSize
	y2 := int(math.Floor(pos.Y+sub.Dimensions.Y)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= gridW {
				return true
			}
			if ty < 0 {
				continue // Surface exit
			}
			if ty >= gridH {
				return true
			}
			if grid[tx][ty] {
				return true
			}
		}
	}
	return false
}

func (sub *ScoutSub) Draw(screen *ebiten.Image, camera *Camera) {
	sx := float32(sub.Pos.X - camera.Pos.X)
	sy := float32(sub.Pos.Y - camera.Pos.Y)
	w := float32(sub.Dimensions.X)
	h := float32(sub.Dimensions.Y)

	// Flip drawing horizontally based on Facing angle (pointing left/right)
	isFacingRight := math.Cos(sub.Facing) >= 0

	subBgClr := color.RGBA{15, 160, 185, 255} // Teal sub body
	domeClr := color.RGBA{80, 205, 255, 180}  // Cyan cockpit dome
	outlineClr := color.RGBA{240, 240, 250, 255}

	// Draw main capsule body
	vector.FillRect(screen, sx+4, sy+4, w-8, h-8, subBgClr, false)
	vector.StrokeRect(screen, sx+4, sy+4, w-8, h-8, 1.5, outlineClr, false)

	// Draw cockpit window dome pointing in facing direction
	if isFacingRight {
		vector.FillRect(screen, sx+w-12, sy+6, 8, h-12, domeClr, false)
		vector.StrokeRect(screen, sx+w-12, sy+6, 8, h-12, 1.0, color.RGBA{255, 255, 255, 255}, false)
		
		// Propeller/rudder at left
		vector.FillRect(screen, sx, sy+h/2.0-8, 4, 16, color.RGBA{220, 100, 30, 255}, false)
	} else {
		vector.FillRect(screen, sx, sy+6, 8, h-12, domeClr, false)
		vector.StrokeRect(screen, sx, sy+6, 8, h-12, 1.0, color.RGBA{255, 255, 255, 255}, false)

		// Propeller/rudder at right
		vector.FillRect(screen, sx+w-4, sy+h/2.0-8, 4, 16, color.RGBA{220, 100, 30, 255}, false)
	}

	// Draw sub engine vent/details
	vector.FillCircle(screen, sx+w/2.0, sy+h/2.0, 5, color.RGBA{20, 30, 50, 255}, false)
}

// ---------------------------------------------------------
// 3. THE HEAVY MECH (Cave Walker/Driller)
// ---------------------------------------------------------

type HeavyMech struct {
	Pos             Vec2
	Vel             Vec2
	Dimensions      Vec2
	Facing          float64
	Health          float64
	MaxHealth       float64
	Battery         float64
	MaxBattery      float64
	Cargo           *Inventory
	IsDrilling      bool
	DrillTimer      int
	TargetDrillNode *ResourceNode
	ThrustersActive bool
}

func NewHeavyMech(x, y float64) *HeavyMech {
	return &HeavyMech{
		Pos:        Vec2{X: x, Y: y},
		Dimensions: Vec2{X: 48, Y: 48},
		Facing:     0.0,
		Health:     200.0,
		MaxHealth:  200.0,
		Battery:    100.0,
		MaxBattery: 100.0,
		Cargo:      NewInventory(8),
	}
}

func (m *HeavyMech) GetPos() Vec2             { return m.Pos }
func (m *HeavyMech) SetPos(pos Vec2)          { m.Pos = pos }
func (m *HeavyMech) GetDimensions() Vec2      { return m.Dimensions }
func (m *HeavyMech) GetHealth() float64       { return m.Health }
func (m *HeavyMech) GetMaxHealth() float64    { return m.MaxHealth }
func (m *HeavyMech) TakeDamage(amount float64) {
	// Heavy Mech ignores 40% of incoming damage
	m.Health -= amount * 0.6
	if m.Health < 0 {
		m.Health = 0
	}
}
func (m *HeavyMech) GetOxygen() float64     { return 100.0 }
func (m *HeavyMech) GetDepthLimit() float64 { return 120.0 } // Crevice depth (all the way)
func (m *HeavyMech) GetCargo() *Inventory   { return m.Cargo }
func (m *HeavyMech) GetPerspective() string { return "cave" }
func (m *HeavyMech) GetName() string        { return "Heavy Mech" }
func (m *HeavyMech) GetBattery() float64    { return m.Battery }
func (m *HeavyMech) GetMaxBattery() float64 { return m.MaxBattery }
func (m *HeavyMech) GetFacing() float64     { return m.Facing }

func (m *HeavyMech) Update(g *Game) {
	if g.ActiveVehicle != m {
		m.Vel.Y += 0.12 // settle to bottom under gravity
		const dragV = 0.95
		m.Vel.Y *= dragV
		m.Vel.X = 0
		m.checkCollisions(g)
		return
	}

	// Update facing angle towards mouse cursor
	cursor := g.Input.Cursor()
	dx := cursor.X - pCenterX(g.player)
	dy := cursor.Y - pCenterY(g.player)
	m.Facing = math.Atan2(dy, dx)

	// Heavy Mech physics (Buoyancy ignored, gravity heavy)
	const gravity = 0.12
	const dragH = 0.88 // Heavy drag in water horizontally
	const dragV = 0.95
	var walkForce = 0.35
	var maxSpeedH = 2.0

	hasPower := m.Battery > 0
	if !hasPower {
		walkForce = 0.08
		maxSpeedH = 0.6
	}

	if g.playerSlowed {
		walkForce *= 0.5
		maxSpeedH *= 0.5
	}

	// 1. Move left/right on floor
	moving := false
	if g.Input.IsKeyPressed(ebiten.KeyA) || g.Input.IsKeyPressed(ebiten.KeyArrowLeft) {
		m.Vel.X -= walkForce
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyD) || g.Input.IsKeyPressed(ebiten.KeyArrowRight) {
		m.Vel.X += walkForce
		moving = true
	}

	// Apply gravity downwards
	m.Vel.Y += gravity

	// 2. Thrusters vertical propulsion (W or Space)
	m.ThrustersActive = false
	if hasPower && (g.Input.IsKeyPressed(ebiten.KeyW) || g.Input.IsKeyPressed(ebiten.KeyArrowUp) || g.Input.IsKeyPressed(ebiten.KeySpace)) {
		m.Vel.Y -= 0.28 // Thrusters counter gravity
		m.Battery -= 0.08
		if m.Battery < 0 {
			m.Battery = 0
		}
		m.ThrustersActive = true
	}

	if moving && hasPower {
		m.Battery -= 0.01
		if m.Battery < 0 {
			m.Battery = 0
		}
	}

	// Friction
	m.Vel.X *= dragH
	m.Vel.Y *= dragV

	// Clamp horizontal speed
	if math.Abs(m.Vel.X) > maxSpeedH {
		m.Vel.X = math.Copysign(maxSpeedH, m.Vel.X)
	}

	// Collisions
	m.checkCollisions(g)

	// Drill arm action timer
	if m.IsDrilling {
		m.DrillTimer--
		if m.DrillTimer <= 0 {
			m.IsDrilling = false
			if m.TargetDrillNode != nil && m.TargetDrillNode.HitsToMine > 0 {
				m.TargetDrillNode.HitsToMine--
				if m.TargetDrillNode.HitsToMine <= 0 {
					// Add to Mech cargo
					m.Cargo.AddItem(m.TargetDrillNode.Type.ItemType(), 1)
					// Remove node from CaveState
					for idx, nd := range g.caveState.Nodes {
						if nd.Tx == m.TargetDrillNode.Tx && nd.Ty == m.TargetDrillNode.Ty {
							g.caveState.Nodes = append(g.caveState.Nodes[:idx], g.caveState.Nodes[idx+1:]...)
							break
						}
					}
				}
			}
			m.TargetDrillNode = nil
		}
	}
}

func (m *HeavyMech) DrillStrike(node *ResourceNode) {
	m.IsDrilling = true
	m.DrillTimer = 15 // Takes 1/4 second to drill 1 hit
	m.TargetDrillNode = node
}

func (m *HeavyMech) checkCollisions(g *Game) {
	newX := m.Pos.X + m.Vel.X
	if m.isSolid(g, Vec2{X: newX, Y: m.Pos.Y}) {
		m.Vel.X = 0
	} else {
		m.Pos.X = newX
	}

	newY := m.Pos.Y + m.Vel.Y
	if m.isSolid(g, Vec2{X: m.Pos.X, Y: newY}) {
		// Sinking fall damage checked if landing hard
		if m.Vel.Y > 4.5 {
			m.TakeDamage((m.Vel.Y - 4.5) * 8.0)
		}
		m.Vel.Y = 0
	} else {
		m.Pos.Y = newY
	}
}

func (m *HeavyMech) isSolid(g *Game, pos Vec2) bool {
	grid := g.caveState.CaveGrid
	if grid == nil {
		return false
	}
	gridW := len(grid)
	gridH := len(grid[0])

	x1 := int(math.Floor(pos.X)) / TileSize
	x2 := int(math.Floor(pos.X+m.Dimensions.X)) / TileSize
	y1 := int(math.Floor(pos.Y)) / TileSize
	y2 := int(math.Floor(pos.Y+m.Dimensions.Y)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= gridW {
				return true
			}
			if ty < 0 {
				continue
			}
			if ty >= gridH {
				return true
			}
			if grid[tx][ty] {
				return true
			}
		}
	}
	return false
}

func (m *HeavyMech) Draw(screen *ebiten.Image, camera *Camera) {
	sx := float32(m.Pos.X - camera.Pos.X)
	sy := float32(m.Pos.Y - camera.Pos.Y)
	w := float32(m.Dimensions.X)
	h := float32(m.Dimensions.Y)

	isFacingRight := math.Cos(m.Facing) >= 0

	mechBodyColor := color.RGBA{218, 98, 16, 255} // Industrial dark orange
	visorColor := color.RGBA{80, 200, 255, 230}
	frameColor := color.RGBA{58, 68, 78, 255}

	// 1. Draw legs
	vector.FillRect(screen, sx+8, sy+h-14, 8, 14, frameColor, false)
	vector.FillRect(screen, sx+w-16, sy+h-14, 8, 14, frameColor, false)
	// Leg joints
	vector.FillCircle(screen, sx+12, sy+h-14, 5, color.RGBA{38, 48, 58, 255}, false)
	vector.FillCircle(screen, sx+w-12, sy+h-14, 5, color.RGBA{38, 48, 58, 255}, false)

	// 2. Draw central main torso cockpit
	vector.FillRect(screen, sx+4, sy+4, w-8, h-16, mechBodyColor, false)
	vector.StrokeRect(screen, sx+4, sy+4, w-8, h-16, 1.5, color.RGBA{250, 160, 50, 255}, false)

	// 3. Draw cockpit glass visor
	if isFacingRight {
		vector.FillRect(screen, sx+w-14, sy+8, 8, 12, visorColor, false)
		// Left arm: claw
		vector.FillRect(screen, sx-6, sy+14, 10, 6, frameColor, false)
		vector.FillRect(screen, sx-6, sy+10, 4, 14, frameColor, false)
		
		// Right arm: drill bit
		drillX := sx + w
		drillY := sy + 18
		if m.IsDrilling {
			// jitter drill while drilling
			drillX += float32(rand.Intn(3) - 1)
			drillY += float32(rand.Intn(3) - 1)
		}
		// Draw drill cone (stacked triangles/lines)
		dx1, dy1 := drillX, drillY
		dx2, dy2 := drillX+12, drillY+4
		dx3, dy3 := drillX, drillY+8
		drawFilledTriangle(screen, dx1, dy1, dx2, dy2, dx3, dy3, color.RGBA{140, 150, 160, 255})
		vector.StrokeLine(screen, dx1, dy1, dx2, dy2, 1.0, color.RGBA{255, 255, 255, 255}, false)
	} else {
		vector.FillRect(screen, sx+6, sy+8, 8, 12, visorColor, false)
		// Right arm: claw
		vector.FillRect(screen, sx+w-4, sy+14, 10, 6, frameColor, false)
		vector.FillRect(screen, sx+w+2, sy+10, 4, 14, frameColor, false)

		// Left arm: drill bit
		drillX := sx - 12
		drillY := sy + 18
		if m.IsDrilling {
			drillX += float32(rand.Intn(3) - 1)
			drillY += float32(rand.Intn(3) - 1)
		}
		dx1, dy1 := drillX, drillY+4
		dx2, dy2 := drillX+12, drillY
		dx3, dy3 := drillX+12, drillY+8
		drawFilledTriangle(screen, dx1, dy1, dx2, dy2, dx3, dy3, color.RGBA{140, 150, 160, 255})
		vector.StrokeLine(screen, dx1, dy1, dx2, dy2, 1.0, color.RGBA{255, 255, 255, 255}, false)
	}

	// 4. Thruster flames if thrusters are engaged
	if m.Battery > 0 && m.ThrustersActive {
		flameX1 := sx + 14
		flameY1 := sy + h - 14
		vector.FillCircle(screen, flameX1, flameY1+float32(rand.Intn(6)), 4, color.RGBA{240, 110, 30, 220}, false)
		
		flameX2 := sx + w - 22
		vector.FillCircle(screen, flameX2, flameY1+float32(rand.Intn(6)), 4, color.RGBA{240, 110, 30, 220}, false)
	}
}
