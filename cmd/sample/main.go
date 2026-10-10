package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/assets"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	tileSize     = 64
	gridW        = 20
	gridH        = 12
)

type Layer int

const (
	LayerBackground Layer = -1
	LayerMidground  Layer = 0
	LayerForeground Layer = 1
)

type Flora struct {
	TileX, TileY int
	Height       float64
	Color        color.RGBA
	SwayPhase    float64
	IsKelp       bool
	Layer        Layer
}

type OreNode struct {
	TileX, TileY int
	Name         string
	Color        color.RGBA
	Highlight    color.RGBA
}

type Player struct {
	X, Y       float64
	Vx, Vy     float64
	Width      float64
	Height     float64
	Facing     float64
	AnimTick   int
	Flashlight bool
}

type SampleGame struct {
	caveGrid [gridW][gridH]bool // Solid Midground (Collidable)
	bgGrid   [gridW][gridH]bool // Background Depth (Passable - swim in front)
	fgGrid   [gridW][gridH]bool // Foreground Depth (Passable - swim behind)

	diverSwimFrames []*ebiten.Image
	tileImages      map[string]*ebiten.Image
	floras          []Flora
	nodes           []OreNode
	ticks           int

	player  Player
	fgAlpha float32
}

func NewSampleGame() *SampleGame {
	g := &SampleGame{
		tileImages: make(map[string]*ebiten.Image),
		player: Player{
			X:          240, // Start in open water above the seabed
			Y:          180,
			Width:      42,
			Height:     24,
			Flashlight: true,
		},
		fgAlpha: 1.0,
	}

	g.initRoomGeometryWithTunnels()
	g.initSpawns()
	g.loadDiverSprites()
	g.preRenderTileAssets()

	return g
}

func (g *SampleGame) initRoomGeometryWithTunnels() {
	// 1. Initial Solid Rock Base
	// Ceiling at y=0, Floor at y=11, Walls at x=0 and x=19
	for x := 0; x < gridW; x++ {
		g.caveGrid[x][0] = true
		g.caveGrid[x][11] = true
	}
	for y := 0; y < gridH; y++ {
		g.caveGrid[0][y] = true
		g.caveGrid[gridW-1][y] = true
	}

	// 2. Main Seabed Surface Flow (Undulating solid ground line around y=5..6)
	// Below y=5..6 starts the solid ground mass
	for x := 1; x < gridW-1; x++ {
		// Surface slope: high on left (y=5), dips to valley at x=5 (y=6), rises at x=10 (y=5), dips at x=15 (y=6)
		fl := 6
		if x >= 8 && x <= 12 {
			fl = 5
		}
		for y := fl; y < gridH; y++ {
			g.caveGrid[x][y] = true
		}
	}

	// 3. Carve Sub-Ground Tunnels Through the Rock
	// Tunnel 1: Descent from Surface Valley 1 (x=4..5, y=6)
	g.caveGrid[4][6] = false
	g.caveGrid[5][6] = false
	g.caveGrid[4][7] = false
	g.caveGrid[5][7] = false

	// Subterranean Horizontal Loop Corridor (y=8..9, spanning x=4 through x=15)
	for x := 4; x <= 15; x++ {
		g.caveGrid[x][8] = false
		g.caveGrid[x][9] = false
	}

	// Tunnel Ascent out to Surface Valley 2 (x=14..15, y=6..7)
	g.caveGrid[14][7] = false
	g.caveGrid[15][7] = false
	g.caveGrid[14][6] = false
	g.caveGrid[15][6] = false

	// Subterranean Deep Grotto Chamber (branching right at x=16..18, y=8..10)
	for x := 16; x <= 18; x++ {
		for y := 8; y <= 10; y++ {
			g.caveGrid[x][y] = false
		}
	}

	// 4. Background Depth (BgGrid)
	// RULE: ALWAYS BACKGROUND inside the sub-ground tunnels!
	// Plus entrance feathering flaring up to the surface mouths
	for x := 1; x < gridW-1; x++ {
		for y := 1; y < gridH-1; y++ {
			if !g.caveGrid[x][y] {
				// Any open space at or below y=6 is inside the subterranean tunnel network!
				if y >= 6 {
					g.bgGrid[x][y] = true
				}
				// Surface mouth alcoves
				if y == 5 && (x == 4 || x == 5 || x == 14 || x == 15) {
					g.bgGrid[x][y] = true
				}
			}
		}
	}

	// 5. Foreground Framing (FgGrid) - removed for now per design feedback
	// (g.fgGrid remains empty/unpopulated)
}

func (g *SampleGame) initSpawns() {
	// A. High-Reward Resources in the Subterranean Grotto
	g.nodes = append(g.nodes, OreNode{
		TileX:     17,
		TileY:     10,
		Name:      "Titanium Geode (Grotto)",
		Color:     color.RGBA{220, 240, 255, 255},
		Highlight: color.RGBA{80, 210, 255, 255},
	})
	g.nodes = append(g.nodes, OreNode{
		TileX:     18,
		TileY:     9,
		Name:      "Copper Vein (Grotto)",
		Color:     color.RGBA{235, 145, 80, 255},
		Highlight: color.RGBA{255, 215, 140, 255},
	})
	g.nodes = append(g.nodes, OreNode{
		TileX:     17,
		TileY:     8,
		Name:      "ShatterBulb O2 Plant (Grotto)",
		Color:     color.RGBA{50, 245, 215, 255},
		Highlight: color.RGBA{255, 255, 255, 255},
	})

	// Surface Mineral
	g.nodes = append(g.nodes, OreNode{
		TileX:     10,
		TileY:     4,
		Name:      "Surface Quartz Outcrop",
		Color:     color.RGBA{240, 245, 255, 255},
		Highlight: color.RGBA{190, 230, 255, 255},
	})

	// B. Passive Flora & Corals on Background (Inside tunnels)
	g.floras = append(g.floras, Flora{TileX: 6, TileY: 7, Height: 45, Color: color.RGBA{26, 92, 58, 200}, SwayPhase: 0.5, IsKelp: true, Layer: LayerBackground})
	g.floras = append(g.floras, Flora{TileX: 8, TileY: 7, Height: 48, Color: color.RGBA{30, 105, 65, 200}, SwayPhase: 1.8, IsKelp: true, Layer: LayerBackground})
	g.floras = append(g.floras, Flora{TileX: 12, TileY: 7, Height: 42, Color: color.RGBA{24, 88, 55, 200}, SwayPhase: 2.9, IsKelp: true, Layer: LayerBackground})
	g.floras = append(g.floras, Flora{TileX: 7, TileY: 9, Height: 18, Color: color.RGBA{135, 65, 95, 210}, SwayPhase: 0, IsKelp: false, Layer: LayerBackground})
	g.floras = append(g.floras, Flora{TileX: 13, TileY: 9, Height: 16, Color: color.RGBA{140, 75, 110, 210}, SwayPhase: 0, IsKelp: false, Layer: LayerBackground})

	// C. Passive Flora on Solid Midground (Surface & tunnel ledges)
	g.floras = append(g.floras, Flora{TileX: 2, TileY: 4, Height: 55, Color: color.RGBA{45, 180, 70, 255}, SwayPhase: 0.8, IsKelp: true, Layer: LayerMidground})
	g.floras = append(g.floras, Flora{TileX: 8, TileY: 4, Height: 50, Color: color.RGBA{40, 165, 65, 255}, SwayPhase: 2.7, IsKelp: true, Layer: LayerMidground})
	g.floras = append(g.floras, Flora{TileX: 12, TileY: 4, Height: 52, Color: color.RGBA{48, 190, 75, 255}, SwayPhase: 4.1, IsKelp: true, Layer: LayerMidground})
	g.floras = append(g.floras, Flora{TileX: 9, TileY: 4, Height: 20, Color: color.RGBA{240, 95, 135, 255}, SwayPhase: 0, IsKelp: false, Layer: LayerMidground})

	// D. Passive Corals on Foreground Arch & Crossing Pillar
	g.floras = append(g.floras, Flora{TileX: 4, TileY: 4, Height: 16, Color: color.RGBA{185, 45, 90, 255}, SwayPhase: 0, IsKelp: false, Layer: LayerForeground})
	g.floras = append(g.floras, Flora{TileX: 14, TileY: 4, Height: 18, Color: color.RGBA{210, 60, 110, 255}, SwayPhase: 0, IsKelp: false, Layer: LayerForeground})
	g.floras = append(g.floras, Flora{TileX: 10, TileY: 7, Height: 14, Color: color.RGBA{170, 40, 85, 255}, SwayPhase: 0, IsKelp: false, Layer: LayerForeground})
}

func (g *SampleGame) loadDiverSprites() {
	swimSheet, err := assets.LoadChromaKeyedImage("diver_swim_sheet")
	if err == nil {
		colRanges := [][2]int{
			{10, 244},
			{266, 500},
			{522, 756},
			{779, 1013},
		}
		rowRanges := [][2]int{
			{10, 250},
			{260, 500},
		}
		for _, rRange := range rowRanges {
			for _, cRange := range colRanges {
				rect := image.Rect(cRange[0], rRange[0], cRange[1], rRange[1])
				frame := ebiten.NewImageFromImage(swimSheet.SubImage(rect))
				g.diverSwimFrames = append(g.diverSwimFrames, frame)
			}
		}
	} else {
		log.Printf("Using fallback procedural diver graphics: %v", err)
	}
}

func (g *SampleGame) preRenderTileAssets() {
	// 1. Background Tile (recessed, darker, tinted navy-slate)
	bgImg := ebiten.NewImage(tileSize, tileSize)
	vector.FillRect(bgImg, 0, 0, tileSize, tileSize, color.RGBA{18, 34, 52, 255}, false)
	for i := 4; i < tileSize; i += 12 {
		vector.FillCircle(bgImg, float32(i), float32(i), 3.0, color.RGBA{14, 26, 42, 255}, false)
	}
	g.tileImages["bg"] = bgImg

	// 2. Solid Midground Tile (full vibrance, crisp boundary)
	solidImg := ebiten.NewImage(tileSize, tileSize)
	vector.FillRect(solidImg, 0, 0, tileSize, tileSize, color.RGBA{58, 72, 88, 255}, false)
	vector.StrokeRect(solidImg, 0, 0, tileSize, tileSize, 1.5, color.RGBA{38, 48, 60, 255}, false)
	for i := 6; i < tileSize-6; i += 14 {
		vector.FillRect(solidImg, float32(i), float32(i), 3, 2, color.RGBA{82, 98, 118, 255}, false)
	}
	g.tileImages["solid"] = solidImg

	// 3. Foreground Tile (dark silhouette with ambient rim highlight)
	fgImg := ebiten.NewImage(tileSize, tileSize)
	vector.FillRect(fgImg, 0, 0, tileSize, tileSize, color.RGBA{10, 14, 22, 255}, false)
	vector.StrokeRect(fgImg, 0, 0, tileSize, tileSize, 1.0, color.RGBA{28, 52, 72, 255}, false)
	g.tileImages["fg"] = fgImg
}

func (g *SampleGame) Update() error {
	g.ticks++

	// 1. Player Movement & Input
	accel := 0.32
	maxSpeed := 5.0

	var inputX, inputY float64
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		inputX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		inputX += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		inputY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		inputY += 1
	}

	if inputX != 0 || inputY != 0 {
		mag := math.Hypot(inputX, inputY)
		inputX /= mag
		inputY /= mag
		g.player.Vx += inputX * accel
		g.player.Vy += inputY * accel
		g.player.Facing = math.Atan2(g.player.Vy, g.player.Vx)
		g.player.AnimTick++
	}

	g.player.Vx *= 0.94
	g.player.Vy *= 0.94

	speed := math.Hypot(g.player.Vx, g.player.Vy)
	if speed > maxSpeed {
		g.player.Vx = (g.player.Vx / speed) * maxSpeed
		g.player.Vy = (g.player.Vy / speed) * maxSpeed
	}

	// 2. Collision with Solid Midground ONLY (CaveGrid)
	newX := g.player.X + g.player.Vx
	if !g.isSolid(newX, g.player.Y, g.player.Width, g.player.Height) {
		g.player.X = newX
	} else {
		g.player.Vx = 0
	}

	newY := g.player.Y + g.player.Vy
	if !g.isSolid(g.player.X, newY, g.player.Width, g.player.Height) {
		g.player.Y = newY
	} else {
		g.player.Vy = 0
	}

	// 3. Toggle Flashlight with [F]
	if ebiten.IsKeyPressed(ebiten.KeyF) && g.ticks%15 == 0 {
		g.player.Flashlight = !g.player.Flashlight
	}

	// 4. Reset with [R]
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.player.X = 240
		g.player.Y = 180
		g.player.Vx = 0
		g.player.Vy = 0
	}

	// 5. Evaluate Proximity to Foreground Arch/Pillar for Smooth Alpha Fading
	playerCenterX := g.player.X + g.player.Width/2
	playerCenterY := g.player.Y + g.player.Height/2
	ptx := int(playerCenterX) / tileSize
	pty := int(playerCenterY) / tileSize

	isNearFG := false
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			tx := ptx + dx
			ty := pty + dy
			if tx >= 0 && tx < gridW && ty >= 0 && ty < gridH && g.fgGrid[tx][ty] {
				isNearFG = true
				break
			}
		}
	}

	targetAlpha := float32(1.0)
	if isNearFG {
		targetAlpha = 0.38
	}
	g.fgAlpha += (targetAlpha - g.fgAlpha) * 0.15

	return nil
}

func (g *SampleGame) isSolid(x, y, w, h float64) bool {
	minTx := int(x) / tileSize
	maxTx := int(x+w) / tileSize
	minTy := int(y) / tileSize
	maxTy := int(y+h) / tileSize

	for tx := minTx; tx <= maxTx; tx++ {
		for ty := minTy; ty <= maxTy; ty++ {
			if tx < 0 || tx >= gridW || ty < 0 || ty >= gridH {
				return true
			}
			if g.caveGrid[tx][ty] {
				return true
			}
		}
	}
	return false
}

func (g *SampleGame) currentLayerName() (string, color.RGBA) {
	pcx := g.player.X + g.player.Width/2
	pcy := g.player.Y + g.player.Height/2
	tx := int(pcx) / tileSize
	ty := int(pcy) / tileSize

	if tx >= 0 && tx < gridW && ty >= 0 && ty < gridH {
		if g.fgGrid[tx][ty] {
			return "BEHIND FOREGROUND ARCH/PILLAR (FgGrid — Swim Behind)", color.RGBA{255, 90, 160, 255}
		}
		if g.bgGrid[tx][ty] {
			if ty >= 8 {
				return "INSIDE SUBTERRANEAN TUNNEL (Always BgGrid — Swim In Front)", color.RGBA{60, 220, 255, 255}
			}
			return "IN FRONT OF ENTRANCE MOUTH (Passable BgGrid)", color.RGBA{70, 200, 240, 255}
		}
	}
	return "OPEN WATER (Surface Seabed Channel)", color.RGBA{140, 240, 160, 255}
}

func (g *SampleGame) Draw(screen *ebiten.Image) {
	// 1. Water Ocean Base
	screen.Fill(color.RGBA{10, 22, 38, 255})

	// 2. LAYER 1: BACKGROUND DEPTH TILES (BgGrid - always inside tunnels)
	op := &ebiten.DrawImageOptions{}
	for tx := 0; tx < gridW; tx++ {
		for ty := 0; ty < gridH; ty++ {
			if g.bgGrid[tx][ty] {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(tx*tileSize), float64(ty*tileSize))
				screen.DrawImage(g.tileImages["bg"], op)
			}
		}
	}

	for _, f := range g.floras {
		if f.Layer == LayerBackground {
			g.drawFlora(screen, f)
		}
	}

	// 3. LAYER 2: SOLID MIDGROUND TILES (CaveGrid)
	for tx := 0; tx < gridW; tx++ {
		for ty := 0; ty < gridH; ty++ {
			if g.caveGrid[tx][ty] {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(tx*tileSize), float64(ty*tileSize))
				screen.DrawImage(g.tileImages["solid"], op)

				if ty > 0 && !g.caveGrid[tx][ty-1] {
					vector.FillRect(screen, float32(tx*tileSize), float32(ty*tileSize), tileSize, 4, color.RGBA{195, 168, 120, 255}, false)
					vector.FillRect(screen, float32(tx*tileSize), float32(ty*tileSize)+4, tileSize, 2, color.RGBA{135, 110, 80, 255}, false)
				}
			}
		}
	}

	for _, n := range g.nodes {
		g.drawOreNode(screen, n)
	}
	for _, f := range g.floras {
		if f.Layer == LayerMidground {
			g.drawFlora(screen, f)
		}
	}

	// 4. PLAYER / DIVER (Swims in front of BgGrid, behind FgGrid)
	g.drawPlayer(screen)

	if g.player.Flashlight {
		g.drawFlashlight(screen)
	}

	// 5. LAYER 3: FOREGROUND TILES (FgGrid - arches & crossing pillars)
	fgOp := &ebiten.DrawImageOptions{}
	fgOp.ColorScale.Scale(1.0, 1.0, 1.0, g.fgAlpha)

	for tx := 0; tx < gridW; tx++ {
		for ty := 0; ty < gridH; ty++ {
			if g.fgGrid[tx][ty] {
				fgOp.GeoM.Reset()
				fgOp.GeoM.Translate(float64(tx*tileSize), float64(ty*tileSize))
				screen.DrawImage(g.tileImages["fg"], fgOp)

				if ty > 0 && !g.fgGrid[tx][ty-1] {
					rimAlpha := uint8(float32(255) * g.fgAlpha)
					vector.FillRect(screen, float32(tx*tileSize), float32(ty*tileSize), tileSize, 2, color.RGBA{45, 95, 125, rimAlpha}, false)
				}
			}
		}
	}

	for _, f := range g.floras {
		if f.Layer == LayerForeground {
			g.drawFlora(screen, f)
		}
	}

	// 6. HUD OVERLAY
	g.drawHUD(screen)
}

func (g *SampleGame) drawPlayer(screen *ebiten.Image) {
	pX := float32(g.player.X + g.player.Width/2)
	pY := float32(g.player.Y + g.player.Height/2)
	facingAngle := g.player.Facing

	if len(g.diverSwimFrames) > 0 {
		frameIdx := (g.player.AnimTick / 6) % len(g.diverSwimFrames)
		frame := g.diverSwimFrames[frameIdx]

		op := &ebiten.DrawImageOptions{}
		b := frame.Bounds()
		fw := float64(b.Dx())
		fh := float64(b.Dy())

		op.GeoM.Translate(-fw/2, -fh/2)
		facingLeft := math.Cos(facingAngle) < 0
		if facingLeft {
			op.GeoM.Scale(-1, 1)
		}

		if g.player.Vy > 0.4 {
			if facingLeft {
				op.GeoM.Rotate(-0.4)
			} else {
				op.GeoM.Rotate(0.4)
			}
		} else if g.player.Vy < -0.4 {
			if facingLeft {
				op.GeoM.Rotate(0.4)
			} else {
				op.GeoM.Rotate(-0.4)
			}
		}

		scale := 52.0 / fw
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(pX), float64(pY))
		screen.DrawImage(frame, op)
	} else {
		vector.FillCircle(screen, pX, pY, 12, color.RGBA{40, 140, 215, 255}, false)
		vx := pX + float32(math.Cos(facingAngle))*8
		vy := pY + float32(math.Sin(facingAngle))*8
		vector.FillCircle(screen, vx, vy, 6, color.RGBA{110, 235, 255, 255}, false)
		tx := pX - float32(math.Cos(facingAngle))*8
		ty := pY - float32(math.Sin(facingAngle))*8
		vector.FillCircle(screen, tx, ty, 7, color.RGBA{245, 210, 45, 255}, false)
	}
}

func (g *SampleGame) drawFlashlight(screen *ebiten.Image) {
	pX := float32(g.player.X + g.player.Width/2)
	pY := float32(g.player.Y + g.player.Height/2)
	angle := g.player.Facing

	beamLength := 240.0
	beamSpread := 0.42

	numSegments := 16
	for i := 0; i < numSegments; i++ {
		t1 := float64(i) / float64(numSegments)
		t2 := float64(i+1) / float64(numSegments)
		a1 := angle - beamSpread/2 + t1*beamSpread
		a2 := angle - beamSpread/2 + t2*beamSpread

		x1 := pX + float32(math.Cos(a1)*beamLength)
		y1 := pY + float32(math.Sin(a1)*beamLength)
		x2 := pX + float32(math.Cos(a2)*beamLength)
		y2 := pY + float32(math.Sin(a2)*beamLength)

		lightClr := color.RGBA{190, 240, 255, 18}
		var path vector.Path
		path.MoveTo(pX, pY)
		path.LineTo(x1, y1)
		path.LineTo(x2, y2)
		path.Close()
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		for idx := range vs {
			vs[idx].ColorR = float32(lightClr.R) / 255
			vs[idx].ColorG = float32(lightClr.G) / 255
			vs[idx].ColorB = float32(lightClr.B) / 255
			vs[idx].ColorA = float32(lightClr.A) / 255
		}
		empty := ebiten.NewImage(3, 3)
		empty.Fill(color.White)
		opts := &ebiten.DrawTrianglesOptions{Blend: ebiten.BlendLighter}
		screen.DrawTriangles(vs, is, empty, opts)
	}
}

func (g *SampleGame) drawFlora(screen *ebiten.Image, f Flora) {
	baseX := float32(f.TileX*tileSize + tileSize/2)
	baseY := float32(f.TileY*tileSize + tileSize)

	if f.IsKelp {
		swayOffset := float32(math.Sin(float64(g.ticks)*0.035+f.SwayPhase) * 12.0)
		for h := float32(0); h < float32(f.Height); h += 3 {
			frac := h / float32(f.Height)
			currSway := swayOffset * frac * frac
			vector.FillCircle(screen, baseX+currSway, baseY-h, 2.5, f.Color, false)
			if int(h)%9 == 0 {
				leafSide := float32(1)
				if int(h)%18 == 0 {
					leafSide = -1
				}
				vector.StrokeLine(screen, baseX+currSway, baseY-h, baseX+currSway+leafSide*8, baseY-h-3, 1.8, f.Color, false)
			}
		}
	} else {
		cSize := float32(f.Height)
		vector.FillCircle(screen, baseX, baseY-cSize/2, cSize/2, f.Color, false)
		vector.FillCircle(screen, baseX-cSize/3, baseY-cSize/3, cSize/3, f.Color, false)
		vector.FillCircle(screen, baseX+cSize/3, baseY-cSize/3, cSize/3, f.Color, false)
	}
}

func (g *SampleGame) drawOreNode(screen *ebiten.Image, n OreNode) {
	nx := float32(n.TileX*tileSize + tileSize/2)
	ny := float32(n.TileY*tileSize + tileSize - 16)

	vector.FillCircle(screen, nx, ny, 12, n.Color, false)
	vector.FillCircle(screen, nx, ny, 6, n.Highlight, false)
	vector.StrokeCircle(screen, nx, ny, 12, 1.5, color.RGBA{20, 30, 45, 255}, false)

	dist := math.Hypot(g.player.X+g.player.Width/2-float64(nx), g.player.Y+g.player.Height/2-float64(ny))
	if dist < 85.0 {
		prompt := fmt.Sprintf("[%s]", n.Name)
		ebitenutil.DebugPrintAt(screen, prompt, int(nx)-len(prompt)*3, int(ny)-24)
	}
}

func (g *SampleGame) drawHUD(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, screenWidth, 48, color.RGBA{8, 14, 24, 220}, false)
	vector.StrokeLine(screen, 0, 48, screenWidth, 48, 1.5, color.RGBA{45, 90, 130, 255}, false)

	layerText, layerCol := g.currentLayerName()
	vector.FillRect(screen, 24, 8, 540, 32, color.RGBA{16, 26, 44, 255}, false)
	vector.StrokeRect(screen, 24, 8, 540, 32, 1.2, layerCol, false)
	ebitenutil.DebugPrintAt(screen, layerText, 36, 16)

	status := fmt.Sprintf("Controls: [WASD / Arrows] Swim  |  [F] Flashlight (%t)  |  [R] Reset  |  FPS: %0.1f", g.player.Flashlight, ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(screen, status, 590, 16)

	vector.FillRect(screen, screenWidth-380, screenHeight-85, 365, 75, color.RGBA{10, 16, 28, 200}, false)
	vector.StrokeRect(screen, screenWidth-380, screenHeight-85, 365, 75, 1.0, color.RGBA{40, 70, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "SUB-GROUND TUNNELS & DEPTH DEMO:", screenWidth-370, screenHeight-80)
	ebitenutil.DebugPrintAt(screen, "* Background (Navy): ALWAYS in tunnels (swim in front)", screenWidth-370, screenHeight-62)
	ebitenutil.DebugPrintAt(screen, "* Midground (Gray-Sand): SOLID terrain (blocks diver)", screenWidth-370, screenHeight-46)
	ebitenutil.DebugPrintAt(screen, "* Foreground (Arch/Pillar): In/around tunnels (fade)", screenWidth-370, screenHeight-30)
}

func (g *SampleGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("SubGame — Playable Sub-Ground Tunnels & Depth Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewSampleGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Game exited with error: %v", err)
	}
}
