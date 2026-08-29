package scene

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/config"
)

func TestMenuPanelLayoutForDetachedPDA(t *testing.T) {
	layout := MenuPanelLayoutFor(true)

	wantH := float64(config.ScreenHeight) - pdaPanelMargin*2
	wantW := wantH * (menuPanelBaseW / menuPanelBaseH)
	if layout.H != wantH {
		t.Fatalf("panel height = %.1f, want %.1f", layout.H, wantH)
	}
	if layout.W != wantW {
		t.Fatalf("panel width = %.1f, want %.1f", layout.W, wantW)
	}
	if layout.Scale != wantH/menuPanelBaseH {
		t.Fatalf("scale = %.4f, want %.4f", layout.Scale, wantH/menuPanelBaseH)
	}
	if layout.X < 0 || layout.Y < 0 {
		t.Fatalf("panel should be centered on screen, got (%.1f, %.1f)", layout.X, layout.Y)
	}
	if layout.X+layout.W > float64(config.ScreenWidth) || layout.Y+layout.H > float64(config.ScreenHeight) {
		t.Fatalf("panel exceeds screen bounds: %.0fx%.0f at (%.0f, %.0f)", layout.W, layout.H, layout.X, layout.Y)
	}
}

func TestMenuPanelLayoutForBaseTerminal(t *testing.T) {
	layout := MenuPanelLayoutFor(false)
	if layout.W != menuPanelBaseW || layout.H != menuPanelBaseH || layout.Scale != 1.0 {
		t.Fatalf("expected base terminal size 800x500 scale 1, got %.0fx%.0f scale %.2f", layout.W, layout.H, layout.Scale)
	}
}
