package cave

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// Quadrant identifiers within a single tile.
const (
	QuadNW = 0
	QuadNE = 1
	QuadSW = 2
	QuadSE = 3
)

// QuadrantShape specifies the geometric shape of a 32x32 quadrant.
type QuadrantShape int

const (
	ShapeEmpty         QuadrantShape = iota
	ShapeFullSolid                   // Completely solid interior
	ShapeOuterCorner                 // Outer rounded/chamfered corner
	ShapeInnerFillet                 // Inner rounded/bracket corner (in empty cell)
	ShapeEdgeTop                     // Exposed horizontal edge at top
	ShapeEdgeBottom                  // Exposed horizontal edge at bottom
	ShapeEdgeLeft                    // Exposed vertical edge at left
	ShapeEdgeRight                   // Exposed vertical edge at right
	ShapeDiagonalSlope               // Full 45° diagonal ramp
)

// QuadrantInfo contains the classified shape and orientation of a 32x32 sub-tile.
type QuadrantInfo struct {
	Shape  QuadrantShape
	Corner int
}

const (
	CornerNW = 0
	CornerNE = 1
	CornerSW = 2
	CornerSE = 3
)

// WreckChamferSize is the pixel size of 45-degree bevels on 90-degree corners.
const WreckChamferSize float32 = 22.0

// WreckFilletSize is the pixel size of 45-degree inner brackets.
const WreckFilletSize float32 = 14.0

// CaveCornerRadius is the baseline radius of outer corner rounding in caves.
const CaveCornerRadius float32 = 20.0

// CaveFilletRadius is the baseline radius of inner corner fillets in caves.
const CaveFilletRadius float32 = 16.0

// IsGridSolid checks if a tile coordinate is solid, treating cave boundaries as solid rock.
func IsGridSolid(grid [][]bool, tx, ty int) bool {
	if grid == nil || len(grid) == 0 {
		return true
	}
	w := len(grid)
	h := len(grid[0])
	if tx < 0 || tx >= w {
		return true
	}
	if ty < 0 {
		return false // Sky / surface water is open
	}
	if ty >= h {
		return true
	}
	return grid[tx][ty]
}

// ClassifyQuadrant determines the geometric shape of quadrant (quad) within tile (tx, ty).
// quad is one of QuadNW, QuadNE, QuadSW, QuadSE.
func ClassifyQuadrant(grid [][]bool, tx, ty int, quad int) QuadrantInfo {
	solid := IsGridSolid(grid, tx, ty)

	var nX, nY int // orthogonal neighbor 1
	var oX, oY int // orthogonal neighbor 2
	var dX, dY int // diagonal neighbor
	var cornerID int

	switch quad {
	case QuadNW:
		nX, nY = tx, ty-1   // North
		oX, oY = tx-1, ty   // West
		dX, dY = tx-1, ty-1 // North-West
		cornerID = CornerNW
	case QuadNE:
		nX, nY = tx, ty-1   // North
		oX, oY = tx+1, ty   // East
		dX, dY = tx+1, ty-1 // North-East
		cornerID = CornerNE
	case QuadSW:
		nX, nY = tx, ty+1   // South
		oX, oY = tx-1, ty   // West
		dX, dY = tx-1, ty+1 // South-West
		cornerID = CornerSW
	case QuadSE:
		nX, nY = tx, ty+1   // South
		oX, oY = tx+1, ty   // East
		dX, dY = tx+1, ty+1 // South-East
		cornerID = CornerSE
	}

	nSolid := IsGridSolid(grid, nX, nY)
	oSolid := IsGridSolid(grid, oX, oY)
	dSolid := IsGridSolid(grid, dX, dY)

	if !solid {
		// Empty tile: check if this quadrant forms an Inner Corner Fillet bridging two solid walls
		if nSolid && oSolid {
			return QuadrantInfo{Shape: ShapeInnerFillet, Corner: cornerID}
		}
		return QuadrantInfo{Shape: ShapeEmpty, Corner: cornerID}
	}

	// Solid tile:
	if !nSolid && !oSolid {
		// Both orthogonal neighbors empty: Outer Corner!
		return QuadrantInfo{Shape: ShapeOuterCorner, Corner: cornerID}
	}

	if !nSolid && oSolid {
		// Exposed edge along the N/S neighbor
		if quad == QuadNW || quad == QuadNE {
			return QuadrantInfo{Shape: ShapeEdgeTop, Corner: cornerID}
		}
		return QuadrantInfo{Shape: ShapeEdgeBottom, Corner: cornerID}
	}

	if nSolid && !oSolid {
		// Exposed edge along the W/E neighbor
		if quad == QuadNW || quad == QuadSW {
			return QuadrantInfo{Shape: ShapeEdgeLeft, Corner: cornerID}
		}
		return QuadrantInfo{Shape: ShapeEdgeRight, Corner: cornerID}
	}

	// nSolid && oSolid:
	if !dSolid {
		return QuadrantInfo{Shape: ShapeFullSolid, Corner: cornerID}
	}

	return QuadrantInfo{Shape: ShapeFullSolid, Corner: cornerID}
}

// deterministicBump calculates stable, continuous boundary offsets based on world coordinates.
func deterministicBump(worldX, worldY float64) float32 {
	s1 := math.Sin(worldX*0.09) * math.Cos(worldY*0.11) * 2.2
	s2 := math.Sin((worldX+worldY)*0.17) * 1.4

	ix := int64(math.Floor(worldX / 8.0))
	iy := int64(math.Floor(worldY / 8.0))
	h := uint64(ix*73856093 ^ iy*19349663)
	h ^= h >> 13
	hFrac := (float64(h%100)/50.0 - 1.0) * 0.8

	return float32(s1 + s2 + hFrac)
}

// DrawOrganicCaveTile renders a single tile (solid or with inner fillets) with natural flowing curves and bumps.
func DrawOrganicCaveTile(
	screen *ebiten.Image,
	grid [][]bool,
	tx, ty int,
	camX, camY float64,
	rockColor, strokeColor color.RGBA,
	interiorDecorator func(sx, sy float32),
) {
	solid := IsGridSolid(grid, tx, ty)
	ts := float32(config.TileSize)
	half := ts / 2.0
	baseSX := float32(tx*config.TileSize - int(camX))
	baseSY := float32(ty*config.TileSize - int(camY))

	// If empty tile, only draw inner corner fillets that bridge adjacent solid walls
	if !solid {
		for q := 0; q < 4; q++ {
			info := ClassifyQuadrant(grid, tx, ty, q)
			if info.Shape == ShapeInnerFillet {
				qx, qy := getQuadOrigin(baseSX, baseSY, half, q)
				drawOrganicInnerFillet(screen, tx, ty, qx, qy, half, info.Corner, camX, camY, rockColor, strokeColor)
			}
		}
		return
	}

	// Solid tile: Draw each of the 4 quadrants
	for q := 0; q < 4; q++ {
		info := ClassifyQuadrant(grid, tx, ty, q)
		qx, qy := getQuadOrigin(baseSX, baseSY, half, q)

		switch info.Shape {
		case ShapeFullSolid:
			vector.FillRect(screen, qx, qy, half, half, rockColor, false)

		case ShapeOuterCorner:
			drawOrganicOuterCorner(screen, tx, ty, qx, qy, half, info.Corner, camX, camY, rockColor, strokeColor)

		case ShapeEdgeTop:
			drawOrganicEdgeTop(screen, tx, ty, qx, qy, half, camX, camY, rockColor, strokeColor)

		case ShapeEdgeBottom:
			drawOrganicEdgeBottom(screen, tx, ty, qx, qy, half, camX, camY, rockColor, strokeColor)

		case ShapeEdgeLeft:
			drawOrganicEdgeLeft(screen, tx, ty, qx, qy, half, camX, camY, rockColor, strokeColor)

		case ShapeEdgeRight:
			drawOrganicEdgeRight(screen, tx, ty, qx, qy, half, camX, camY, rockColor, strokeColor)
		}
	}

	// Apply optional interior decorator (e.g. strata, basalt cracks, moss, mineral veins)
	if interiorDecorator != nil {
		interiorDecorator(baseSX, baseSY)
	}
}

func getQuadOrigin(baseSX, baseSY, half float32, quad int) (float32, float32) {
	switch quad {
	case QuadNW:
		return baseSX, baseSY
	case QuadNE:
		return baseSX + half, baseSY
	case QuadSW:
		return baseSX, baseSY + half
	case QuadSE:
		return baseSX + half, baseSY + half
	}
	return baseSX, baseSY
}

// drawOrganicOuterCorner renders a quadrant where two orthogonal sides are open water, curving smoothly inward.
func drawOrganicOuterCorner(
	screen *ebiten.Image,
	tx, ty int,
	qx, qy, half float32,
	cornerID int,
	camX, camY float64,
	rockColor, strokeColor color.RGBA,
) {
	r := CaveCornerRadius

	var innerFillX, innerFillY, innerFillW, innerFillH float32
	var p1X, p1Y, p2X, p2Y float32
	var cornerX, cornerY float32

	switch cornerID {
	case CornerNW:
		cornerX, cornerY = qx, qy
		p1X, p1Y = qx, qy+r
		p2X, p2Y = qx+r, qy
		innerFillX, innerFillY = qx+r, qy+r
		innerFillW, innerFillH = half-r, half-r
		vector.FillRect(screen, qx+r, qy, half-r, r, rockColor, false)
		vector.FillRect(screen, qx, qy+r, r, half-r, rockColor, false)
		vector.FillRect(screen, innerFillX, innerFillY, innerFillW, innerFillH, rockColor, false)

	case CornerNE:
		cornerX, cornerY = qx+half, qy
		p1X, p1Y = qx+half-r, qy
		p2X, p2Y = qx+half, qy+r
		innerFillX, innerFillY = qx, qy+r
		innerFillW, innerFillH = half-r, half-r
		vector.FillRect(screen, qx, qy, half-r, r, rockColor, false)
		vector.FillRect(screen, qx+half-r, qy+r, r, half-r, rockColor, false)
		vector.FillRect(screen, innerFillX, innerFillY, innerFillW, innerFillH, rockColor, false)

	case CornerSW:
		cornerX, cornerY = qx, qy+half
		p1X, p1Y = qx, qy+half-r
		p2X, p2Y = qx+r, qy+half
		innerFillX, innerFillY = qx+r, qy
		innerFillW, innerFillH = half-r, half-r
		vector.FillRect(screen, qx+r, qy+half-r, half-r, r, rockColor, false)
		vector.FillRect(screen, qx, qy, r, half-r, rockColor, false)
		vector.FillRect(screen, innerFillX, innerFillY, innerFillW, innerFillH, rockColor, false)

	case CornerSE:
		cornerX, cornerY = qx+half, qy+half
		p1X, p1Y = qx+half-r, qy+half
		p2X, p2Y = qx+half, qy+half-r
		innerFillX, innerFillY = qx, qy
		innerFillW, innerFillH = half-r, half-r
		vector.FillRect(screen, qx, qy+half-r, half-r, r, rockColor, false)
		vector.FillRect(screen, qx+half-r, qy, r, half-r, rockColor, false)
		vector.FillRect(screen, innerFillX, innerFillY, innerFillW, innerFillH, rockColor, false)
	}

	worldMidX := float64(cornerX) + camX
	worldMidY := float64(cornerY) + camY
	bump := deterministicBump(worldMidX, worldMidY)

	midX := (p1X+p2X)*0.5 + bump*0.7
	midY := (p1Y+p2Y)*0.5 + bump*0.7

	fillTriangle(screen, p1X, p1Y, p2X, p2Y, innerFillX, innerFillY, rockColor)
	fillTriangle(screen, p1X, p1Y, p2X, p2Y, midX, midY, rockColor)

	vector.StrokeLine(screen, p1X, p1Y, midX, midY, 1.2, strokeColor, false)
	vector.StrokeLine(screen, midX, midY, p2X, p2Y, 1.2, strokeColor, false)

	switch cornerID {
	case CornerNW:
		strokeBumpyLine(screen, qx+r, qy, qx+half, qy, camX, camY, strokeColor)
		strokeBumpyLine(screen, qx, qy+r, qx, qy+half, camX, camY, strokeColor)
	case CornerNE:
		strokeBumpyLine(screen, qx, qy, qx+half-r, qy, camX, camY, strokeColor)
		strokeBumpyLine(screen, qx+half, qy+r, qx+half, qy+half, camX, camY, strokeColor)
	case CornerSW:
		strokeBumpyLine(screen, qx+r, qy+half, qx+half, qy+half, camX, camY, strokeColor)
		strokeBumpyLine(screen, qx, qy, qx, qy+half-r, camX, camY, strokeColor)
	case CornerSE:
		strokeBumpyLine(screen, qx, qy+half, qx+half-r, qy+half, camX, camY, strokeColor)
		strokeBumpyLine(screen, qx+half, qy, qx+half, qy+half-r, camX, camY, strokeColor)
	}
}

// drawOrganicInnerFillet renders a concave rock fillet in an empty water quadrant that bridges two solid perpendicular walls.
func drawOrganicInnerFillet(
	screen *ebiten.Image,
	tx, ty int,
	qx, qy, half float32,
	cornerID int,
	camX, camY float64,
	rockColor, strokeColor color.RGBA,
) {
	r := CaveFilletRadius
	var cornerX, cornerY, p1X, p1Y, p2X, p2Y float32

	switch cornerID {
	case CornerNW:
		cornerX, cornerY = qx, qy
		p1X, p1Y = qx, qy+r
		p2X, p2Y = qx+r, qy
	case CornerNE:
		cornerX, cornerY = qx+half, qy
		p1X, p1Y = qx+half-r, qy
		p2X, p2Y = qx+half, qy+r
	case CornerSW:
		cornerX, cornerY = qx, qy+half
		p1X, p1Y = qx, qy+half-r
		p2X, p2Y = qx+r, qy+half
	case CornerSE:
		cornerX, cornerY = qx+half, qy+half
		p1X, p1Y = qx+half-r, qy+half
		p2X, p2Y = qx+half, qy+half-r
	}

	worldMidX := float64(cornerX) + camX
	worldMidY := float64(cornerY) + camY
	bump := deterministicBump(worldMidX, worldMidY)

	inwardX := (cornerX + p1X + p2X) / 3.0 + bump*0.5
	inwardY := (cornerY + p1Y + p2Y) / 3.0 + bump*0.5

	fillTriangle(screen, cornerX, cornerY, p1X, p1Y, inwardX, inwardY, rockColor)
	fillTriangle(screen, cornerX, cornerY, p2X, p2Y, inwardX, inwardY, rockColor)

	vector.StrokeLine(screen, p1X, p1Y, inwardX, inwardY, 1.2, strokeColor, false)
	vector.StrokeLine(screen, inwardX, inwardY, p2X, p2Y, 1.2, strokeColor, false)
}

func drawOrganicEdgeTop(screen *ebiten.Image, tx, ty int, qx, qy, half float32, camX, camY float64, rockColor, strokeColor color.RGBA) {
	vector.FillRect(screen, qx, qy, half, half, rockColor, false)
	strokeBumpyLine(screen, qx, qy, qx+half, qy, camX, camY, strokeColor)
}

func drawOrganicEdgeBottom(screen *ebiten.Image, tx, ty int, qx, qy, half float32, camX, camY float64, rockColor, strokeColor color.RGBA) {
	vector.FillRect(screen, qx, qy, half, half, rockColor, false)
	strokeBumpyLine(screen, qx, qy+half, qx+half, qy+half, camX, camY, strokeColor)
}

func drawOrganicEdgeLeft(screen *ebiten.Image, tx, ty int, qx, qy, half float32, camX, camY float64, rockColor, strokeColor color.RGBA) {
	vector.FillRect(screen, qx, qy, half, half, rockColor, false)
	strokeBumpyLine(screen, qx, qy, qx, qy+half, camX, camY, strokeColor)
}

func drawOrganicEdgeRight(screen *ebiten.Image, tx, ty int, qx, qy, half float32, camX, camY float64, rockColor, strokeColor color.RGBA) {
	vector.FillRect(screen, qx, qy, half, half, rockColor, false)
	strokeBumpyLine(screen, qx+half, qy, qx+half, qy+half, camX, camY, strokeColor)
}

func strokeBumpyLine(screen *ebiten.Image, x1, y1, x2, y2 float32, camX, camY float64, strokeColor color.RGBA) {
	const segs = 3
	lastX, lastY := x1, y1

	dx := (x2 - x1) / float32(segs)
	dy := (y2 - y1) / float32(segs)

	lenLine := float32(math.Hypot(float64(x2-x1), float64(y2-y1)))
	if lenLine < 0.1 {
		return
	}
	normX := -(y2 - y1) / lenLine
	normY := (x2 - x1) / lenLine

	for i := 1; i <= segs; i++ {
		targetX := x1 + dx*float32(i)
		targetY := y1 + dy*float32(i)

		if i < segs {
			worldX := float64(targetX) + camX
			worldY := float64(targetY) + camY
			b := deterministicBump(worldX, worldY)
			targetX += normX * b
			targetY += normY * b
		}

		vector.StrokeLine(screen, lastX, lastY, targetX, targetY, 1.0, strokeColor, false)
		lastX, lastY = targetX, targetY
	}
}

// fillTriangle fills a triangle defined by three 2D vertices using small trapezoids.
func fillTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, clr color.RGBA) {
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	if y1 > y3 {
		x1, x3 = x3, x1
		y1, y3 = y3, y1
	}
	if y2 > y3 {
		x2, x3 = x3, x2
		y2, y3 = y3, y2
	}

	totalH := y3 - y1
	if totalH <= 0.2 {
		return
	}

	for y := y1; y <= y3; y += 1.0 {
		secondHalf := y > y2 || y2 == y1
		var segmentH float32
		if secondHalf {
			segmentH = y3 - y2
		} else {
			segmentH = y2 - y1
		}
		if segmentH <= 0.05 {
			continue
		}

		alpha := (y - y1) / totalH
		var beta float32
		if secondHalf {
			beta = (y - y2) / segmentH
		} else {
			beta = (y - y1) / segmentH
		}

		ax := x1 + (x3-x1)*alpha
		var bx float32
		if secondHalf {
			bx = x2 + (x3-x2)*beta
		} else {
			bx = x1 + (x2-x1)*beta
		}

		if ax > bx {
			ax, bx = bx, ax
		}
		vector.FillRect(screen, ax, y, bx-ax+1.0, 1.0, clr, false)
	}
}

// DrawWreckageTile renders a single wreckage tile with 45° chamfers, diagonal ramps, rivets, and hazard stripes.
func DrawWreckageTile(
	screen *ebiten.Image,
	grid [][]bool,
	tx, ty int,
	camX, camY float64,
	rockColor, strokeColor, stripeColor color.RGBA,
) {
	solid := IsGridSolid(grid, tx, ty)
	ts := float32(config.TileSize)
	half := ts / 2.0
	baseSX := float32(tx*config.TileSize - int(camX))
	baseSY := float32(ty*config.TileSize - int(camY))

	// If empty tile, check for 45° inner corner brackets
	if !solid {
		for q := 0; q < 4; q++ {
			info := ClassifyQuadrant(grid, tx, ty, q)
			if info.Shape == ShapeInnerFillet {
				qx, qy := getQuadOrigin(baseSX, baseSY, half, q)
				drawWreckageInnerBracket(screen, qx, qy, half, info.Corner, rockColor, strokeColor)
			}
		}
		return
	}

	// Solid tile:
	for q := 0; q < 4; q++ {
		info := ClassifyQuadrant(grid, tx, ty, q)
		qx, qy := getQuadOrigin(baseSX, baseSY, half, q)

		switch info.Shape {
		case ShapeFullSolid:
			vector.FillRect(screen, qx, qy, half, half, rockColor, false)
			vector.StrokeRect(screen, qx, qy, half, half, 0.6, strokeColor, false)

		case ShapeOuterCorner:
			drawWreckageOuterChamfer(screen, qx, qy, half, info.Corner, rockColor, strokeColor, stripeColor)

		case ShapeEdgeTop:
			vector.FillRect(screen, qx, qy, half, half, rockColor, false)
			vector.StrokeLine(screen, qx, qy, qx+half, qy, 1.2, strokeColor, false)
			drawWreckageHazardStripe(screen, qx, qy, half, half, 0, stripeColor)

		case ShapeEdgeBottom:
			vector.FillRect(screen, qx, qy, half, half, rockColor, false)
			vector.StrokeLine(screen, qx, qy+half, qx+half, qy+half, 1.2, strokeColor, false)
			drawWreckageHazardStripe(screen, qx, qy, half, half, 2, stripeColor)

		case ShapeEdgeLeft:
			vector.FillRect(screen, qx, qy, half, half, rockColor, false)
			vector.StrokeLine(screen, qx, qy, qx, qy+half, 1.2, strokeColor, false)
			drawWreckageHazardStripe(screen, qx, qy, half, half, 3, stripeColor)

		case ShapeEdgeRight:
			vector.FillRect(screen, qx, qy, half, half, rockColor, false)
			vector.StrokeLine(screen, qx+half, qy, qx+half, qy+half, 1.2, strokeColor, false)
			drawWreckageHazardStripe(screen, qx, qy, half, half, 1, stripeColor)
		}
	}
}

// drawWreckageOuterChamfer cuts the protruding 90° corner at 45° with an angled plate and rivets.
func drawWreckageOuterChamfer(
	screen *ebiten.Image,
	qx, qy, half float32,
	cornerID int,
	rockColor, strokeColor, stripeColor color.RGBA,
) {
	c := WreckChamferSize
	var p1X, p1Y, p2X, p2Y float32
	var innerX, innerY, innerW, innerH float32

	switch cornerID {
	case CornerNW:
		p1X, p1Y = qx, qy+c
		p2X, p2Y = qx+c, qy
		innerX, innerY = qx+c, qy+c
		innerW, innerH = half-c, half-c
		vector.FillRect(screen, qx+c, qy, half-c, c, rockColor, false)
		vector.FillRect(screen, qx, qy+c, c, half-c, rockColor, false)
		vector.FillRect(screen, innerX, innerY, innerW, innerH, rockColor, false)
		fillTriangle(screen, p1X, p1Y, p2X, p2Y, innerX, innerY, rockColor)
		vector.StrokeLine(screen, p1X, p1Y, p2X, p2Y, 1.4, strokeColor, false)
		vector.FillCircle(screen, p1X+1.5, p1Y, 1.2, strokeColor, false)
		vector.FillCircle(screen, p2X, p2Y+1.5, 1.2, strokeColor, false)
		vector.StrokeLine(screen, p1X+4, p1Y, p2X, p2Y+4, 1.8, stripeColor, false)

	case CornerNE:
		p1X, p1Y = qx+half-c, qy
		p2X, p2Y = qx+half, qy+c
		innerX, innerY = qx, qy+c
		innerW, innerH = half-c, half-c
		vector.FillRect(screen, qx, qy, half-c, c, rockColor, false)
		vector.FillRect(screen, qx+half-c, qy+c, c, half-c, rockColor, false)
		vector.FillRect(screen, innerX, innerY, innerW, innerH, rockColor, false)
		fillTriangle(screen, p1X, p1Y, p2X, p2Y, qx+half-c, qy+c, rockColor)
		vector.StrokeLine(screen, p1X, p1Y, p2X, p2Y, 1.4, strokeColor, false)
		vector.FillCircle(screen, p1X, p1Y+1.5, 1.2, strokeColor, false)
		vector.FillCircle(screen, p2X-1.5, p2Y, 1.2, strokeColor, false)
		vector.StrokeLine(screen, p1X, p1Y+4, p2X-4, p2Y, 1.8, stripeColor, false)

	case CornerSW:
		p1X, p1Y = qx, qy+half-c
		p2X, p2Y = qx+c, qy+half
		innerX, innerY = qx+c, qy
		innerW, innerH = half-c, half-c
		vector.FillRect(screen, qx+c, qy+half-c, half-c, c, rockColor, false)
		vector.FillRect(screen, qx, qy, c, half-c, rockColor, false)
		vector.FillRect(screen, innerX, innerY, innerW, innerH, rockColor, false)
		fillTriangle(screen, p1X, p1Y, p2X, p2Y, qx+c, qy+half-c, rockColor)
		vector.StrokeLine(screen, p1X, p1Y, p2X, p2Y, 1.4, strokeColor, false)
		vector.FillCircle(screen, p1X+1.5, p1Y, 1.2, strokeColor, false)
		vector.FillCircle(screen, p2X, p2Y-1.5, 1.2, strokeColor, false)
		vector.StrokeLine(screen, p1X+4, p1Y, p2X, p2Y-4, 1.8, stripeColor, false)

	case CornerSE:
		p1X, p1Y = qx+half-c, qy+half
		p2X, p2Y = qx+half, qy+half-c
		innerX, innerY = qx, qy
		innerW, innerH = half-c, half-c
		vector.FillRect(screen, qx, qy+half-c, half-c, c, rockColor, false)
		vector.FillRect(screen, qx+half-c, qy, c, half-c, rockColor, false)
		vector.FillRect(screen, innerX, innerY, innerW, innerH, rockColor, false)
		fillTriangle(screen, p1X, p1Y, p2X, p2Y, qx+half-c, qy+half-c, rockColor)
		vector.StrokeLine(screen, p1X, p1Y, p2X, p2Y, 1.4, strokeColor, false)
		vector.FillCircle(screen, p1X, p1Y-1.5, 1.2, strokeColor, false)
		vector.FillCircle(screen, p2X-1.5, p2Y, 1.2, strokeColor, false)
		vector.StrokeLine(screen, p1X, p1Y-4, p2X-4, p2Y, 1.8, stripeColor, false)
	}
}

// drawWreckageInnerBracket renders a 45° angled structural bracket inside an empty water quadrant.
func drawWreckageInnerBracket(
	screen *ebiten.Image,
	qx, qy, half float32,
	cornerID int,
	rockColor, strokeColor color.RGBA,
) {
	b := WreckFilletSize
	var cornerX, cornerY, p1X, p1Y, p2X, p2Y float32

	switch cornerID {
	case CornerNW:
		cornerX, cornerY = qx, qy
		p1X, p1Y = qx, qy+b
		p2X, p2Y = qx+b, qy
	case CornerNE:
		cornerX, cornerY = qx+half, qy
		p1X, p1Y = qx+half-b, qy
		p2X, p2Y = qx+half, qy+b
	case CornerSW:
		cornerX, cornerY = qx, qy+half
		p1X, p1Y = qx, qy+half-b
		p2X, p2Y = qx+b, qy+half
	case CornerSE:
		cornerX, cornerY = qx+half, qy+half
		p1X, p1Y = qx+half-b, qy+half
		p2X, p2Y = qx+half, qy+half-b
	}

	fillTriangle(screen, cornerX, cornerY, p1X, p1Y, p2X, p2Y, rockColor)
	vector.StrokeLine(screen, p1X, p1Y, p2X, p2Y, 1.4, strokeColor, false)
	midX := (p1X + p2X + cornerX) / 3.0
	midY := (p1Y + p2Y + cornerY) / 3.0
	vector.FillCircle(screen, midX, midY, 1.0, strokeColor, false)
}

func drawWreckageHazardStripe(screen *ebiten.Image, qx, qy, w, h float32, side int, stripeColor color.RGBA) {
	switch side {
	case 0: // Top
		vector.StrokeLine(screen, qx+4, qy+10, qx+14, qy, 1.8, stripeColor, false)
		vector.StrokeLine(screen, qx+18, qy+10, qx+28, qy, 1.8, stripeColor, false)
	case 2: // Bottom
		vector.StrokeLine(screen, qx+4, qy+h, qx+14, qy+h-10, 1.8, stripeColor, false)
		vector.StrokeLine(screen, qx+18, qy+h, qx+28, qy+h-10, 1.8, stripeColor, false)
	case 3: // Left
		vector.StrokeLine(screen, qx, qy+4, qx+10, qy+14, 1.8, stripeColor, false)
		vector.StrokeLine(screen, qx, qy+18, qx+10, qy+28, 1.8, stripeColor, false)
	case 1: // Right
		vector.StrokeLine(screen, qx+w-10, qy+4, qx+w, qy+14, 1.8, stripeColor, false)
		vector.StrokeLine(screen, qx+w-10, qy+18, qx+w, qy+28, 1.8, stripeColor, false)
	}
}

// CheckWreckageSolid evaluates 45° chamfers and sloped geometry for a bounding box [x, y, w, h].
// Returns true if penetrating, and provides the outward collision normal (nx, ny) if hitting a 45° slope.
func CheckWreckageSolid(grid [][]bool, tx, ty int, x, y, w, h float64) (isSolid bool, isSlope bool, nx, ny float64) {
	solid := IsGridSolid(grid, tx, ty)
	ts := float64(config.TileSize)
	tileX := float64(tx * config.TileSize)
	tileY := float64(ty * config.TileSize)

	overlapLeft := math.Max(x, tileX)
	overlapRight := math.Min(x+w, tileX+ts)
	overlapTop := math.Max(y, tileY)
	overlapBottom := math.Min(y+h, tileY+ts)

	if overlapRight <= overlapLeft || overlapBottom <= overlapTop {
		return false, false, 0, 0
	}

	// 1. If empty tile, check if it penetrates an inner 45° bracket
	if !solid {
		b := float64(WreckFilletSize)
		if ClassifyQuadrant(grid, tx, ty, QuadNW).Shape == ShapeInnerFillet {
			if x < tileX+b && y < tileY+b {
				if (x-tileX)+(y-tileY) < b {
					return true, true, 0.7071, 0.7071
				}
			}
		}
		if ClassifyQuadrant(grid, tx, ty, QuadNE).Shape == ShapeInnerFillet {
			if x+w > tileX+ts-b && y < tileY+b {
				if (tileX+ts-(x+w))+(y-tileY) < b {
					return true, true, -0.7071, 0.7071
				}
			}
		}
		if ClassifyQuadrant(grid, tx, ty, QuadSW).Shape == ShapeInnerFillet {
			if x < tileX+b && y+h > tileY+ts-b {
				if (x-tileX)+(tileY+ts-(y+h)) < b {
					return true, true, 0.7071, -0.7071
				}
			}
		}
		if ClassifyQuadrant(grid, tx, ty, QuadSE).Shape == ShapeInnerFillet {
			if x+w > tileX+ts-b && y+h > tileY+ts-b {
				if (tileX+ts-(x+w))+(tileY+ts-(y+h)) < b {
					return true, true, -0.7071, -0.7071
				}
			}
		}
		return false, false, 0, 0
	}

	// 2. Solid tile: Check each quadrant for 45° outer chamfers
	c := float64(WreckChamferSize)

	if ClassifyQuadrant(grid, tx, ty, QuadNW).Shape == ShapeOuterCorner {
		if overlapRight <= tileX+c && overlapBottom <= tileY+c {
			relX := overlapRight - tileX
			relY := overlapBottom - tileY
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, -0.7071, -0.7071
		}
	}

	if ClassifyQuadrant(grid, tx, ty, QuadNE).Shape == ShapeOuterCorner {
		if overlapLeft >= tileX+ts-c && overlapBottom <= tileY+c {
			relX := tileX + ts - overlapLeft
			relY := overlapBottom - tileY
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, 0.7071, -0.7071
		}
	}

	if ClassifyQuadrant(grid, tx, ty, QuadSW).Shape == ShapeOuterCorner {
		if overlapRight <= tileX+c && overlapTop >= tileY+ts-c {
			relX := overlapRight - tileX
			relY := tileY + ts - overlapTop
			if relX+relY <= c {
				return false, false, 0, 0
			}
			return true, true, -0.7071, 0.7071
		}
	}

	if ClassifyQuadrant(grid, tx, ty, QuadSE).Shape == ShapeOuterCorner {
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
