package vehicle

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// SonarPulse defines gameplay tuning values for an emitted sonar ping.
type SonarPulse struct {
	DurationTicks int
	RadiusStep    float64
}

// InputSource provides the subset of input APIs vehicle logic needs.
type InputSource interface {
	Cursor() gvec.Vec2
	IsKeyJustPressed(k ebiten.Key) bool
	IsKeyPressed(k ebiten.Key) bool
}

// DrillableResource represents a cave node that can be mined by the heavy mech.
type DrillableResource interface {
	item.Item
	GetTilePos() (int, int)
	GetHitsToMine() int
	SetHitsToMine(hits int)
	GetColor() color.Color
	GetRecipeResultName() string
}

// GameCommand is a sealed interface for fire-and-forget mutations vehicles
// request from the game. Commands are queued during vehicle Update calls and
// processed by game.Update() after all vehicles have ticked.
type GameCommand interface{ gameCommand() }

// ActivateSonarCmd fires a sonar ping originating from Source.
type ActivateSonarCmd struct {
	Source gvec.Vec2
	Pulse  SonarPulse
	Bright bool
}

// RemoveCaveNodeCmd removes the resource node at tile position (TX, TY).
type RemoveCaveNodeCmd struct {
	TX, TY int
}

// UnlockRecipeCmd requests the game to unlock a recipe by result item name.
type UnlockRecipeCmd struct {
	RecipeResultName string
}

// SpawnBubbleCmd spawns a bubble particle at Pos.
type SpawnBubbleCmd struct {
	Pos gvec.Vec2
}

// SpawnDebrisCmd spawns a debris particle at Pos with the given Color.
type SpawnDebrisCmd struct {
	Pos   gvec.Vec2
	Color color.RGBA
}

// TriggerShakeCmd requests a screen shake of the given Duration and Intensity.
type TriggerShakeCmd struct {
	Duration  int
	Intensity float64
}

// SetWarningCmd requests a user-facing warning popup with the given message, duration, and level.
type SetWarningCmd struct {
	Message  string
	Duration int
	Level    int
}

func (ActivateSonarCmd) gameCommand()  {}
func (RemoveCaveNodeCmd) gameCommand() {}
func (UnlockRecipeCmd) gameCommand()   {}
func (SpawnBubbleCmd) gameCommand()    {}
func (SpawnDebrisCmd) gameCommand()    {}
func (TriggerShakeCmd) gameCommand()   {}
func (SetWarningCmd) gameCommand()     {}

// Runtime exposes world/game state that vehicles need without importing package game.
// Synchronous queries return values immediately; mutations are submitted via Emit
// and applied by the game after all vehicle ticks complete.
type Runtime interface {
	// Queries — synchronous, must return a value this tick.
	TimeOfDay() float64
	IsActiveVehicle(v Vehicle) bool
	Input() InputSource
	PlayerScreenCenter() gvec.Vec2
	PlayerSlowed() bool
	PlayerStunned() bool
	IsOverworldSolidAt(tx, ty int) bool
	IsCaveSolidAt(tx, ty int) bool
	CanUseSonar() bool
	BaseStationPos() (pos gvec.Vec2, size gvec.Vec2)

	// Emit queues a fire-and-forget command to be processed by the game
	// after all vehicles have finished updating.
	Emit(cmd GameCommand)
}
