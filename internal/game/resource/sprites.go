package resource

import (
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/assets"
)

var (
	TitaniumSprite *ebiten.Image
	CopperSprite   *ebiten.Image
	QuartzSprite   *ebiten.Image
	AbyssalSprite  *ebiten.Image
	spritesLoaded  bool
	gPath          vector.Path
)

// LoadAssets preloads and chroma-keys all resource crystal sprites.
func LoadAssets() {
	sheet, err := assets.LoadChromaKeyedImage("ore_sheet")
	if err != nil {
		log.Printf("Error: Failed to load ore sheet: %v", err)
		return
	}

	bounds := sheet.Bounds()

	if bounds.Dx() == 3584 && bounds.Dy() == 1184 {
		// Specific coordinate slice for the user's high-res generated ore sheet
		TitaniumSprite = sheet.SubImage(image.Rect(832, 330, 1312, 856)).(*ebiten.Image)
		CopperSprite = sheet.SubImage(image.Rect(1327, 330, 1809, 856)).(*ebiten.Image)
		QuartzSprite = sheet.SubImage(image.Rect(1827, 330, 2308, 856)).(*ebiten.Image)
		AbyssalSprite = sheet.SubImage(image.Rect(2328, 330, 2790, 856)).(*ebiten.Image)
	} else {
		// General fallback: slice into 4 equal columns
		w := bounds.Dx() / 4
		h := bounds.Dy()
		TitaniumSprite = sheet.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
		CopperSprite = sheet.SubImage(image.Rect(w, 0, w*2, h)).(*ebiten.Image)
		QuartzSprite = sheet.SubImage(image.Rect(w*2, 0, w*3, h)).(*ebiten.Image)
		AbyssalSprite = sheet.SubImage(image.Rect(w*3, 0, w*4, h)).(*ebiten.Image)
	}
	spritesLoaded = true
}

// drawNodeBase renders the shared backing block behind all resource nodes.
func drawNodeBase(screen *ebiten.Image, tx, ty int, camX, camY float64) (float32, float32) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))
	vector.FillRect(screen, sx, sy, TileSize, TileSize, color.RGBA{25, 22, 30, 255}, false)
	vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, color.RGBA{45, 40, 52, 255}, false)
	return sx, sy
}

// Helper to darken a color
func darkenColor(c color.Color, factor float32) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(float32(r>>8) * factor),
		G: uint8(float32(g>>8) * factor),
		B: uint8(float32(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

// Helper to blend two colors
func blendColor(c1, c2 color.Color, t float32) color.RGBA {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return color.RGBA{
		R: uint8((1.0-t)*float32(r1>>8) + t*float32(r2>>8)),
		G: uint8((1.0-t)*float32(g1>>8) + t*float32(g2>>8)),
		B: uint8((1.0-t)*float32(b1>>8) + t*float32(b2>>8)),
		A: uint8((1.0-t)*float32(a1>>8) + t*float32(a2>>8)),
	}
}

// rotateVec rotates a 2D vector by an angle in radians
func rotateVec(v [2]float32, angle float32) [2]float32 {
	cosA := float32(math.Cos(float64(angle)))
	sinA := float32(math.Sin(float64(angle)))
	return [2]float32{
		v[0]*cosA - v[1]*sinA,
		v[0]*sinA + v[1]*cosA,
	}
}

// localToScreen transforms a local point (lx, ly) to screen space based on growth basis
func localToScreen(cx, cy float32, lx, ly float32, dirVec, perpVec [2]float32) (float32, float32) {
	return cx + lx*perpVec[0] + ly*dirVec[0], cy + lx*perpVec[1] + ly*dirVec[1]
}

// drawShard draws a single 3D crystal shard growing in local space
func drawShard(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, length, width float32, shadowColor, highlightColor color.Color) {
	// Left Face
	gPath.Reset()
	lx0, ly0 := localToScreen(cx, cy, 0, 0, dirVec, perpVec)
	lx1, ly1 := localToScreen(cx, cy, -width/2, 0, dirVec, perpVec)
	lx2, ly2 := localToScreen(cx, cy, -width/2, length*0.4, dirVec, perpVec)
	lx3, ly3 := localToScreen(cx, cy, 0, length, dirVec, perpVec)
	lx4, ly4 := localToScreen(cx, cy, 0, length*0.45, dirVec, perpVec)

	gPath.MoveTo(lx0, ly0)
	gPath.LineTo(lx1, ly1)
	gPath.LineTo(lx2, ly2)
	gPath.LineTo(lx3, ly3)
	gPath.LineTo(lx4, ly4)
	gPath.Close()
	var shadowOpts vector.DrawPathOptions
	shadowOpts.ColorScale.ScaleWithColor(shadowColor)
	vector.FillPath(screen, &gPath, nil, &shadowOpts)

	// Right Face
	gPath.Reset()
	rx0, ry0 := localToScreen(cx, cy, 0, 0, dirVec, perpVec)
	rx1, ry1 := localToScreen(cx, cy, 0, length*0.45, dirVec, perpVec)
	rx2, ry2 := localToScreen(cx, cy, 0, length, dirVec, perpVec)
	rx3, ry3 := localToScreen(cx, cy, width/2, length*0.4, dirVec, perpVec)
	rx4, ry4 := localToScreen(cx, cy, width/2, 0, dirVec, perpVec)

	gPath.MoveTo(rx0, ry0)
	gPath.LineTo(rx1, ry1)
	gPath.LineTo(rx2, ry2)
	gPath.LineTo(rx3, ry3)
	gPath.LineTo(rx4, ry4)
	gPath.Close()
	var highlightOpts vector.DrawPathOptions
	highlightOpts.ColorScale.ScaleWithColor(highlightColor)
	vector.FillPath(screen, &gPath, nil, &highlightOpts)
}

// drawCrystalCluster draws 3 crystal shards in a cluster
func drawCrystalCluster(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color, isSpiky bool) {
	baseLength := float32(28.0) * scale
	baseWidth := float32(11.0) * scale
	if isSpiky {
		baseLength = float32(34.0) * scale
		baseWidth = float32(7.0) * scale
	}

	// 1. Left shard (rotated by -0.42 radians, slightly shorter)
	leftDir := rotateVec(dirVec, -0.42)
	leftPerp := rotateVec(perpVec, -0.42)
	drawShard(screen, cx, cy, leftDir, leftPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	// 2. Right shard (rotated by +0.42 radians, slightly shorter)
	rightDir := rotateVec(dirVec, 0.42)
	rightPerp := rotateVec(perpVec, 0.42)
	drawShard(screen, cx, cy, rightDir, rightPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	// 3. Center shard (straight, full size)
	drawShard(screen, cx, cy, dirVec, perpVec, baseLength, baseWidth, shadowColor, highlightColor)
}

// drawNodule draws 4 overlapping sphere-like bumps to represent a nodule
func drawNodule(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, mineralColor, coreColor color.Color) {
	R := float32(12.0) * scale
	if R < 2.0 {
		R = 2.0
	}

	type bump struct {
		lx, ly float32
		r      float32
	}

	// Layering order: back to front (Left & Right, then Center, then Top/Tip)
	bumps := []bump{
		{-R * 0.45, R * 0.3, R * 0.75}, // Left
		{R * 0.45, R * 0.3, R * 0.75},  // Right
		{0, R * 0.5, R},                // Center
		{0, R * 0.85, R * 0.6},         // Top/Tip
	}

	for _, b := range bumps {
		bx, by := localToScreen(cx, cy, b.lx, b.ly, dirVec, perpVec)
		// 1. Fill base dark bump
		vector.FillCircle(screen, bx, by, b.r, darkenColor(mineralColor, 0.9), false)
		// 2. Stroke outline
		vector.StrokeCircle(screen, bx, by, b.r, 1.0, darkenColor(mineralColor, 0.5), false)

		// 3. Draw shiny highlight offset
		hx, hy := localToScreen(cx, cy, b.lx-b.r*0.25, b.ly+b.r*0.25, dirVec, perpVec)
		hr := b.r * 0.28
		if hr < 1.0 {
			hr = 1.0
		}
		vector.FillCircle(screen, hx, hy, hr, blendColor(coreColor, color.White, 0.6), false)
	}
}

// drawQuartzNeedles draws thin, long glowing quartz needles
func drawQuartzNeedles(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color) {
	baseLength := float32(36.0) * scale
	baseWidth := float32(4.5) * scale

	// Draw 4 needles pointing at various angles
	angles := []float32{-0.5, -0.18, 0.2, 0.55}
	lengths := []float32{0.7, 1.0, 0.85, 0.65}

	for i, angle := range angles {
		d := rotateVec(dirVec, angle)
		p := rotateVec(perpVec, angle)
		drawShard(screen, cx, cy, d, p, baseLength*lengths[i], baseWidth, shadowColor, highlightColor)
	}
}

// drawMineral renders the mineral based on its type and attachment direction
func drawMineral(screen *ebiten.Image, tx, ty int, camX, camY float64, hitsToMine int, mineralColor, coreColor color.Color, attachDir AttachDirection, mineralName string) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))

	// Determine basis vectors based on AttachDirection
	var cx, cy float32
	var dirVec, perpVec [2]float32

	switch attachDir {
	case AttachTop:
		cx = sx + float32(TileSize)/2.0
		cy = sy
		dirVec = [2]float32{0, 1}
		perpVec = [2]float32{-1, 0}
	case AttachLeft:
		cx = sx
		cy = sy + float32(TileSize)/2.0
		dirVec = [2]float32{1, 0}
		perpVec = [2]float32{0, -1}
	case AttachRight:
		cx = sx + float32(TileSize)
		cy = sy + float32(TileSize)/2.0
		dirVec = [2]float32{-1, 0}
		perpVec = [2]float32{0, 1}
	case AttachBottom:
		cx = sx + float32(TileSize)/2.0
		cy = sy + float32(TileSize)
		dirVec = [2]float32{0, -1}
		perpVec = [2]float32{1, 0}
	default: // AttachNone (fallback / icon center)
		cx = sx + float32(TileSize)/2.0
		cy = sy + float32(TileSize)/2.0 + 8.0 // offset slightly down so growing up centers it
		dirVec = [2]float32{0, -1}
		perpVec = [2]float32{1, 0}
	}

	// Scale size based on hits left
	scale := float32(hitsToMine) / 3.0
	if scale < 0.35 {
		scale = 0.35
	}

	shadowColor := darkenColor(mineralColor, 0.82)
	highlightColor := blendColor(mineralColor, coreColor, 0.65)

	switch mineralName {
	case "Copper":
		drawNodule(screen, cx, cy, dirVec, perpVec, scale, mineralColor, coreColor)
	case "Quartz":
		drawQuartzNeedles(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor)
	case "Abyssal Ore":
		// Glowing purple crystal shards
		drawSpikyCrystal := true
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, drawSpikyCrystal)
	default: // Titanium / default
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, false)
	}
}

// drawMineralIcon renders the mineral crystal or nodule centered at cx, cy with a custom size.
func drawMineralIcon(screen *ebiten.Image, cx, cy, size float32, mineralColor, coreColor color.Color, mineralName string) {
	scale := size / 40.0
	if scale < 0.2 {
		scale = 0.2
	}

	dirVec := [2]float32{0, -1}
	perpVec := [2]float32{1, 0}

	shadowColor := darkenColor(mineralColor, 0.82)
	highlightColor := blendColor(mineralColor, coreColor, 0.65)

	switch mineralName {
	case "Copper":
		drawNodule(screen, cx, cy, dirVec, perpVec, scale, mineralColor, coreColor)
	case "Quartz":
		drawQuartzNeedles(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor)
	case "Abyssal Ore":
		drawSpikyCrystal := true
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, drawSpikyCrystal)
	default: // Titanium / default
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, false)
	}
}

func drawCracks(screen *ebiten.Image, sx, sy float32, hitsToMine int) {
	if hitsToMine >= 3 {
		// No cracks at full health (3 hits remaining)
		return
	}

	// Crack color: dark charcoal/black to represent fracture lines overlaying the mineral
	crackColor := color.RGBA{15, 15, 20, 235}

	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0

	// Stage 1 (1 hit taken, 2 remaining): draw primary cracks radiating from center
	if hitsToMine <= 2 {
		vector.StrokeLine(screen, cx-14, cy-12, cx-3, cy-2, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx-3, cy-2, cx+12, cy-14, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx-3, cy-2, cx-6, cy+14, 3.0, crackColor, false)
	}

	// Stage 2 (2 hits taken, 1 remaining): add detailed, longer branching fractures
	if hitsToMine <= 1 {
		// Branches from Stage 1 cracks
		vector.StrokeLine(screen, cx-14, cy-12, cx-22, cy-8, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx+12, cy-14, cx+20, cy-20, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx-6, cy+14, cx-14, cy+22, 2.0, crackColor, false)
		vector.StrokeLine(screen, cx-6, cy+14, cx+8, cy+12, 2.0, crackColor, false)

		// Second independent fracture cluster
		vector.StrokeLine(screen, cx+14, cy+4, cx+3, cy-8, 3.0, crackColor, false)
		vector.StrokeLine(screen, cx+3, cy-8, cx-8, cy-5, 3.0, crackColor, false)
	}
}

func drawNodeSprite(screen *ebiten.Image, tx, ty int, camX, camY float64, sprite *ebiten.Image, hitsToMine int) bool {
	if sprite == nil {
		return false
	}
	sx := float64(tx*TileSize - int(camX))
	sy := float64(ty*TileSize - int(camY))

	// Draw the node background block under the sprite
	drawNodeBase(screen, tx, ty, camX, camY)

	op := &ebiten.DrawImageOptions{}

	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	// Scale the sprite to match the full TileSize (64x64) without shrinking
	baseScaleX := float64(TileSize) / spriteW
	baseScaleY := float64(TileSize) / spriteH

	// Center the sprite on the origin (0,0) before scaling
	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	// Scale to full tile size
	op.GeoM.Scale(baseScaleX, baseScaleY)
	// Translate to screen tile coordinates + center offset
	op.GeoM.Translate(sx+float64(TileSize)/2.0, sy+float64(TileSize)/2.0)

	screen.DrawImage(sprite, op)

	// Draw overlay cracks representing node damage state
	drawCracks(screen, float32(sx), float32(sy), hitsToMine)

	return true
}

func drawNodeIconSprite(screen *ebiten.Image, cx, cy, size float32, sprite *ebiten.Image) bool {
	if sprite == nil {
		return false
	}
	op := &ebiten.DrawImageOptions{}
	spriteW := float64(sprite.Bounds().Dx())
	spriteH := float64(sprite.Bounds().Dy())

	op.GeoM.Translate(-spriteW/2.0, -spriteH/2.0)
	op.GeoM.Scale(float64(size)/spriteW, float64(size)/spriteH)
	op.GeoM.Translate(float64(cx), float64(cy))

	screen.DrawImage(sprite, op)
	return true
}
