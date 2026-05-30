package game

import "github.com/jaredwarren/SubGame/internal/game/scene"

// InputSource is an alias for scene.InputSource.
type InputSource = scene.InputSource

// EbitenInput is an alias for scene.EbitenInput.
type EbitenInput = scene.EbitenInput

// MockInput is an alias for scene.MockInput.
type MockInput = scene.MockInput

// NewEbitenInput creates a new EbitenInput.
func NewEbitenInput() *EbitenInput { return scene.NewEbitenInput() }

// NewMockInput creates a new MockInput for testing.
func NewMockInput() *MockInput { return scene.NewMockInput() }
