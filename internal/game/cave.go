package game

import (
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// -----------------------------------------------------------------------------
// DIVER ANIMATION SPRITESHEET CONFIGURATION
// -----------------------------------------------------------------------------
// Adjust these variables to manually configure how the diver spritesheet is
// sliced, scaled, and offset.
// -----------------------------------------------------------------------------

// DiverDrawWidth defines the targeted width of the diver sprite on screen.
// Decrease this value if the diver is too big; increase to make it larger.
var DiverDrawWidth = 36.0

// Slicing parameters:
// By default, Rows 0, 1, and 3 are divided into an 8-column grid.
// Row 2 (Mining) is divided into a 6-column grid because the frames are wider (including the tool).
var (
	DiverGridCols     = 8 // Columns in standard rows (0, 1, 3)
	DiverGridRows     = 4 // Total rows in the spritesheet
	DiverMineGridCols = 6 // Columns in the mining row (Row 2)
)

// Column mappings:
// Map each animation frame to a specific column index on the sheet (0-indexed).
var (
	DiverIdleCols   = []int{0, 1, 2, 3}             // Row 0: 4 frames
	DiverSwimCols   = []int{0, 1, 2, 3, 4, 5, 6, 7} // Row 1: 8 frames
	DiverMineCols   = []int{0, 1, 2, 3}             // Row 2: 4 frames inside 6-col grid
	DiverDamageCols = []int{0}                      // Row 3: 1 frame
)

// Slicing offsets:
// Adjust these if you need to offset the entire grid horizontally or vertically.
var (
	DiverXOffset = 0 // Shift all slices horizontally by this many pixels
	DiverYOffset = 0 // Shift all slices vertically by this many pixels
)

// CaveScene manages the side-view cave swimming controls, collision, and rendering.
type CaveScene struct {
	ActiveCave Cave
	CaveGrid   [][]bool
	Nodes      []resource.Resource
	Entities   []CaveEntity
	IsShallow  bool

	// Pre-allocated Draw parameters to prevent allocations in the hot path
	shaderOpts    ebiten.DrawRectShaderOptions
	uniforms      map[string]any
	lightSource   []float32
	flashlightDir []float32
	sonarSource   []float32
	entranceLight []float32

	offscreen *ebiten.Image

	// Spritesheet animations
	diverSheet       *ebiten.Image
	diverIdleFrames  []*ebiten.Image
	diverSwimFrames  []*ebiten.Image
	diverMineFrames  []*ebiten.Image
	diverDamageFrame *ebiten.Image
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
	cs.loadDiverSheet()
	return cs
}

// loadDiverSheet searches for assets/textures/diver_sheet.png, removes the green background, and slices it into animation frames.
func (c *CaveScene) loadDiverSheet() {
	paths := []string{
		"assets/textures/diver_sheet.png",
		"/Users/jaredwarren/src/github.com/jaredwarren/SubGame/assets/textures/diver_sheet.png",
		"../../assets/textures/diver_sheet.png",
		"../assets/textures/diver_sheet.png",
	}

	var file *os.File
	var err error
	for _, p := range paths {
		file, err = os.Open(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Printf("Warning: Failed to open assets/textures/diver_sheet.png: %v", err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Warning: Failed to decode assets/textures/diver_sheet.png: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	// Chroma keying: set green background to transparent (Alpha 0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			clr := img.At(x, y)
			r, g, b, a := clr.RGBA()
			ru := uint8(r >> 8)
			gu := uint8(g >> 8)
			bu := uint8(b >> 8)
			au := uint8(a >> 8)

			// Green background keying: dominant green channel, low red and blue channels
			if gu > 140 && ru < 100 && bu < 100 {
				rgba.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				rgba.SetRGBA(x, y, color.RGBA{ru, gu, bu, au})
			}
		}
	}

	sheet := ebiten.NewImageFromImage(rgba)
	c.diverSheet = sheet

	// Dynamically calculate default frame sizes based on grid configuration
	frameW := bounds.Dx() / DiverGridCols
	frameH := bounds.Dy() / DiverGridRows

	// -------------------------------------------------------------------------
	// OPTION A: Config-driven Slicing (Default)
	// Uses the variables defined at the top of this file.
	// -------------------------------------------------------------------------

	// Row 0: Idle
	for _, col := range DiverIdleCols {
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+0, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH)
		c.diverIdleFrames = append(c.diverIdleFrames, sheet.SubImage(rect).(*ebiten.Image))
	}

	// Row 1: Swim
	for _, col := range DiverSwimCols {
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+frameH, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH*2)
		c.diverSwimFrames = append(c.diverSwimFrames, sheet.SubImage(rect).(*ebiten.Image))
	}

	// Row 2: Mine
	mineFrameW := bounds.Dx() / DiverMineGridCols
	for _, col := range DiverMineCols {
		rect := image.Rect(DiverXOffset+col*mineFrameW, DiverYOffset+frameH*2, DiverXOffset+(col+1)*mineFrameW, DiverYOffset+frameH*3)
		c.diverMineFrames = append(c.diverMineFrames, sheet.SubImage(rect).(*ebiten.Image))
	}

	// Row 3: Damage
	if len(DiverDamageCols) > 0 {
		col := DiverDamageCols[0]
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+frameH*3, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH*4)
		c.diverDamageFrame = sheet.SubImage(rect).(*ebiten.Image)
	}

	// -------------------------------------------------------------------------
	// OPTION B: Manual Frame Rectangles (Pixel-Perfect Override)
	// If the sheet columns are not uniform (e.g. have gaps or spacing),
	// uncomment the block below and specify the exact pixel coordinate boxes.
	// -------------------------------------------------------------------------
	/*
		// Clear default slices first
		c.diverIdleFrames = c.diverIdleFrames[:0]
		c.diverSwimFrames = c.diverSwimFrames[:0]
		c.diverMineFrames = c.diverMineFrames[:0]

		// 1. Row 0: Idle (4 frames) - Set image.Rect(x1, y1, x2, y2)
		c.diverIdleFrames = []*ebiten.Image{
			sheet.SubImage(image.Rect(0, 0, 352, 384)).(*ebiten.Image),
			sheet.SubImage(image.Rect(352, 0, 704, 384)).(*ebiten.Image),
			sheet.SubImage(image.Rect(704, 0, 1056, 384)).(*ebiten.Image),
			sheet.SubImage(image.Rect(1056, 0, 1408, 384)).(*ebiten.Image),
		}

		// 2. Row 1: Swim (8 frames)
		c.diverSwimFrames = []*ebiten.Image{
			sheet.SubImage(image.Rect(0, 384, 352, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(352, 384, 704, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(704, 384, 1056, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(1056, 384, 1408, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(1408, 384, 1760, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(1760, 384, 2112, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(2112, 384, 2464, 768)).(*ebiten.Image),
			sheet.SubImage(image.Rect(2464, 384, 2816, 768)).(*ebiten.Image),
		}

		// 3. Row 2: Mine (4 frames) - 6-column grid with 469px width cells
		c.diverMineFrames = []*ebiten.Image{
			sheet.SubImage(image.Rect(0, 768, 469, 1152)).(*ebiten.Image),
			sheet.SubImage(image.Rect(469, 768, 938, 1152)).(*ebiten.Image),
			sheet.SubImage(image.Rect(938, 768, 1408, 1152)).(*ebiten.Image),
			sheet.SubImage(image.Rect(1408, 768, 1877, 1152)).(*ebiten.Image),
		}

		// 4. Row 3: Damage (1 frame)
		c.diverDamageFrame = sheet.SubImage(image.Rect(0, 1152, 352, 1536)).(*ebiten.Image)
	*/
}

func (c *CaveScene) OnEnter(g *Game) {
	g.currentState = StateCave
}

func (c *CaveScene) OnExit(g *Game) {}

// Update handles player input, side-scroller swimming physics, and checks exit transitions.
func (c *CaveScene) Update(g *Game) error {
	// Reset electrical tracking timer; active Electro-Weavers will update it
	g.WeaverTrackingTimer = 0.0

	// Spawn passive plankton / marine snow particles in the viewport
	planktonCount := 0
	for _, p := range g.Particles {
		if p.Type == ParticlePlankton {
			planktonCount++
		}
	}
	if planktonCount < 120 {
		for i := 0; i < 2; i++ {
			rx := g.camera.Pos.X + rand.Float64()*float64(ScreenWidth)
			ry := g.camera.Pos.Y + rand.Float64()*float64(ScreenHeight)
			if ry >= 0 {
				g.SpawnPlankton(rx, ry)
			}
		}
	}

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
		p.IsMining = true
		p.MiningAnimTimer = 24

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

		// Find clicked passive creature (fish/crab) to catch
		for i, ent := range c.Entities {
			if !ent.IsActive() {
				continue
			}
			if creature, ok := ent.(PassiveCreature); ok {
				pos := ent.GetPos()
				dims := ent.GetDimensions()
				if worldX >= pos.X && worldX < pos.X+dims.X && worldY >= pos.Y && worldY < pos.Y+dims.Y {
					playerCenter := gvec.Vec2{X: p.Pos.X + p.Width/2, Y: p.Pos.Y + p.Height/2}
					if creature.CanCatch(playerCenter) {
						harvestedItem := creature.GetHarvestedItem()
						if p.Inventory.AddItem(harvestedItem, 1) {
							ent.SetActive(false)
							c.Entities = append(c.Entities[:i], c.Entities[i+1:]...)
							g.MineWarning = "Caught " + harvestedItem.GetName() + "!"
							g.MineWarningTimer = 90
						} else {
							g.MineWarning = "Inventory full!"
							g.MineWarningTimer = 90
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
	pScreenX := p.Pos.X + p.Width/2.0 - g.camera.Pos.X
	pScreenY := p.Pos.Y + p.Height/2.0 - g.camera.Pos.Y
	dx := cursor.X - pScreenX
	dy := cursor.Y - pScreenY
	p.Facing = math.Atan2(dy, dx)

	// Movement physics settings
	speedProps := p.Speed["cave"]
	var swimForce = speedProps.Acceleration
	var maxSpeed = speedProps.TopSpeed
	var buoyancy = p.Buoyancy  // Upward force (drifts up slowly if not swimming down)
	var drag = speedProps.Drag // Drag resistance in water // TODO: move to player property, so I can add a suite upgrade to reduce drag later.

	isSprinting := g.Input.IsKeyPressed(ebiten.KeyShift)
	if isSprinting && p.CurrentStamina > 0 {
		swimForce *= 1.5
		maxSpeed *= 1.6
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
func (c *CaveScene) checkCollisions(p *player.Player) {
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
func (c *CaveScene) Draw(g *Game, finalScreen *ebiten.Image) {
	// Allocate offscreen texture if nil
	if c.offscreen == nil {
		c.offscreen = ebiten.NewImage(ScreenWidth, ScreenHeight)
	}
	c.offscreen.Clear()

	// Capture the target drawing buffer as 'screen' to redirect all drawing calls
	screen := c.offscreen

	// Camera offset from camera controller
	camX := g.camera.Pos.X
	camY := g.camera.Pos.Y

	// Cave ambient backdrop — fill first so sky can paint over it
	maxDepth := 6000.0
	if c.CaveGrid != nil && len(c.CaveGrid[0]) > 0 {
		maxDepth = float64(len(c.CaveGrid[0]) * TileSize)
	}
	mult := GetOverworldLightMultiplier(g.TimeOfDay)
	if c.ActiveCave != nil {
		c.ActiveCave.DrawBackground(screen, camY, maxDepth, mult)
	} else if c.IsShallow {
		// Fallback (should not happen if ActiveCave is initialized)
		// Surface base color
		baseR := float64(10) + float64(30)*mult
		baseG := float64(40) + float64(80)*mult
		baseB := float64(100) + float64(80)*mult
		maxDarken := 0.45 + (1.0-mult)*0.45
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
		if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == CaveVoid {
			skyColor = color.RGBA{2, 3, 6, 255}
		} else if c.IsShallow {
			skyColor = getSkyColor(g.TimeOfDay)
		} else {
			skyColor = color.RGBA{14, 52, 115, 255} // Darker deep water for deep caves
		}
		vector.FillRect(screen, 0, 0, ScreenWidth, float32(-camY), skyColor, false)
	}

	// Draw a distinct line at the water surface boundary (Y = 0) if visible
	surfaceY := float32(-camY)
	if surfaceY >= 0 && surfaceY < float32(ScreenHeight) {
		var lineColor color.RGBA
		if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == CaveVoid {
			lineColor = color.RGBA{2, 3, 6, 255}
		} else if c.IsShallow {
			lineColor = color.RGBA{220, 240, 255, 255} // foam white
		} else {
			lineColor = color.RGBA{30, 80, 160, 255} // dark water surface for deep cave
		}
		vector.StrokeLine(screen, 0, surfaceY, ScreenWidth, surfaceY, 3.0, lineColor, false)
	}

	// Draw background particles (plankton/marine snow) behind the rock tiles
	c.drawBackgroundParticles(g, screen)

	if c.CaveGrid != nil && c.ActiveCave != nil {
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

		c.ActiveCave.DrawTiles(screen, camX, camY, startTileX, startTileY, endTileX, endTileY)
	}

	// --- Phase 5: Render Resource Nodes ---
	for _, node := range c.Nodes {
		node.Draw(screen, camX, camY)
	}

	isPiloting := g.ActiveVehicle != nil

	// Draw player character at screen coordinates (handles camera tracking/shake/lag)
	pX := float32(g.player.Pos.X + g.player.Width/2.0 - g.camera.Pos.X)
	pY := float32(g.player.Pos.Y + g.player.Height/2.0 - g.camera.Pos.Y)
	facingAngle := g.player.Facing

	if isPiloting {
		facingAngle = g.ActiveVehicle.GetFacing()
	}

	if !isPiloting {
		var activeFrame *ebiten.Image
		p := g.player

		if c.diverSheet != nil {
			if p.IsDamaged {
				activeFrame = c.diverDamageFrame
			} else if p.IsMining {
				// Calculate frame index relative to when the swing started
				// (24 total ticks, 4 frames = 6 ticks per frame)
				elapsed := 24 - p.MiningAnimTimer
				frameIdx := elapsed / 6
				if frameIdx < 0 {
					frameIdx = 0
				}
				if frameIdx > 3 {
					frameIdx = 3
				}
				if frameIdx < len(c.diverMineFrames) {
					activeFrame = c.diverMineFrames[frameIdx]
				}
			} else if math.Hypot(p.Vel.X, p.Vel.Y) > 0.2 {
				frameIdx := (p.AnimTick / 5) % 8
				if frameIdx < len(c.diverSwimFrames) {
					activeFrame = c.diverSwimFrames[frameIdx]
				}
			} else {
				frameIdx := (p.AnimTick / 10) % 4
				if frameIdx < len(c.diverIdleFrames) {
					activeFrame = c.diverIdleFrames[frameIdx]
				}
			}
		}

		if activeFrame != nil {
			op := &ebiten.DrawImageOptions{}
			frameW := activeFrame.Bounds().Dx()
			frameH := activeFrame.Bounds().Dy()

			// Scale the sprite relative to the standard idle frame width to keep the diver's
			// body size consistent across all animations. This allows the wider tool/pickaxe
			// in the mining frames to extend outward naturally without shrinking the diver's body.
			baseFrameW := float64(frameW)
			baseFrameH := float64(frameH)
			if len(c.diverIdleFrames) > 0 {
				baseFrameW = float64(c.diverIdleFrames[0].Bounds().Dx())
				baseFrameH = float64(c.diverIdleFrames[0].Bounds().Dy())
			}

			// Center the frame on the origin before drawing.
			// By translating by -baseFrameW/2 rather than -frameW/2, we align the diver's
			// body horizontally in all frames (standard or wider mining frames). The extra width
			// for the pickaxe/tool on the right side of the mining frame will project outward.
			op.GeoM.Translate(-baseFrameW/2.0, -baseFrameH/2.0)

			// Flip horizontally if facing left
			isFacingLeft := math.Cos(facingAngle) < 0
			if isFacingLeft {
				op.GeoM.Scale(-1, 1)
			}

			scale := DiverDrawWidth / baseFrameW
			op.GeoM.Scale(scale, scale)

			// Translate to screen coordinates
			op.GeoM.Translate(float64(pX), float64(pY))

			screen.DrawImage(activeFrame, op)
		} else {
			// Fallback: original vector drawing
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
	}

	// --- Phase 4: Dynamic Lighting Shader Mask ---
	if LightShader != nil && !c.IsShallow && !g.DebugDisableLightShader {
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
		// Determine base flashlight stats
		baseRadius := float32(360.0)
		baseAngle := float32(math.Pi / 7.5) // ~24 degrees half-cone

		// Calculate depth-based decay (gets narrower and dimmer as depth increases)
		maxDepth := 6000.0
		if len(c.CaveGrid) > 0 && len(c.CaveGrid[0]) > 0 {
			maxDepth = float64(len(c.CaveGrid[0]) * TileSize)
		}

		depth := g.player.Pos.Y
		if depth < 0 {
			depth = 0
		}
		depthFrac := depth / maxDepth
		if depthFrac > 1.0 {
			depthFrac = 1.0
		}

		// Decay factors based on upgrades
		maxRadiusDecay := float32(0.65) // 65% decay without upgrades
		maxAngleDecay := float32(0.50)  // 50% decay without upgrades

		// Check player upgrades using MaxOxygen cache proxy
		if g.player.MaxOxygen >= 240.0 {
			// Ultra High Capacity: 15% radius decay, 10% angle decay
			maxRadiusDecay = 0.15
			maxAngleDecay = 0.10
		} else if g.player.MaxOxygen >= 160.0 {
			// High Capacity: 35% radius decay, 25% angle decay
			maxRadiusDecay = 0.35
			maxAngleDecay = 0.25
		}

		// Apply decay
		radius := baseRadius * (1.0 - float32(depthFrac)*maxRadiusDecay)
		angle := baseAngle * (1.0 - float32(depthFrac)*maxAngleDecay)

		c.uniforms["LightRadius"] = radius
		c.uniforms["ConeHalfAngle"] = angle
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
		entranceActive := float32(1.0)
		if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == CaveVoid {
			entranceActive = 0.0
		}
		c.uniforms["EntranceActive"] = entranceActive

		screen.DrawRectShader(ScreenWidth, ScreenHeight, LightShader, &c.shaderOpts)
	}

	// --- Phase 4: Bioluminescent Highlights (drawn on top of the shader mask) ---
	if !c.IsShallow {
		c.drawBioluminescence(g, screen, camX, camY)
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

	// Collect active Brimstone Siphon screen coordinates for heat wave shader
	var ventPositions [16]float32
	var ventCount float32 = 0
	for _, ent := range c.Entities {
		if siphon, ok := ent.(*BrimstoneSiphon); ok && siphon.IsActive() {
			if siphon.Timer >= 60 && ventCount < 8 {
				vx := float32(siphon.Pos.X - g.camera.Pos.X + siphon.Dimensions.X/2.0)
				vy := float32(siphon.Pos.Y - g.camera.Pos.Y + siphon.Dimensions.Y/2.0)
				idx := int(ventCount) * 2
				ventPositions[idx] = vx
				ventPositions[idx+1] = vy
				ventCount++
			}
		}
	}

	// Apply water shimmer & heat distortion displacement shader
	if WaterDisplacementShader != nil && !g.DebugDisableWaterShader {
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = c.offscreen
		op.Uniforms = map[string]any{
			"Time":          float32(g.Ticks),
			"VentPositions": ventPositions,
			"VentCount":     ventCount,
			"SurfaceY":      float32(-g.camera.Pos.Y),
		}
		finalScreen.DrawRectShader(ScreenWidth, ScreenHeight, WaterDisplacementShader, op)
	} else {
		finalScreen.DrawImage(c.offscreen, nil)
	}
}

// drawBioluminescence renders glowing flora and spores that remain visible in the pitch dark.
func (c *CaveScene) drawBioluminescence(g *Game, screen *ebiten.Image, camX, camY float64) {
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

					// Draw soft outer light emission with a pulsing effect over time
					pulse := float32(math.Cos(float64(g.Ticks)*0.015+float64(hash))) * 1.5
					radius := float32(5.0) + pulse
					if radius < 2.0 {
						radius = 2.0
					}
					vector.FillCircle(screen, sx, sy, radius, color.RGBA{glowColor.R, glowColor.G, glowColor.B, 70}, false)
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

// drawBackgroundParticles renders the plankton/marine snow behind the rock tiles.
func (c *CaveScene) drawBackgroundParticles(g *Game, screen *ebiten.Image) {
	camX := g.camera.Pos.X
	camY := g.camera.Pos.Y

	for _, p := range g.Particles {
		if p.Type != ParticlePlankton {
			continue
		}

		if p.Pos.Y < 0 {
			// Skip plankton in the air/sky (Y < 0 is above water surface)
			continue
		}

		sx := float32(p.Pos.X - camX)
		sy := float32(p.Pos.Y - camY)

		clr := p.Color
		// Fade in for the first 10% of life, then fade out for the rest
		opacity := p.Life
		if p.Life > 0.9 {
			opacity = (1.0 - p.Life) * 10.0
		}
		clr.A = uint8(float64(clr.A) * opacity)

		// Draw plankton as soft small squares/pixels
		vector.FillRect(screen, sx-p.Size/2.0, sy-p.Size/2.0, p.Size, p.Size, clr, false)
	}
}
