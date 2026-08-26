package entity

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Constants matching cave.CaveType to prevent circular import between cave and entity packages.
const (
	CoralBiomeShallow  = 0
	CoralBiomeTrench   = 1
	CoralBiomeWreckage = 2
	CoralBiomeShock    = 4
	CoralBiomeThermo   = 5
)

// Attachment sides for surface-aligned coral.
const (
	CoralAttachFloor   = "floor"
	CoralAttachCeiling = "ceiling"
	CoralAttachLeft    = "left"
	CoralAttachRight   = "right"
)

// Variant indices within a biome (spawn rolls 0..count-1).
const (
	CoralShallowStaghorn = 0
	CoralShallowBrain    = 1
	CoralShallowTube     = 2

	CoralTrenchFan  = 0
	CoralTrenchBulb = 1

	CoralWreckageBarnacle = 0
	CoralWreckageTubes    = 1

	CoralShockSpire  = 0
	CoralShockBranch = 1

	CoralThermoSpikes = 0
	CoralThermoVent   = 1
)

const (
	CoralShallowVariantCount = 3
	CoralBiomeVariantCount   = 2 // trench / wreckage / shock / thermo
)

// CoralVariantCount is how many looks spawn for a biome (Intn-friendly).
func CoralVariantCount(biome int) int {
	if biome == CoralBiomeShallow {
		return CoralShallowVariantCount
	}
	return CoralBiomeVariantCount
}

// coralDefaultVariant matches the former switch default when Variant is out of range.
var coralDefaultVariant = map[int]int{
	CoralBiomeShallow:  CoralShallowTube,
	CoralBiomeTrench:   CoralTrenchBulb,
	CoralBiomeWreckage: CoralWreckageTubes,
	CoralBiomeShock:    CoralShockBranch,
	CoralBiomeThermo:   CoralThermoVent,
}

// whitePixel is reused by DrawTriangles fills — avoids per-coral NewImage each frame.
var whitePixel = ebiten.NewImage(1, 1)

func init() {
	whitePixel.Fill(color.White)
}

// Coral is a decorative, surface-aligned marine growth that spawns in caves.
type Coral struct {
	BaseEntity
	Biome      int
	Attachment string // CoralAttachFloor / Ceiling / Left / Right
	Variant    int    // Per-biome look index (see CoralShallow* / CoralTrench* / …)
	SwayPhase  float64
	RandOffset float64
}

// NewCoral creates a new Coral entity.
func NewCoral(x, y float64, biome int, attachment string, variant int, r *rand.Rand) *Coral {
	return &Coral{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 24, Y: 24},
			Active:     true,
		},
		Biome:      biome,
		Attachment: attachment,
		Variant:    variant,
		SwayPhase:  r.Float64() * math.Pi * 2,
		RandOffset: r.Float64() * 100.0,
	}
}

func (c *Coral) Update(gr Runtime) {
	c.SwayPhase += 0.035
}

type coralStyleKey struct {
	Biome   int
	Variant int
}

type coralDrawer func(c *Coral, screen *ebiten.Image, bx, by float32)

var coralDrawers = map[coralStyleKey]coralDrawer{
	{CoralBiomeShallow, CoralShallowStaghorn}: drawCoralStaghorn,
	{CoralBiomeShallow, CoralShallowBrain}:    drawCoralBrain,
	{CoralBiomeShallow, CoralShallowTube}:     drawCoralTube,

	{CoralBiomeTrench, CoralTrenchFan}:  drawCoralFan,
	{CoralBiomeTrench, CoralTrenchBulb}: drawCoralBulbStalk,

	{CoralBiomeWreckage, CoralWreckageBarnacle}: drawCoralBarnacle,
	{CoralBiomeWreckage, CoralWreckageTubes}:    drawCoralMetalTubes,

	{CoralBiomeShock, CoralShockSpire}:  drawCoralCrystalSpire,
	{CoralBiomeShock, CoralShockBranch}: drawCoralElectricBranch,

	{CoralBiomeThermo, CoralThermoSpikes}: drawCoralObsidianSpikes,
	{CoralBiomeThermo, CoralThermoVent}:   drawCoralMagmaVent,
}

// project transforms local coordinates (relative to attachment surface base) into screen coordinates.
// localX: offset along the surface (-halfWidth to +halfWidth)
// localY: distance extending outward from the surface (0 to height)
func (c *Coral) project(localX, localY float32, baseScreenX, baseScreenY float32) (float32, float32) {
	switch c.Attachment {
	case CoralAttachCeiling:
		return baseScreenX + localX, baseScreenY + localY
	case CoralAttachLeft:
		return baseScreenX + localY, baseScreenY + localX
	case CoralAttachRight:
		return baseScreenX - localY, baseScreenY + localX
	default: // CoralAttachFloor
		return baseScreenX + localX, baseScreenY - localY
	}
}

func (c *Coral) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)

	var bx, by float32
	switch c.Attachment {
	case CoralAttachCeiling:
		bx, by = sx+sw/2.0, sy
	case CoralAttachLeft:
		bx, by = sx, sy+sh/2.0
	case CoralAttachRight:
		bx, by = sx+sw, sy+sh/2.0
	default: // CoralAttachFloor
		bx, by = sx+sw/2.0, sy+sh
	}

	key := coralStyleKey{c.Biome, c.Variant}
	draw, ok := coralDrawers[key]
	if !ok {
		if def, has := coralDefaultVariant[c.Biome]; has {
			draw, ok = coralDrawers[coralStyleKey{c.Biome, def}]
		}
	}
	if ok {
		draw(c, screen, bx, by)
	}
}

func fillCoralPath(screen *ebiten.Image, p *vector.Path, clr color.RGBA, alpha float32) {
	op := &ebiten.DrawTrianglesOptions{AntiAlias: false}
	vertices, indices := p.AppendVerticesAndIndicesForFilling(nil, nil)
	r := float32(clr.R) / 255
	g := float32(clr.G) / 255
	b := float32(clr.B) / 255
	for i := range vertices {
		vertices[i].ColorR = r
		vertices[i].ColorG = g
		vertices[i].ColorB = b
		vertices[i].ColorA = alpha
	}
	screen.DrawTriangles(vertices, indices, whitePixel, op)
}

func drawCoralStaghorn(c *Coral, screen *ebiten.Image, bx, by float32) {
	mainClr := color.RGBA{255, 110, 130, 255}
	tipClr := color.RGBA{255, 180, 195, 255}
	sway := float32(math.Sin(c.SwayPhase)) * 3.0

	mx, my := c.project(sway*0.5, 10, bx, by)
	vector.StrokeLine(screen, bx, by, mx, my, 2.5, mainClr, false)

	lx, ly := c.project(-6+sway*0.8, 18, bx, by)
	vector.StrokeLine(screen, mx, my, lx, ly, 1.8, mainClr, false)
	vector.FillCircle(screen, lx, ly, 2.2, tipClr, false)

	rx, ry := c.project(6+sway*0.8, 18, bx, by)
	vector.StrokeLine(screen, mx, my, rx, ry, 1.8, mainClr, false)
	vector.FillCircle(screen, rx, ry, 2.2, tipClr, false)
}

func drawCoralBrain(c *Coral, screen *ebiten.Image, bx, by float32) {
	mainClr := color.RGBA{255, 160, 60, 255}
	strokeClr := color.RGBA{210, 110, 30, 255}

	cx1, cy1 := c.project(-4, 6, bx, by)
	vector.FillCircle(screen, cx1, cy1, 6, mainClr, false)

	cx2, cy2 := c.project(4, 6, bx, by)
	vector.FillCircle(screen, cx2, cy2, 6, mainClr, false)

	cx3, cy3 := c.project(0, 9, bx, by)
	vector.FillCircle(screen, cx3, cy3, 8, mainClr, false)

	vector.StrokeCircle(screen, cx3, cy3, 5, 0.8, strokeClr, false)
	vector.StrokeCircle(screen, cx3, cy3, 8, 0.8, strokeClr, false)
}

func drawCoralTube(c *Coral, screen *ebiten.Image, bx, by float32) {
	tubeClr := color.RGBA{210, 115, 220, 255}
	rimClr := color.RGBA{170, 75, 180, 255}
	openClr := color.RGBA{80, 20, 90, 255}

	tx1, ty1 := c.project(-5, 12, bx, by)
	bx1, by1 := c.project(-5, 0, bx, by)
	vector.StrokeLine(screen, bx1, by1, tx1, ty1, 4.5, tubeClr, false)
	vector.FillCircle(screen, tx1, ty1, 2.2, openClr, false)
	vector.StrokeCircle(screen, tx1, ty1, 2.2, 0.8, rimClr, false)

	tx2, ty2 := c.project(0, 18, bx, by)
	bx2, by2 := c.project(0, 0, bx, by)
	vector.StrokeLine(screen, bx2, by2, tx2, ty2, 5.0, tubeClr, false)
	vector.FillCircle(screen, tx2, ty2, 2.5, openClr, false)
	vector.StrokeCircle(screen, tx2, ty2, 2.5, 0.8, rimClr, false)

	tx3, ty3 := c.project(5, 10, bx, by)
	bx3, by3 := c.project(5, 0, bx, by)
	vector.StrokeLine(screen, bx3, by3, tx3, ty3, 4.0, tubeClr, false)
	vector.FillCircle(screen, tx3, ty3, 2.0, openClr, false)
	vector.StrokeCircle(screen, tx3, ty3, 2.0, 0.8, rimClr, false)
}

func drawCoralFan(c *Coral, screen *ebiten.Image, bx, by float32) {
	pulse := float32(math.Sin(c.SwayPhase))*0.5 + 0.5
	tealVal := uint8(150 + 105*pulse)
	fanClr := color.RGBA{0, tealVal, 190, uint8(160 + 95*pulse)}
	glowClr := color.RGBA{100, 255, 230, uint8(80 * pulse)}

	tips := []struct{ lx, ly float32 }{
		{-10, 15},
		{-5, 19},
		{0, 20},
		{5, 19},
		{10, 15},
	}
	for _, t := range tips {
		tx, ty := c.project(t.lx, t.ly, bx, by)
		vector.StrokeLine(screen, bx, by, tx, ty, 1.2, fanClr, false)
		vector.FillCircle(screen, tx, ty, 3.0, glowClr, false)
		vector.FillCircle(screen, tx, ty, 1.0, color.White, false)
	}
}

func drawCoralBulbStalk(c *Coral, screen *ebiten.Image, bx, by float32) {
	pulse := float32(math.Sin(c.SwayPhase))*0.5 + 0.5
	stalkClr := color.RGBA{20, 80, 100, 255}
	blueVal := uint8(200 + 55*pulse)
	bulbClr := color.RGBA{0, blueVal, 255, 255}
	glowClr := color.RGBA{0, 120, 255, uint8(60 + 60*pulse)}
	sway := float32(math.Cos(c.SwayPhase)) * 2.0

	tx1, ty1 := c.project(sway*0.4, 8, bx, by)
	vector.StrokeLine(screen, bx, by, tx1, ty1, 1.8, stalkClr, false)

	tx2, ty2 := c.project(sway, 16, bx, by)
	vector.StrokeLine(screen, tx1, ty1, tx2, ty2, 1.5, stalkClr, false)

	vector.FillCircle(screen, tx2, ty2, 7.0+pulse*2.0, glowClr, false)
	vector.FillCircle(screen, tx2, ty2, 3.5, bulbClr, false)
	vector.FillCircle(screen, tx2, ty2, 1.2, color.White, false)
}

func drawCoralBarnacle(c *Coral, screen *ebiten.Image, bx, by float32) {
	rustClr := color.RGBA{165, 85, 35, 255}
	darkClr := color.RGBA{50, 45, 45, 255}
	rimClr := color.RGBA{105, 110, 115, 255}

	p := &vector.Path{}
	p0x, p0y := c.project(-8, 0, bx, by)
	p1x, p1y := c.project(-3, 10, bx, by)
	p2x, p2y := c.project(3, 10, bx, by)
	p3x, p3y := c.project(8, 0, bx, by)

	p.MoveTo(p0x, p0y)
	p.LineTo(p1x, p1y)
	p.LineTo(p2x, p2y)
	p.LineTo(p3x, p3y)
	p.Close()
	fillCoralPath(screen, p, rustClr, 1.0)

	cx, cy := c.project(0, 10, bx, by)
	vector.FillCircle(screen, cx, cy, 3.0, darkClr, false)
	vector.StrokeCircle(screen, cx, cy, 3.0, 0.8, rimClr, false)
}

func drawCoralMetalTubes(c *Coral, screen *ebiten.Image, bx, by float32) {
	metalClr := color.RGBA{95, 100, 105, 255}
	stripeClr := color.RGBA{215, 130, 30, 220}

	tx1, ty1 := c.project(-4, 12, bx, by)
	bx1, by1 := c.project(-4, 0, bx, by)
	vector.StrokeLine(screen, bx1, by1, tx1, ty1, 4.0, metalClr, false)
	sx1, sy1 := c.project(-4, 6, bx, by)
	vector.FillCircle(screen, sx1, sy1, 2.2, stripeClr, false)

	tx2, ty2 := c.project(3, 16, bx, by)
	bx2, by2 := c.project(3, 0, bx, by)
	vector.StrokeLine(screen, bx2, by2, tx2, ty2, 4.5, metalClr, false)
	sx2, sy2 := c.project(3, 8, bx, by)
	vector.FillCircle(screen, sx2, sy2, 2.5, stripeClr, false)
}

func drawCoralCrystalSpire(c *Coral, screen *ebiten.Image, bx, by float32) {
	pulse := float32(math.Sin(c.SwayPhase*1.5))*0.5 + 0.5
	purpClr := color.RGBA{160, 45, 230, 220}
	cyanClr := color.RGBA{0, 230, 255, 255}
	glowClr := color.RGBA{0, 210, 255, uint8(50 + 80*pulse)}

	p := &vector.Path{}
	p0x, p0y := c.project(0, 0, bx, by)
	p1x, p1y := c.project(-7, 8, bx, by)
	p2x, p2y := c.project(0, 20, bx, by)
	p3x, p3y := c.project(7, 8, bx, by)

	p.MoveTo(p0x, p0y)
	p.LineTo(p1x, p1y)
	p.LineTo(p2x, p2y)
	p.LineTo(p3x, p3y)
	p.Close()
	fillCoralPath(screen, p, purpClr, 0.9)

	cx1, cy1 := c.project(0, 3, bx, by)
	cx2, cy2 := c.project(0, 17, bx, by)
	vector.StrokeLine(screen, cx1, cy1, cx2, cy2, 1.5, cyanClr, false)

	tipX, tipY := c.project(0, 20, bx, by)
	vector.FillCircle(screen, tipX, tipY, 4.0+pulse*3.0, glowClr, false)
	vector.FillCircle(screen, tipX, tipY, 1.2, color.White, false)
}

func drawCoralElectricBranch(c *Coral, screen *ebiten.Image, bx, by float32) {
	pulse := float32(math.Sin(c.SwayPhase*1.5))*0.5 + 0.5
	cyanClr := color.RGBA{0, 240, 255, 255}
	purpClr := color.RGBA{140, 50, 210, 255}
	sparkClr := color.RGBA{255, 255, 255, uint8(180 + 75*pulse)}

	x1, y1 := c.project(-4, 7, bx, by)
	vector.StrokeLine(screen, bx, by, x1, y1, 2.0, purpClr, false)

	x2, y2 := c.project(3, 14, bx, by)
	vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, purpClr, false)

	x3, y3 := c.project(-2, 19, bx, by)
	vector.StrokeLine(screen, x2, y2, x3, y3, 1.2, cyanClr, false)

	vector.FillCircle(screen, x3, y3, 3.5, sparkClr, false)
	if pulse > 0.8 {
		sx, sy := c.project(-2+float32(math.Sin(c.RandOffset))*4.0, 19+4.0, bx, by)
		vector.StrokeLine(screen, x3, y3, sx, sy, 0.8, cyanClr, false)
	}
}

func drawCoralObsidianSpikes(c *Coral, screen *ebiten.Image, bx, by float32) {
	pulse := float32(math.Sin(c.SwayPhase*1.2))*0.5 + 0.5
	blackClr := color.RGBA{30, 26, 26, 255}
	rimClr := color.RGBA{50, 42, 42, 255}
	glowClr := color.RGBA{255, 80, 10, uint8(180 + 75*pulse)}

	p1 := &vector.Path{}
	p1aX, p1aY := c.project(-7, 0, bx, by)
	p1bX, p1bY := c.project(-2, 18, bx, by)
	p1cX, p1cY := c.project(3, 0, bx, by)
	p1.MoveTo(p1aX, p1aY)
	p1.LineTo(p1bX, p1bY)
	p1.LineTo(p1cX, p1cY)
	p1.Close()
	fillCoralPath(screen, p1, blackClr, 1.0)

	vector.StrokeLine(screen, p1aX, p1aY, p1bX, p1bY, 0.8, rimClr, false)
	vector.StrokeLine(screen, p1bX, p1bY, p1cX, p1cY, 0.8, rimClr, false)

	p2 := &vector.Path{}
	p2aX, p2aY := c.project(1, 0, bx, by)
	p2bX, p2bY := c.project(6, 12, bx, by)
	p2cX, p2cY := c.project(9, 0, bx, by)
	p2.MoveTo(p2aX, p2aY)
	p2.LineTo(p2bX, p2bY)
	p2.LineTo(p2cX, p2cY)
	p2.Close()
	fillCoralPath(screen, p2, blackClr, 1.0)
	vector.StrokeLine(screen, p2aX, p2aY, p2bX, p2bY, 0.8, rimClr, false)
	vector.StrokeLine(screen, p2bX, p2bY, p2cX, p2cY, 0.8, rimClr, false)

	cx1, cy1 := c.project(-4, 4, bx, by)
	cx2, cy2 := c.project(-2.5, 12, bx, by)
	vector.StrokeLine(screen, cx1, cy1, cx2, cy2, 1.2, glowClr, false)
}

func drawCoralMagmaVent(c *Coral, screen *ebiten.Image, bx, by float32) {
	pulse := float32(math.Sin(c.SwayPhase*1.2))*0.5 + 0.5
	ventClr := color.RGBA{40, 32, 32, 255}
	lavaClr := color.RGBA{255, 60, 0, 255}
	glowClr := color.RGBA{255, 100, 10, uint8(100 + 80*pulse)}

	cx, cy := c.project(0, 5, bx, by)
	vector.FillCircle(screen, cx, cy, 7.5, ventClr, false)
	vector.StrokeCircle(screen, cx, cy, 7.5, 0.8, color.RGBA{65, 55, 55, 255}, false)

	mx, my := c.project(0, 7, bx, by)
	vector.FillCircle(screen, mx, my, 3.5, lavaClr, false)
	vector.FillCircle(screen, mx, my, 6.0+pulse*3.0, glowClr, false)

	for i := 0; i < 3; i++ {
		h := hashCoords(int(c.Pos.X)+i, int(c.Pos.Y))
		eyOffset := float32(math.Mod(c.SwayPhase*10.0+float64(h%10), 12.0))
		exOffset := float32(math.Sin(c.SwayPhase+float64(h))) * 3.0
		ex, ey := c.project(exOffset, 7.0+eyOffset, bx, by)
		vector.FillRect(screen, ex-0.6, ey-0.6, 1.2, 1.2, color.RGBA{255, 140, 30, uint8(255 * (1.0 - eyOffset/12.0))}, false)
	}
}

func hashCoords(tx, ty int) uint64 {
	x := (int64(tx) << 32) | (int64(uint32(ty)))
	u := uint64(x)
	u ^= u >> 33
	u *= 0xff51afd7ed558ccd
	u ^= u >> 33
	u *= 0xc4ceb9fe1a85ec53
	u ^= u >> 33
	return u
}
