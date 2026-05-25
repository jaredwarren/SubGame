package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// CaveState manages the side-view cave swimming controls, collision, and rendering.
type CaveState struct {
	Player   *Player
	CaveGrid [][]bool
	Nodes    []ResourceNode
}

// NewCaveState creates a new CaveState instance.
func NewCaveState(player *Player) *CaveState {
	return &CaveState{
		Player: player,
		Nodes:  []ResourceNode{},
	}
}

// Update handles player input, side-scroller swimming physics, and checks exit transitions.
// Returns the next State, and true if a state transition occurred.
func (c *CaveState) Update(g *Game) (State, bool) {
	p := c.Player

	// Exit cave back to Overworld if player swims past the surface (Y <= 0)
	if p.Y < -8 {
		return StateOverworld, true
	}

	// --- Phase 5: Mining Strike Mechanic ---
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		camX := p.X - ScreenWidth/2 + p.Width/2
		camY := p.Y - ScreenHeight/2 + p.Height/2
		worldX := camX + float64(mx)
		worldY := camY + float64(my)

		mtx := int(worldX) / TileSize
		mty := int(worldY) / TileSize

		// Find clicked resource node
		for i := range c.Nodes {
			node := &c.Nodes[i]
			if node.Tx == mtx && node.Ty == mty && node.HitsToMine > 0 {
				// Distance check (Reach limit of ~96 pixels = 1.5 tiles)
				px := p.X + p.Width/2
				py := p.Y + p.Height/2
				nx := float64(node.Tx*TileSize + TileSize/2)
				ny := float64(node.Ty*TileSize + TileSize/2)
				dist := math.Hypot(px-nx, py-ny)

				if dist <= 96.0 {
					if node.Type == ResourceAbyssalOre {
						g.MineWarning = "Requires Heavy Mech Drill Arm to harvest"
						g.MineWarningTimer = 120
						continue
					}
					node.HitsToMine--
					if node.HitsToMine <= 0 {
						// Add to inventory
						p.Inventory.AddItem(node.Type.ItemType(), 1)
						// Drop node from active list
						c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
					}
					break // Strike hits only one node per click
				}
			}
		}
	}

	// Flashlight follows the mouse cursor - update player's facing angle
	mx, my := ebiten.CursorPosition()
	dx := float64(mx) - pCenterX(p)
	dy := float64(my) - pCenterY(p)
	p.Facing = math.Atan2(dy, dx)

	// Movement physics settings
	var swimForce = 0.15
	var maxSpeed = 3.5
	const buoyancy = -0.04 // Upward force (drifts up slowly if not swimming down)
	const drag = 0.92      // Drag resistance in water

	isSprinting := ebiten.IsKeyPressed(ebiten.KeyShift)
	if isSprinting && p.CurrentStamina > 0 {
		swimForce = 0.28
		maxSpeed = 5.5
	}

	// Apply Fins upgrade speed boost (35% increase)
	if p.Inventory.HasItem(ItemFins, 1) {
		swimForce *= 1.35
		maxSpeed *= 1.35
	}

	swimming := false

	// Apply movement force based on inputs
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		p.Vy -= swimForce
		swimming = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		p.Vy += swimForce
		swimming = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.Vx -= swimForce
		swimming = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		p.Vx += swimForce
		swimming = true
	}

	// Apply constant buoyancy
	p.Vy += buoyancy

	// Apply water friction (drag)
	p.Vx *= drag
	p.Vy *= drag

	// Speed clamp
	speed := math.Sqrt(p.Vx*p.Vx + p.Vy*p.Vy)
	if speed > maxSpeed {
		p.Vx = (p.Vx / speed) * maxSpeed
		p.Vy = (p.Vy / speed) * maxSpeed
	}

	// AABB Collision check and position updates
	c.checkCollisions()

	// Update stats (is in cave, checks sprinting if keys are pressed and player is moving)
	isMoving := speed > 0.1
	p.UpdateStats(true, isSprinting && isMoving && swimming)

	return StateCave, false
}

// checkCollisions resolves AABB collision with cave walls in X and Y directions.
func (c *CaveState) checkCollisions() {
	p := c.Player

	// X Axis collision
	newX := p.X + p.Vx
	if c.isSolid(newX, p.Y, p.Width, p.Height) {
		p.Vx = 0
	} else {
		p.X = newX
	}

	// Y Axis collision
	newY := p.Y + p.Vy
	if c.isSolid(p.X, newY, p.Width, p.Height) {
		p.Vy = 0
	} else {
		p.Y = newY
	}
}

// isSolid checks if the proposed bounding box overlaps with solid cave tiles.
func (c *CaveState) isSolid(x, y, w, h float64) bool {
	if c.CaveGrid == nil {
		return false
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])

	x1 := int(math.Floor(x)) / TileSize
	x2 := int(math.Floor(x+w)) / TileSize
	y1 := int(math.Floor(y)) / TileSize
	y2 := int(math.Floor(y+h)) / TileSize

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			// Bounds check: side boundaries are solid
			if tx < 0 || tx >= gridW {
				return true
			}
			// Let player swim past top boundary for surface exit
			if ty < 0 {
				continue
			}
			// Bottom boundary is solid
			if ty >= gridH {
				return true
			}
			if c.CaveGrid[tx][ty] {
				return true
			}
		}
	}
	return false
}

// Draw renders the cave scene, solid tiles, and player assets.
func (c *CaveState) Draw(screen *ebiten.Image, camera *Camera, g *Game) {
	// Camera offset from camera controller
	camX := camera.X
	camY := camera.Y

	// Render ocean sky color if camera rises above the cave mouth (Y < 0)
	if camY < 0 {
		skyColor := color.RGBA{14, 52, 115, 255}
		vector.DrawFilledRect(screen, 0, 0, ScreenWidth, float32(-camY), skyColor, false)
	}

	// Cave ambient backdrop color
	caveColor := color.RGBA{10, 8, 16, 255}
	screen.Fill(caveColor)

	if c.CaveGrid != nil {
		gridW := len(c.CaveGrid)
		gridH := len(c.CaveGrid[0])

		// Calculate visible tiles in viewport
		startTileX := int(camX) / TileSize
		endTileX := (int(camX) + ScreenWidth) / TileSize + 1
		startTileY := int(camY) / TileSize
		endTileY := (int(camY) + ScreenHeight) / TileSize + 1

		// Clamp bounds
		if startTileX < 0 {
			startTileX = 0
		}
		if endTileX > gridW {
			endTileX = gridW
		}
		if startTileY < 0 {
			startTileY = 0
		}
		if endTileY > gridH {
			endTileY = gridH
		}

		for tx := startTileX; tx < endTileX; tx++ {
			for ty := startTileY; ty < endTileY; ty++ {
				if c.CaveGrid[tx][ty] {
					sx := float32(tx*TileSize - int(camX))
					sy := float32(ty*TileSize - int(camY))

					// Calculate depth ratio for visual shading (deeper rocks are darker)
					depthRatio := float64(ty) / float64(gridH)
					r := uint8(math.Max(12, 45-30*depthRatio))
					g := uint8(math.Max(10, 40-30*depthRatio))
					b := uint8(math.Max(18, 50-30*depthRatio))

					rockColor := color.RGBA{r, g, b, 255}
					strokeColor := color.RGBA{r + 12, g + 12, b + 12, 255}

					vector.DrawFilledRect(screen, sx, sy, TileSize, TileSize, rockColor, false)
					vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, strokeColor, false)
				}
			}
		}
	}

	// --- Phase 5: Render Resource Nodes ---
	for i := range c.Nodes {
		c.Nodes[i].Draw(screen, camX, camY)
	}

	isPiloting := g.ActiveVehicle != nil

	// Draw player character centered on screen (only if not piloting a vehicle)
	pX := float32(pCenterX(c.Player))
	pY := float32(pCenterY(c.Player))
	facingAngle := c.Player.Facing

	if isPiloting {
		facingAngle = g.ActiveVehicle.(*ScoutSub).Facing // Default or specific vehicle casting, let's keep it robust
		// We can get facing angle based on the type of active vehicle
		if sub, ok := g.ActiveVehicle.(*ScoutSub); ok {
			facingAngle = sub.Facing
		} else if mech, ok := g.ActiveVehicle.(*HeavyMech); ok {
			facingAngle = mech.Facing
		}
	}

	if !isPiloting {
		// Draw yellow oxygen tank on the back (opposite to facing direction)
		tankAngle := facingAngle + math.Pi
		tx := pX + float32(math.Cos(tankAngle))*8
		ty := pY + float32(math.Sin(tankAngle))*8
		tankColor := color.RGBA{240, 220, 50, 255}
		vector.DrawFilledCircle(screen, tx, ty, 6, tankColor, false)

		// Draw the player body
		playerBodyColor := color.RGBA{220, 95, 45, 255}
		vector.DrawFilledCircle(screen, pX, pY, 9, playerBodyColor, false)

		// Draw helmet glass visor pointing in the Facing direction
		vx := pX + float32(math.Cos(facingAngle))*6
		vy := pY + float32(math.Sin(facingAngle))*6
		visorColor := color.RGBA{80, 200, 255, 200}
		vector.DrawFilledCircle(screen, vx, vy, 5, visorColor, false)
	}

	// --- Phase 4: Dynamic Lighting Shader Mask ---
	if LightShader != nil {
		op := &ebiten.DrawRectShaderOptions{}
		
		var sonarSourceX, sonarSourceY float32
		var sonarRadius float32
		if g.SonarTimer > 0 {
			sonarSourceX = float32(g.SonarSourceX - g.camera.X)
			sonarSourceY = float32(g.SonarSourceY - g.camera.Y)
			sonarRadius = float32(g.SonarRadius)
		}

		op.Uniforms = map[string]interface{}{
			"LightSource":    []float32{pX, pY},
			"FlashlightDir":  []float32{float32(math.Cos(facingAngle)), float32(math.Sin(facingAngle))},
			"LightRadius":    float32(360.0),                     // Reach of flashlight
			"ConeHalfAngle":  float32(math.Pi / 7.5),             // ~24 degree half-angle (48 degree beam)
			"PersonalRadius": float32(65.0),                      // Direct glow around player
			"AmbientColor":   c.getAmbientColor(),
			"SonarSource":    []float32{sonarSourceX, sonarSourceY},
			"SonarRadius":    sonarRadius,
		}
		screen.DrawRectShader(ScreenWidth, ScreenHeight, LightShader, op)
	}

	// --- Phase 4: Bioluminescent Highlights (drawn on top of the shader mask) ---
	c.drawBioluminescence(screen, camX, camY)
}

// drawBioluminescence renders glowing flora and spores that remain visible in the pitch dark.
func (c *CaveState) drawBioluminescence(screen *ebiten.Image, camX, camY float64) {
	if c.CaveGrid == nil {
		return
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])

	startTileX := int(camX) / TileSize
	endTileX := (int(camX) + ScreenWidth) / TileSize + 1
	startTileY := int(camY) / TileSize
	endTileY := (int(camY) + ScreenHeight) / TileSize + 1

	if startTileX < 0 {
		startTileX = 0
	}
	if endTileX > gridW {
		endTileX = gridW
	}
	if startTileY < 0 {
		startTileY = 0
	}
	if endTileY > gridH {
		endTileY = gridH
	}

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.CaveGrid[tx][ty] {
				// Deterministic pseudo-random pattern for placing glowing spores
				hash := (tx*31 + ty*17) % 17
				if hash == 0 {
					sx := float32(tx*TileSize-int(camX)) + float32(TileSize)/2.0
					sy := float32(ty*TileSize-int(camY)) + float32(TileSize)/2.0

					// Alternate colors for variety
					var glowColor color.RGBA
					if (tx+ty)%2 == 0 {
						glowColor = color.RGBA{0, 245, 210, 255} // Bioluminescent cyan
					} else {
						glowColor = color.RGBA{245, 75, 140, 255} // Bioluminescent hot pink
					}

					// Draw soft outer light emission
					vector.DrawFilledCircle(screen, sx, sy, 5.0, color.RGBA{glowColor.R, glowColor.G, glowColor.B, 70}, false)
					// Draw hot white central core
					vector.DrawFilledCircle(screen, sx, sy, 1.5, color.RGBA{255, 255, 255, 255}, false)
				}
			}
		}
	}
}

func (c *CaveState) getAmbientColor() []float32 {
	if LightCaveForDebug {
		return []float32{0.02, 0.02, 0.03, 0.15} // Light translucency for debugging (85% visibility)
	}
	return []float32{0.01, 0.01, 0.03, 0.97} // Crushing deep-sea darkness mask
}
