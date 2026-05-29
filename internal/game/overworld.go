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
	speedProp := p.Speed["overworld"]
	var accel = speedProp.Acceleration
	var maxSpeed = speedProp.TopSpeed
	isSprinting := g.Input.IsKeyPressed(ebiten.KeyShift)

	if isSprinting && p.CurrentStamina > 0 {
		accel *= 1.5
		maxSpeed *= 1.5
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
	if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
		if g.Input.IsKeyJustPressed(ebiten.KeyE) {
			g.EnterCave(tx, ty)
			return nil
		}
	} else {
		tile := o.World.OverworldMap[tx][ty]
		if tile == world.TileTrench || tile == world.TileWater || tile == world.TileWreckage {
			// Dive only if not near base station (so E can be used to open terminal)
			if g.Input.IsKeyJustPressed(ebiten.KeyE) && g.baseStation.DistanceToPlayer(p) >= 100.0 {
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
				continue // Out of bounds is NOT solid
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

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			sx := float32(tx*TileSize - int(camX))
			sy := float32(ty*TileSize - int(camY))

			// Clamp coordinates to find the nearest in-bounds tile for base color
			clampedTx := tx
			if clampedTx < 0 {
				clampedTx = 0
			} else if clampedTx >= o.World.Width {
				clampedTx = o.World.Width - 1
			}

			clampedTy := ty
			if clampedTy < 0 {
				clampedTy = 0
			} else if clampedTy >= o.World.Height {
				clampedTy = o.World.Height - 1
			}

			// Distance from boundary (positive if outside, 0 if on boundary, negative if inside)
			var distFromBorder float64
			if tx < 0 || tx >= o.World.Width || ty < 0 || ty >= o.World.Height {
				dx := 0.0
				if tx < 0 {
					dx = float64(-tx)
				} else if tx >= o.World.Width {
					dx = float64(tx - o.World.Width + 1)
				}
				dy := 0.0
				if ty < 0 {
					dy = float64(-ty)
				} else if ty >= o.World.Height {
					dy = float64(ty - o.World.Height + 1)
				}
				distFromBorder = math.Max(dx, dy)
			} else {
				dxInside := float64(tx)
				if float64(o.World.Width-1-tx) < dxInside {
					dxInside = float64(o.World.Width - 1 - tx)
				}
				dyInside := float64(ty)
				if float64(o.World.Height-1-ty) < dyInside {
					dyInside = float64(o.World.Height - 1 - ty)
				}
				distFromBorder = -math.Min(dxInside, dyInside)
			}

			// Transition starts 1 tile inside (-1.0) and finishes 3 tiles outside (3.0)
			t := (distFromBorder + 1.0) / 4.0
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}

			var baseClr color.RGBA
			var baseStrokeClr color.RGBA

			switch o.World.OverworldMap[clampedTx][clampedTy] {
			case world.TileWater:
				dist := o.World.LandDist[clampedTx][clampedTy]
				const maxDist = 15
				lerpT := float64(dist) / float64(maxDist)
				if lerpT > 1.0 {
					lerpT = 1.0
				}
				r := uint8(28 - lerpT*20)
				g := uint8(85 - lerpT*53)
				b := uint8(165 - lerpT*83)
				baseClr = color.RGBA{r, g, b, 255}
				baseStrokeClr = color.RGBA{r + 8, g + 10, b + 15, 255}
			case world.TileLand:
				baseClr = color.RGBA{38, 142, 85, 255}
				baseStrokeClr = color.RGBA{48, 160, 98, 255}
			case world.TileTrench:
				baseClr = color.RGBA{6, 18, 42, 255}
				baseStrokeClr = color.RGBA{10, 26, 58, 255}
			case world.TileWreckage:
				baseClr = color.RGBA{45, 52, 60, 255}
				baseStrokeClr = color.RGBA{110, 80, 50, 255}
			}

			voidClr := color.RGBA{4, 6, 12, 255}
			voidStrokeClr := color.RGBA{8, 12, 20, 255}

			rv := uint8(float64(baseClr.R)*(1.0-t) + float64(voidClr.R)*t)
			gv := uint8(float64(baseClr.G)*(1.0-t) + float64(voidClr.G)*t)
			bv := uint8(float64(baseClr.B)*(1.0-t) + float64(voidClr.B)*t)
			tileClr := color.RGBA{rv, gv, bv, 255}

			srv := uint8(float64(baseStrokeClr.R)*(1.0-t) + float64(voidStrokeClr.R)*t)
			sgv := uint8(float64(baseStrokeClr.G)*(1.0-t) + float64(voidStrokeClr.G)*t)
			sbv := uint8(float64(baseStrokeClr.B)*(1.0-t) + float64(voidStrokeClr.B)*t)
			strokeClr := color.RGBA{srv, sgv, sbv, 255}

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
		if pTileX < 0 || pTileX >= o.World.Width || pTileY < 0 || pTileY >= o.World.Height {
			promptText := "Press [E] to Dive into Void"
			promptX := float32(pCenterX(g.player)) - 95
			promptY := float32(pCenterY(g.player)) - 40
			vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
		} else {
			tile := o.World.OverworldMap[pTileX][pTileY]
			if tile == world.TileTrench || tile == world.TileWater || tile == world.TileWreckage {
				promptText := "Press [E] to Dive"
				if tile == world.TileWreckage {
					promptText = "Press [E] to Salvage Wreckage"
				}
				promptX := float32(pCenterX(g.player)) - 95
				promptY := float32(pCenterY(g.player)) - 40
				vector.FillRect(screen, promptX, promptY, 190, 25, color.RGBA{0, 0, 0, 180}, false)
				ebitenutil.DebugPrintAt(screen, promptText, int(promptX)+10, int(promptY)+4)
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
