package scene

import (
	"image/color"
	"testing"

	"github.com/jaredwarren/SubGame/internal/world"
)

func TestComputeTileColors_Sand(t *testing.T) {
	// For a land tile adjacent to water (waterDist == 1), it must return beach sand colors.
	// Since we are within bounds (0, 0) of a 100x100 world, the distance from border is negative,
	// so the void fade factor t is 0. With light multiplier = 1.0, we expect exact sand colors.
	fill, stroke := ComputeTileColors(10, 10, world.TileLand, 0, 1, 100, 100, 1.0)

	expectedFill := color.RGBA{232, 212, 165, 255}
	expectedStroke := color.RGBA{215, 195, 145, 255}

	if fill != expectedFill {
		t.Errorf("expected sand fill color %+v, got %+v", expectedFill, fill)
	}
	if stroke != expectedStroke {
		t.Errorf("expected sand stroke color %+v, got %+v", expectedStroke, stroke)
	}
}

func TestComputeTileColors_GrassGradient(t *testing.T) {
	// For a land tile inland (waterDist > 1), it must use the grass gradient.
	// We verify that different water distances result in different grass colors.
	fill2, _ := ComputeTileColors(10, 10, world.TileLand, 0, 2, 100, 100, 1.0)
	fill4, _ := ComputeTileColors(10, 10, world.TileLand, 0, 4, 100, 100, 1.0)

	if fill2 == fill4 {
		t.Errorf("expected different grass colors for waterDist 2 and 4, but got same color %+v", fill2)
	}
}

func TestComputeTileColors_WaterGradient(t *testing.T) {
	// For a water tile, color changes based on distance to nearest land.
	fillClose, _ := ComputeTileColors(10, 10, world.TileWater, 1, 0, 100, 100, 1.0)
	fillFar, _ := ComputeTileColors(10, 10, world.TileWater, 15, 0, 100, 100, 1.0)

	if fillClose == fillFar {
		t.Errorf("expected different water colors for landDist 1 and 15, but got same color %+v", fillClose)
	}
}

func TestComputeTileColors_VoidBorderFade(t *testing.T) {
	// Test a tile far outside the world boundaries (e.g. tx = -10).
	// Distance from border is 10.
	// The void factor t = (10 + 1) / 4 = 2.75, which clamps to 1.0.
	// When t is 1.0, the color must be voidClr = color.RGBA{4, 6, 12, 255}
	// with light multiplier applied.
	fill, stroke := ComputeTileColors(-10, 10, world.TileWater, 0, 0, 100, 100, 1.0)

	expectedFill := color.RGBA{4, 6, 12, 255}
	expectedStroke := color.RGBA{8, 12, 20, 255}

	if fill != expectedFill {
		t.Errorf("expected void fill color %+v, got %+v", expectedFill, fill)
	}
	if stroke != expectedStroke {
		t.Errorf("expected void stroke color %+v, got %+v", expectedStroke, stroke)
	}
}

func TestComputeTileColors_LightMultiplier(t *testing.T) {
	// Test the effect of the light multiplier on sand tile colors.
	fillFull, _ := ComputeTileColors(10, 10, world.TileLand, 0, 1, 100, 100, 1.0)
	fillHalf, _ := ComputeTileColors(10, 10, world.TileLand, 0, 1, 100, 100, 0.5)

	expectedHalfR := uint8(float64(fillFull.R) * 0.5)
	expectedHalfG := uint8(float64(fillFull.G) * 0.5)
	expectedHalfB := uint8(float64(fillFull.B) * 0.5)

	if fillHalf.R != expectedHalfR || fillHalf.G != expectedHalfG || fillHalf.B != expectedHalfB {
		t.Errorf("expected half light color to be scaled by 0.5: expected R=%d, G=%d, B=%d, got %+v",
			expectedHalfR, expectedHalfG, expectedHalfB, fillHalf)
	}
}
