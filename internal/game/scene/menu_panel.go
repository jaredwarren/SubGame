package scene

import "github.com/jaredwarren/SubGame/internal/game/config"

const (
	menuPanelBaseW = 800.0
	menuPanelBaseH = 500.0
	pdaPanelMargin = 16.0
)

// MenuPanelLayout describes the base menu / PDA popup frame and a uniform scale
// factor relative to the original 800×500 design.
type MenuPanelLayout struct {
	X, Y, W, H float64
	Scale      float64
}

// MenuPanelLayoutFor returns panel dimensions. Detached PDA popups use nearly the
// full screen height with width scaled to preserve the original aspect ratio.
func MenuPanelLayoutFor(detached bool) MenuPanelLayout {
	w, h := menuPanelBaseW, menuPanelBaseH
	if detached {
		h = float64(config.ScreenHeight) - pdaPanelMargin*2
		w = h * (menuPanelBaseW / menuPanelBaseH)
		maxW := float64(config.ScreenWidth) - pdaPanelMargin*2
		if w > maxW {
			w = maxW
			h = w * (menuPanelBaseH / menuPanelBaseW)
		}
	}
	x := (float64(config.ScreenWidth) - w) / 2.0
	y := (float64(config.ScreenHeight) - h) / 2.0
	return MenuPanelLayout{X: x, Y: y, W: w, H: h, Scale: h / menuPanelBaseH}
}

func (l MenuPanelLayout) S(v float64) float64 { return v * l.Scale }

func (l MenuPanelLayout) SF(v float32) float32 { return float32(l.S(float64(v))) }

func (l MenuPanelLayout) contentHeight() float64 { return l.H - l.S(115) }
