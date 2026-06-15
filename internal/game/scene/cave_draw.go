package scene

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/game/shader"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// draw renders the cave scene, solid tiles, and player assets.
func (c *CaveScene) draw(g CaveContext, finalScreen *ebiten.Image) {
	if c.offscreen == nil {
		c.offscreen = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	}
	c.offscreen.Clear()

	if c.scrollActive {
		c.drawScrollTransition(g)
	} else {
		cam := g.GetCamera()
		c.drawScene(g, c.offscreen, c.ActiveCave, c.CaveGrid, c.Nodes, c.Entities, g.GetActiveTrenchKey(), cam.Pos.X, cam.Pos.Y, false)
		c.applyLighting(g)
		if !c.IsShallow {
			c.drawBioluminescence(g, c.offscreen, cam.Pos.X, cam.Pos.Y)
		}
	}

	c.applyWaterDisplacement(g, finalScreen)
}

func (c *CaveScene) drawScrollTransition(g CaveContext) {
	if c.offscreenOld == nil {
		c.offscreenOld = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	}
	if c.offscreenNew == nil {
		c.offscreenNew = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	}
	c.offscreenOld.Clear()
	c.offscreenNew.Clear()

	// 1. Draw old scene
	c.drawScene(g, c.offscreenOld, c.oldCave, c.oldCaveGrid, c.oldNodes, c.oldEntities, c.oldTrenchKey, c.oldCamX, c.oldCamY, true)

	// 2. Draw new scene
	c.drawScene(g, c.offscreenNew, c.newCave, c.newCaveGrid, c.newNodes, c.newEntities, c.newTrenchKey, c.newCamX, c.newCamY, true)

	// 3. Slide them on c.offscreen
	t := float64(c.scrollTimer) / 45.0
	t = math.Sin(t * math.Pi / 2.0)

	opOld := &ebiten.DrawImageOptions{}
	opNew := &ebiten.DrawImageOptions{}

	if c.scrollDir == 1 { // Scrolling right
		opOld.GeoM.Translate(-t*float64(config.ScreenWidth), 0)
		opNew.GeoM.Translate((1.0-t)*float64(config.ScreenWidth), 0)
	} else { // Scrolling left
		opOld.GeoM.Translate(t*float64(config.ScreenWidth), 0)
		opNew.GeoM.Translate((t-1.0)*float64(config.ScreenWidth), 0)
	}

	c.offscreen.DrawImage(c.offscreenOld, opOld)
	c.offscreen.DrawImage(c.offscreenNew, opNew)

	// 4. Draw the player or active vehicle at the interpolated position
	p := g.GetPlayer()
	isPiloting := g.GetActiveVehicle() != nil
	cam := g.GetCamera()

	var width float64 = p.Width
	if isPiloting {
		width = g.GetActiveVehicle().GetDimensions().X
	}

	var screenStartX float64
	var screenEndX float64
	if c.scrollDir == 1 { // Scrolling right
		screenStartX = float64(config.ScreenWidth) - width
		screenEndX = 20.0
	} else { // Scrolling left
		screenStartX = 20.0
		screenEndX = float64(config.ScreenWidth) - width - 20.0
	}

	interpolatedX := (1.0-t)*screenStartX + t*screenEndX

	if isPiloting {
		v := g.GetActiveVehicle()
		oldPos := v.GetPos()
		v.SetPos(gvec.Vec2{X: interpolatedX, Y: oldPos.Y})
		v.Draw(c.offscreen, 0, cam.Pos.Y)
		v.SetPos(oldPos)
	} else {
		pX := float32(interpolatedX + p.Width/2.0)
		pY := float32(p.Pos.Y + p.Height/2.0 - cam.Pos.Y)
		c.drawPlayer(c.offscreen, p, pX, pY)
	}
}

func (c *CaveScene) drawPlayer(screen *ebiten.Image, p *player.Player, pX, pY float32) {
	facingAngle := p.Facing

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

func (c *CaveScene) applyLighting(g CaveContext) {
	// Enable lighting for Shock Kelp Cave even though it is classified as shallow
	isShallowWithoutLight := c.IsShallow && (c.ActiveCave == nil || (c.ActiveCave.GetCaveType() != cave.CaveShockKelp && c.ActiveCave.GetCaveType() != cave.CaveThermo))

	if shader.LightShader == nil || isShallowWithoutLight || g.IsDebugLightShaderDisabled() {
		return
	}

	cam := g.GetCamera()
	sonar := g.GetSonar()
	p := g.GetPlayer()
	isPiloting := g.GetActiveVehicle() != nil
	facingAngle := p.Facing
	if isPiloting {
		facingAngle = g.GetActiveVehicle().GetFacing()
	}
	pX := float32(p.Pos.X + p.Width/2.0 - cam.Pos.X)
	pY := float32(p.Pos.Y + p.Height/2.0 - cam.Pos.Y)

	var sonarSourceX, sonarSourceY, sonarRadius float32
	if sonar.Timer > 0 {
		sonarSourceX = float32(sonar.SourceX - cam.Pos.X)
		sonarSourceY = float32(sonar.SourceY - cam.Pos.Y)
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

	entranceX := float32(float64(len(c.CaveGrid)/2*config.TileSize) + config.TileSize/2.0 - cam.Pos.X)
	entranceY := float32(0.0 - cam.Pos.Y)

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
	c.Uniforms["AmbientColor"] = c.getAmbientColor(g.GetTimeOfDay())
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

	var lavaPositions [16]float32
	var lavaCount float32 = 0
	for _, ent := range c.Entities {
		if siphon, ok := ent.(*entity.BrimstoneSiphon); ok && siphon.IsActive() && lavaCount < 8 {
			idx := int(lavaCount) * 2
			lavaPositions[idx] = float32(siphon.Pos.X - cam.Pos.X + siphon.Dimensions.X/2.0)
			lavaPositions[idx+1] = float32(siphon.Pos.Y - cam.Pos.Y + siphon.Dimensions.Y/2.0)
			lavaCount++
		}
	}
	c.Uniforms["LavaPositions"] = lavaPositions
	c.Uniforms["LavaCount"] = lavaCount

	c.offscreen.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.LightShader, &c.shaderOpts)
}

func (c *CaveScene) applyWaterDisplacement(g CaveContext, finalScreen *ebiten.Image) {
	cam := g.GetCamera()
	if shader.WaterDisplacementShader != nil && !g.IsDebugWaterShaderDisabled() {
		var ventPositions [16]float32
		var ventCount float32 = 0
		for _, ent := range c.Entities {
			if siphon, ok := ent.(*entity.BrimstoneSiphon); ok && siphon.IsActive() && siphon.Timer >= 60 && ventCount < 8 {
				idx := int(ventCount) * 2
				ventPositions[idx] = float32(siphon.Pos.X - cam.Pos.X + siphon.Dimensions.X/2.0)
				ventPositions[idx+1] = float32(siphon.Pos.Y - cam.Pos.Y + siphon.Dimensions.Y/2.0)
				ventCount++
			}
		}

		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = c.offscreen
		op.Uniforms = map[string]any{
			"Time":          float32(g.GetTicks()),
			"VentPositions": ventPositions,
			"VentCount":     ventCount,
			"SurfaceY":      float32(-cam.Pos.Y),
		}
		finalScreen.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.WaterDisplacementShader, op)
	} else {
		finalScreen.DrawImage(c.offscreen, nil)
	}
}

func (c *CaveScene) drawScene(g CaveContext, screen *ebiten.Image, activeCave cave.Cave, caveGrid [][]bool, nodes []resource.Resource, entities []entity.CaveEntity, trenchKey string, camX, camY float64, hidePlayer bool) {
	maxDepth := 6000.0
	if caveGrid != nil && len(caveGrid[0]) > 0 {
		maxDepth = float64(len(caveGrid[0]) * config.TileSize)
	}
	mult := GetOverworldLightMultiplier(g.GetTimeOfDay())
	if activeCave != nil {
		activeCave.DrawBackground(screen, camY, maxDepth, mult)
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
		if activeCave != nil && activeCave.GetCaveType() == cave.CaveVoid {
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
		if activeCave != nil && activeCave.GetCaveType() == cave.CaveVoid {
			lineColor = color.RGBA{2, 3, 6, 255}
		} else if c.IsShallow {
			lineColor = color.RGBA{220, 240, 255, 255}
		} else {
			lineColor = color.RGBA{30, 80, 160, 255}
		}
		vector.StrokeLine(screen, 0, surfaceY, float32(config.ScreenWidth), surfaceY, 3.0, lineColor, false)
	}

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

	if caveGrid != nil && activeCave != nil {
		gridW := len(caveGrid)
		gridH := len(caveGrid[0])
		startTileX := max(int(camX)/config.TileSize, 0)
		endTileX := min((int(camX)+config.ScreenWidth)/config.TileSize+1, gridW)
		startTileY := max(int(camY)/config.TileSize, 0)
		endTileY := min((int(camY)+config.ScreenHeight)/config.TileSize+1, gridH)
		activeCave.DrawTiles(screen, camX, camY, startTileX, startTileY, endTileX, endTileY)
	}

	for _, node := range nodes {
		node.Draw(screen, camX, camY)
	}

	for _, v := range g.GetCaveVehicles(trenchKey) {
		if hidePlayer && v == g.GetActiveVehicle() {
			continue
		}
		v.Draw(screen, camX, camY)
	}

	p := g.GetPlayer()
	isPiloting := g.GetActiveVehicle() != nil

	if !hidePlayer && !isPiloting {
		pX := float32(p.Pos.X + p.Width/2.0 - camX)
		pY := float32(p.Pos.Y + p.Height/2.0 - camY)
		c.drawPlayer(screen, p, pX, pY)
	}

	mockCam := &camera.Camera{}
	mockCam.Pos.X = camX
	mockCam.Pos.Y = camY
	for _, ent := range entities {
		ent.Draw(screen, mockCam, g.GetTimeOfDay())
	}
}

func (c *CaveScene) drawBioluminescence(g CaveContext, screen *ebiten.Image, camX, camY float64) {
	if c.CaveGrid == nil {
		return
	}
	gridW := len(c.CaveGrid)
	gridH := len(c.CaveGrid[0])
	startTileX := max(int(camX)/config.TileSize, 0)
	endTileX := min((int(camX)+config.ScreenWidth)/config.TileSize+1, gridW)
	startTileY := max(int(camY)/config.TileSize, 0)
	endTileY := min((int(camY)+config.ScreenHeight)/config.TileSize+1, gridH)

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
