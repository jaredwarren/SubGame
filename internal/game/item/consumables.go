package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// RawFish represents the raw fish consumable item.
type RawFish struct{}

func (f *RawFish) GetName() string       { return "Raw Fish" }
func (f *RawFish) GetMaxStack() int      { return 5 }
func (f *RawFish) GetColor() color.Color { return color.RGBA{70, 140, 180, 255} }
func (f *RawFish) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, f.GetName(), cx, cy, size) {
		return
	}
	// Draw fish body (oval and tail)
	vector.FillCircle(screen, cx, cy, size/3.5, f.GetColor(), false)
	// Tail
	var path vector.Path
	path.MoveTo(cx-size/3.5, cy)
	path.LineTo(cx-size/1.8, cy-size/4.0)
	path.LineTo(cx-size/1.8, cy+size/4.0)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(f.GetColor())
	vector.FillPath(screen, &path, nil, &opts)
	// Eye
	vector.FillCircle(screen, cx+size/6.0, cy-size/10.0, 2.0, color.White, false)
}
func (f *RawFish) GetHealthRestore() float64  { return 0.0 }
func (f *RawFish) GetStaminaRestore() float64 { return 5.0 }

// CookedFish represents the cooked fish consumable item.
type CookedFish struct{}

func (f *CookedFish) GetName() string       { return "Cooked Fish" }
func (f *CookedFish) GetMaxStack() int      { return 5 }
func (f *CookedFish) GetColor() color.Color { return color.RGBA{170, 110, 60, 255} }
func (f *CookedFish) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, f.GetName(), cx, cy, size) {
		return
	}
	// Golden brown cooked fish
	vector.FillCircle(screen, cx, cy, size/3.5, f.GetColor(), false)
	var path vector.Path
	path.MoveTo(cx-size/3.5, cy)
	path.LineTo(cx-size/1.8, cy-size/4.0)
	path.LineTo(cx-size/1.8, cy+size/4.0)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(f.GetColor())
	vector.FillPath(screen, &path, nil, &opts)
	// Grill lines
	vector.StrokeLine(screen, cx, cy-size/6.0, cx-size/6.0, cy+size/6.0, 1.5, color.RGBA{100, 60, 30, 255}, false)
	vector.StrokeLine(screen, cx+size/8.0, cy-size/6.0, cx-size/12.0, cy+size/6.0, 1.5, color.RGBA{100, 60, 30, 255}, false)
}
func (f *CookedFish) GetHealthRestore() float64  { return 25.0 }
func (f *CookedFish) GetStaminaRestore() float64 { return 15.0 }

// RawCrab represents the raw crab consumable item.
type RawCrab struct{}

func (c *RawCrab) GetName() string       { return "Raw Crab" }
func (c *RawCrab) GetMaxStack() int      { return 5 }
func (c *RawCrab) GetColor() color.Color { return color.RGBA{180, 50, 50, 255} }
func (c *RawCrab) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, c.GetName(), cx, cy, size) {
		return
	}
	// Crab body circle
	vector.FillCircle(screen, cx, cy, size/4.0, c.GetColor(), false)
	// Claws
	vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	// Little eyes
	vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.White, false)
	vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.White, false)
}
func (c *RawCrab) GetHealthRestore() float64  { return 0.0 }
func (c *RawCrab) GetStaminaRestore() float64 { return 8.0 }

// CookedCrab represents the cooked crab consumable item.
type CookedCrab struct{}

func (c *CookedCrab) GetName() string       { return "Cooked Crab" }
func (c *CookedCrab) GetMaxStack() int      { return 5 }
func (c *CookedCrab) GetColor() color.Color { return color.RGBA{240, 90, 50, 255} }
func (c *CookedCrab) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, c.GetName(), cx, cy, size) {
		return
	}
	// Orange-red cooked crab
	vector.FillCircle(screen, cx, cy, size/4.0, c.GetColor(), false)
	vector.FillRect(screen, cx-size/2.5, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillRect(screen, cx+size/2.5-size/5.0, cy-size/4.0, size/5.0, size/5.0, c.GetColor(), false)
	vector.FillCircle(screen, cx-size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
	vector.FillCircle(screen, cx+size/10.0, cy-size/4.0, 1.5, color.RGBA{255, 230, 200, 255}, false)
}
func (c *CookedCrab) GetHealthRestore() float64  { return 20.0 }
func (c *CookedCrab) GetStaminaRestore() float64 { return 20.0 }
