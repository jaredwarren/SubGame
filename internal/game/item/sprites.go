package item

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
	iconSprites map[string]*ebiten.Image
	iconsLoaded bool
	gPath       vector.Path
)

// LoadAssets preloads and chroma-keys all item icon sprites.
func LoadAssets() {
	iconSprites = make(map[string]*ebiten.Image)

	sheet, err := assets.LoadChromaKeyedImage("item_icons")
	if err != nil {
		log.Printf("Error: Failed to load item icons: %v", err)
		return
	}

	bounds := sheet.Bounds()

	cellSize := 316
	startOffset := 76
	if bounds.Dx() != 2048 || bounds.Dy() != 2048 {
		cellSize = bounds.Dx() / 6
		startOffset = 0
	}

	itemCoords := map[string][2]int{
		"Titanium":                   {0, 0},
		"Copper":                     {1, 0},
		"Quartz":                     {2, 0},
		"Abyssal Ore":                {3, 0},
		"Scrap Metal":                {4, 0},
		"Electronic Waste":           {5, 0},
		"High Capacity O2 Tank":      {0, 1},
		"Ultra High Capacity O2 Tank":{1, 1},
		"Propulsion Fins":            {2, 1},
		"Scanner Tool":               {3, 1},
		"Solar Array Module":         {4, 1},
		"Solar Array MKII Module":    {5, 1},
		"Storage Vault Module":       {0, 2},
		"Storage Vault MKII Module":  {1, 2},
		"Scout Sub Kit":              {2, 2},
		"Heavy Mech Kit":             {3, 2},
		"Sonar Amplifier":            {4, 2},
		"Power Cell":                 {5, 2},
		"Thermal Generator":          {0, 4},
		"Escape Rocket":              {1, 4},
		"Raw Fish":                   {2, 4},
		"Cooked Fish":                {3, 4},
		"Raw Crab":                   {4, 4},
		"Cooked Crab":                {5, 4},
		"Sonic Decoy":                {0, 3},
		"Chemical Deterrent":         {1, 3},
		"Decoy Launcher Module":      {2, 3},
		"Chemical Discharger Module": {3, 3},
	}

	for name, coord := range itemCoords {
		col, row := coord[0], coord[1]
		x0 := startOffset + col*cellSize
		y0 := startOffset + row*cellSize
		x1 := x0 + cellSize
		y1 := y0 + cellSize

		if x1 <= bounds.Max.X && y1 <= bounds.Max.Y {
			sub := sheet.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
			iconSprites[name] = sub
		}
	}
	iconsLoaded = true
}

func drawItemIconSprite(screen *ebiten.Image, name string, cx, cy, size float32) bool {
	sprite, ok := iconSprites[name]
	if !ok || sprite == nil {
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

// DrawItemIconSprite wraps the internal drawItemIconSprite function so other packages can render item icons.
func DrawItemIconSprite(screen *ebiten.Image, name string, cx, cy, size float32) bool {
	return drawItemIconSprite(screen, name, cx, cy, size)
}

func darkenColor(c color.Color, factor float32) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(float32(r>>8) * factor),
		G: uint8(float32(g>>8) * factor),
		B: uint8(float32(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

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

func rotateVec(v [2]float32, angle float32) [2]float32 {
	cosA := float32(math.Cos(float64(angle)))
	sinA := float32(math.Sin(float64(angle)))
	return [2]float32{
		v[0]*cosA - v[1]*sinA,
		v[0]*sinA + v[1]*cosA,
	}
}

func localToScreen(cx, cy float32, lx, ly float32, dirVec, perpVec [2]float32) (float32, float32) {
	return cx + lx*perpVec[0] + ly*dirVec[0], cy + lx*perpVec[1] + ly*dirVec[1]
}

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

func drawCrystalCluster(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color, isSpiky bool) {
	baseLength := float32(28.0) * scale
	baseWidth := float32(11.0) * scale
	if isSpiky {
		baseLength = float32(34.0) * scale
		baseWidth = float32(7.0) * scale
	}

	leftDir := rotateVec(dirVec, -0.42)
	leftPerp := rotateVec(perpVec, -0.42)
	drawShard(screen, cx, cy, leftDir, leftPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	rightDir := rotateVec(dirVec, 0.42)
	rightPerp := rotateVec(perpVec, 0.42)
	drawShard(screen, cx, cy, rightDir, rightPerp, baseLength*0.75, baseWidth*0.8, shadowColor, highlightColor)

	drawShard(screen, cx, cy, dirVec, perpVec, baseLength, baseWidth, shadowColor, highlightColor)
}

func drawNodule(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, mineralColor, coreColor color.Color) {
	R := float32(12.0) * scale
	if R < 2.0 {
		R = 2.0
	}

	type bump struct {
		lx, ly float32
		r      float32
	}

	bumps := []bump{
		{-R * 0.45, R * 0.3, R * 0.75},
		{R * 0.45, R * 0.3, R * 0.75},
		{0, R * 0.5, R},
		{0, R * 0.85, R * 0.6},
	}

	for _, b := range bumps {
		bx, by := localToScreen(cx, cy, b.lx, b.ly, dirVec, perpVec)
		vector.FillCircle(screen, bx, by, b.r, darkenColor(mineralColor, 0.9), false)
		vector.StrokeCircle(screen, bx, by, b.r, 1.0, darkenColor(mineralColor, 0.5), false)

		hx, hy := localToScreen(cx, cy, b.lx-b.r*0.25, b.ly+b.r*0.25, dirVec, perpVec)
		hr := b.r * 0.28
		if hr < 1.0 {
			hr = 1.0
		}
		vector.FillCircle(screen, hx, hy, hr, blendColor(coreColor, color.White, 0.6), false)
	}
}

func drawQuartzNeedles(screen *ebiten.Image, cx, cy float32, dirVec, perpVec [2]float32, scale float32, shadowColor, highlightColor color.Color) {
	baseLength := float32(36.0) * scale
	baseWidth := float32(4.5) * scale

	angles := []float32{-0.5, -0.18, 0.2, 0.55}
	lengths := []float32{0.7, 1.0, 0.85, 0.65}

	for i, angle := range angles {
		d := rotateVec(dirVec, angle)
		p := rotateVec(perpVec, angle)
		drawShard(screen, cx, cy, d, p, baseLength*lengths[i], baseWidth, shadowColor, highlightColor)
	}
}

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
	default:
		drawCrystalCluster(screen, cx, cy, dirVec, perpVec, scale, shadowColor, highlightColor, false)
	}
}
