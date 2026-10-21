package cave

import (
	"testing"
)

func TestClassifyQuadrant(t *testing.T) {
	// 3x3 grid:
	// row 0: W S W (water, solid, water)
	// row 1: S S S (solid, solid, solid)
	// row 2: W S W (water, solid, water)
	grid := [][]bool{
		{false, true, false},
		{true, true, true},
		{false, true, false},
	}

	// Center tile (1, 1) is solid.
	// Its top (1, 0) is solid. Its bottom (1, 2) is solid.
	// Its left (0, 1) is solid. Its right (2, 1) is solid.
	// Its corners (0, 0), (2, 0), (0, 2), (2, 2) are false.
	// All 4 quadrants should be ShapeFullSolid because orthogonal neighbors are solid!
	for q := 0; q < 4; q++ {
		info := ClassifyQuadrant(grid, 1, 1, q)
		if info.Shape != ShapeFullSolid {
			t.Errorf("center tile quadrant %d expected ShapeFullSolid, got %v", q, info.Shape)
		}
	}

	// Isolated solid block at (1, 1) with all neighbors false
	isoGrid := [][]bool{
		{false, false, false},
		{false, true, false},
		{false, false, false},
	}
	// All 4 quadrants should be ShapeOuterCorner!
	for q := 0; q < 4; q++ {
		info := ClassifyQuadrant(isoGrid, 1, 1, q)
		if info.Shape != ShapeOuterCorner {
			t.Errorf("isolated tile quadrant %d expected ShapeOuterCorner, got %v", q, info.Shape)
		}
	}

	// L-shaped rock with corner at (1, 1) being water
	// row 0: S S
	// row 1: S W
	lGrid := [][]bool{
		{true, true},
		{true, false},
	}
	// Tile (1, 1) is water.
	// Its North (1, 0) is solid, West (0, 1) is solid.
	// NW quadrant of tile (1, 1) should be ShapeInnerFillet!
	infoNW := ClassifyQuadrant(lGrid, 1, 1, QuadNW)
	if infoNW.Shape != ShapeInnerFillet {
		t.Errorf("L-bracket empty corner expected ShapeInnerFillet, got %v", infoNW.Shape)
	}
}

func TestCheckWreckageSolidOuterChamfer(t *testing.T) {
	// Isolated solid tile at (1, 1)
	isoGrid := [][]bool{
		{false, false, false},
		{false, true, false},
		{false, false, false},
	}

	tileX := 64.0
	tileY := 64.0

	// Box deep in the cut-off NW corner: e.g. at (tileX + 2, tileY + 2, 8, 8)
	// (x - tileX) + (y - tileY) = 2 + 2 = 4 <= 22 (WreckChamferSize)
	solid, isSlope, _, _ := CheckWreckageSolid(isoGrid, 1, 1, tileX+2, tileY+2, 8, 8)
	if solid {
		t.Errorf("expected box in cut-off chamfer corner to NOT be solid, but got solid=true")
	}

	// Box deep inside center of the tile: at (tileX + 25, tileY + 25, 10, 10)
	solidCenter, _, _, _ := CheckWreckageSolid(isoGrid, 1, 1, tileX+25, tileY+25, 10, 10)
	if !solidCenter {
		t.Errorf("expected center of solid tile to be solid, but got solid=false")
	}

	_ = isSlope
}

func TestCheckWreckageInnerBracket(t *testing.T) {
	// L-shaped rock with corner at (1, 1) being water
	lGrid := [][]bool{
		{true, true},
		{true, false},
	}
	tileX := 64.0
	tileY := 64.0

	// Box deep in corner (tileX+1, tileY+1, 4, 4)
	// (1 + 1) = 2 < WreckFilletSize (14.0)
	solid, isSlope, nx, ny := CheckWreckageSolid(lGrid, 1, 1, tileX+1, tileY+1, 4, 4)
	if !solid || !isSlope {
		t.Errorf("expected inner bracket to be solid slope, got solid=%v isSlope=%v", solid, isSlope)
	}
	if nx <= 0 || ny <= 0 {
		t.Errorf("expected outward normal pointing down-right for NW bracket, got nx=%f ny=%f", nx, ny)
	}
}

