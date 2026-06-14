package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
)

// GameWonScene represents the screen shown when the player successfully wins.
type GameWonScene struct{}

// NewGameWonScene creates a new GameWonScene.
func NewGameWonScene() *GameWonScene {
	return &GameWonScene{}
}

// GameWonContext defines the narrow context interface required by GameWonScene.
type GameWonContext interface {
	GetInput() InputSource
	Respawn()
	SetCurrentState(s State)
}

func (s *GameWonScene) OnEnter(g GameContext) {
	s.onEnter(g)
}

func (s *GameWonScene) onEnter(g GameWonContext) {
	g.SetCurrentState(StateGameWon)
}

func (s *GameWonScene) OnExit(g GameContext) {}

func (s *GameWonScene) Update(g GameContext) error {
	return s.update(g)
}

func (s *GameWonScene) update(g GameWonContext) error {
	if g.GetInput().IsKeyJustPressed(ebiten.KeyEnter) {
		g.Respawn()
	}
	return nil
}

func (s *GameWonScene) Draw(g GameContext, screen *ebiten.Image) {
	s.draw(g, screen)
}

func (s *GameWonScene) draw(g GameWonContext, screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 8, G: 24, B: 38, A: 255})

	panelW := float32(500)
	panelH := float32(300)
	panelX := float32(config.ScreenWidth-int(panelW)) / 2.0
	panelY := float32(config.ScreenHeight-int(panelH)) / 2.0

	vector.FillRect(screen, panelX, panelY, panelW, panelH, color.RGBA{14, 38, 28, 240}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2.0, color.RGBA{220, 180, 50, 255}, false)
	vector.StrokeLine(screen, panelX+20, panelY+60, panelX+panelW-20, panelY+60, 1.5, color.RGBA{220, 180, 50, 180}, false)

	ebitenutil.DebugPrintAt(screen, "★ SUCCESSFUL ESCAPE ★", int(panelX)+160, int(panelY)+25)

	lines := []string{
		"You have constructed and launched the Escape Rocket!",
		"Breaking through the ocean ceiling, you leave the deep",
		"abyssal trenches of SubGame behind.",
		"",
		"Rescue telemetry has established contact with the orbiting",
		"freighter. You are safe at last.",
		"",
		"Thanks for playing!",
		"",
		"Press [ENTER] to return to the surface and start anew.",
	}

	textStartY := int(panelY) + 80
	for i, line := range lines {
		offsetX := (int(panelW) - len(line)*6) / 2
		if offsetX < 20 {
			offsetX = 20
		}
		ebitenutil.DebugPrintAt(screen, line, int(panelX)+offsetX+1, textStartY+i*18+1)
		ebitenutil.DebugPrintAt(screen, line, int(panelX)+offsetX, textStartY+i*18)
	}
}
