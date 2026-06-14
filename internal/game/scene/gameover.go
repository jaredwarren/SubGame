package scene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// GameOverScene represents the screen shown when the player dies.
type GameOverScene struct{}

// NewGameOverScene creates a new GameOverScene.
func NewGameOverScene() *GameOverScene {
	return &GameOverScene{}
}

// GameOverContext defines the narrow context interface required by GameOverScene.
type GameOverContext interface {
	GetInput() InputSource
	GetDeathReason() string
	Respawn()
	SetCurrentState(s State)
}

func (s *GameOverScene) OnEnter(g GameContext) {
	s.onEnter(g)
}

func (s *GameOverScene) onEnter(g GameOverContext) {
	g.SetCurrentState(StateGameOver)
}

func (s *GameOverScene) OnExit(g GameContext) {}

func (s *GameOverScene) Update(g GameContext) error {
	return s.update(g)
}

func (s *GameOverScene) update(g GameOverContext) error {
	if g.GetInput().IsKeyJustPressed(ebiten.KeyEnter) {
		g.Respawn()
	}
	return nil
}

func (s *GameOverScene) Draw(g GameContext, screen *ebiten.Image) {
	s.draw(g, screen)
}

func (s *GameOverScene) draw(g GameOverContext, screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 50, G: 10, B: 10, A: 255})
	msg := "Your hull cracked or you ran out of oxygen."
	if reason := g.GetDeathReason(); reason != "" {
		msg = reason
	}
	ebitenutil.DebugPrint(screen, "GAME OVER\n\n"+msg+"\n\nPress ENTER to respawn.")
}
