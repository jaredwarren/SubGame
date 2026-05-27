package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// CaveScene manages the side-view cave swimming controls, collision, and rendering.
type CaveScene struct {
	CaveGrid  [][]bool
	Nodes     []resource.Resource
	Entities  []CaveEntity
	IsShallow bool

	// Pre-allocated Draw parameters to prevent allocations in the hot path
	shaderOpts    ebiten.DrawRectShaderOptions
	uniforms      map[string]any
	lightSource   []float32
	flashlightDir []float32
	sonarSource   []float32
	entranceLight []float32
}

// NewCaveScene creates a new CaveScene instance.
func NewCaveScene() *CaveScene {
	cs := &CaveScene{
		Nodes:         []resource.Resource{},
		Entities:      []CaveEntity{},
		uniforms:      make(map[string]any),
		lightSource:   make([]float32, 2),
		flashlightDir: make([]float32, 2),
		sonarSource:   make([]float32, 2),
		entranceLight: make([]float32, 2),
	}
	cs.shaderOpts.Uniforms = cs.uniforms
	return cs
}

func (c *CaveScene) OnEnter(g *Game) {
	g.currentState = StateCave
}

func (c *CaveScene) OnExit(g *Game) {}

// Update handles player input, side-scroller swimming physics, and checks exit transitions.
func (c *CaveScene) Update(g *Game) error {
	// Reset electrical tracking timer; active Electro-Weavers will update it
	g.WeaverTrackingTimer = 0.0

	// Update cave entities
	for _, ent := range c.Entities {
		ent.Update(g, c)
	}

	// Clean up deactivated entities
	activeCount := 0
	for _, ent := range c.Entities {
		if ent.IsActive() {
			c.Entities[activeCount] = ent
			activeCount++
		}
	}
	c.Entities = c.Entities[:activeCount]
	g.caveEntities[g.activeTrenchKey] = c.Entities

	if g.ActiveVehicle != nil {
		// Heavy Mech drilling handler
		if mech, ok := g.ActiveVehicle.(*vehicle.HeavyMech); ok && !mech.IsDrilling {
			if g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				cursor := g.Input.Cursor()
				camX := mech.Pos.X - ScreenWidth/2 + mech.Dimensions.X/2
				camY := mech.Pos.Y - ScreenHeight/2 + mech.Dimensions.Y/2
				worldX := camX + cursor.X
				worldY := camY + cursor.Y

				mtx := int(worldX) / TileSize
				mty := int(worldY) / TileSize

				for i := 0; i < len(c.Nodes); i++ {
					node := c.Nodes[i]
					nodeTx, nodeTy := node.GetTilePos()
					if nodeTx == mtx && nodeTy == mty && node.GetHitsToMine() > 0 {
						px := mech.Pos.X + mech.Dimensions.X/2
						py := mech.Pos.Y + mech.Dimensions.Y/2
						nx := float64(nodeTx*TileSize + TileSize/2)
						ny := float64(nodeTy*TileSize + TileSize/2)
						dist := math.Hypot(px-nx, py-ny)

						if dist <= 120.0 {
							mech.DrillStrike(node)
							break
						}
					}
				}
			}
		}
		return nil
	}

	p := g.player

	// Exit cave back to Overworld if player swims past the surface (Y <= -8)
	if p.Pos.Y < -8 {
		g.ExitCave()
		return nil
	}

	// --- Phase 5: Mining Strike Mechanic ---
	if g.Input.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := g.Input.Cursor()
		camX := p.Pos.X - ScreenWidth/2 + p.Width/2
		camY := p.Pos.Y - ScreenHeight/2 + p.Height/2
		worldX := camX + cursor.X
		worldY := camY + cursor.Y

		mtx := int(worldX) / TileSize
		mty := int(worldY) / TileSize

		// Find clicked Shatter-bulb
		for _, ent := range c.Entities {
			if ent.GetType() == EntShatterBulb && ent.IsActive() {
				pos := ent.GetPos()
				dims := ent.GetDimensions()
				if worldX >= pos.X && worldX < pos.X+dims.X && worldY >= pos.Y && worldY < pos.Y+dims.Y {
					px := p.Pos.X + p.Width/2
					py := p.Pos.Y + p.Height/2
					dist := math.Hypot(px-(pos.X+dims.X/2), py-(pos.Y+dims.Y/2))
					if dist <= 96.0 {
						if bulb, ok := ent.(*ShatterBulb); ok {
							bulb.Pop(g, c)
						}
						break
					}
				}
			}
		}

		// Find clicked resource node
		for i := 0; i < len(c.Nodes); i++ {
			node := c.Nodes[i]
			nodeTx, nodeTy := node.GetTilePos()
			if nodeTx == mtx && nodeTy == mty && node.GetHitsToMine() > 0 {
				// Distance check (Reach limit of ~96 pixels = 1.5 tiles)
				px := p.Pos.X + p.Width/2
				py := p.Pos.Y + p.Height/2
				nx := float64(nodeTx*TileSize + TileSize/2)
				ny := float64(nodeTy*TileSize + TileSize/2)
				dist := math.Hypot(px-nx, py-ny)

				if dist <= 96.0 {
					if node.RequiresMech() {
						g.MineWarning = "Requires Heavy Mech Drill Arm to harvest"
						g.MineWarningTimer = 120
						continue
					}
					node.SetHitsToMine(node.GetHitsToMine() - 1)

					// Spawn mineral debris particles
					nodeColor := color.RGBA{150, 150, 150, 255}
					if cRgba, ok := node.GetColor().(color.RGBA); ok {
						nodeColor = cRgba
					}
					g.SpawnDebris(nx, ny, nodeColor)

					if node.GetHitsToMine() <= 0 {
						// Add to inventory
						p.Inventory.AddItem(node, 1)
						// Drop node from active list
						c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
						i--
					}
					break // Strike hits only one node per click
				}
			}
		}
	}

	// Flashlight follows the mouse cursor - update player's facing angle
	cursor := g.Input.Cursor()
	dx := cursor.X - pCenterX(p)
	dy := cursor.Y - pCenterY(p)
	p.Facing = math.Atan2(dy, dx)

	// Movement physics settings
	var swimForce = 0.15
	var maxSpeed = 3.5
	const buoyancy = -0.04 // Upward force (drifts up slowly if not swimming down)
	const drag = 0.92      // Drag resistance in water

	isSprinting := g.Input.IsKeyPressed(ebiten.KeyShift)
	if isSprinting && p.CurrentStamina > 0 {
		swimForce = 0.28
		maxSpeed = 5.5
	}

	// Apply Fins upgrade speed boost (35% increase)
	if p.HasFins {
		swimForce *= 1.35
		maxSpeed *= 1.35
	}

	// Apply Nerve-Mat slow debuff (50% reduction)
	if g.playerSlowed {
		swimForce *= 0.5
		maxSpeed *= 0.5
	}

	swimming := false

	// Apply movement force based on inputs
	if g.Input.IsKeyPressed(ebiten.KeyW) || g.Input.IsKeyPressed(ebiten.KeyArrowUp) {
		p.Vel.Y -= swimForce
		swimming = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyS) || g.Input.IsKeyPressed(ebiten.KeyArrowDown) {
		p.Vel.Y += swimForce
		swimming = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyA) || g.Input.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.Vel.X -= swimForce
		swimming = true
	}
	if g.Input.IsKeyPressed(ebiten.KeyD) || g.Input.IsKeyPressed(ebiten.KeyArrowRight) {
		p.Vel.X += swimForce
		swimming = true
	}

	// Apply constant buoyancy
	p.Vel.Y += buoyancy

	// Apply water friction (drag)
	p.Vel = p.Vel.Scale(drag)

	// Speed clamp
	speed := p.Vel.Length()
	if speed > maxSpeed {
		p.Vel = p.Vel.Scale(maxSpeed / speed)
	}

	// Emit bubbles occasionally while swimming fast
	if swimming && speed > 2.0 && rand.Float64() < 0.12 {
		g.SpawnBubble(p.Pos.X+p.Width/2.0, p.Pos.Y+p.Height/2.0)
	}

	// AABB Collision check and position updates
	c.checkCollisions(p)

	// Update stats (is in cave, checks sprinting if keys are pressed and player is moving)
	isMoving := speed > 0.1
	p.UpdateStats(true, isSprinting && isMoving && swimming)

	return nil
}

// checkCollisions resolves AABB collision with cave walls in X and Y directions.
func (c *CaveScene) checkCollisions(p *Player) {
	// X Axis collision
	newX := p.Pos.X + p.Vel.X
	if c.isSolid(newX, p.Pos.Y, p.Width, p.Height) {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}

	// Y Axis collision
	newY := p.Pos.Y + p.Vel.Y
	if c.isSolid(p.Pos.X, newY, p.Width, p.Height) {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}

// isSolid checks if the proposed bounding box overlaps with solid cave tiles.
func (c *CaveScene) isSolid(x, y, w, h float64) bool {
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

// getSkyColor returns a sky color appropriate for the given time-of-day tick,
// interpolating between night, dawn, day, and dusk phases.
func getSkyColor(timeOfDay float64) color.RGBA {
	// Key colors
	nightSky := [3]float64{20, 30, 70}  // Deep dark blue
	dawnSky := [3]float64{255, 160, 80} // Warm orange sunrise
	daySky := [3]float64{140, 200, 255} // Bright sky blue (daytime)
	duskSky := [3]float64{220, 100, 60} // Warm orange sunset

	lerpF := func(a, b [3]float64, t float64) color.RGBA {
		return color.RGBA{
			R: uint8(a[0] + (b[0]-a[0])*t),
			G: uint8(a[1] + (b[1]-a[1])*t),
			B: uint8(a[2] + (b[2]-a[2])*t),
			A: 255,
		}
	}

	switch {
	case timeOfDay < 1200: // Dawn: night → day
		return lerpF(nightSky, dawnSky, timeOfDay/1200.0)
	case timeOfDay < 2400: // Post-dawn: dawn → day
		return lerpF(dawnSky, daySky, (timeOfDay-1200.0)/1200.0)
	case timeOfDay < 6000: // Full day
		return lerpF(daySky, daySky, 0)
	case timeOfDay < 7200: // Dusk: day → dusk
		return lerpF(daySky, duskSky, (timeOfDay-6000.0)/1200.0)
	case timeOfDay < 8400: // Post-dusk: dusk → night
		return lerpF(duskSky, nightSky, (timeOfDay-7200.0)/1200.0)
	default: // Night
		return lerpF(nightSky, nightSky, 0)
	}
}

// Draw renders the cave scene, solid tiles, and player assets.
func (c *CaveScene) Draw(g *Game, screen *ebiten.Image) {
	// Camera offset from camera controller
	camX := g.camera.Pos.X
	camY := g.camera.Pos.Y

	// Cave ambient backdrop — fill first so sky can paint over it
	if c.IsShallow {
		// Draw a depth gradient: bright near the surface, darker at the floor.
		// The maximum darkness at the floor is time-of-day dependent.
		mult := GetOverworldLightMultiplier(g.TimeOfDay)

		// Surface base color
		baseR := float64(10) + float64(30)*mult  // 10 (night) → 40 (day)
		baseG := float64(40) + float64(80)*mult  // 40 (night) → 120 (day)
		baseB := float64(100) + float64(80)*mult // 100 (night) → 180 (day)

		// Max fraction of brightness to remove at the floor:
		// full day → 0.45 (still reasonably lit), full night → 0.90 (near total dark)
		maxDarken := 0.45 + (1.0-mult)*0.45

		// Cave floor depth in world pixels
		maxDepth := float64(30 * TileSize) // fallback
		if len(c.CaveGrid) > 0 {
			maxDepth = float64(len(c.CaveGrid[0]) * TileSize)
		}

		// Draw horizontal strips, each slightly darker than the one above
		const stripH = float32(6)
		for sy := float32(0); sy < float32(ScreenHeight); sy += stripH {
			worldY := camY + float64(sy)
			depthFrac := 0.0
			if worldY > 0 {
				depthFrac = worldY / maxDepth
				if depthFrac > 1 {
					depthFrac = 1
				}
			}
			darkFactor := 1.0 - depthFrac*maxDarken
			sc := color.RGBA{
				R: uint8(baseR * darkFactor),
				G: uint8(baseG * darkFactor),
				B: uint8(baseB * darkFactor),
				A: 255,
			}
			vector.FillRect(screen, 0, sy, float32(ScreenWidth), stripH, sc, false)
		}
	} else {
		screen.Fill(color.RGBA{10, 8, 16, 255})
	}

	// Render sky color above the water surface (camY < 0 means camera is above Y=0)
	if camY < 0 {
		var skyColor color.RGBA
		if c.IsShallow {
			skyColor = getSkyColor(g.TimeOfDay)
		} else {
			skyColor = color.RGBA{14, 52, 115, 255} // Darker deep water for deep caves
		}
		vector.FillRect(screen, 0, 0, ScreenWidth, float32(-camY), skyColor, false)
	}

	// Draw a distinct line at the water surface boundary (Y = 0) if visible
	surfaceY := float32(-camY)
	if surfaceY >= 0 && surfaceY < float32(ScreenHeight) {
		lineColor := color.RGBA{220, 240, 255, 255} // foam white
		if !c.IsShallow {
			lineColor = color.RGBA{30, 80, 160, 255} // dark water surface for deep cave
		}
		vector.StrokeLine(screen, 0, surfaceY, ScreenWidth, surfaceY, 3.0, lineColor, false)
	}

	if c.CaveGrid != nil {
		gridW := len(c.CaveGrid)
		gridH := len(c.CaveGrid[0])

		// Calculate visible tiles in viewport
		startTileX := int(camX) / TileSize
		endTileX := (int(camX)+ScreenWidth)/TileSize + 1
		startTileY := int(camY) / TileSize
		endTileY := (int(camY)+ScreenHeight)/TileSize + 1

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

					var rockColor, strokeColor color.RGBA
					if c.IsShallow {
						// Sandy reef rock
						rockColor = color.RGBA{180, 155, 100, 255}
						strokeColor = color.RGBA{210, 185, 120, 255}
					} else if ty < 40 {
						// Biome 1: Mid-Depth (Cyan/Teal) - Luminous Pneumatophore Grotto
						bandRatio := float64(ty) / 40.0
						r := uint8(math.Max(8, 22-14*bandRatio))
						g := uint8(math.Max(24, 64-40*bandRatio))
						b := uint8(math.Max(32, 78-46*bandRatio))
						rockColor = color.RGBA{r, g, b, 255}
						strokeColor = color.RGBA{r + 20, g + 40, b + 48, 255}
					} else if ty < 80 {
						// Biome 2: Deep (Dark Grey/Orange) - Silicate Smoker Trenches
						bandRatio := float64(ty-40) / 40.0
						r := uint8(math.Max(25, 45-20*bandRatio))
						g := uint8(math.Max(20, 32-12*bandRatio))
						b := uint8(math.Max(18, 26-8*bandRatio))
						rockColor = color.RGBA{r, g, b, 255}
						strokeColor = color.RGBA{uint8(math.Max(80, 150-70*bandRatio)), 65, 40, 255}
					} else {
						// Biome 3: Abyssal (Vantablack/White) - Benthic Brine-Falls
						rockColor = color.RGBA{5, 5, 8, 255}
						strokeColor = color.RGBA{210, 210, 220, 255}
					}

					vector.FillRect(screen, sx, sy, TileSize, TileSize, rockColor, false)
					vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, strokeColor, false)
				}
			}
		}
	}

	// --- Phase 5: Render Resource Nodes ---
	for _, node := range c.Nodes {
		node.Draw(screen, camX, camY)
	}

	isPiloting := g.ActiveVehicle != nil

	// Draw player character centered on screen (only if not piloting a vehicle)
	pX := float32(pCenterX(g.player))
	pY := float32(pCenterY(g.player))
	facingAngle := g.player.Facing

	if isPiloting {
		facingAngle = g.ActiveVehicle.GetFacing()
	}

	if !isPiloting {
		// Draw yellow oxygen tank on the back (opposite to facing direction)
		tankAngle := facingAngle + math.Pi
		tx := pX + float32(math.Cos(tankAngle))*8
		ty := pY + float32(math.Sin(tankAngle))*8
		tankColor := color.RGBA{240, 220, 50, 255}
		vector.FillCircle(screen, tx, ty, 6, tankColor, false)

		// Draw the player body
		playerBodyColor := color.RGBA{220, 95, 45, 255}
		vector.FillCircle(screen, pX, pY, 9, playerBodyColor, false)

		// Draw helmet glass visor pointing in the Facing direction
		vx := pX + float32(math.Cos(facingAngle))*6
		vy := pY + float32(math.Sin(facingAngle))*6
		visorColor := color.RGBA{80, 200, 255, 200}
		vector.FillCircle(screen, vx, vy, 5, visorColor, false)
	}

	// --- Phase 4: Dynamic Lighting Shader Mask ---
	if LightShader != nil && !c.IsShallow {
		var sonarSourceX, sonarSourceY float32
		var sonarRadius float32
		if g.Sonar.Timer > 0 {
			sonarSourceX = float32(g.Sonar.SourceX - g.camera.Pos.X)
			sonarSourceY = float32(g.Sonar.SourceY - g.camera.Pos.Y)
			sonarRadius = float32(g.Sonar.Radius)
		}

		var fDirX, fDirY float32
		if g.FlashlightOn {
			fDirX = float32(math.Cos(facingAngle))
			fDirY = float32(math.Sin(facingAngle))

			// Flickering effect based on Electro-Weaver tracking intensity
			if g.WeaverTrackingTimer > 0 {
				flickerChance := (g.WeaverTrackingTimer / 300.0) * 0.20 // up to 20% flicker chance
				if rand.Float64() < flickerChance {
					fDirX, fDirY = 0, 0
				}
			}
		}

		// Calculate screen-space coordinates of cave entrance
		entranceX := float32(float64(len(c.CaveGrid)/2*TileSize) + TileSize/2.0 - g.camera.Pos.X)
		entranceY := float32(0.0 - g.camera.Pos.Y)

		// Update pre-allocated uniforms and arrays to avoid heap allocation
		c.lightSource[0], c.lightSource[1] = pX, pY
		c.flashlightDir[0], c.flashlightDir[1] = fDirX, fDirY
		c.sonarSource[0], c.sonarSource[1] = sonarSourceX, sonarSourceY
		c.entranceLight[0], c.entranceLight[1] = entranceX, entranceY

		c.uniforms["LightSource"] = c.lightSource
		c.uniforms["FlashlightDir"] = c.flashlightDir
		c.uniforms["LightRadius"] = float32(360.0)
		c.uniforms["ConeHalfAngle"] = float32(math.Pi / 7.5)
		c.uniforms["PersonalRadius"] = float32(65.0)
		c.uniforms["AmbientColor"] = c.getAmbientColor(c.IsShallow, g.TimeOfDay)
		c.uniforms["SonarSource"] = c.sonarSource
		c.uniforms["SonarRadius"] = sonarRadius
		sonarBright := float32(1.0)
		sonarFadeLimit := float32(1200.0)
		if g.Sonar.Bright {
			sonarBright = 2.5
			sonarFadeLimit = 3000.0
		}
		c.uniforms["SonarBright"] = sonarBright
		c.uniforms["SonarFadeLimit"] = sonarFadeLimit
		c.uniforms["EntranceLight"] = c.entranceLight
		c.uniforms["EntranceActive"] = float32(1.0)

		screen.DrawRectShader(ScreenWidth, ScreenHeight, LightShader, &c.shaderOpts)
	}

	// --- Phase 4: Bioluminescent Highlights (drawn on top of the shader mask) ---
	if !c.IsShallow {
		c.drawBioluminescence(screen, camX, camY)
	}

	// --- Phase 8: Render Biome Entities ---
	for _, ent := range c.Entities {
		ent.Draw(screen, g.camera, g.TimeOfDay)
	}

	// Draw popped bulb sound wave ripple circle
	if g.SoundWaveTimer > 0 {
		alpha := float32(g.SoundWaveTimer) / 60.0
		clr := color.RGBA{245, 120, 20, uint8(200 * alpha)}
		vector.StrokeCircle(screen, float32(g.SoundWaveX-g.camera.Pos.X), float32(g.SoundWaveY-g.camera.Pos.Y), float32(g.SoundWaveRadius), 2.0, clr, false)
	}
}

// drawBioluminescence renders glowing flora and spores that remain visible in the pitch dark.
func (c *CaveScene) drawBioluminescence(screen *ebiten.Image, camX, camY float64) {
	if c.CaveGrid == nil {
		return
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])

	startTileX := int(camX) / TileSize
	endTileX := (int(camX)+ScreenWidth)/TileSize + 1
	startTileY := int(camY) / TileSize
	endTileY := (int(camY)+ScreenHeight)/TileSize + 1

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
				hash := (tx*31 + ty*17) % 17
				if hash == 0 {
					sx := float32(tx*TileSize-int(camX)) + float32(TileSize)/2.0
					sy := float32(ty*TileSize-int(camY)) + float32(TileSize)/2.0

					var glowColor color.RGBA
					if (tx+ty)%2 == 0 {
						glowColor = color.RGBA{0, 245, 210, 255} // Bioluminescent cyan
					} else {
						glowColor = color.RGBA{245, 75, 140, 255} // Bioluminescent hot pink
					}

					// Draw soft outer light emission
					vector.FillCircle(screen, sx, sy, 5.0, color.RGBA{glowColor.R, glowColor.G, glowColor.B, 70}, false)
					// Draw hot white central core
					vector.FillCircle(screen, sx, sy, 1.5, color.RGBA{255, 255, 255, 255}, false)
				}
			}
		}
	}
}

func (c *CaveScene) getAmbientColor(isShallow bool, timeOfDay float64) []float32 {
	if LightCaveForDebug {
		return []float32{0.02, 0.02, 0.03, 0.15} // Light translucency for debugging (85% visibility)
	}
	if isShallow {
		// Ambient mask: low alpha = bright (day), high alpha = dark (night).
		// Use the overworld light multiplier: 1.0 = full day, 0.2 = deepest night.
		mult := GetOverworldLightMultiplier(timeOfDay)
		// Map mult [0.2, 1.0] -> alpha [0.75, 0.15]  (night darker, day bright)
		alpha := float32(0.75 - (mult-0.2)/0.8*0.60)
		return []float32{0.04, 0.06, 0.12, alpha}
	}
	return []float32{0.01, 0.01, 0.03, 0.97} // Crushing deep-sea darkness mask
}
