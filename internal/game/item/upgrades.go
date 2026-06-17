package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DecoyLauncher represents the vehicle decoy launcher upgrade module.
type DecoyLauncher struct{}

func (l *DecoyLauncher) GetName() string       { return "Decoy Launcher Module" }
func (l *DecoyLauncher) GetMaxStack() int      { return 1 }
func (l *DecoyLauncher) GetColor() color.Color { return color.RGBA{110, 120, 130, 255} }
func (l *DecoyLauncher) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, l.GetName(), cx, cy, size) {
		return
	}
	// Draw tube launcher fallback
	vector.FillRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, l.GetColor(), false)
	vector.StrokeRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, 1.5, color.RGBA{220, 220, 220, 255}, false)
	// Green activation diode
	vector.FillCircle(screen, cx, cy-size/4.0, 3, color.RGBA{50, 240, 100, 255}, false)
}
func (l *DecoyLauncher) IsVehicleUpgrade() bool { return true }

// ChemicalDischarger represents the vehicle chemical discharger upgrade module.
type ChemicalDischarger struct{}

func (d *ChemicalDischarger) GetName() string       { return "Chemical Discharger Module" }
func (d *ChemicalDischarger) GetMaxStack() int      { return 1 }
func (d *ChemicalDischarger) GetColor() color.Color { return color.RGBA{130, 80, 180, 255} }
func (d *ChemicalDischarger) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, d.GetName(), cx, cy, size) {
		return
	}
	// Draw double nozzle/violet canister fallback
	vector.FillRect(screen, cx-size/3.0, cy-size/2.0, size*0.6, size, d.GetColor(), false)
	// Nozzles
	vector.FillRect(screen, cx-size/4.0, cy-size/1.8, size/6.0, size/4.0, color.RGBA{80, 80, 90, 255}, false)
	vector.FillRect(screen, cx+size/12.0, cy-size/1.8, size/6.0, size/4.0, color.RGBA{80, 80, 90, 255}, false)
}
func (d *ChemicalDischarger) IsVehicleUpgrade() bool { return true }

// SonarAmplifier represents the vehicle sonar amplifier upgrade.
type SonarAmplifier struct{}

func (s *SonarAmplifier) GetName() string       { return "Sonar Amplifier" }
func (s *SonarAmplifier) GetMaxStack() int      { return 1 }
func (s *SonarAmplifier) GetColor() color.Color { return color.RGBA{0, 240, 255, 255} }
func (s *SonarAmplifier) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, s.GetName(), cx, cy, size) {
		return
	}
	// A nice cyan/white concentric rings icon representing sonar amplification
	vector.StrokeCircle(screen, cx, cy, size/2.0, 2.0, s.GetColor(), false)
	vector.StrokeCircle(screen, cx, cy, size/3.5, 1.5, color.RGBA{255, 255, 255, 200}, false)
	vector.FillCircle(screen, cx, cy, 3, s.GetColor(), false)
}
func (s *SonarAmplifier) IsVehicleUpgrade() bool { return true }

// PowerCell represents the vehicle battery power cell item.
type PowerCell struct{}

func (p *PowerCell) GetName() string       { return "Power Cell" }
func (p *PowerCell) GetMaxStack() int      { return 5 }
func (p *PowerCell) GetColor() color.Color { return color.RGBA{220, 180, 40, 255} }
func (p *PowerCell) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, p.GetName(), cx, cy, size) {
		return
	}
	// Draw a yellow cylinder battery cell with a light grey top tip
	vector.FillRect(screen, cx-size/4.0, cy-size/3.0, size/2.0, size*0.7, p.GetColor(), false)
	vector.FillRect(screen, cx-size/8.0, cy-size/2.0, size/4.0, size/6.0, color.RGBA{180, 190, 200, 255}, false)
}

// ThermalGenerator represents the vehicle thermal generator upgrade.
type ThermalGenerator struct{}

func (t *ThermalGenerator) GetName() string       { return "Thermal Generator" }
func (t *ThermalGenerator) GetMaxStack() int      { return 1 }
func (t *ThermalGenerator) GetColor() color.Color { return color.RGBA{235, 100, 50, 255} }
func (t *ThermalGenerator) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, t.GetName(), cx, cy, size) {
		return
	}
	// Draw a diamond container with an inner orange flame/core
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.5, t.GetColor(), false)
	vector.FillCircle(screen, cx, cy, size/4.0, color.RGBA{255, 120, 0, 255}, false)
}
func (t *ThermalGenerator) IsVehicleUpgrade() bool { return true }
