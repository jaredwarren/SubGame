package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/world"
)

// OverworldState manages the top-down surface sailing view.
type OverworldState struct {
	Player *Player
	World  *world.World
}

// NewOverworldState creates a new OverworldState.
func NewOverworldState(player *Player, world *world.World) *OverworldState {
	return &OverworldState{
		Player: player,
		World:  world,
	}
}

// Update handles input, movement physics, and checks state transition triggers.
// Returns the next State, and true if a state transition should occur.
func (o *OverworldState) Update() (State, bool) {
	p := o.Player

	// On foot swimming in Overworld
	var accel = 0.08
	var maxSpeed = 1.6
	isSprinting := ebiten.IsKeyPressed(ebiten.KeyShift)

	if isSprinting && p.CurrentStamina > 0 {
		accel = 0.16
		maxSpeed = 2.6
	}

	// Apply Fins upgrade speed boost (30% increase)
	if p.Inventory.HasItem(ItemFins, 1) {
		accel *= 1.30
		maxSpeed *= 1.30
	}

	// Direct WASD movement (no steering inertia when swimming on foot)
	moving := false
	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		dy -= 1.0
		moving = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		dy += 1.0
		moving = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1.0
		moving = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1.0
		moving = true
	}

	if moving {
		angle := math.Atan2(dy, dx)
		p.Facing = angle // Visor faces swim direction
		p.Vx += math.Cos(angle) * accel
		p.Vy += math.Sin(angle) * accel
	}

	// Apply fluid friction (drag) to the swimming player
	const drag = 0.88
	p.Vx *= drag
	p.Vy *= drag

	// Speed clamp
	speed := math.Sqrt(p.Vx*p.Vx + p.Vy*p.Vy)
	if speed > maxSpeed {
		p.Vx = (p.Vx / speed) * maxSpeed
		p.Vy = (p.Vy / speed) * maxSpeed
	}

	// AABB Collision check and position updates
	o.checkCollisions()

	// Update stats (not in cave, check if sprinting and actually moving)
	isMoving := speed > 0.1
	p.UpdateStats(false, isSprinting && isMoving && moving)

	// Check if player is overlapping a Trench or Water tile to trigger dive transition
	tx := int(p.X+p.Width/2) / TileSize
	ty := int(p.Y+p.Height/2) / TileSize
	if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
		tile := o.World.OverworldMap[tx][ty]
		if tile == world.TileTrench || tile == world.TileWater {
			if ebiten.IsKeyPressed(ebiten.KeyE) {
				return StateCave, true
			}
		}
	}

	return StateOverworld, false
}

// checkCollisions resolves AABB collision with land tiles in X and Y directions.
func (o *OverworldState) checkCollisions() {
	p := o.Player

	// X Axis collision
	newX := p.X + p.Vx
	if o.isSolid(newX, p.Y, p.Width, p.Height) {
		p.Vx = 0
	} else {
		p.X = newX
	}

	// Y Axis collision
	newY := p.Y + p.Vy
	if o.isSolid(p.X, newY, p.Width, p.Height) {
		p.Vy = 0
	} else {
		p.Y = newY
	}
}

// isSolid checks if the proposed bounding box overlaps with solid land.
func (o *OverworldState) isSolid(x, y, w, h float64) bool {
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
func (o *OverworldState) Draw(screen *ebiten.Image, camera *Camera, isPiloting bool) {
	// Camera offset from camera controller
	camX := camera.X
	camY := camera.Y

	// Render tiles that fall within screen viewport bounds
	startTileX := int(camX) / TileSize
	endTileX := (int(camX) + ScreenWidth) / TileSize + 1
	startTileY := int(camY) / TileSize
	endTileY := (int(camY) + ScreenHeight) / TileSize + 1

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

			var tileClr color.Color
			var strokeClr color.Color

			switch o.World.OverworldMap[tx][ty] {
			case world.TileWater:
				tileClr = color.RGBA{14, 52, 115, 255}
				strokeClr = color.RGBA{22, 64, 135, 255}
			case world.TileLand:
				tileClr = color.RGBA{38, 142, 85, 255} // Reef green
				strokeClr = color.RGBA{48, 160, 98, 255}
			case world.TileTrench:
				tileClr = color.RGBA{6, 18, 42, 255} // Deep sinkhole
				strokeClr = color.RGBA{10, 26, 58, 255}
			}

			vector.DrawFilledRect(screen, sx, sy, TileSize, TileSize, tileClr, false)
			vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, strokeClr, false)
		}
	}

	// If player is sitting on a Trench, display dive prompt overlay (only if not piloting a vehicle)
	if !isPiloting {
		pTileX := int(o.Player.X+o.Player.Width/2) / TileSize
		pTileY := int(o.Player.Y+o.Player.Height/2) / TileSize
		if pTileX >= 0 && pTileX < o.World.Width && pTileY >= 0 && pTileY < o.World.Height {
			tile := o.World.OverworldMap[pTileX][pTileY]
			if tile == world.TileTrench || tile == world.TileWater {
				promptX := float32(pCenterX(o.Player)) - 80
				promptY := float32(pCenterY(o.Player)) - 40
				vector.DrawFilledRect(screen, promptX, promptY, 160, 25, color.RGBA{0, 0, 0, 180}, false)
				ebitenutil.DebugPrintAt(screen, "Press [E] to Dive", int(promptX)+25, int(promptY)+4)
			}
		}

		// Draw the player swimming as a small diver circle centered on screen
		pX := float32(pCenterX(o.Player))
		pY := float32(pCenterY(o.Player))

		// Player body circle
		vector.DrawFilledCircle(screen, pX, pY, 8.0, color.RGBA{220, 95, 45, 255}, false)
		// Helmet glass pointing in Facing direction
		vx := pX + float32(math.Cos(o.Player.Facing))*5
		vy := pY + float32(math.Sin(o.Player.Facing))*5
		vector.DrawFilledCircle(screen, vx, vy, 4.0, color.RGBA{80, 200, 255, 200}, false)
	}
}

// Global empty white image for custom vector triangle rendering
var emptyImage = ebiten.NewImage(3, 3)

func init() {
	emptyImage.Fill(color.White)
}

// drawFilledTriangle fills a 2D triangle using Ebitengine DrawTriangles.
func drawFilledTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, clr color.Color) {
	r, g, b, a := clr.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff

	vertices := []ebiten.Vertex{
		{DstX: x1, DstY: y1, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af},
		{DstX: x2, DstY: y2, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af},
		{DstX: x3, DstY: y3, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af},
	}
	indices := []uint16{0, 1, 2}
	screen.DrawTriangles(vertices, indices, emptyImage, nil)
}
