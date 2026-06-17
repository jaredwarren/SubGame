package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Titanium represents the titanium mineral item.
type Titanium struct{}

func (t *Titanium) GetName() string       { return "Titanium" }
func (t *Titanium) GetMaxStack() int      { return 10 }
func (t *Titanium) GetColor() color.Color { return color.RGBA{168, 178, 188, 255} }
func (t *Titanium) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{220, 230, 240, 255}
	drawMineralIcon(screen, cx, cy, size, t.GetColor(), coreColor, "Titanium")
}

// Copper represents the copper mineral item.
type Copper struct{}

func (c *Copper) GetName() string       { return "Copper" }
func (c *Copper) GetMaxStack() int      { return 10 }
func (c *Copper) GetColor() color.Color { return color.RGBA{218, 118, 48, 255} }
func (c *Copper) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{240, 160, 80, 255}
	drawMineralIcon(screen, cx, cy, size, c.GetColor(), coreColor, "Copper")
}

// Quartz represents the quartz mineral item.
type Quartz struct{}

func (q *Quartz) GetName() string       { return "Quartz" }
func (q *Quartz) GetMaxStack() int      { return 10 }
func (q *Quartz) GetColor() color.Color { return color.RGBA{48, 218, 245, 255} }
func (q *Quartz) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{220, 250, 255, 255}
	drawMineralIcon(screen, cx, cy, size, q.GetColor(), coreColor, "Quartz")
}

// AbyssalOre represents the abyssal ore mineral item.
type AbyssalOre struct{}

func (a *AbyssalOre) GetName() string       { return "Abyssal Ore" }
func (a *AbyssalOre) GetMaxStack() int      { return 10 }
func (a *AbyssalOre) GetColor() color.Color { return color.RGBA{148, 48, 218, 255} }
func (a *AbyssalOre) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	coreColor := color.RGBA{230, 180, 255, 255}
	drawMineralIcon(screen, cx, cy, size, a.GetColor(), coreColor, "Abyssal Ore")
}

// ScrapMetal represents the scrap metal mineral item.
type ScrapMetal struct{}

func (s *ScrapMetal) GetName() string       { return "Scrap Metal" }
func (s *ScrapMetal) GetMaxStack() int      { return 10 }
func (s *ScrapMetal) GetColor() color.Color { return color.RGBA{140, 110, 95, 255} }
func (s *ScrapMetal) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, s.GetName(), cx, cy, size) {
		return
	}
	// Draw an angled metallic sheet
	var path vector.Path
	path.MoveTo(cx-size/3.0, cy-size/3.0)
	path.LineTo(cx+size/3.0, cy-size/4.0)
	path.LineTo(cx+size/4.0, cy+size/3.0)
	path.LineTo(cx-size/3.0, cy+size/4.0)
	path.Close()
	var opts vector.DrawPathOptions
	opts.ColorScale.ScaleWithColor(s.GetColor())
	vector.FillPath(screen, &path, nil, &opts)
	vector.StrokeLine(screen, cx-size/3.0, cy, cx+size/3.0, cy-size/10.0, 1.5, color.RGBA{180, 150, 130, 255}, false)
}

// ElectronicWaste represents the electronic waste mineral item.
type ElectronicWaste struct{}

func (e *ElectronicWaste) GetName() string       { return "Electronic Waste" }
func (e *ElectronicWaste) GetMaxStack() int      { return 10 }
func (e *ElectronicWaste) GetColor() color.Color { return color.RGBA{70, 130, 90, 255} }
func (e *ElectronicWaste) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, e.GetName(), cx, cy, size) {
		return
	}
	// Draw a green circuit board chip
	vector.FillRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, e.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.2, cy-size/3.0, size/1.1, size/1.5, 1.0, color.RGBA{120, 200, 140, 255}, false)
	// Draw microchip core
	vector.FillRect(screen, cx-size/6.0, cy-size/6.0, size/3.0, size/3.0, color.RGBA{40, 40, 40, 255}, false)
	// Tiny copper pin details
	vector.FillRect(screen, cx-size/3.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
	vector.FillRect(screen, cx, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
	vector.FillRect(screen, cx+size/4.0, cy-size/2.5, size/15.0, size/10.0, color.RGBA{220, 150, 50, 255}, false)
}
