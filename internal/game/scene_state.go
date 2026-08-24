package game

// stateForScene derives the authoritative State enum from the active scene pointer.
// TransitionTo sets currentState from this after OnEnter — scenes must not call SetCurrentState.
func (g *Game) stateForScene(s Scene) State {
	switch s {
	case nil:
		return g.currentState
	case g.titleState:
		return StateTitle
	case g.introState:
		return StateIntro
	case g.overworldState:
		return StateOverworld
	case g.caveState:
		return StateCave
	case g.baseMenu:
		return StateBaseMenu
	case g.gameOverState:
		return StateGameOver
	case g.gameWonState:
		return StateGameWon
	case g.pauseState:
		return StatePause
	default:
		return g.currentState
	}
}
