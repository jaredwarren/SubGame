package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jaredwarren/SubGame/internal/game/gvec"
)

// InputSource abstracts user input polling for decoupled game logic and unit testing.
type InputSource interface {
	Update()
	Cursor() gvec.Vec2
	IsKeyJustPressed(k ebiten.Key) bool
	IsKeyPressed(k ebiten.Key) bool
	IsMouseButtonJustPressed(b ebiten.MouseButton) bool
}

// EbitenInput implements InputSource by polling the live Ebitengine input APIs.
// It caches values during Update() to allow safe access from both Update and Draw threads.
type EbitenInput struct {
	cursor           gvec.Vec2
	justPressedKeys  map[ebiten.Key]bool
	pressedKeys      map[ebiten.Key]bool
	justPressedMouse map[ebiten.MouseButton]bool
}

// NewEbitenInput creates a new EbitenInput manager.
func NewEbitenInput() *EbitenInput {
	return &EbitenInput{
		justPressedKeys:  make(map[ebiten.Key]bool),
		pressedKeys:      make(map[ebiten.Key]bool),
		justPressedMouse: make(map[ebiten.MouseButton]bool),
	}
}

// Update polls all relevant input states. Call this once at the start of Game.Update().
func (e *EbitenInput) Update() {
	mx, my := ebiten.CursorPosition()
	e.cursor = gvec.Vec2{X: float64(mx), Y: float64(my)}

	// List of all keys used in the game
	keys := []ebiten.Key{
		ebiten.KeyW, ebiten.KeyS, ebiten.KeyA, ebiten.KeyD,
		ebiten.KeyArrowUp, ebiten.KeyArrowDown, ebiten.KeyArrowLeft, ebiten.KeyArrowRight,
		ebiten.KeyShift, ebiten.KeySpace, ebiten.KeyT, ebiten.KeyTab,
		ebiten.KeyO, ebiten.KeyC, ebiten.KeyM, ebiten.KeyG, ebiten.KeyF, ebiten.KeyE,
		ebiten.KeyQ, ebiten.KeyEnter, ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
	}

	for _, k := range keys {
		e.justPressedKeys[k] = inpututil.IsKeyJustPressed(k)
		e.pressedKeys[k] = ebiten.IsKeyPressed(k)
	}

	buttons := []ebiten.MouseButton{
		ebiten.MouseButtonLeft,
		ebiten.MouseButtonRight,
	}
	for _, b := range buttons {
		e.justPressedMouse[b] = inpututil.IsMouseButtonJustPressed(b)
	}
}

// Cursor returns the cached screen space mouse cursor position.
func (e *EbitenInput) Cursor() gvec.Vec2 {
	return e.cursor
}

// IsKeyJustPressed returns true only on the first frame a key is pressed down.
func (e *EbitenInput) IsKeyJustPressed(k ebiten.Key) bool {
	return e.justPressedKeys[k]
}

// IsKeyPressed returns true every frame a key is held down.
func (e *EbitenInput) IsKeyPressed(k ebiten.Key) bool {
	return e.pressedKeys[k]
}

// IsMouseButtonJustPressed returns true only on the first frame a mouse button is pressed.
func (e *EbitenInput) IsMouseButtonJustPressed(b ebiten.MouseButton) bool {
	return e.justPressedMouse[b]
}

// MockInput provides a mock implementation of InputSource for testing.
type MockInput struct {
	CursorPos        gvec.Vec2
	JustPressedKeys  map[ebiten.Key]bool
	PressedKeys      map[ebiten.Key]bool
	JustPressedMouse map[ebiten.MouseButton]bool
}

// NewMockInput creates a new MockInput instance.
func NewMockInput() *MockInput {
	return &MockInput{
		JustPressedKeys:  make(map[ebiten.Key]bool),
		PressedKeys:      make(map[ebiten.Key]bool),
		JustPressedMouse: make(map[ebiten.MouseButton]bool),
	}
}

func (m *MockInput) Update()                                            {}
func (m *MockInput) Cursor() gvec.Vec2                                  { return m.CursorPos }
func (m *MockInput) IsKeyJustPressed(k ebiten.Key) bool                 { return m.JustPressedKeys[k] }
func (m *MockInput) IsKeyPressed(k ebiten.Key) bool                     { return m.PressedKeys[k] }
func (m *MockInput) IsMouseButtonJustPressed(b ebiten.MouseButton) bool { return m.JustPressedMouse[b] }
