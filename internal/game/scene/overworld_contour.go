package scene

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

const (
	// OverworldCornerRadius is the outer headland corner rounding radius.
	OverworldCornerRadius float32 = 20.0

	// OverworldFilletRadius is the inner bay inlet concave fillet radius.
	OverworldFilletRadius float32 = 16.0

	// OverworldShelfWidth is the width of the outer wet coastal breaker shelf rim.
	OverworldShelfWidth float32 = 7.0
)

// IsWorldLand checks if a tile coordinate is solid land. Boundaries outside the world are treated as open void ocean.
func IsWorldLand(w *world.World, tx, ty int) bool {
	if w == nil {
		return false
	}
	if tx < 0 || tx >= w.Width || ty < 0 || ty >= w.Height {
		return false
	}
	return w.OverworldMap[tx][ty] == world.TileLand
}

// ClassifyOverworldQuadrant determines the geometric shape of quadrant (quad) within overworld tile (tx, ty).
func ClassifyOverworldQuadrant(w *world.World, tx, ty int, quad int) cave.QuadrantInfo {
	solid := IsWorldLand(w, tx, ty)

	var nX, nY int // orthogonal neighbor 1
	var oX, oY int // orthogonal neighbor 2
	var dX, dY int // diagonal neighbor
	var cornerID int

	switch quad {
	case cave.QuadNW:
		nX, nY = tx, ty-1
		oX, oY = tx-1, ty
		dX, dY = tx-1, ty-1
		cornerID = cave.CornerNW
	case cave.QuadNE:
		nX, nY = tx, ty-1
		oX, oY = tx+1, ty
		dX, dY = tx+1, ty-1
		cornerID = cave.CornerNE
	case cave.QuadSW:
		nX, nY = tx, ty+1
		oX, oY = tx-1, ty
		dX, dY = tx-1, ty+1
		cornerID = cave.CornerSW
	case cave.QuadSE:
		nX, nY = tx, ty+1
		oX, oY = tx+1, ty
		dX, dY = tx+1, ty+1
		cornerID = cave.CornerSE
	}

	nSolid := IsWorldLand(w, nX, nY)
	oSolid := IsWorldLand(w, oX, oY)
	dSolid := IsWorldLand(w, dX, dY)

	if !solid {
		if nSolid && oSolid {
			return cave.QuadrantInfo{Shape: cave.ShapeInnerFillet, Corner: cornerID}
		}
		return cave.QuadrantInfo{Shape: cave.ShapeEmpty, Corner: cornerID}
	}

	if !nSolid && !oSolid {
		return cave.QuadrantInfo{Shape: cave.ShapeOuterCorner, Corner: cornerID}
	}

	if !nSolid && oSolid {
		if quad == cave.QuadNW || quad == cave.QuadNE {
			return cave.QuadrantInfo{Shape: cave.ShapeEdgeTop, Corner: cornerID}
		}
		return cave.QuadrantInfo{Shape: cave.ShapeEdgeBottom, Corner: cornerID}
	}

	if nSolid && !oSolid {
		if quad == cave.QuadNW || quad == cave.QuadSW {
			return cave.QuadrantInfo{Shape: cave.ShapeEdgeLeft, Corner: cornerID}
		}
		return cave.QuadrantInfo{Shape: cave.ShapeEdgeRight, Corner: cornerID}
	}

	if !dSolid {
		return cave.QuadrantInfo{Shape: cave.ShapeFullSolid, Corner: cornerID}
	}

	return cave.QuadrantInfo{Shape: cave.ShapeFullSolid, Corner: cornerID}
}

// overworldEdgeBump computes deterministic continuous sinusoidal edge undulation based on world coordinates.
func overworldEdgeBump(worldCoord float32) float32 {
	w := float64(worldCoord)
	s1 := math.Sin(w * 0.09)
	s2 := math.Sin(w*0.23 + 1.2)
	return float32(s1*1.5 + s2*0.8)
}

// DrawOverworldLandTile renders a contoured land tile or water tile with inner fillets.
func DrawOverworldLandTile(target *ebiten.Image, w *world.World, tx, ty int, offsetX, offsetY float64, landBase, landStroke, coastBase, coastStroke color.RGBA) {
	half := float32(config.TileSize) / 2.0 // 32.0
	sx := float32(tx*config.TileSize) - float32(offsetX)
	sy := float32(ty*config.TileSize) - float32(offsetY)
	worldX := float32(tx * config.TileSize)
	worldY := float32(ty * config.TileSize)

	quads := [4]struct {
		id int
		qx float32
		qy float32
		wx float32
		wy float32
	}{
		{cave.QuadNW, sx, sy, worldX, worldY},
		{cave.QuadNE, sx + half, sy, worldX + half, worldY},
		{cave.QuadSW, sx, sy + half, worldX, worldY + half},
		{cave.QuadSE, sx + half, sy + half, worldX + half, worldY + half},
	}

	for _, q := range quads {
		info := ClassifyOverworldQuadrant(w, tx, ty, q.id)
		drawOverworldQuadrant(target, info, q.qx, q.qy, half, q.wx, q.wy, landBase, landStroke, coastBase, coastStroke)
	}
}

func drawOverworldQuadrant(target *ebiten.Image, info cave.QuadrantInfo, qx, qy, half, wx, wy float32, landBase, landStroke, coastBase, coastStroke color.RGBA) {
	shelf := OverworldShelfWidth

	switch info.Shape {
	case cave.ShapeEmpty:
		return

	case cave.ShapeFullSolid:
		vector.FillRect(target, qx, qy, half, half, landBase, false)

	case cave.ShapeEdgeTop:
		vector.FillRect(target, qx, qy, half, half, coastBase, false)
		vector.FillRect(target, qx, qy+shelf, half, half-shelf, landBase, false)
		drawOverworldWavyHoriz(target, qx, qy, half, wx, wy, 1.4, coastStroke)

	case cave.ShapeEdgeBottom:
		vector.FillRect(target, qx, qy, half, half, coastBase, false)
		vector.FillRect(target, qx, qy, half, half-shelf, landBase, false)
		drawOverworldWavyHoriz(target, qx, qy+half, half, wx, wy+half, 1.4, coastStroke)

	case cave.ShapeEdgeLeft:
		vector.FillRect(target, qx, qy, half, half, coastBase, false)
		vector.FillRect(target, qx+shelf, qy, half-shelf, half, landBase, false)
		drawOverworldWavyVert(target, qx, qy, half, wx, wy, 1.4, coastStroke)

	case cave.ShapeEdgeRight:
		vector.FillRect(target, qx, qy, half, half, coastBase, false)
		vector.FillRect(target, qx, qy, half-shelf, half, landBase, false)
		drawOverworldWavyVert(target, qx+half, qy, half, wx+half, wy, 1.4, coastStroke)

	case cave.ShapeOuterCorner:
		drawOverworldOuterCorner(target, qx, qy, half, wx, wy, info.Corner, landBase, landStroke, coastBase, coastStroke)

	case cave.ShapeInnerFillet:
		drawOverworldInnerFillet(target, qx, qy, half, info.Corner, landBase, landStroke, coastBase, coastStroke)
	}
}

func drawOverworldWavyHoriz(target *ebiten.Image, qx, qy, half, wx, wy, strokeWidth float32, strokeColor color.RGBA) {
	const segments = 4
	step := half / float32(segments)
	for i := 0; i < segments; i++ {
		x1 := qx + float32(i)*step
		x2 := qx + float32(i+1)*step
		w1 := wx + float32(i)*step
		w2 := wx + float32(i+1)*step
		y1 := qy + overworldEdgeBump(w1)
		y2 := qy + overworldEdgeBump(w2)
		vector.StrokeLine(target, x1, y1, x2, y2, strokeWidth, strokeColor, false)
	}
}

func drawOverworldWavyVert(target *ebiten.Image, qx, qy, half, wx, wy, strokeWidth float32, strokeColor color.RGBA) {
	const segments = 4
	step := half / float32(segments)
	for i := 0; i < segments; i++ {
		y1 := qy + float32(i)*step
		y2 := qy + float32(i+1)*step
		w1 := wy + float32(i)*step
		w2 := wy + float32(i+1)*step
		x1 := qx + overworldEdgeBump(w1)
		x2 := qx + overworldEdgeBump(w2)
		vector.StrokeLine(target, x1, y1, x2, y2, strokeWidth, strokeColor, false)
	}
}

func drawOverworldWavyHorizFoam(target *ebiten.Image, qx, qy, half, wx, wy, strokeWidth float32, clr color.RGBA) {
	const segments = 4
	step := half / float32(segments)
	for i := 0; i < segments; i++ {
		x1 := qx + float32(i)*step
		x2 := qx + float32(i+1)*step
		w1 := wx + float32(i)*step
		w2 := wx + float32(i+1)*step
		y1 := qy + overworldEdgeBump(w1)
		y2 := qy + overworldEdgeBump(w2)
		vector.StrokeLine(target, x1, y1, x2, y2, strokeWidth, clr, false)
	}
}

func drawOverworldWavyVertFoam(target *ebiten.Image, qx, qy, half, wx, wy, strokeWidth float32, clr color.RGBA) {
	const segments = 4
	step := half / float32(segments)
	for i := 0; i < segments; i++ {
		y1 := qy + float32(i)*step
		y2 := qy + float32(i+1)*step
		w1 := wy + float32(i)*step
		w2 := wy + float32(i+1)*step
		x1 := qx + overworldEdgeBump(w1)
		x2 := qx + overworldEdgeBump(w2)
		vector.StrokeLine(target, x1, y1, x2, y2, strokeWidth, clr, false)
	}
}

func drawArcStroke(target *ebiten.Image, cx, cy, radius float32, startAngle, endAngle float64, steps int, strokeWidth float32, clr color.RGBA) {
	angleStep := (endAngle - startAngle) / float64(steps)
	var prevX, prevY float32
	for i := 0; i <= steps; i++ {
		theta := startAngle + float64(i)*angleStep
		px := cx + float32(math.Cos(theta))*radius
		py := cy + float32(math.Sin(theta))*radius
		if i > 0 {
			vector.StrokeLine(target, prevX, prevY, px, py, strokeWidth, clr, false)
		}
		prevX, prevY = px, py
	}
}

func fillArcFan(target *ebiten.Image, cx, cy, radius float32, startAngle, endAngle float64, steps int, clr color.RGBA) {
	angleStep := (endAngle - startAngle) / float64(steps)
	var prevX, prevY float32
	for i := 0; i <= steps; i++ {
		theta := startAngle + float64(i)*angleStep
		px := cx + float32(math.Cos(theta))*radius
		py := cy + float32(math.Sin(theta))*radius
		if i > 0 {
			fillTriangleOverworld(target, cx, cy, prevX, prevY, px, py, clr)
		}
		prevX, prevY = px, py
	}
}

func drawOverworldOuterCorner(target *ebiten.Image, qx, qy, half, wx, wy float32, corner int, landBase, landStroke, coastBase, coastStroke color.RGBA) {
	r := OverworldCornerRadius
	shelf := OverworldShelfWidth
	rIn := r - shelf
	const steps = 8

	var cx, cy float32
	var startAngle, endAngle float64

	switch corner {
	case cave.CornerNW:
		cx, cy = qx+r, qy+r
		startAngle, endAngle = math.Pi, 1.5*math.Pi

		// 1. Wet coastal shelf layer: 3 inner rects and arc fan
		vector.FillRect(target, qx+r, qy, half-r, r, coastBase, false)
		vector.FillRect(target, qx, qy+r, r, half-r, coastBase, false)
		vector.FillRect(target, qx+r, qy+r, half-r, half-r, coastBase, false)
		fillArcFan(target, cx, cy, r, startAngle, endAngle, steps, coastBase)

		// 2. Inland basalt core: inset rects and inner arc fan
		vector.FillRect(target, qx+r, qy+shelf, half-r, rIn, landBase, false)
		vector.FillRect(target, qx+shelf, qy+r, rIn, half-r, landBase, false)
		vector.FillRect(target, qx+r, qy+r, half-r, half-r, landBase, false)
		fillArcFan(target, cx, cy, rIn, startAngle, endAngle, steps, landBase)

		// 3. Curved arc contour stroke
		drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.4, coastStroke)

		// 4. Straight exposed quadrant edges
		drawOverworldWavyHoriz(target, qx+r, qy, half-r, wx+r, wy, 1.4, coastStroke)
		drawOverworldWavyVert(target, qx, qy+r, half-r, wx, wy+r, 1.4, coastStroke)

	case cave.CornerNE:
		cx, cy = qx+half-r, qy+r
		startAngle, endAngle = 1.5*math.Pi, 2.0*math.Pi

		vector.FillRect(target, qx, qy, half-r, r, coastBase, false)
		vector.FillRect(target, qx+half-r, qy+r, r, half-r, coastBase, false)
		vector.FillRect(target, qx, qy+r, half-r, half-r, coastBase, false)
		fillArcFan(target, cx, cy, r, startAngle, endAngle, steps, coastBase)

		vector.FillRect(target, qx, qy+shelf, half-r, rIn, landBase, false)
		vector.FillRect(target, qx+half-r, qy+r, rIn, half-r, landBase, false)
		vector.FillRect(target, qx, qy+r, half-r, half-r, landBase, false)
		fillArcFan(target, cx, cy, rIn, startAngle, endAngle, steps, landBase)

		drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.4, coastStroke)
		drawOverworldWavyHoriz(target, qx, qy, half-r, wx, wy, 1.4, coastStroke)
		drawOverworldWavyVert(target, qx+half, qy+r, half-r, wx+half, wy+r, 1.4, coastStroke)

	case cave.CornerSW:
		cx, cy = qx+r, qy+half-r
		startAngle, endAngle = 0.5*math.Pi, math.Pi

		vector.FillRect(target, qx, qy, r, half-r, coastBase, false)
		vector.FillRect(target, qx+r, qy, half-r, half-r, coastBase, false)
		vector.FillRect(target, qx+r, qy+half-r, half-r, r, coastBase, false)
		fillArcFan(target, cx, cy, r, startAngle, endAngle, steps, coastBase)

		vector.FillRect(target, qx+shelf, qy, rIn, half-r, landBase, false)
		vector.FillRect(target, qx+r, qy, half-r, half-r, landBase, false)
		vector.FillRect(target, qx+r, qy+half-r, half-r, rIn, landBase, false)
		fillArcFan(target, cx, cy, rIn, startAngle, endAngle, steps, landBase)

		drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.4, coastStroke)
		drawOverworldWavyHoriz(target, qx+r, qy+half, half-r, wx+r, wy+half, 1.4, coastStroke)
		drawOverworldWavyVert(target, qx, qy, half-r, wx, wy, 1.4, coastStroke)

	case cave.CornerSE:
		cx, cy = qx+half-r, qy+half-r
		startAngle, endAngle = 0.0, 0.5*math.Pi

		vector.FillRect(target, qx, qy, half-r, half-r, coastBase, false)
		vector.FillRect(target, qx+half-r, qy, r, half-r, coastBase, false)
		vector.FillRect(target, qx, qy+half-r, half-r, r, coastBase, false)
		fillArcFan(target, cx, cy, r, startAngle, endAngle, steps, coastBase)

		vector.FillRect(target, qx, qy, half-r, half-r, landBase, false)
		vector.FillRect(target, qx+half-r, qy, rIn, half-r, landBase, false)
		vector.FillRect(target, qx, qy+half-r, half-r, rIn, landBase, false)
		fillArcFan(target, cx, cy, rIn, startAngle, endAngle, steps, landBase)

		drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.4, coastStroke)
		drawOverworldWavyHoriz(target, qx, qy+half, half-r, wx, wy+half, 1.4, coastStroke)
		drawOverworldWavyVert(target, qx+half, qy, half-r, wx+half, wy, 1.4, coastStroke)
	}
}

func drawOverworldInnerFillet(target *ebiten.Image, qx, qy, half float32, corner int, landBase, landStroke, coastBase, coastStroke color.RGBA) {
	r := OverworldFilletRadius
	shelf := OverworldShelfWidth
	const steps = 8

	var cx, cy float32
	var cornerX, cornerY float32
	var startAngle, endAngle float64

	switch corner {
	case cave.CornerNW:
		cornerX, cornerY = qx, qy
		cx, cy = qx+r, qy+r
		startAngle, endAngle = math.Pi, 1.5*math.Pi
	case cave.CornerNE:
		cornerX, cornerY = qx+half, qy
		cx, cy = qx+half-r, qy+r
		startAngle, endAngle = 1.5*math.Pi, 2.0*math.Pi
	case cave.CornerSW:
		cornerX, cornerY = qx, qy+half
		cx, cy = qx+r, qy+half-r
		startAngle, endAngle = 0.5*math.Pi, math.Pi
	case cave.CornerSE:
		cornerX, cornerY = qx+half, qy+half
		cx, cy = qx+half-r, qy+half-r
		startAngle, endAngle = 0.0, 0.5*math.Pi
	}

	// 1. Wet shelf fillet fan from (cornerX, cornerY) to arc
	angleStep := (endAngle - startAngle) / float64(steps)
	var prevX, prevY float32
	for i := 0; i <= steps; i++ {
		theta := startAngle + float64(i)*angleStep
		px := cx + float32(math.Cos(theta))*r
		py := cy + float32(math.Sin(theta))*r
		if i > 0 {
			fillTriangleOverworld(target, cornerX, cornerY, prevX, prevY, px, py, coastBase)
		}
		prevX, prevY = px, py
	}

	// 2. Inland basalt core fillet fan (inset towards cornerX, cornerY)
	rIn := r - shelf*0.5
	if rIn > 4.0 {
		var prevInX, prevInY float32
		for i := 0; i <= steps; i++ {
			theta := startAngle + float64(i)*angleStep
			px := cx + float32(math.Cos(theta))*rIn
			py := cy + float32(math.Sin(theta))*rIn
			if i > 0 {
				fillTriangleOverworld(target, cornerX, cornerY, prevInX, prevInY, px, py, landBase)
			}
			prevInX, prevInY = px, py
		}
	}

	// 3. Smooth concave arc stroke facing the water
	drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.4, coastStroke)
}

func fillTriangleOverworld(screen *ebiten.Image, x0, y0, x1, y1, x2, y2 float32, clr color.RGBA) {
	vs := []ebiten.Vertex{
		{DstX: x0, DstY: y0, ColorR: float32(clr.R) / 255, ColorG: float32(clr.G) / 255, ColorB: float32(clr.B) / 255, ColorA: float32(clr.A) / 255},
		{DstX: x1, DstY: y1, ColorR: float32(clr.R) / 255, ColorG: float32(clr.G) / 255, ColorB: float32(clr.B) / 255, ColorA: float32(clr.A) / 255},
		{DstX: x2, DstY: y2, ColorR: float32(clr.R) / 255, ColorG: float32(clr.G) / 255, ColorB: float32(clr.B) / 255, ColorA: float32(clr.A) / 255},
	}
	indices := []uint16{0, 1, 2}
	screen.DrawTriangles(vs, indices, whiteSubPixelImage(), &ebiten.DrawTrianglesOptions{})
}

var whiteSubPixel *ebiten.Image

func whiteSubPixelImage() *ebiten.Image {
	if whiteSubPixel == nil {
		whiteSubPixel = ebiten.NewImage(1, 1)
		whiteSubPixel.Fill(color.White)
	}
	return whiteSubPixel
}

// DrawOverworldCoastalFoam paints pulsing dynamic surf wash arcs along contoured coastlines.
func DrawOverworldCoastalFoam(target *ebiten.Image, w *world.World, tx, ty int, camX, camY float64, ticks int, mult float64) {
	isLand := IsWorldLand(w, tx, ty)
	half := float32(config.TileSize) / 2.0
	sx := float32(tx*config.TileSize) - float32(camX)
	sy := float32(ty*config.TileSize) - float32(camY)
	worldX := float32(tx * config.TileSize)
	worldY := float32(ty * config.TileSize)

	// Pulsing rhythm synced to world coordinates and game ticks
	phase := math.Sin(float64(ticks)*0.07 + float64(tx*7+ty*11)*0.3)
	if phase < -0.3 {
		return
	}
	alphaVal := float64(phase+0.3) / 1.3
	if alphaVal > 1.0 {
		alphaVal = 1.0
	}
	alpha := uint8(alphaVal * 170.0 * mult)
	if alpha < 12 {
		return
	}
	foamClr := color.RGBA{235, 246, 255, alpha}
	pulse := float32(phase+0.3) * 1.5 // 0.0 to 1.95 px offset outward into water

	quads := [4]struct {
		id int
		qx float32
		qy float32
		wx float32
		wy float32
	}{
		{cave.QuadNW, sx, sy, worldX, worldY},
		{cave.QuadNE, sx + half, sy, worldX + half, worldY},
		{cave.QuadSW, sx, sy + half, worldX, worldY + half},
		{cave.QuadSE, sx + half, sy + half, worldX + half, worldY + half},
	}

	for _, q := range quads {
		info := ClassifyOverworldQuadrant(w, tx, ty, q.id)
		if isLand {
			switch info.Shape {
			case cave.ShapeEdgeTop:
				drawOverworldWavyHorizFoam(target, q.qx, q.qy-pulse, half, q.wx, q.wy, 1.5, foamClr)
			case cave.ShapeEdgeBottom:
				drawOverworldWavyHorizFoam(target, q.qx, q.qy+half+pulse, half, q.wx, q.wy+half, 1.5, foamClr)
			case cave.ShapeEdgeLeft:
				drawOverworldWavyVertFoam(target, q.qx-pulse, q.qy, half, q.wx, q.wy, 1.5, foamClr)
			case cave.ShapeEdgeRight:
				drawOverworldWavyVertFoam(target, q.qx+half+pulse, q.qy, half, q.wx+half, q.wy, 1.5, foamClr)
			case cave.ShapeOuterCorner:
				steps := 8
				r := OverworldCornerRadius + pulse
				var cx, cy float32
				var startAngle, endAngle float64
				switch info.Corner {
				case cave.CornerNW:
					cx, cy = q.qx+OverworldCornerRadius, q.qy+OverworldCornerRadius
					startAngle, endAngle = math.Pi, 1.5*math.Pi
					drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.5, foamClr)
					drawOverworldWavyHorizFoam(target, q.qx+OverworldCornerRadius, q.qy-pulse, half-OverworldCornerRadius, q.wx+OverworldCornerRadius, q.wy, 1.5, foamClr)
					drawOverworldWavyVertFoam(target, q.qx-pulse, q.qy+OverworldCornerRadius, half-OverworldCornerRadius, q.wx, q.wy+OverworldCornerRadius, 1.5, foamClr)
				case cave.CornerNE:
					cx, cy = q.qx+half-OverworldCornerRadius, q.qy+OverworldCornerRadius
					startAngle, endAngle = 1.5*math.Pi, 2.0*math.Pi
					drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.5, foamClr)
					drawOverworldWavyHorizFoam(target, q.qx, q.qy-pulse, half-OverworldCornerRadius, q.wx, q.wy, 1.5, foamClr)
					drawOverworldWavyVertFoam(target, q.qx+half+pulse, q.qy+OverworldCornerRadius, half-OverworldCornerRadius, q.wx+half, q.wy+OverworldCornerRadius, 1.5, foamClr)
				case cave.CornerSW:
					cx, cy = q.qx+OverworldCornerRadius, q.qy+half-OverworldCornerRadius
					startAngle, endAngle = 0.5*math.Pi, math.Pi
					drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.5, foamClr)
					drawOverworldWavyHorizFoam(target, q.qx+OverworldCornerRadius, q.qy+half+pulse, half-OverworldCornerRadius, q.wx+OverworldCornerRadius, q.wy+half, 1.5, foamClr)
					drawOverworldWavyVertFoam(target, q.qx-pulse, q.qy, half-OverworldCornerRadius, q.wx, q.wy, 1.5, foamClr)
				case cave.CornerSE:
					cx, cy = q.qx+half-OverworldCornerRadius, q.qy+half-OverworldCornerRadius
					startAngle, endAngle = 0.0, 0.5*math.Pi
					drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.5, foamClr)
					drawOverworldWavyHorizFoam(target, q.qx, q.qy+half+pulse, half-OverworldCornerRadius, q.wx, q.wy+half, 1.5, foamClr)
					drawOverworldWavyVertFoam(target, q.qx+half+pulse, q.qy, half-OverworldCornerRadius, q.wx+half, q.wy, 1.5, foamClr)
				}
			}
		} else if info.Shape == cave.ShapeInnerFillet {
			steps := 8
			r := OverworldFilletRadius - pulse
			var cx, cy float32
			var startAngle, endAngle float64
			switch info.Corner {
			case cave.CornerNW:
				cx, cy = q.qx+OverworldFilletRadius, q.qy+OverworldFilletRadius
				startAngle, endAngle = math.Pi, 1.5*math.Pi
			case cave.CornerNE:
				cx, cy = q.qx+half-OverworldFilletRadius, q.qy+OverworldFilletRadius
				startAngle, endAngle = 1.5*math.Pi, 2.0*math.Pi
			case cave.CornerSW:
				cx, cy = q.qx+OverworldFilletRadius, q.qy+half-OverworldFilletRadius
				startAngle, endAngle = 0.5*math.Pi, math.Pi
			case cave.CornerSE:
				cx, cy = q.qx+half-OverworldFilletRadius, q.qy+half-OverworldFilletRadius
				startAngle, endAngle = 0.0, 0.5*math.Pi
			}
			drawArcStroke(target, cx, cy, r, startAngle, endAngle, steps, 1.5, foamClr)
		}
	}
}

// CheckOverworldSolid evaluates contoured coastlines against a bounding box [x, y, w, h].
// Returns isSolid, and if hitting a 45°/rounded slope, returns isSlope=true and surface normal (nx, ny).
func CheckOverworldSolid(worldObj *world.World, tx, ty int, x, y, wDim, hDim float64) (isSolid bool, isSlope bool, nx, ny float64) {
	solid := IsWorldLand(worldObj, tx, ty)
	ts := float64(config.TileSize)
	tileX := float64(tx * config.TileSize)
	tileY := float64(ty * config.TileSize)

	overlapLeft := math.Max(x, tileX)
	overlapRight := math.Min(x+wDim, tileX+ts)
	overlapTop := math.Max(y, tileY)
	overlapBottom := math.Min(y+hDim, tileY+ts)

	if overlapRight <= overlapLeft || overlapBottom <= overlapTop {
		return false, false, 0, 0
	}

	// 1. Water tile: check if bounding box intersects an inner rock fillet
	if !solid {
		r := float64(OverworldFilletRadius)
		if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadNW).Shape == cave.ShapeInnerFillet {
			if x < tileX+r && y < tileY+r {
				if (x-tileX)+(y-tileY) < r {
					return true, true, 0.7071, 0.7071
				}
			}
		}
		if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadNE).Shape == cave.ShapeInnerFillet {
			if x+wDim > tileX+ts-r && y < tileY+r {
				if (tileX+ts-(x+wDim))+(y-tileY) < r {
					return true, true, -0.7071, 0.7071
				}
			}
		}
		if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadSW).Shape == cave.ShapeInnerFillet {
			if x < tileX+r && y+hDim > tileY+ts-r {
				if (x-tileX)+(tileY+ts-(y+hDim)) < r {
					return true, true, 0.7071, -0.7071
				}
			}
		}
		if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadSE).Shape == cave.ShapeInnerFillet {
			if x+wDim > tileX+ts-r && y+hDim > tileY+ts-r {
				if (tileX+ts-(x+wDim))+(tileY+ts-(y+hDim)) < r {
					return true, true, -0.7071, -0.7071
				}
			}
		}
		return false, false, 0, 0
	}

	// 2. Solid land tile: Check outer corners. The cut-off triangle is passable water!
	c := float64(OverworldCornerRadius)

	if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadNW).Shape == cave.ShapeOuterCorner {
		if overlapRight <= tileX+c && overlapBottom <= tileY+c {
			relX := overlapRight - tileX
			relY := overlapBottom - tileY
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, -0.7071, -0.7071
		}
	}

	if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadNE).Shape == cave.ShapeOuterCorner {
		if overlapLeft >= tileX+ts-c && overlapBottom <= tileY+c {
			relX := tileX + ts - overlapLeft
			relY := overlapBottom - tileY
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, 0.7071, -0.7071
		}
	}

	if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadSW).Shape == cave.ShapeOuterCorner {
		if overlapRight <= tileX+c && overlapTop >= tileY+ts-c {
			relX := overlapRight - tileX
			relY := tileY + ts - overlapTop
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, -0.7071, 0.7071
		}
	}

	if ClassifyOverworldQuadrant(worldObj, tx, ty, cave.QuadSE).Shape == cave.ShapeOuterCorner {
		if overlapLeft >= tileX+ts-c && overlapTop >= tileY+ts-c {
			relX := tileX + ts - overlapLeft
			relY := tileY + ts - overlapTop
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, 0.7071, 0.7071
		}
	}

	return true, false, 0, 0
}

// MovePlayerOverworld handles player and boat movement with 45° and curved coastline sliding deflection.
func (o *OverworldScene) MovePlayerOverworld(p *player.Player, dims gvec.Vec2, isSolid func(gvec.Vec2) bool, isPiloting bool) {
	newPos := p.Pos.Add(p.Vel)
	if !isSolid(newPos) {
		p.Pos = newPos
		return
	}

	x1, x2, y1, y2 := tileRange(newPos.X, newPos.Y, dims.X, dims.Y, config.TileSize)
	var hitSlope bool
	var normalX, normalY float64

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if isSol, isSlp, nx, ny := CheckOverworldSolid(o.World, tx, ty, newPos.X, newPos.Y, dims.X, dims.Y); isSol && isSlp {
				hitSlope = true
				normalX += nx
				normalY += ny
			}
		}
	}

	if hitSlope {
		lenN := math.Hypot(normalX, normalY)
		if lenN > 0.001 {
			normalX /= lenN
			normalY /= lenN

			dot := p.Vel.X*normalX + p.Vel.Y*normalY
			if dot < 0 {
				speed := math.Hypot(p.Vel.X, p.Vel.Y)
				p.Vel.X -= dot * normalX
				p.Vel.Y -= dot * normalY

				deflectedPos := p.Pos.Add(p.Vel)
				if !isSolid(deflectedPos) {
					p.Pos = deflectedPos

					// Play hull scrape sound effect if moving in vehicle at speed
					if isPiloting && speed > 1.2 {
						audio.Get().PlaySFXVaried("sfx/sub_hull_creak.wav", 0.45, 0.08)
					}
					return
				}
			}
		}
	}

	gvec.MoveAxisSeparated(&p.Pos, &p.Vel, dims, isSolid, nil, nil)
}
