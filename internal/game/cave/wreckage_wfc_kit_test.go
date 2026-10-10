package cave

import (
	"math/rand"
	"testing"

	"github.com/jaredwarren/SubGame/internal/pcghash"
)

func TestWreckageRelaxWFCDeterministic(t *testing.T) {
	prev := pcghash.UseRelaxWFC
	pcghash.UseRelaxWFC = true
	defer func() { pcghash.UseRelaxWFC = prev }()

	gen := func(seed int64) [][]bool {
		return GenerateWreckageGridWithShip(rand.New(rand.NewSource(seed)), 0)
	}

	a := gen(42)
	b := gen(42)
	if len(a) != CaveWidth || len(a[0]) != CaveHeight {
		t.Fatalf("bad size %dx%d", len(a), len(a[0]))
	}
	for x := 0; x < CaveWidth; x++ {
		for y := 0; y < CaveHeight; y++ {
			if a[x][y] != b[x][y] {
				t.Fatalf("nondeterministic at %d,%d", x, y)
			}
		}
	}

	// Entrance / shaft must stay open for diving.
	if a[CaveWidth/2][0] {
		t.Fatal("expected open entrance at top center")
	}
	shaftOpen := false
	for y := 10; y < 40; y++ {
		if !a[CaveWidth/2][y] {
			shaftOpen = true
			break
		}
	}
	if !shaftOpen {
		t.Fatal("expected open central shaft")
	}

	// At least one open room cell outside shaft/corridor masks.
	roomOpen := 0
	for x := 4; x < 24; x++ {
		for y := 5; y < 20; y++ {
			if !a[x][y] && !isWreckageCorridorOrShaftTile(x, y) {
				roomOpen++
			}
		}
	}
	if roomOpen < 8 {
		t.Fatalf("expected carved room cells in upper-left bay, got %d open", roomOpen)
	}
}

func TestWreckageLegacyUnchangedWhenFlagOff(t *testing.T) {
	prev := pcghash.UseRelaxWFC
	pcghash.UseRelaxWFC = false
	defer func() { pcghash.UseRelaxWFC = prev }()

	a := GenerateWreckageGridWithShip(rand.New(rand.NewSource(7)), 0)
	b := GenerateWreckageGridWithShip(rand.New(rand.NewSource(7)), 0)
	for x := 0; x < CaveWidth; x++ {
		for y := 0; y < CaveHeight; y++ {
			if a[x][y] != b[x][y] {
				t.Fatalf("legacy nondeterministic at %d,%d", x, y)
			}
		}
	}
}

func TestCarveBayWithRelaxWFCConnectsDeck(t *testing.T) {
	grid := make([][]bool, CaveWidth)
	for x := range grid {
		grid[x] = make([]bool, CaveHeight)
		for y := range grid[x] {
			grid[x][y] = true
		}
	}
	// Pre-carve deck corridor the bay should connect to.
	for x := 4; x < 26; x++ {
		grid[x][24] = false
	}

	ok := carveBayWithRelaxWFC(grid, 99, 4, 24, 5, 22, 24)
	if !ok {
		t.Fatal("carveBayWithRelaxWFC failed")
	}
	linked := false
	for x := 4; x <= 24; x++ {
		if !grid[x][23] && !grid[x][24] {
			linked = true
			break
		}
	}
	if !linked {
		t.Fatal("expected bay linked to deck at y=24")
	}
}
