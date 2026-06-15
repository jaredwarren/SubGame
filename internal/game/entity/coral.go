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

// Coral is a decorative, surface-aligned marine growth that spawns in caves.
type Coral struct {
	BaseEntity
	Biome      int
	Attachment string // "floor", "ceiling", "left", "right"
	Variant    int    // Spawning variation (0, 1, 2 depending on biome)
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
	// Increment animation phases
	c.SwayPhase += 0.035
}

// project transforms local coordinates (relative to attachment surface base) into screen coordinates.
// localX: offset along the surface (-halfWidth to +halfWidth)
// localY: distance extending outward from the surface (0 to height)
func (c *Coral) project(localX, localY float32, baseScreenX, baseScreenY float32) (float32, float32) {
	switch c.Attachment {
	case "ceiling":
		return baseScreenX + localX, baseScreenY + localY
	case "left":
		return baseScreenX + localY, baseScreenY + localX
	case "right":
		return baseScreenX - localY, baseScreenY + localX
	default: // "floor"
		return baseScreenX + localX, baseScreenY - localY
	}
}

func (c *Coral) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(c.Pos.X - camera.Pos.X)
	sy := float32(c.Pos.Y - camera.Pos.Y)
	sw := float32(c.Dimensions.X)
	sh := float32(c.Dimensions.Y)

	// Determine base position along the attached wall/floor/ceiling
	var bx, by float32
	switch c.Attachment {
	case "ceiling":
		bx, by = sx+sw/2.0, sy
	case "left":
		bx, by = sx, sy+sh/2.0
	case "right":
		bx, by = sx+sw, sy+sh/2.0
	default: // "floor"
		bx, by = sx+sw/2.0, sy+sh
	}

	switch c.Biome {
	case CoralBiomeShallow:
		c.drawShallow(screen, bx, by)
	case CoralBiomeTrench:
		c.drawTrench(screen, bx, by)
	case CoralBiomeWreckage:
		c.drawWreckage(screen, bx, by)
	case CoralBiomeShock:
		c.drawShock(screen, bx, by)
	case CoralBiomeThermo:
		c.drawThermo(screen, bx, by)
	}
}

func (c *Coral) drawShallow(screen *ebiten.Image, bx, by float32) {
	switch c.Variant {
	case 0:
		// Staghorn (Branching) Coral: Coral pink/rose color, sways slightly, light tips
		mainClr := color.RGBA{255, 110, 130, 255}
		tipClr := color.RGBA{255, 180, 195, 255}
		sway := float32(math.Sin(c.SwayPhase)) * 3.0

		// Main stem
		mx, my := c.project(sway*0.5, 10, bx, by)
		vector.StrokeLine(screen, bx, by, mx, my, 2.5, mainClr, false)

		// Left branch
		lx, ly := c.project(-6+sway*0.8, 18, bx, by)
		vector.StrokeLine(screen, mx, my, lx, ly, 1.8, mainClr, false)
		vector.FillCircle(screen, lx, ly, 2.2, tipClr, false)

		// Right branch
		rx, ry := c.project(6+sway*0.8, 18, bx, by)
		vector.StrokeLine(screen, mx, my, rx, ry, 1.8, mainClr, false)
		vector.FillCircle(screen, rx, ry, 2.2, tipClr, false)

	case 1:
		// Brain (Mound) Coral: Orange-yellow, ribbed dome texture
		mainClr := color.RGBA{255, 160, 60, 255}
		strokeClr := color.RGBA{210, 110, 30, 255}

		// Draw overlapping circles to create dome
		cx1, cy1 := c.project(-4, 6, bx, by)
		vector.FillCircle(screen, cx1, cy1, 6, mainClr, false)

		cx2, cy2 := c.project(4, 6, bx, by)
		vector.FillCircle(screen, cx2, cy2, 6, mainClr, false)

		cx3, cy3 := c.project(0, 9, bx, by)
		vector.FillCircle(screen, cx3, cy3, 8, mainClr, false)

		// Rib texture lines
		vector.StrokeCircle(screen, cx3, cy3, 5, 0.8, strokeClr, false)
		vector.StrokeCircle(screen, cx3, cy3, 8, 0.8, strokeClr, false)

	default:
		// Tube Coral: Lavender cylinders with dark openings
		tubeClr := color.RGBA{210, 115, 220, 255}
		rimClr := color.RGBA{170, 75, 180, 255}
		openClr := color.RGBA{80, 20, 90, 255}

		// Tube 1 (left)
		tx1, ty1 := c.project(-5, 12, bx, by)
		bx1, by1 := c.project(-5, 0, bx, by)
		vector.StrokeLine(screen, bx1, by1, tx1, ty1, 4.5, tubeClr, false)
		vector.FillCircle(screen, tx1, ty1, 2.2, openClr, false)
		vector.StrokeCircle(screen, tx1, ty1, 2.2, 0.8, rimClr, false)

		// Tube 2 (center)
		tx2, ty2 := c.project(0, 18, bx, by)
		bx2, by2 := c.project(0, 0, bx, by)
		vector.StrokeLine(screen, bx2, by2, tx2, ty2, 5.0, tubeClr, false)
		vector.FillCircle(screen, tx2, ty2, 2.5, openClr, false)
		vector.StrokeCircle(screen, tx2, ty2, 2.5, 0.8, rimClr, false)

		// Tube 3 (right)
		tx3, ty3 := c.project(5, 10, bx, by)
		bx3, by3 := c.project(5, 0, bx, by)
		vector.StrokeLine(screen, bx3, by3, tx3, ty3, 4.0, tubeClr, false)
		vector.FillCircle(screen, tx3, ty3, 2.0, openClr, false)
		vector.StrokeCircle(screen, tx3, ty3, 2.0, 0.8, rimClr, false)
	}
}

func (c *Coral) drawTrench(screen *ebiten.Image, bx, by float32) {
	// Bioluminescent deep-sea corals
	pulse := float32(math.Sin(c.SwayPhase)) * 0.5 + 0.5 // 0.0 to 1.0

	switch c.Variant {
	case 0:
		// Fan Coral: Teal, radiating ribs that pulse in brightness/alpha
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

	default:
		// Bulb Stalk: Swaying stem with a glowing blue bulb at the tip
		stalkClr := color.RGBA{20, 80, 100, 255}
		blueVal := uint8(200 + 55*pulse)
		bulbClr := color.RGBA{0, blueVal, 255, 255}
		glowClr := color.RGBA{0, 120, 255, uint8(60 + 60*pulse)}
		sway := float32(math.Cos(c.SwayPhase)) * 2.0

		// Stem curve
		tx1, ty1 := c.project(sway*0.4, 8, bx, by)
		vector.StrokeLine(screen, bx, by, tx1, ty1, 1.8, stalkClr, false)

		tx2, ty2 := c.project(sway, 16, bx, by)
		vector.StrokeLine(screen, tx1, ty1, tx2, ty2, 1.5, stalkClr, false)

		// Bulb glow & center
		vector.FillCircle(screen, tx2, ty2, 7.0+pulse*2.0, glowClr, false)
		vector.FillCircle(screen, tx2, ty2, 3.5, bulbClr, false)
		vector.FillCircle(screen, tx2, ty2, 1.2, color.White, false)
	}
}

func (c *Coral) drawWreckage(screen *ebiten.Image, bx, by float32) {
	// Industrial wreckage barnacles/metal-encrusting tube growths
	switch c.Variant {
	case 0:
		// Barnacle Mounds: Cratered rusty volcano cones
		rustClr := color.RGBA{165, 85, 35, 255}
		darkClr := color.RGBA{50, 45, 45, 255}
		rimClr := color.RGBA{105, 110, 115, 255}

		// Draw main barnacle cone
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

		op := &ebiten.DrawTrianglesOptions{AntiAlias: false}
		vertices, indices := p.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vertices {
			vertices[i].ColorR = float32(rustClr.R) / 255
			vertices[i].ColorG = float32(rustClr.G) / 255
			vertices[i].ColorB = float32(rustClr.B) / 255
			vertices[i].ColorA = 1.0
		}
		// Render filled polygon
		img := ebiten.NewImage(1, 1)
		img.Fill(color.White)
		screen.DrawTriangles(vertices, indices, img, op)

		// Draw opening/crater at the top
		cx, cy := c.project(0, 10, bx, by)
		vector.FillCircle(screen, cx, cy, 3.0, darkClr, false)
		vector.StrokeCircle(screen, cx, cy, 3.0, 0.8, rimClr, false)

	default:
		// Industrial Encrusted Metal Tubes: Straight tubes with rusted safety stripes
		metalClr := color.RGBA{95, 100, 105, 255}
		stripeClr := color.RGBA{215, 130, 30, 220}

		// Left tube
		tx1, ty1 := c.project(-4, 12, bx, by)
		bx1, by1 := c.project(-4, 0, bx, by)
		vector.StrokeLine(screen, bx1, by1, tx1, ty1, 4.0, metalClr, false)
		// Rusted stripe
		sx1, sy1 := c.project(-4, 6, bx, by)
		vector.FillCircle(screen, sx1, sy1, 2.2, stripeClr, false)

		// Right tube
		tx2, ty2 := c.project(3, 16, bx, by)
		bx2, by2 := c.project(3, 0, bx, by)
		vector.StrokeLine(screen, bx2, by2, tx2, ty2, 4.5, metalClr, false)
		// Rusted stripe
		sx2, sy2 := c.project(3, 8, bx, by)
		vector.FillCircle(screen, sx2, sy2, 2.5, stripeClr, false)
	}
}

func (c *Coral) drawShock(screen *ebiten.Image, bx, by float32) {
	// Electric crystalline corals in Shock Kelp cave
	pulse := float32(math.Sin(c.SwayPhase * 1.5)) * 0.5 + 0.5 // High frequency pulse

	switch c.Variant {
	case 0:
		// Crystal Spire: Crystalline purple diamond spires with electric cyan centers
		purpClr := color.RGBA{160, 45, 230, 220}
		cyanClr := color.RGBA{0, 230, 255, 255}
		glowClr := color.RGBA{0, 210, 255, uint8(50 + 80*pulse)}

		// Shard path
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

		op := &ebiten.DrawTrianglesOptions{AntiAlias: false}
		vertices, indices := p.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vertices {
			vertices[i].ColorR = float32(purpClr.R) / 255
			vertices[i].ColorG = float32(purpClr.G) / 255
			vertices[i].ColorB = float32(purpClr.B) / 255
			vertices[i].ColorA = 0.9
		}
		img := ebiten.NewImage(1, 1)
		img.Fill(color.White)
		screen.DrawTriangles(vertices, indices, img, op)

		// Glowing core line
		cx1, cy1 := c.project(0, 3, bx, by)
		cx2, cy2 := c.project(0, 17, bx, by)
		vector.StrokeLine(screen, cx1, cy1, cx2, cy2, 1.5, cyanClr, false)

		// Outer spark glow
		tipX, tipY := c.project(0, 20, bx, by)
		vector.FillCircle(screen, tipX, tipY, 4.0+pulse*3.0, glowClr, false)
		vector.FillCircle(screen, tipX, tipY, 1.2, color.White, false)

	default:
		// Electric Branch: Zig-zag purple/cyan lightning-like stalk
		cyanClr := color.RGBA{0, 240, 255, 255}
		purpClr := color.RGBA{140, 50, 210, 255}
		sparkClr := color.RGBA{255, 255, 255, uint8(180 + 75*pulse)}

		// Segment 1
		x1, y1 := c.project(-4, 7, bx, by)
		vector.StrokeLine(screen, bx, by, x1, y1, 2.0, purpClr, false)

		// Segment 2
		x2, y2 := c.project(3, 14, bx, by)
		vector.StrokeLine(screen, x1, y1, x2, y2, 1.5, purpClr, false)

		// Segment 3
		x3, y3 := c.project(-2, 19, bx, by)
		vector.StrokeLine(screen, x2, y2, x3, y3, 1.2, cyanClr, false)

		// Arc Spark
		vector.FillCircle(screen, x3, y3, 3.5, sparkClr, false)
		if pulse > 0.8 {
			// Random spark line jutting out
			sx, sy := c.project(-2+float32(math.Sin(c.RandOffset))*4.0, 19+4.0, bx, by)
			vector.StrokeLine(screen, x3, y3, sx, sy, 0.8, cyanClr, false)
		}
	}
}

func (c *Coral) drawThermo(screen *ebiten.Image, bx, by float32) {
	// Volcanic magma-resistant corals in thermo cave
	pulse := float32(math.Sin(c.SwayPhase * 1.2)) * 0.5 + 0.5 // Magma breathing pulse

	switch c.Variant {
	case 0:
		// Obsidian Spikes: Sharp black basalt triangles with glowing lava cracks
		blackClr := color.RGBA{30, 26, 26, 255}
		rimClr := color.RGBA{50, 42, 42, 255}
		glowClr := color.RGBA{255, 80, 10, uint8(180 + 75*pulse)}

		// Spike 1 (large)
		p1 := &vector.Path{}
		p1a_x, p1a_y := c.project(-7, 0, bx, by)
		p1b_x, p1b_y := c.project(-2, 18, bx, by)
		p1c_x, p1c_y := c.project(3, 0, bx, by)
		p1.MoveTo(p1a_x, p1a_y)
		p1.LineTo(p1b_x, p1b_y)
		p1.LineTo(p1c_x, p1c_y)
		p1.Close()

		op := &ebiten.DrawTrianglesOptions{AntiAlias: false}
		vertices, indices := p1.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vertices {
			vertices[i].ColorR = float32(blackClr.R) / 255
			vertices[i].ColorG = float32(blackClr.G) / 255
			vertices[i].ColorB = float32(blackClr.B) / 255
			vertices[i].ColorA = 1.0
		}
		img := ebiten.NewImage(1, 1)
		img.Fill(color.White)
		screen.DrawTriangles(vertices, indices, img, op)

		// Stroke boundary to make it distinct
		vector.StrokeLine(screen, p1a_x, p1a_y, p1b_x, p1b_y, 0.8, rimClr, false)
		vector.StrokeLine(screen, p1b_x, p1b_y, p1c_x, p1c_y, 0.8, rimClr, false)

		// Spike 2 (small)
		p2 := &vector.Path{}
		p2a_x, p2a_y := c.project(1, 0, bx, by)
		p2b_x, p2b_y := c.project(6, 12, bx, by)
		p2c_x, p2c_y := c.project(9, 0, bx, by)
		p2.MoveTo(p2a_x, p2a_y)
		p2.LineTo(p2b_x, p2b_y)
		p2.LineTo(p2c_x, p2c_y)
		p2.Close()

		vertices2, indices2 := p2.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vertices2 {
			vertices2[i].ColorR = float32(blackClr.R) / 255
			vertices2[i].ColorG = float32(blackClr.G) / 255
			vertices2[i].ColorB = float32(blackClr.B) / 255
			vertices2[i].ColorA = 1.0
		}
		screen.DrawTriangles(vertices2, indices2, img, op)
		vector.StrokeLine(screen, p2a_x, p2a_y, p2b_x, p2b_y, 0.8, rimClr, false)
		vector.StrokeLine(screen, p2b_x, p2b_y, p2c_x, p2c_y, 0.8, rimClr, false)

		// Glowing cracks in large spike
		cx1, cy1 := c.project(-4, 4, bx, by)
		cx2, cy2 := c.project(-2.5, 12, bx, by)
		vector.StrokeLine(screen, cx1, cy1, cx2, cy2, 1.2, glowClr, false)

	default:
		// Magma/Ember Vent Nodule: Round dark nodule venting pulsing lava sparks
		ventClr := color.RGBA{40, 32, 32, 255}
		lavaClr := color.RGBA{255, 60, 0, 255}
		glowClr := color.RGBA{255, 100, 10, uint8(100 + 80*pulse)}

		// Draw base dome
		cx, cy := c.project(0, 5, bx, by)
		vector.FillCircle(screen, cx, cy, 7.5, ventClr, false)
		vector.StrokeCircle(screen, cx, cy, 7.5, 0.8, color.RGBA{65, 55, 55, 255}, false)

		// Draw molten center mouth
		mx, my := c.project(0, 7, bx, by)
		vector.FillCircle(screen, mx, my, 3.5, lavaClr, false)
		vector.FillCircle(screen, mx, my, 6.0+pulse*3.0, glowClr, false)

		// Draw tiny rising embers
		for i := 0; i < 3; i++ {
			h := hashCoords(int(c.Pos.X)+i, int(c.Pos.Y))
			eyOffset := float32(math.Mod(c.SwayPhase*10.0+float64(h%10), 12.0))
			exOffset := float32(math.Sin(c.SwayPhase+float64(h))) * 3.0
			ex, ey := c.project(exOffset, 7.0+eyOffset, bx, by)
			vector.FillRect(screen, ex-0.6, ey-0.6, 1.2, 1.2, color.RGBA{255, 140, 30, uint8(255 * (1.0 - eyOffset/12.0))}, false)
		}
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
