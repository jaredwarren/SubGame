package game

import "github.com/jaredwarren/SubGame/internal/game/scene"

// InputSource is an alias for scene.InputSource.
type InputSource = scene.InputSource

// EbitenInput is an alias for scene.EbitenInput.
type EbitenInput = scene.EbitenInput

// MockInput is an alias for scene.MockInput.
type MockInput = scene.MockInput

// TouchControls is an alias for scene.TouchControls.
type TouchControls = scene.TouchControls

// CombinedInput is an alias for scene.CombinedInput.
type CombinedInput = scene.CombinedInput

// TouchContext is an alias for scene.TouchContext.
type TouchContext = scene.TouchContext

// Touch context values selecting which virtual controls are shown.
const (
	TouchContextHidden      = scene.TouchContextHidden
	TouchContextOnFoot      = scene.TouchContextOnFoot
	TouchContextCave        = scene.TouchContextCave
	TouchContextDriving     = scene.TouchContextDriving
	TouchContextCaveDriving = scene.TouchContextCaveDriving
	TouchContextInventory   = scene.TouchContextInventory
	TouchContextMenu        = scene.TouchContextMenu
)

// NewEbitenInput creates a new EbitenInput.
func NewEbitenInput() *EbitenInput { return scene.NewEbitenInput() }

// NewTouchControls creates the virtual touch control overlay.
func NewTouchControls() *TouchControls { return scene.NewTouchControls() }

// NewCombinedInput merges physical input with virtual touch controls.
func NewCombinedInput(base *EbitenInput, touch *TouchControls) *CombinedInput {
	return scene.NewCombinedInput(base, touch)
}

// NewMockInput creates a new MockInput for testing.
func NewMockInput() *MockInput { return scene.NewMockInput() }

// HUDHotbarSlotAt returns the hotbar slot index (0..4) at screen coordinates (x, y), or -1 if none.
var HUDHotbarSlotAt = scene.HUDHotbarSlotAt

// HUDHotbarSlotRect returns the screen-space bounds for a given hotbar slot.
var HUDHotbarSlotRect = scene.HUDHotbarSlotRect

// HUDVehicleLightButtonHit returns true if (x, y) coordinates fall within the driving light button.
var HUDVehicleLightButtonHit = scene.HUDVehicleLightButtonHit

// HUDVehicleLightButtonRect returns screen-space bounding box for the driving headlight toggle button.
var HUDVehicleLightButtonRect = scene.HUDVehicleLightButtonRect

