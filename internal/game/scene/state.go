package scene

// State represents the current game scene/view.
type State int

const (
	StateTitle State = iota
	StateOverworld
	StateCave
	StateBaseMenu
	StateGameOver
	StateGameWon
)

// String returns the string representation of the State.
func (s State) String() string {
	switch s {
	case StateTitle:
		return "Title"
	case StateOverworld:
		return "Overworld"
	case StateCave:
		return "Cave"
	case StateBaseMenu:
		return "Base Menu"
	case StateGameOver:
		return "Game Over"
	case StateGameWon:
		return "Game Won"
	default:
		return "Unknown"
	}
}
