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

	// Steering rotation (A/D or Left/Right arrows)
	const turnSpeed = 0.05
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.Facing -= turnSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		p.Facing += turnSpeed
	}

	// Steering speed limits and acceleration (W/S or Up/Down arrows)
	var accel = 0.12
	var maxSpeed = 4.0
	isSprinting := ebiten.IsKeyPressed(ebiten.KeyShift)

	if isSprinting && p.CurrentStamina > 0 {
		accel = 0.25
		maxSpeed = 6.5
	}

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		p.Vx += math.Cos(p.Facing) * accel
		p.Vy += math.Sin(p.Facing) * accel
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		p.Vx -= math.Cos(p.Facing) * (accel * 0.4)
		p.Vy -= math.Sin(p.Facing) * (accel * 0.4)
	}

	// Apply fluid friction (drag) to the boat
	const drag = 0.95
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
	p.UpdateStats(false, isSprinting && isMoving)

	// Check if player is overlapping a Trench tile to trigger dive transition
	tx := int(p.X+p.Width/2) / TileSize
	ty := int(p.Y+p.Height/2) / TileSize
	if tx >= 0 && tx < o.World.Width && ty >= 0 && ty < o.World.Height {
		if o.World.OverworldMap[tx][ty] == world.TileTrench {
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
func (o *OverworldState) Draw(screen *ebiten.Image, camera *Camera) {
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

	// If player is sitting on a Trench, display dive prompt overlay
	pTileX := int(o.Player.X+o.Player.Width/2) / TileSize
	pTileY := int(o.Player.Y+o.Player.Height/2) / TileSize
	if pTileX >= 0 && pTileX < o.World.Width && pTileY >= 0 && pTileY < o.World.Height {
		if o.World.OverworldMap[pTileX][pTileY] == world.TileTrench {
			promptX := float32(pCenterX(o.Player)) - 80
			promptY := float32(pCenterY(o.Player)) - 40
			vector.DrawFilledRect(screen, promptX, promptY, 160, 25, color.RGBA{0, 0, 0, 180}, false)
			ebitenutil.DebugPrintAt(screen, "Press [E] to Dive", int(promptX)+25, int(promptY)+4)
		}
	}

	// Draw the player's boat centered on screen
	pX := float32(pCenterX(o.Player))
	pY := float32(pCenterY(o.Player))
	const size = 16.0

	x1 := pX + float32(math.Cos(o.Player.Facing))*size*1.5
	y1 := pY + float32(math.Sin(o.Player.Facing))*size*1.5
	x2 := pX + float32(math.Cos(o.Player.Facing+math.Pi*0.8))*size
	y2 := pY + float32(math.Sin(o.Player.Facing+math.Pi*0.8))*size
	x3 := pX + float32(math.Cos(o.Player.Facing-math.Pi*0.8))*size
	y3 := pY + float32(math.Sin(o.Player.Facing-math.Pi*0.8))*size

	boatColor := color.RGBA{230, 238, 248, 255}
	drawFilledTriangle(screen, x1, y1, x2, y2, x3, y3, boatColor)

	cabinColor := color.RGBA{45, 90, 120, 255}
	cx1 := pX + float32(math.Cos(o.Player.Facing))*size*0.4
	cy1 := pY + float32(math.Sin(o.Player.Facing))*size*0.4
	cx2 := pX + float32(math.Cos(o.Player.Facing+math.Pi*0.75))*size*0.6
	cy2 := pY + float32(math.Sin(o.Player.Facing+math.Pi*0.75))*size*0.6
	cx3 := pX + float32(math.Cos(o.Player.Facing-math.Pi*0.75))*size*0.6
	cy3 := pY + float32(math.Sin(o.Player.Facing-math.Pi*0.75))*size*0.6
	drawFilledTriangle(screen, cx1, cy1, cx2, cy2, cx3, cy3, cabinColor)
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
