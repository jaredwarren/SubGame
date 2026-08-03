package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// PauseContext defines context required for PauseScene.
type PauseContext interface {
	GetInput() InputSource
	SetCurrentState(s State)
	GetCurrentState() State
	SaveGame() error
	TransitionToOverworld()
	TransitionToCave()
	TransitionToTitle()
	SetMineWarning(msg string, duration, level int)
}

// PauseScene renders the interactive pause & settings overlay.
type PauseScene struct {
	PriorState State
}

// NewPauseScene instantiates a PauseScene.
func NewPauseScene() *PauseScene {
	return &PauseScene{}
}

func (p *PauseScene) OnEnter(g GameContext) {
	g.SetCurrentState(StatePause)
}

func (p *PauseScene) OnExit(g GameContext) {}

func (p *PauseScene) Update(g GameContext) error {
	return p.update(g)
}

func (p *PauseScene) update(g GameContext) error {
	inp := g.GetInput()

	if inp.IsKeyJustPressed(ebiten.KeyEscape) {
		p.resume(g)
		return nil
	}

	if inp.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cursor := inp.Cursor()
		mx, my := float64(cursor.X), float64(cursor.Y)

		const (
			btnW = 220.0
			btnH = 46.0
		)
		centerX := float64(config.ScreenWidth-btnW) / 2.0
		startY := 240.0

		// Resume button
		if mx >= centerX && mx < centerX+btnW && my >= startY && my < startY+btnH {
			p.resume(g)
			return nil
		}

		// Save Game button
		if mx >= centerX && mx < centerX+btnW && my >= startY+60 && my < startY+60+btnH {
			if err := g.SaveGame(); err == nil {
				g.SetMineWarning("GAME SAVED", 120, 1)
			} else {
				g.SetMineWarning("SAVE FAILED", 120, 3)
			}
			return nil
		}

		// Return to Title button
		if mx >= centerX && mx < centerX+btnW && my >= startY+120 && my < startY+120+btnH {
			g.TransitionToTitle()
			return nil
		}
	}

	return nil
}

func (p *PauseScene) resume(g GameContext) {
	if p.PriorState == StateCave {
		g.TransitionToCave()
	} else {
		g.TransitionToOverworld()
	}
}

func (p *PauseScene) Draw(g GameContext, screen *ebiten.Image) {
	// Dark semi-transparent background overlay
	vector.FillRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, color.RGBA{R: 5, G: 15, B: 25, A: 210}, false)

	// Pause title text
	title := "P A U S E D"
	titleX := (config.ScreenWidth - len(title)*12) / 2
	ebitenutil.DebugPrintAt(screen, title, titleX, 160)

	inp := g.GetInput()
	cursor := inp.Cursor()
	mx, my := float64(cursor.X), float64(cursor.Y)

	const (
		btnW = 220.0
		btnH = 46.0
	)
	centerX := float64(config.ScreenWidth-btnW) / 2.0
	startY := 240.0

	buttons := []struct {
		label string
		y     float64
	}{
		{"RESUME", startY},
		{"SAVE GAME", startY + 60},
		{"TITLE MENU", startY + 120},
	}

	for _, b := range buttons {
		isHovered := mx >= centerX && mx < centerX+btnW && my >= b.y && my < b.y+btnH

		bgColor := color.RGBA{R: 12, G: 32, B: 54, A: 220}
		borderColor := color.RGBA{R: 45, G: 140, B: 210, A: 255}

		if isHovered {
			bgColor = color.RGBA{R: 24, G: 60, B: 100, A: 255}
			borderColor = color.RGBA{R: 80, G: 220, B: 255, A: 255}
			vector.StrokeRect(screen, float32(centerX-2), float32(b.y-2), float32(btnW+4), float32(btnH+4), 1.0, color.RGBA{R: 80, G: 220, B: 255, A: 100}, false)
		}

		vector.FillRect(screen, float32(centerX), float32(b.y), float32(btnW), float32(btnH), bgColor, false)
		vector.StrokeRect(screen, float32(centerX), float32(b.y), float32(btnW), float32(btnH), 2.0, borderColor, false)

		lblX := int(centerX + (btnW-float64(len(b.label)*6))/2.0)
		lblY := int(b.y + (btnH-12)/2.0)
		ebitenutil.DebugPrintAt(screen, b.label, lblX, lblY)
	}
}
