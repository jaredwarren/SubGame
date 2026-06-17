package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// O2TankHC represents the high capacity oxygen tank upgrade.
type O2TankHC struct{}

func (o *O2TankHC) GetName() string       { return "High Capacity O2 Tank" }
func (o *O2TankHC) GetMaxStack() int      { return 1 }
func (o *O2TankHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, o.GetName(), cx, cy, size) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}
func (o *O2TankHC) IsPlayerUpgrade() bool     { return true }
func (o *O2TankHC) GetMaxO2Capacity() float64 { return 60.0 }

// O2TankUHC represents the ultra high capacity oxygen tank upgrade.
type O2TankUHC struct{}

func (o *O2TankUHC) GetName() string       { return "Ultra High Capacity O2 Tank" }
func (o *O2TankUHC) GetMaxStack() int      { return 1 }
func (o *O2TankUHC) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (o *O2TankUHC) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, o.GetName(), cx, cy, size) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, o.GetColor(), false)
}
func (o *O2TankUHC) IsPlayerUpgrade() bool     { return true }
func (o *O2TankUHC) GetMaxO2Capacity() float64 { return 140.0 }

// Fins represents the propulsion fins player speed upgrade.
type Fins struct{}

func (f *Fins) GetName() string       { return "Propulsion Fins" }
func (f *Fins) GetMaxStack() int      { return 1 }
func (f *Fins) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (f *Fins) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, f.GetName(), cx, cy, size) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, f.GetColor(), false)
}
func (f *Fins) IsPlayerUpgrade() bool { return true }

func (s *Fins) GetSpeedUpgrade() map[string]Speed {
	return map[string]Speed{
		"overworld": {
			Drag:         0.92,
			Acceleration: 0.12,
			TopSpeed:     2.6,
		},
		"cave": {
			Drag:         0.96,
			Acceleration: 0.30,
			TopSpeed:     6.5,
		},
	}
}

// Scanner represents the scanner tool player upgrade.
type Scanner struct{}

func (s *Scanner) GetName() string       { return "Scanner Tool" }
func (s *Scanner) GetMaxStack() int      { return 1 }
func (s *Scanner) GetColor() color.Color { return color.RGBA{98, 198, 148, 255} }
func (s *Scanner) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, s.GetName(), cx, cy, size) {
		return
	}
	vector.FillCircle(screen, cx, cy, size/2.0, s.GetColor(), false)
}
func (s *Scanner) IsPlayerUpgrade() bool { return true }
