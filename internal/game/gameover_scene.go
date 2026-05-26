package game

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

func (s *GameOverScene) OnEnter(g *Game) {
	g.currentState = StateGameOver
}

func (s *GameOverScene) OnExit(g *Game) {}

func (s *GameOverScene) Update(g *Game) error {
	if g.Input.IsKeyJustPressed(ebiten.KeyEnter) {
		g.Respawn()
	}
	return nil
}

func (s *GameOverScene) Draw(g *Game, screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 50, G: 10, B: 10, A: 255})
	ebitenutil.DebugPrint(screen, "GAME OVER\n\nYour hull cracked or you ran out of oxygen.\n\nPress ENTER to respawn.")
}
