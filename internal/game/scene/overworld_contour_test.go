package scene

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/player"
	"github.com/jaredwarren/SubGame/internal/gvec"
	"github.com/jaredwarren/SubGame/internal/world"
)

func newTestWorld(w, h int) *world.World {
	m := make([][]world.TileType, w)
	for x := range m {
		m[x] = make([]world.TileType, h)
		for y := range m[x] {
			m[x][y] = world.TileWater
		}
	}
	return &world.World{
		Width:        w,
		Height:       h,
		OverworldMap: m,
	}
}

func TestOverworldQuadrantClassification(t *testing.T) {
	w := newTestWorld(10, 10)
	// Place an isolated land tile at (5, 5)
	w.OverworldMap[5][5] = world.TileLand

	// All 4 quadrants must be outer corners (headland points)
	nw := ClassifyOverworldQuadrant(w, 5, 5, cave.QuadNW)
	ne := ClassifyOverworldQuadrant(w, 5, 5, cave.QuadNE)
	sw := ClassifyOverworldQuadrant(w, 5, 5, cave.QuadSW)
	se := ClassifyOverworldQuadrant(w, 5, 5, cave.QuadSE)

	if nw.Shape != cave.ShapeOuterCorner || nw.Corner != cave.CornerNW {
		t.Fatalf("expected NW outer corner, got %+v", nw)
	}
	if ne.Shape != cave.ShapeOuterCorner || ne.Corner != cave.CornerNE {
		t.Fatalf("expected NE outer corner, got %+v", ne)
	}
	if sw.Shape != cave.ShapeOuterCorner || sw.Corner != cave.CornerSW {
		t.Fatalf("expected SW outer corner, got %+v", sw)
	}
	if se.Shape != cave.ShapeOuterCorner || se.Corner != cave.CornerSE {
		t.Fatalf("expected SE outer corner, got %+v", se)
	}

	// Now place land at (5, 6) and (6, 5) to test inner fillet in water at (6, 6)
	w.OverworldMap[5][6] = world.TileLand
	w.OverworldMap[6][5] = world.TileLand
	// (6, 6) is water, surrounded by land to North (6, 5) and West (5, 6)
	fillet := ClassifyOverworldQuadrant(w, 6, 6, cave.QuadNW)
	if fillet.Shape != cave.ShapeInnerFillet {
		t.Fatalf("expected ShapeInnerFillet in water tile at (6, 6), got %+v", fillet)
	}
}

func TestOverworldOuterCornerPassable(t *testing.T) {
	w := newTestWorld(10, 10)
	w.OverworldMap[5][5] = world.TileLand

	ts := float64(config.TileSize) // 64.0
	tileX := 5.0 * ts
	tileY := 5.0 * ts

	// The NW outer corner has an OverworldCornerRadius (20px) cut off.
	// A small 4x4 bounding box situated at (tileX + 2, tileY + 2) is inside the cut triangle:
	// relX + relY = 2 + 2 = 4 <= 20, so it must NOT collide.
	isSol, _, _, _ := CheckOverworldSolid(w, 5, 5, tileX+2, tileY+2, 4, 4)
	if isSol {
		t.Fatalf("expected cut-off headland point at (%.1f, %.1f) to be passable", tileX+2, tileY+2)
	}

	// Deep inside the land tile (tileX + 32, tileY + 32), it must be solid.
	isSolDeep, _, _, _ := CheckOverworldSolid(w, 5, 5, tileX+32, tileY+32, 4, 4)
	if !isSolDeep {
		t.Fatalf("expected center of land tile at (%.1f, %.1f) to be solid", tileX+32, tileY+32)
	}
}

func TestOverworldCoastlineSlidingDeflection(t *testing.T) {
	w := newTestWorld(10, 10)
	w.OverworldMap[5][5] = world.TileLand

	scene := &OverworldScene{
		World: w,
	}

	ts := float64(config.TileSize)
	tileX := 5.0 * ts
	tileY := 5.0 * ts

	// Position player just outside the NW headland corner of tile (5, 5).
	// Boat moves southeast into the headland at an angle (+10 X, +2 Y).
	p := &player.Player{
		Pos:    gvec.Vec2{X: tileX - 4, Y: tileY - 4},
		Vel:    gvec.Vec2{X: 10, Y: 2},
		Width:  12,
		Height: 12,
	}

	initialPos := p.Pos
	scene.CheckCollisions(p, nil)

	// In the contoured overworld, hitting the 45°/rounded headland corner deflects
	// player velocity tangentially along the coast, advancing along the shoreline.
	if p.Pos.X <= initialPos.X || p.Pos.Y >= initialPos.Y {
		t.Fatalf("expected player to slide up-right along headland contour, initial: %v, new: %v", initialPos, p.Pos)
	}
}

func TestOverworldInnerFilletCollision(t *testing.T) {
	w := newTestWorld(10, 10)
	w.OverworldMap[5][6] = world.TileLand
	w.OverworldMap[6][5] = world.TileLand
	// Tile (6, 6) is water, but its NW quadrant has an inner rock fillet bridging (5,6) and (6,5)
	ts := float64(config.TileSize)
	tileX := 6.0 * ts
	tileY := 6.0 * ts

	// Deep inside the inner fillet (tileX + 2, tileY + 2):
	// relX + relY = 2 + 2 = 4 < 16 (Fillet radius) => MUST be solid!
	isSol, isSlope, nx, ny := CheckOverworldSolid(w, 6, 6, tileX+2, tileY+2, 4, 4)
	if !isSol || !isSlope {
		t.Fatalf("expected inner fillet to be solid slope, got isSol=%v, isSlope=%v", isSol, isSlope)
	}
	if nx <= 0 || ny <= 0 {
		t.Fatalf("expected normal pointing southeast into open water, got (%f, %f)", nx, ny)
	}

	// In open water of tile (6, 6) outside the fillet (tileX + 24, tileY + 24): MUST be passable!
	isSolWater, _, _, _ := CheckOverworldSolid(w, 6, 6, tileX+24, tileY+24, 4, 4)
	if isSolWater {
		t.Fatalf("expected open water in tile (6, 6) to be passable")
	}
}
