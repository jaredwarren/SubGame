package scene

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
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/shader"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// DiverDrawWidth defines the targeted width of the diver sprite on screen.
var DiverDrawWidth = 36.0

var (
	DiverGridCols     = 8
	DiverGridRows     = 4
	DiverMineGridCols = 6
)

var (
	DiverIdleCols   = []int{0, 1, 2, 3}
	DiverSwimCols   = []int{0, 1, 2, 3, 4, 5, 6, 7}
	DiverMineCols   = []int{0, 1, 2, 3}
	DiverDamageCols = []int{0}
)

var (
	DiverXOffset = 0
	DiverYOffset = 0
)

// CaveScene manages the side-view cave swimming controls, collision, and rendering.
type CaveScene struct {
	ActiveCave cave.Cave
	CaveGrid   [][]bool
	Nodes      []resource.Resource
	Entities   []entity.CaveEntity
	IsShallow  bool

	shaderOpts    ebiten.DrawRectShaderOptions
	Uniforms      map[string]any
	lightSource   []float32
	flashlightDir []float32
	sonarSource   []float32
	entranceLight []float32

	offscreen *ebiten.Image

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
		Entities:      []entity.CaveEntity{},
		Uniforms:      make(map[string]any),
		lightSource:   make([]float32, 2),
		flashlightDir: make([]float32, 2),
		sonarSource:   make([]float32, 2),
		entranceLight: make([]float32, 2),
	}
	cs.shaderOpts.Uniforms = cs.Uniforms
	cs.loadDiverSheet()
	return cs
}

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
	defer func() { _ = file.Close() }()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Warning: Failed to decode assets/textures/diver_sheet.png: %v", err)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			clr := img.At(x, y)
			r, g, b, a := clr.RGBA()
			ru := uint8(r >> 8)
			gu := uint8(g >> 8)
			bu := uint8(b >> 8)
			au := uint8(a >> 8)
			if gu > 140 && ru < 100 && bu < 100 {
				rgba.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				rgba.SetRGBA(x, y, color.RGBA{ru, gu, bu, au})
			}
		}
	}

	sheet := ebiten.NewImageFromImage(rgba)
	c.diverSheet = sheet

	frameW := bounds.Dx() / DiverGridCols
	frameH := bounds.Dy() / DiverGridRows

	for _, col := range DiverIdleCols {
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH)
		c.diverIdleFrames = append(c.diverIdleFrames, sheet.SubImage(rect).(*ebiten.Image))
	}
	for _, col := range DiverSwimCols {
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+frameH, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH*2)
		c.diverSwimFrames = append(c.diverSwimFrames, sheet.SubImage(rect).(*ebiten.Image))
	}
	mineFrameW := bounds.Dx() / DiverMineGridCols
	for _, col := range DiverMineCols {
		rect := image.Rect(DiverXOffset+col*mineFrameW, DiverYOffset+frameH*2, DiverXOffset+(col+1)*mineFrameW, DiverYOffset+frameH*3)
		c.diverMineFrames = append(c.diverMineFrames, sheet.SubImage(rect).(*ebiten.Image))
	}
	if len(DiverDamageCols) > 0 {
		col := DiverDamageCols[0]
		rect := image.Rect(DiverXOffset+col*frameW, DiverYOffset+frameH*3, DiverXOffset+(col+1)*frameW, DiverYOffset+frameH*4)
		c.diverDamageFrame = sheet.SubImage(rect).(*ebiten.Image)
	}
}

func (c *CaveScene) OnEnter(g GameContext) {
	g.SetCurrentState(StateCave)
}

func (c *CaveScene) OnExit(g GameContext) {}

// Update handles player input, side-scroller swimming physics, and checks exit transitions.
func (c *CaveScene) Update(g GameContext) error {
	g.SetWeaverTrackingTimer(0.0)

	cam := g.GetCamera()
	particles := g.GetParticles()
	planktonCount := 0
	for _, p := range particles {
		if p.Type == particle.ParticlePlankton {
			planktonCount++
		}
	}
	if planktonCount < 120 {
		for i := 0; i < 2; i++ {
			rx := cam.Pos.X + rand.Float64()*float64(config.ScreenWidth)
			ry := cam.Pos.Y + rand.Float64()*float64(config.ScreenHeight)
			if ry >= 0 {
				g.SpawnPlankton(rx, ry)
			}
		}
	}

	entityRuntime := g.NewEntityRuntime()
	for _, ent := range c.Entities {
		ent.Update(entityRuntime, c.CaveGrid)
	}

	activeCount := 0
	for _, ent := range c.Entities {
		if ent.IsActive() {
			c.Entities[activeCount] = ent
			activeCount++
		}
	}
	c.Entities = c.Entities[:activeCount]
	g.SetCaveEntities(g.GetActiveTrenchKey(), c.Entities)

	g.DrainEntityCommands(entityRuntime)

	activeVehicle := g.GetActiveVehicle()
	inp := g.GetInput()

	if activeVehicle != nil {
		if mech, ok := activeVehicle.(*vehicle.HeavyMech); ok && !mech.IsDrilling {
			if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				cursor := inp.Cursor()
				camX := mech.Pos.X - config.ScreenWidth/2 + mech.Dimensions.X/2
				camY := mech.Pos.Y - config.ScreenHeight/2 + mech.Dimensions.Y/2
				worldX := camX + cursor.X
				worldY := camY + cursor.Y

				mtx := int(worldX) / config.TileSize
				mty := int(worldY) / config.TileSize

				for i := 0; i < len(c.Nodes); i++ {
					node := c.Nodes[i]
					nodeTx, nodeTy := node.GetTilePos()
					if nodeTx == mtx && nodeTy == mty && node.GetHitsToMine() > 0 {
						px := mech.Pos.X + mech.Dimensions.X/2
						py := mech.Pos.Y + mech.Dimensions.Y/2
						nx := float64(nodeTx*config.TileSize + config.TileSize/2)
						ny := float64(nodeTy*config.TileSize + config.TileSize/2)
						if math.Hypot(px-nx, py-ny) <= 120.0 {
							mech.DrillStrike(node)
							break
						}
					}
				}
			}
		}
		return nil
	}

	p := g.GetPlayer()

	if p.Pos.Y < -8 {
		g.ExitCave()
		return nil
	}

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		p.IsMining = true
		p.MiningAnimTimer = 24

		cursor := inp.Cursor()
		camX := p.Pos.X - config.ScreenWidth/2 + p.Width/2
		camY := p.Pos.Y - config.ScreenHeight/2 + p.Height/2
		worldX := camX + cursor.X
		worldY := camY + cursor.Y

		mtx := int(worldX) / config.TileSize
		mty := int(worldY) / config.TileSize

		for _, ent := range c.Entities {
			if ent.GetType() == entity.EntShatterBulb && ent.IsActive() {
				pos := ent.GetPos()
				dims := ent.GetDimensions()
				if worldX >= pos.X && worldX < pos.X+dims.X && worldY >= pos.Y && worldY < pos.Y+dims.Y {
					px := p.Pos.X + p.Width/2
					py := p.Pos.Y + p.Height/2
					if math.Hypot(px-(pos.X+dims.X/2), py-(pos.Y+dims.Y/2)) <= 96.0 {
						if bulb, ok := ent.(*entity.ShatterBulb); ok {
							bulb.Pop(entityRuntime)
						}
						break
					}
				}
			}
		}

		for i, ent := range c.Entities {
			if !ent.IsActive() {
				continue
			}
			if creature, ok := ent.(entity.PassiveCreature); ok {
				pos := ent.GetPos()
				dims := ent.GetDimensions()
				if worldX >= pos.X && worldX < pos.X+dims.X && worldY >= pos.Y && worldY < pos.Y+dims.Y {
					playerCenter := gvec.Vec2{X: p.Pos.X + p.Width/2, Y: p.Pos.Y + p.Height/2}
					if creature.CanCatch(playerCenter) {
						harvestedItem := creature.GetHarvestedItem()
						if p.Inventory.AddItem(harvestedItem, 1) {
							ent.SetActive(false)
							c.Entities = append(c.Entities[:i], c.Entities[i+1:]...)
							g.SetMineWarning("Caught "+harvestedItem.GetName()+"!", 90)
						} else {
							g.SetMineWarning("Inventory full!", 90)
						}
						break
					}
				}
			}
		}

		for i := 0; i < len(c.Nodes); i++ {
			node := c.Nodes[i]
			nodeTx, nodeTy := node.GetTilePos()
			if nodeTx == mtx && nodeTy == mty && node.GetHitsToMine() > 0 {
				px := p.Pos.X + p.Width/2
				py := p.Pos.Y + p.Height/2
				nx := float64(nodeTx*config.TileSize + config.TileSize/2)
				ny := float64(nodeTy*config.TileSize + config.TileSize/2)

				if math.Hypot(px-nx, py-ny) <= 96.0 {
					if node.RequiresMech() {
						g.SetMineWarning("Requires Heavy Mech Drill Arm to harvest", 120)
						continue
					}
					node.SetHitsToMine(node.GetHitsToMine() - 1)

					nodeColor := color.RGBA{150, 150, 150, 255}
					if cRgba, ok := node.GetColor().(color.RGBA); ok {
						nodeColor = cRgba
					}
					g.SpawnDebris(nx, ny, nodeColor)

					if node.GetHitsToMine() <= 0 {
						p.Inventory.AddItem(node, 1)
						c.Nodes = append(c.Nodes[:i], c.Nodes[i+1:]...)
					}
					break
				}
			}
		}
	}

	cursor := inp.Cursor()
	pScreenX := p.Pos.X + p.Width/2.0 - cam.Pos.X
	pScreenY := p.Pos.Y + p.Height/2.0 - cam.Pos.Y
	p.Facing = math.Atan2(cursor.Y-pScreenY, cursor.X-pScreenX)

	speedProps := p.Speed["cave"]
	swimForce := speedProps.Acceleration
	maxSpeed := speedProps.TopSpeed
	buoyancy := p.Buoyancy
	drag := speedProps.Drag

	isSprinting := inp.IsKeyPressed(ebiten.KeyShift)
	if isSprinting && p.CurrentStamina > 0 {
		swimForce *= 1.5
		maxSpeed *= 1.6
	}
	if g.IsPlayerSlowed() {
		swimForce *= 0.5
		maxSpeed *= 0.5
	}

	swimming := false
	if inp.IsKeyPressed(ebiten.KeyW) || inp.IsKeyPressed(ebiten.KeyArrowUp) {
		p.Vel.Y -= swimForce
		swimming = true
	}
	if inp.IsKeyPressed(ebiten.KeyS) || inp.IsKeyPressed(ebiten.KeyArrowDown) {
		p.Vel.Y += swimForce
		swimming = true
	}
	if inp.IsKeyPressed(ebiten.KeyA) || inp.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.Vel.X -= swimForce
		swimming = true
	}
	if inp.IsKeyPressed(ebiten.KeyD) || inp.IsKeyPressed(ebiten.KeyArrowRight) {
		p.Vel.X += swimForce
		swimming = true
	}

	p.Vel.Y += buoyancy
	p.Vel = p.Vel.Scale(drag)

	speed := p.Vel.Length()
	if speed > maxSpeed {
		p.Vel = p.Vel.Scale(maxSpeed / speed)
	}

	if swimming && speed > 2.0 && rand.Float64() < 0.12 {
		g.SpawnBubble(p.Pos.X+p.Width/2.0, p.Pos.Y+p.Height/2.0)
	}

	c.checkCollisions(p)

	isMoving := speed > 0.1
	p.UpdateStats(true, isSprinting && isMoving && swimming)

	return nil
}

func (c *CaveScene) checkCollisions(p *player.Player) {
	newX := p.Pos.X + p.Vel.X
	if c.IsSolid(newX, p.Pos.Y, p.Width, p.Height) {
		p.Vel.X = 0
	} else {
		p.Pos.X = newX
	}
	newY := p.Pos.Y + p.Vel.Y
	if c.IsSolid(p.Pos.X, newY, p.Width, p.Height) {
		p.Vel.Y = 0
	} else {
		p.Pos.Y = newY
	}
}

// IsSolid checks if the proposed bounding box overlaps with solid cave tiles.
func (c *CaveScene) IsSolid(x, y, w, h float64) bool {
	if c.CaveGrid == nil {
		return false
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])

	x1 := int(math.Floor(x)) / config.TileSize
	x2 := int(math.Floor(x+w)) / config.TileSize
	y1 := int(math.Floor(y)) / config.TileSize
	y2 := int(math.Floor(y+h)) / config.TileSize

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
			if c.CaveGrid[tx][ty] {
				return true
			}
		}
	}
	return false
}

// Draw renders the cave scene, solid tiles, and player assets.
func (c *CaveScene) Draw(g GameContext, finalScreen *ebiten.Image) {
	if c.offscreen == nil {
		c.offscreen = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	}
	c.offscreen.Clear()
	screen := c.offscreen

	cam := g.GetCamera()
	camX := cam.Pos.X
	camY := cam.Pos.Y

	maxDepth := 6000.0
	if c.CaveGrid != nil && len(c.CaveGrid[0]) > 0 {
		maxDepth = float64(len(c.CaveGrid[0]) * config.TileSize)
	}
	mult := GetOverworldLightMultiplier(g.GetTimeOfDay())
	if c.ActiveCave != nil {
		c.ActiveCave.DrawBackground(screen, camY, maxDepth, mult)
	} else if c.IsShallow {
		baseR := float64(10) + float64(30)*mult
		baseG := float64(40) + float64(80)*mult
		baseB := float64(100) + float64(80)*mult
		maxDarken := 0.45 + (1.0-mult)*0.45
		const stripH = float32(6)
		for sy := float32(0); sy < float32(config.ScreenHeight); sy += stripH {
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
			vector.FillRect(screen, 0, sy, float32(config.ScreenWidth), stripH, sc, false)
		}
	} else {
		screen.Fill(color.RGBA{10, 8, 16, 255})
	}

	if camY < 0 {
		var skyColor color.RGBA
		if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == cave.CaveVoid {
			skyColor = color.RGBA{2, 3, 6, 255}
		} else if c.IsShallow {
			skyColor = getSkyColor(g.GetTimeOfDay())
		} else {
			skyColor = color.RGBA{14, 52, 115, 255}
		}
		vector.FillRect(screen, 0, 0, float32(config.ScreenWidth), float32(-camY), skyColor, false)
	}

	surfaceY := float32(-camY)
	if surfaceY >= 0 && surfaceY < float32(config.ScreenHeight) {
		var lineColor color.RGBA
		if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == cave.CaveVoid {
			lineColor = color.RGBA{2, 3, 6, 255}
		} else if c.IsShallow {
			lineColor = color.RGBA{220, 240, 255, 255}
		} else {
			lineColor = color.RGBA{30, 80, 160, 255}
		}
		vector.StrokeLine(screen, 0, surfaceY, float32(config.ScreenWidth), surfaceY, 3.0, lineColor, false)
	}

	c.drawBackgroundParticles(g, screen)

	if c.CaveGrid != nil && c.ActiveCave != nil {
		gridW := len(c.CaveGrid)
		gridH := len(c.CaveGrid[0])
		startTileX := max0(int(camX)/config.TileSize, 0)
		endTileX := min0((int(camX)+config.ScreenWidth)/config.TileSize+1, gridW)
		startTileY := max0(int(camY)/config.TileSize, 0)
		endTileY := min0((int(camY)+config.ScreenHeight)/config.TileSize+1, gridH)
		c.ActiveCave.DrawTiles(screen, camX, camY, startTileX, startTileY, endTileX, endTileY)
	}

	for _, node := range c.Nodes {
		node.Draw(screen, camX, camY)
	}

	p := g.GetPlayer()
	isPiloting := g.GetActiveVehicle() != nil

	pX := float32(p.Pos.X + p.Width/2.0 - camX)
	pY := float32(p.Pos.Y + p.Height/2.0 - camY)
	facingAngle := p.Facing
	if isPiloting {
		facingAngle = g.GetActiveVehicle().GetFacing()
	}

	if !isPiloting {
		var activeFrame *ebiten.Image
		if c.diverSheet != nil {
			if p.IsDamaged {
				activeFrame = c.diverDamageFrame
			} else if p.IsMining {
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
			baseFrameW := float64(activeFrame.Bounds().Dx())
			baseFrameH := float64(activeFrame.Bounds().Dy())
			if len(c.diverIdleFrames) > 0 {
				baseFrameW = float64(c.diverIdleFrames[0].Bounds().Dx())
				baseFrameH = float64(c.diverIdleFrames[0].Bounds().Dy())
			}
			op.GeoM.Translate(-baseFrameW/2.0, -baseFrameH/2.0)
			if math.Cos(facingAngle) < 0 {
				op.GeoM.Scale(-1, 1)
			}
			scale := DiverDrawWidth / baseFrameW
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(pX), float64(pY))
			screen.DrawImage(activeFrame, op)
		} else {
			tankAngle := facingAngle + math.Pi
			tx := pX + float32(math.Cos(tankAngle))*8
			ty := pY + float32(math.Sin(tankAngle))*8
			vector.FillCircle(screen, tx, ty, 6, color.RGBA{240, 220, 50, 255}, false)
			vector.FillCircle(screen, pX, pY, 9, color.RGBA{220, 95, 45, 255}, false)
			vx := pX + float32(math.Cos(facingAngle))*6
			vy := pY + float32(math.Sin(facingAngle))*6
			vector.FillCircle(screen, vx, vy, 5, color.RGBA{80, 200, 255, 200}, false)
		}
	}

	sonar := g.GetSonar()
	if shader.LightShader != nil && !c.IsShallow && !g.IsDebugLightShaderDisabled() {
		var sonarSourceX, sonarSourceY, sonarRadius float32
		if sonar.Timer > 0 {
			sonarSourceX = float32(sonar.SourceX - camX)
			sonarSourceY = float32(sonar.SourceY - camY)
			sonarRadius = float32(sonar.Radius)
		}

		var fDirX, fDirY float32
		weaverTimer := g.GetWeaverTrackingTimer()
		if g.IsFlashlightOn() {
			fDirX = float32(math.Cos(facingAngle))
			fDirY = float32(math.Sin(facingAngle))
			if weaverTimer > 0 && rand.Float64() < (weaverTimer/300.0)*0.20 {
				fDirX, fDirY = 0, 0
			}
		}

		entranceX := float32(float64(len(c.CaveGrid)/2*config.TileSize) + config.TileSize/2.0 - camX)
		entranceY := float32(0.0 - camY)

		c.lightSource[0], c.lightSource[1] = pX, pY
		c.flashlightDir[0], c.flashlightDir[1] = fDirX, fDirY
		c.sonarSource[0], c.sonarSource[1] = sonarSourceX, sonarSourceY
		c.entranceLight[0], c.entranceLight[1] = entranceX, entranceY

		c.Uniforms["LightSource"] = c.lightSource
		c.Uniforms["FlashlightDir"] = c.flashlightDir

		maxDepthF := 6000.0
		if len(c.CaveGrid) > 0 && len(c.CaveGrid[0]) > 0 {
			maxDepthF = float64(len(c.CaveGrid[0]) * config.TileSize)
		}
		depth := p.Pos.Y
		if depth < 0 {
			depth = 0
		}
		depthFrac := depth / maxDepthF
		if depthFrac > 1.0 {
			depthFrac = 1.0
		}
		maxRadiusDecay := float32(0.65)
		maxAngleDecay := float32(0.50)
		if p.MaxOxygen >= 240.0 {
			maxRadiusDecay, maxAngleDecay = 0.15, 0.10
		} else if p.MaxOxygen >= 160.0 {
			maxRadiusDecay, maxAngleDecay = 0.35, 0.25
		}
		radius := float32(360.0) * (1.0 - float32(depthFrac)*maxRadiusDecay)
		angle := float32(math.Pi/7.5) * (1.0 - float32(depthFrac)*maxAngleDecay)

		c.Uniforms["LightRadius"] = radius
		c.Uniforms["ConeHalfAngle"] = angle
		c.Uniforms["PersonalRadius"] = float32(65.0)
		c.Uniforms["AmbientColor"] = c.getAmbientColor(c.IsShallow, g.GetTimeOfDay())
		c.Uniforms["SonarSource"] = c.sonarSource
		c.Uniforms["SonarRadius"] = sonarRadius
		sonarBright := float32(1.0)
		sonarFadeLimit := float32(1200.0)
		if sonar.Bright {
			sonarBright, sonarFadeLimit = 2.5, 3000.0
		}
		c.Uniforms["SonarBright"] = sonarBright
		c.Uniforms["SonarFadeLimit"] = sonarFadeLimit
		c.Uniforms["EntranceLight"] = c.entranceLight
		entranceActive := float32(1.0)
		if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == cave.CaveVoid {
			entranceActive = 0.0
		}
		c.Uniforms["EntranceActive"] = entranceActive

		screen.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.LightShader, &c.shaderOpts)
	}

	if !c.IsShallow {
		c.drawBioluminescence(g, screen, camX, camY)
	}

	for _, ent := range c.Entities {
		ent.Draw(screen, cam, g.GetTimeOfDay())
	}

	swTimer, swX, swY, swRadius := g.GetSoundWaveState()
	if swTimer > 0 {
		alpha := float32(swTimer) / 60.0
		clr := color.RGBA{245, 120, 20, uint8(200 * alpha)}
		vector.StrokeCircle(screen, float32(swX-camX), float32(swY-camY), float32(swRadius), 2.0, clr, false)
	}

	var ventPositions [16]float32
	var ventCount float32 = 0
	for _, ent := range c.Entities {
		if siphon, ok := ent.(*entity.BrimstoneSiphon); ok && siphon.IsActive() && siphon.Timer >= 60 && ventCount < 8 {
			idx := int(ventCount) * 2
			ventPositions[idx] = float32(siphon.Pos.X - camX + siphon.Dimensions.X/2.0)
			ventPositions[idx+1] = float32(siphon.Pos.Y - camY + siphon.Dimensions.Y/2.0)
			ventCount++
		}
	}

	if shader.WaterDisplacementShader != nil && !g.IsDebugWaterShaderDisabled() {
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = c.offscreen
		op.Uniforms = map[string]any{
			"Time":          float32(g.GetTicks()),
			"VentPositions": ventPositions,
			"VentCount":     ventCount,
			"SurfaceY":      float32(-camY),
		}
		finalScreen.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.WaterDisplacementShader, op)
	} else {
		finalScreen.DrawImage(c.offscreen, nil)
	}
}

func (c *CaveScene) drawBioluminescence(g GameContext, screen *ebiten.Image, camX, camY float64) {
	if c.CaveGrid == nil {
		return
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])
	startTileX := max0(int(camX)/config.TileSize, 0)
	endTileX := min0((int(camX)+config.ScreenWidth)/config.TileSize+1, gridW)
	startTileY := max0(int(camY)/config.TileSize, 0)
	endTileY := min0((int(camY)+config.ScreenHeight)/config.TileSize+1, gridH)

	for tx := startTileX; tx < endTileX; tx++ {
		for ty := startTileY; ty < endTileY; ty++ {
			if c.CaveGrid[tx][ty] {
				hash := (tx*31 + ty*17) % 17
				if hash == 0 {
					sx := float32(tx*config.TileSize-int(camX)) + float32(config.TileSize)/2.0
					sy := float32(ty*config.TileSize-int(camY)) + float32(config.TileSize)/2.0

					var glowColor color.RGBA
					if (tx+ty)%2 == 0 {
						glowColor = color.RGBA{0, 245, 210, 255}
					} else {
						glowColor = color.RGBA{245, 75, 140, 255}
					}

					pulse := float32(math.Cos(g.GetTicks()*0.015+float64(hash))) * 1.5
					radius := float32(5.0) + pulse
					if radius < 2.0 {
						radius = 2.0
					}
					vector.FillCircle(screen, sx, sy, radius, color.RGBA{glowColor.R, glowColor.G, glowColor.B, 70}, false)
					vector.FillCircle(screen, sx, sy, 1.5, color.RGBA{255, 255, 255, 255}, false)
				}
			}
		}
	}
}

func getSkyColor(timeOfDay float64) color.RGBA {
	nightSky := [3]float64{20, 30, 70}
	dawnSky := [3]float64{255, 160, 80}
	daySky := [3]float64{140, 200, 255}
	duskSky := [3]float64{220, 100, 60}

	lerpF := func(a, b [3]float64, t float64) color.RGBA {
		return color.RGBA{
			R: uint8(a[0] + (b[0]-a[0])*t),
			G: uint8(a[1] + (b[1]-a[1])*t),
			B: uint8(a[2] + (b[2]-a[2])*t),
			A: 255,
		}
	}

	switch {
	case timeOfDay < 1200:
		return lerpF(nightSky, dawnSky, timeOfDay/1200.0)
	case timeOfDay < 2400:
		return lerpF(dawnSky, daySky, (timeOfDay-1200.0)/1200.0)
	case timeOfDay < 6000:
		return lerpF(daySky, daySky, 0)
	case timeOfDay < 7200:
		return lerpF(daySky, duskSky, (timeOfDay-6000.0)/1200.0)
	case timeOfDay < 8400:
		return lerpF(duskSky, nightSky, (timeOfDay-7200.0)/1200.0)
	default:
		return lerpF(nightSky, nightSky, 0)
	}
}

func (c *CaveScene) getAmbientColor(isShallow bool, timeOfDay float64) []float32 {
	if config.LightCaveForDebug {
		return []float32{0.02, 0.02, 0.03, 0.15}
	}
	if isShallow {
		mult := GetOverworldLightMultiplier(timeOfDay)
		alpha := float32(0.75 - (mult-0.2)/0.8*0.60)
		return []float32{0.04, 0.06, 0.12, alpha}
	}
	return []float32{0.01, 0.01, 0.03, 0.97}
}

func (c *CaveScene) drawBackgroundParticles(g GameContext, screen *ebiten.Image) {
	camX := g.GetCamera().Pos.X
	camY := g.GetCamera().Pos.Y

	for _, p := range g.GetParticles() {
		if p.Type != particle.ParticlePlankton || p.Pos.Y < 0 {
			continue
		}
		sx := float32(p.Pos.X - camX)
		sy := float32(p.Pos.Y - camY)
		clr := p.Color
		opacity := p.Life
		if p.Life > 0.9 {
			opacity = (1.0 - p.Life) * 10.0
		}
		clr.A = uint8(float64(clr.A) * opacity)
		vector.FillRect(screen, sx-p.Size/2.0, sy-p.Size/2.0, p.Size, p.Size, clr, false)
	}
}

func max0(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min0(a, b int) int {
	if a < b {
		return a
	}
	return b
}
