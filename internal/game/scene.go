package game

import "github.com/hajimehoshi/ebiten/v2"

// Scene represents a distinct game state or view (e.g. Overworld, Cave, Menu, Game Over).
type Scene interface {
	// Update advances the scene logical state by one tick.
	Update(g *Game) error

	// Draw renders the scene graphics to the screen. It must not modify game state.
	Draw(g *Game, screen *ebiten.Image)

	// OnEnter is called once when the scene becomes active.
	OnEnter(g *Game)

	// OnExit is called once when the scene is replaced or deactivated.
	OnExit(g *Game)
}
