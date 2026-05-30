package game

import "github.com/jaredwarren/SubGame/internal/game/scene"

// State is an alias for scene.State.
type State = scene.State

// Re-export state constants so existing game-package code and tests continue to compile.
const (
	StateTitle     State = scene.StateTitle
	StateOverworld State = scene.StateOverworld
	StateCave      State = scene.StateCave
	StateBaseMenu  State = scene.StateBaseMenu
	StateGameOver  State = scene.StateGameOver
	StateGameWon   State = scene.StateGameWon
)
