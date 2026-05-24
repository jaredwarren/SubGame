package game

// State represents the current game scene/view.
type State int

const (
	StateOverworld State = iota
	StateCave
	StateBaseMenu
	StateGameOver
)

// String returns the string representation of the State.
func (s State) String() string {
	switch s {
	case StateOverworld:
		return "Overworld"
	case StateCave:
		return "Cave"
	case StateBaseMenu:
		return "Base Menu"
	case StateGameOver:
		return "Game Over"
	default:
		return "Unknown"
	}
}
