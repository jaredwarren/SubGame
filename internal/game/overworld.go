package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldScene manages the top-down surface sailing view.
type OverworldScene struct {
	World *world.World
}

// NewOverworldScene creates a new OverworldScene.
func NewOverworldScene(world *world.World) *OverworldScene {
	return &OverworldScene{
		World: world,
	}
}

func (o *OverworldScene) OnEnter(g *Game) {
	g.currentState = StateOverworld
}

func (o *OverworldScene) OnExit(g *Game) {}

// Update handles input, movement physics, and checks state transition triggers.
func (o *OverworldScene) Update(g *Game) error {
	if g.ActiveVehicle != nil {
		return nil
	}

	p := g.player

	// On foot swimming in Overworld
	var accel = 0.08
	var maxSpeed = 1.6
	isSprinting := g.Input.IsKeyPressed(ebiten.KeyShift)

	if isSprinting && p.CurrentStamina > 0 {
		accel = 0.16
		maxSpeed = 2.6
	}

	// Apply Fins upgrade speed boost (30% increase)
	if p.HasFins {
		accel *= 1.30
		maxSpeed *= 1.30
	}

	// Direct WASD movement (no steering inertia when swimming on foot)
	moving := false
	var dx, dy float64
	if g.Input.IsKeyPressed(ebiten.KeyW) || g.Input.IsKeyPressed(ebiten.KeyArrowUp) {
		dy -= 1.0
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyS) || g.Input.IsKeyPressed(ebiten.KeyArrowDown) {
		dy += 1.0
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyA) || g.Input.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1.0
		moving = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyD) || g.Input.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1.0
		moving = true
	}

	if moving {
		angle := math.Atan2(dy, dx)
		p.Facing = angle // Visor faces swim direction
		p.Vel.X += math.Cos(angle) * accel
		p.Vel.Y += math.Sin(angle) * accel
	}

	// Apply fluid friction (drag) to the swimming player
	const drag = 0.88
	p.Vel = p.Vel.Scale(drag)

	// Speed clamp
	speed := p.Vel.Length()
	if speed > maxSpeed {
		p.Vel = p.Vel.Scale(maxSpeed / speed)
	}

	// AABB Collision check and position updates
	o.checkCollisions(p)

	// Update stats (not in cave, check if sprinting and actually moving)
	isMoving := speed > 0.1
	p.UpdateStats(false, isSprinting && isMoving && moving)

	// Check if player is overlapping a Trench or Water tile to trigger dive transition
	tx := int(p.Pos.X+p.Width/2) / TileSize
	ty := int(p.Pos.Y+p.Height/2) / TileSize
	if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
		tile := o.World.OverworldMap[tx][ty]
		if tile == world.TileTrench || tile == world.TileWater {
			// Dive only if not near base station (so E can be used to open terminal)
			if g.Input.IsKeyPressed(ebiten.KeyE) && g.baseStation.DistanceToPlayer(p) >= 100.0 {
				g.EnterCave(tx, ty)
				return nil
			}
		}
	}

	return nil
}

// checkCollisions resolves AABB collision with land tiles in X and Y directions.
func (o *OverworldScene) checkCollisions(p *Player) {
	// X Axis collision
	newX := p.Pos.X + p.Vel.X
	if o.isSolid(newX, p.Pos.Y, p.Width, p.Height) {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}

	// Y Axis collision
	newY := p.Pos.Y + p.Vel.Y
	if o.isSolid(p.Pos.X, newY, p.Width, p.Height) {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}

// isSolid checks if the proposed bounding box overlaps with solid land.
func (o *OverworldScene) isSolid(x, y, w, h float64) bool {
	x1 := int(math.Floor(x)) / TileSize
	x2 := int(math.Floor(x+w)) / TileSize
	y1 := int(math.Floor(y)) / TileSize
	y2 := int(math.Floor(y+h)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
				return true // Out of bounds is solid
			}
			if o.World.OverworldMap[tx][ty] == world.TileLand {
				return true
			}
		}
	}
	return false
}

// Draw renders the overworld tiles in a viewport centered on the player.
func (o *OverworldScene) Draw(g *Game, screen *ebiten.Image) {
	camera := g.camera
	isPiloting := g.ActiveVehicle != nil

	// Camera offset from camera controller
	camX := camera.Pos.X
	camY := camera.Pos.Y

	// Render tiles that fall within screen viewport bounds
	startTileX := int(camX) / TileSize
	endTileX := (int(camX)+ScreenWidth)/TileSize + 1
	startTileY := int(camY) / TileSize
	endTileY := (int(camY)+ScreenHeight)/TileSize + 1

	// Clamp bounds
	if startTileX < 0 {
		startTileX = 0
	}
	if endTileX > o.World.Width {
		endTileX = o.World.Width
	}
	if startTileY < 0 {
		startTileY = 0
	}
	if endTileY > o.World.Height {
		endTileY = o.World.Height
	}

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			sx := float32(tx*TileSize - int(camX))
			sy := float32(ty*TileSize - int(camY))

			var tileClr color.RGBA
			var strokeClr color.RGBA

			switch o.World.OverworldMap[tx][ty] {
			case world.TileWater:
				// Lerp water color based on distance to land:
				// Close to land = lighter coastal blue, far = darker deep ocean
				dist := o.World.LandDist[tx][ty]
				const maxDist = 15 // tiles at which water reaches full dark
				t := float64(dist) / float64(maxDist)
				if t > 1.0 {
					t = 1.0
				}
				// Coastal (light):  RGBA{28, 85, 165}
				// Deep (dark):      RGBA{8, 32, 82}
				r := uint8(28 - t*20)
				g := uint8(85 - t*53)
				b := uint8(165 - t*83)
				tileClr = color.RGBA{r, g, b, 255}
				strokeClr = color.RGBA{r + 8, g + 10, b + 15, 255}
			case world.TileLand:
				tileClr = color.RGBA{38, 142, 85, 255} // Reef green
				strokeClr = color.RGBA{48, 160, 98, 255}
			case world.TileTrench:
				tileClr = color.RGBA{6, 18, 42, 255} // Deep sinkhole
				strokeClr = color.RGBA{10, 26, 58, 255}
			}

			// Apply daytime/night light multipliers
			mult := GetOverworldLightMultiplier(g.TimeOfDay)
			tileClr = applyLight(tileClr, mult)
			strokeClr = applyLight(strokeClr, mult)

			vector.FillRect(screen, sx, sy, TileSize, TileSize, tileClr, false)
			vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, strokeClr, false)
		}
	}

	// If player is sitting on a Trench, display dive prompt overlay (only if not piloting a vehicle)
	if !isPiloting {
		pTileX := int(g.player.Pos.X+g.player.Width/2) / TileSize
		pTileY := int(g.player.Pos.Y+g.player.Height/2) / TileSize
		if pTileX >= 0 && pTileX < o.World.Width && pTileY >= 0 && pTileY < o.World.Height {
			tile := o.World.OverworldMap[pTileX][pTileY]
			if tile == world.TileTrench || tile == world.TileWater {
				promptX := float32(pCenterX(g.player)) - 80
				promptY := float32(pCenterY(g.player)) - 40
				vector.FillRect(screen, promptX, promptY, 160, 25, color.RGBA{0, 0, 0, 180}, false)
				ebitenutil.DebugPrintAt(screen, "Press [E] to Dive", int(promptX)+25, int(promptY)+4)
			}
		}

		// Draw the player swimming as a small diver circle centered on screen
		pX := float32(pCenterX(g.player))
		pY := float32(pCenterY(g.player))

		// Player body circle
		vector.FillCircle(screen, pX, pY, 8.0, color.RGBA{220, 95, 45, 255}, false)
		// Helmet glass pointing in Facing direction
		vx := pX + float32(math.Cos(g.player.Facing))*5
		vy := pY + float32(math.Sin(g.player.Facing))*5
		vector.FillCircle(screen, vx, vy, 4.0, color.RGBA{80, 200, 255, 200}, false)
	}
}

// Global empty white image for custom vector triangle rendering
var emptyImage = ebiten.NewImage(3, 3)

func init() {
	emptyImage.Fill(color.White)
}

// Pre-allocated buffers for drawing triangles to eliminate garbage collection overhead.
var (
	triangleVertices = make([]ebiten.Vertex, 3)
	triangleIndices  = []uint16{0, 1, 2}
)

// GetOverworldLightMultiplier returns a light multiplier based on TimeOfDay.
func GetOverworldLightMultiplier(timeOfDay float64) float64 {
	// Dawn (ticks 0 to 1200): Lerp multiplier from 0.2 to 1.0
	if timeOfDay >= 0 && timeOfDay < 1200 {
		return 0.2 + (timeOfDay/1200.0)*0.8
	}
	// Day (ticks 1200 to 6000): Constant 1.0
	if timeOfDay >= 1200 && timeOfDay < 6000 {
		return 1.0
	}
	// Dusk (ticks 6000 to 7200): Lerp multiplier from 1.0 to 0.2
	if timeOfDay >= 6000 && timeOfDay < 7200 {
		return 1.0 - ((timeOfDay-6000.0)/1200.0)*0.8
	}
	// Night (ticks 7200 to 14400): Constant 0.2
	return 0.2
}

func applyLight(c color.RGBA, mult float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * mult),
		G: uint8(float64(c.G) * mult),
		B: uint8(float64(c.B) * mult),
		A: c.A,
	}
}
