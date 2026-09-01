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
		c.applyLighting(g)
	} else {
		cam := g.GetCamera()
		c.drawScene(g, c.offscreen, c.ActiveCave, c.CaveGrid, c.Nodes, c.Entities, g.GetActiveTrenchKey(), cam.Pos.X, cam.Pos.Y, false)
		c.applyLighting(g)
		if !c.IsShallow {
			c.drawBioluminescence(g, c.offscreen, cam.Pos.X, cam.Pos.Y)
		}
	}

	c.applyPostFX(g, finalScreen)
}

// Edge-blur strength vs depth (0 = surface, 1 = cave floor).
// This is mix amount for real scene defocus near the frame — not a dark mask.
const (
	caveEdgeBlurNear      = 0.45 // stronger soft focus near the surface
	caveEdgeBlurDeep      = 1.0  // full mix deep
	caveEdgeBlurDepthSat  = 0.40 // full strength a bit earlier
	caveEdgeBlurFalloffPx = 280.0
	caveEdgeBlurMaxRadius = 11.0 // wider sample radius = more obvious blur
)

// applyPostFX runs water displacement then soft edge defocus into finalScreen.
func (c *CaveScene) applyPostFX(g CaveContext, finalScreen *ebiten.Image) {
	src := c.offscreen
	useEdgeBlur := shader.EdgeBlurShader != nil

	// If we need edge blur after water, water must land in an intermediate.
	waterTarget := finalScreen
	if useEdgeBlur {
		if c.postFX == nil {
			c.postFX = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
		}
		c.postFX.Clear()
		waterTarget = c.postFX
	}

	cam := g.GetCamera()
	if shader.WaterDisplacementShader != nil && !g.IsDebugWaterShaderDisabled() {
		var ventPositions [16]float32
		var ventCount float32 = 0
		for _, ent := range c.Entities {
			if siphon, ok := ent.(*entity.BrimstoneSiphon); ok && siphon.IsActive() && siphon.Timer >= entity.BrimstoneSiphonArchetype.ActiveStartFrame && ventCount < 8 {
				idx := int(ventCount) * 2
				ventPositions[idx] = float32(siphon.Pos.X - cam.Pos.X + siphon.Dimensions.X/2.0)
				ventPositions[idx+1] = float32(siphon.Pos.Y - cam.Pos.Y + siphon.Dimensions.Y/2.0)
				ventCount++
			}
		}

		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = src
		op.Uniforms = map[string]any{
			"Time":          float32(g.GetTicks()),
			"VentPositions": ventPositions,
			"VentCount":     ventCount,
			"SurfaceY":      float32(-cam.Pos.Y),
		}
		waterTarget.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.WaterDisplacementShader, op)
	} else {
		waterTarget.DrawImage(src, nil)
	}

	if !useEdgeBlur {
		c.drawInkScreenVignette(g, finalScreen)
		return
	}

	// Depth-driven strength: soft near surface, more locked-in deep.
	maxDepth := 6000.0
	if len(c.CaveGrid) > 0 && len(c.CaveGrid[0]) > 0 {
		maxDepth = float64(len(c.CaveGrid[0]) * config.TileSize)
	}
	depth := g.GetPlayer().Pos.Y
	if depth < 0 {
		depth = 0
	}
	depthFrac := depth / maxDepth
	if depthFrac > 1 {
		depthFrac = 1
	}
	d := depthFrac / caveEdgeBlurDepthSat
	if d > 1 {
		d = 1
	}
	d = math.Sqrt(d)
	strength := caveEdgeBlurNear + d*(caveEdgeBlurDeep-caveEdgeBlurNear)

	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = c.postFX
	op.Uniforms = map[string]any{
		"FalloffPx": float32(caveEdgeBlurFalloffPx),
		"MaxBlurPx": float32(caveEdgeBlurMaxRadius),
		"Strength":  float32(strength),
	}
	finalScreen.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.EdgeBlurShader, op)
	c.drawInkScreenVignette(g, finalScreen)
}

func (c *CaveScene) drawInkScreenVignette(g CaveContext, screen *ebiten.Image) {
	p := g.GetPlayer()
	if p == nil || p.SlowTimer <= 0 {
		return
	}

	alpha := float64(p.SlowTimer) / 180.0
	if alpha > 1.0 {
		alpha = 1.0
	}

	w := float32(config.ScreenWidth)
	h := float32(config.ScreenHeight)

	// Soft, layered translucent ink colors with gentle, non-harsh contrast
	outerHaze := color.RGBA{22, 25, 45, uint8(alpha * 38)}
	midPlume := color.RGBA{18, 20, 36, uint8(alpha * 65)}
	innerInk := color.RGBA{14, 15, 28, uint8(alpha * 92)}

	// Helper to draw a feathered, organic soft ink splotch
	drawInkSpot := func(cx, cy, baseRad float32) {
		r := baseRad * float32(alpha)
		if r <= 1.0 {
			return
		}
		// Soft outer gradient fringe
		vector.FillCircle(screen, cx, cy, r*1.35, outerHaze, false)
		// Mid translucent ink body
		vector.FillCircle(screen, cx, cy, r, midPlume, false)
		// Dense core
		vector.FillCircle(screen, cx-r*0.1, cy-r*0.1, r*0.65, innerInk, false)
	}

	// 1. Broad soft corner blooms (significantly enlarged for atmospheric vignette)
	cornerSize := float32(340.0)
	drawInkSpot(0, 0, cornerSize)
	drawInkSpot(w, 0, cornerSize)
	drawInkSpot(0, h, cornerSize)
	drawInkSpot(w, h, cornerSize)

	// 2. Large border blooms replacing harsh rectangular strips
	edgeSpots := []struct {
		cx, cy float32
		rad    float32
	}{
		// Top border
		{w * 0.22, -10, 230}, {w * 0.50, -30, 260}, {w * 0.78, -10, 225},
		// Bottom border
		{w * 0.25, h + 10, 235}, {w * 0.52, h + 30, 265}, {w * 0.80, h + 10, 230},
		// Left border
		{-10, h * 0.35, 220}, {-25, h * 0.65, 230},
		// Right border
		{w + 10, h * 0.32, 225}, {w + 25, h * 0.68, 235},
	}
	for _, es := range edgeSpots {
		drawInkSpot(es.cx, es.cy, es.rad)
	}

	// 3. Large organic ink spots & splatters reaching across the screen
	screenSpots := []struct {
		xRatio, yRatio float32
		rad            float32
	}{
		{0.18, 0.24, 155},
		{0.12, 0.45, 135},
		{0.28, 0.16, 145},
		{0.82, 0.22, 160},
		{0.88, 0.48, 140},
		{0.72, 0.18, 150},
		{0.16, 0.76, 165},
		{0.26, 0.85, 140},
		{0.84, 0.74, 170},
		{0.74, 0.84, 145},
		// Mid-field plumes
		{0.34, 0.32, 95},
		{0.66, 0.28, 100},
		{0.38, 0.70, 98},
		{0.62, 0.68, 102},
	}
	for _, ss := range screenSpots {
		drawInkSpot(w*ss.xRatio, h*ss.yRatio, ss.rad)
	}
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

	switch c.scrollDir {
	case 1: // Scrolling right
		opOld.GeoM.Translate(-t*float64(config.ScreenWidth), 0)
		opNew.GeoM.Translate((1.0-t)*float64(config.ScreenWidth), 0)
	case -1: // Scrolling left
		opOld.GeoM.Translate(t*float64(config.ScreenWidth), 0)
		opNew.GeoM.Translate((t-1.0)*float64(config.ScreenWidth), 0)
	case 2: // Scrolling down
		opOld.GeoM.Translate(0, -t*float64(config.ScreenHeight))
		opNew.GeoM.Translate(0, (1.0-t)*float64(config.ScreenHeight))
	case -2: // Scrolling up
		opOld.GeoM.Translate(0, t*float64(config.ScreenHeight))
		opNew.GeoM.Translate(0, (t-1.0)*float64(config.ScreenHeight))
	}

	c.offscreen.DrawImage(c.offscreenOld, opOld)
	c.offscreen.DrawImage(c.offscreenNew, opNew)

	// 4. Draw the player or active vehicle at the interpolated position
	p := g.GetPlayer()
	isPiloting := g.GetActiveVehicle() != nil

	var width float64 = p.Width
	var height float64 = p.Height
	if isPiloting {
		dims := g.GetActiveVehicle().GetDimensions()
		width = dims.X
		height = dims.Y
	}

	var screenStartX, screenEndX float64
	var screenStartY, screenEndY float64

	switch c.scrollDir {
	case 1: // Scrolling right
		screenStartX = float64(config.ScreenWidth) - width
		screenEndX = 20.0
		screenStartY = p.Pos.Y - c.oldCamY
		screenEndY = screenStartY
	case -1: // Scrolling left
		screenStartX = 20.0
		screenEndX = float64(config.ScreenWidth) - width - 20.0
		screenStartY = p.Pos.Y - c.oldCamY
		screenEndY = screenStartY
	case 2: // Scrolling down (into deep Shock Kelp cave)
		screenStartX = p.Pos.X - c.oldCamX
		screenEndX = float64(config.ScreenWidth)/2.0 - width/2.0
		screenStartY = p.Pos.Y - c.oldCamY
		screenEndY = float64(config.TileSize * 2)
	case -2: // Scrolling up (back into shallow seabed)
		screenStartX = p.Pos.X - c.oldCamX
		chasmMinX, chasmMaxX, chasmTriggerY := float64(0), float64(0), float64(0)
		if chasm, ok := c.newCave.(cave.ChasmProvider); ok && chasm.HasFloorChasm() {
			chasmMinX, chasmMaxX, chasmTriggerY = chasm.GetChasmBounds()
		}
		if chasmMaxX > chasmMinX {
			screenEndX = (chasmMinX+chasmMaxX)/2.0 - width/2.0 - c.newCamX
		} else {
			screenEndX = float64(config.ScreenWidth)/2.0 - width/2.0
		}
		screenStartY = p.Pos.Y - c.oldCamY
		screenEndY = chasmTriggerY - height - 8.0 - c.newCamY
	}

	interpolatedX := (1.0-t)*screenStartX + t*screenEndX
	interpolatedY := (1.0-t)*screenStartY + t*screenEndY

	if isPiloting {
		v := g.GetActiveVehicle()
		oldPos := v.GetPos()
		v.SetPos(gvec.Vec2{X: interpolatedX, Y: interpolatedY})
		v.Draw(c.offscreen, 0, 0)
		v.SetPos(oldPos)
	} else {
		pX := float32(interpolatedX + p.Width/2.0)
		pY := float32(interpolatedY + p.Height/2.0)
		c.drawPlayer(c.offscreen, p, pX, pY)
	}
}

func (c *CaveScene) drawPlayer(screen *ebiten.Image, p *player.Player, pX, pY float32) {
	facingAngle := p.Facing

	var activeFrame *ebiten.Image
	if c.diverSheet != nil || len(c.diverSwimFrames) > 0 {
		if p.IsMining {
			elapsed := 24 - p.MiningAnimTimer
			numMineFrames := len(c.diverMineFrames)
			if numMineFrames > 0 {
				frameIdx := elapsed * numMineFrames / 24
				if frameIdx < 0 {
					frameIdx = 0
				}
				if frameIdx >= numMineFrames {
					frameIdx = numMineFrames - 1
				}
				activeFrame = c.diverMineFrames[frameIdx]
			}
		} else if math.Hypot(p.Vel.X, p.Vel.Y) > 0.2 {
			if len(c.diverSwimFrames) > 0 {
				frameIdx := (p.AnimTick / 5) % len(c.diverSwimFrames)
				activeFrame = c.diverSwimFrames[frameIdx]
			}
		} else {
			if len(c.diverIdleFrames) > 0 {
				frameIdx := (p.AnimTick / 15) % len(c.diverIdleFrames)
				activeFrame = c.diverIdleFrames[frameIdx]
			}
		}
	}

	if activeFrame != nil {
		op := &ebiten.DrawImageOptions{}
		b := activeFrame.Bounds()
		baseFrameW := float64(b.Dx())
		baseFrameH := float64(b.Dy())
		minX := float64(b.Min.X)
		minY := float64(b.Min.Y)

		op.GeoM.Translate(-minX-baseFrameW/2.0, -minY-baseFrameH/2.0)
		facingLeft := math.Cos(facingAngle) < 0
		if facingLeft {
			op.GeoM.Scale(-1, 1)
		}
		if p.Vel.Y > 0.2 && p.Vel.Y > math.Abs(p.Vel.X)*0.8 {
			if facingLeft {
				op.GeoM.Rotate(-math.Pi / 2.0)
			} else {
				op.GeoM.Rotate(math.Pi / 2.0)
			}
		}
		scale := DiverDrawWidth / baseFrameW
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(pX), float64(pY))

		if p.IsDamaged {
			if (p.DamageAnimTimer/4)%2 == 0 {
				op.ColorScale.Scale(1.35, 0.70, 0.70, 0.75)
			}
		}

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
	if shader.LightShader == nil || g.IsDebugLightShaderDisabled() {
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
	ambient := c.getAmbientColor(g.GetTimeOfDay())
	if c.ActiveCave != nil && c.ActiveCave.GetCaveType() == cave.CaveOrganicTrench {
		// Deep oceanic darkness (0.91 at top, deepening to 0.96 near the bottom)
		ambient[3] = float32(0.91 + depthFrac*0.05)
	}
	c.Uniforms["AmbientColor"] = ambient
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

	var biolumPositions [64]float32
	var biolumColors [64]float32
	var biolumCount float32 = 0
	for _, ent := range c.Entities {
		if !ent.IsActive() || biolumCount >= 16 {
			continue
		}
		if emitter, ok := ent.(entity.PointLightEmitter); ok {
			pos, rad, r, g, b, intensity := emitter.PointLight()
			if rad > 0 && intensity > 0 {
				idx4 := int(biolumCount) * 4
				biolumPositions[idx4] = float32(pos.X - cam.Pos.X)
				biolumPositions[idx4+1] = float32(pos.Y - cam.Pos.Y)
				biolumPositions[idx4+2] = float32(rad)
				biolumPositions[idx4+3] = float32(intensity)

				biolumColors[idx4] = r
				biolumColors[idx4+1] = g
				biolumColors[idx4+2] = b
				biolumColors[idx4+3] = 1.0
				biolumCount++
			}
		}
	}
	c.Uniforms["BiolumPositions"] = biolumPositions
	c.Uniforms["BiolumColors"] = biolumColors
	c.Uniforms["BiolumCount"] = biolumCount

	c.offscreen.DrawRectShader(config.ScreenWidth, config.ScreenHeight, shader.LightShader, &c.shaderOpts)
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

	isSurfaceCave := c.IsShallow && (activeCave == nil || activeCave.GetCaveType() == cave.CaveOrganicShallow)

	if camY < 0 {
		var skyColor color.RGBA
		if isSurfaceCave {
			skyColor = getSkyColor(g.GetTimeOfDay())
		} else if activeCave != nil && activeCave.GetCaveType() == cave.CaveShockKelp {
			skyColor = color.RGBA{15, 12, 22, 255}
		} else if activeCave != nil && activeCave.GetCaveType() == cave.CaveVoid {
			skyColor = color.RGBA{2, 3, 6, 255}
		} else {
			skyColor = color.RGBA{10, 8, 16, 255}
		}
		vector.FillRect(screen, 0, 0, float32(config.ScreenWidth), float32(-camY), skyColor, false)
	}

	if isSurfaceCave {
		surfaceY := float32(-camY)
		if surfaceY >= 0 && surfaceY < float32(config.ScreenHeight) {
			lineColor := color.RGBA{220, 240, 255, 255}
			vector.StrokeLine(screen, 0, surfaceY, float32(config.ScreenWidth), surfaceY, 3.0, lineColor, false)
		}
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
