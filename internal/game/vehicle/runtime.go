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
}

// Runtime exposes world/game state that vehicles need without importing package game.
type Runtime interface {
	TimeOfDay() float64
	IsActiveVehicle(v Vehicle) bool
	Input() InputSource
	PlayerScreenCenter() gvec.Vec2
	PlayerSlowed() bool
	IsOverworldSolidAt(tx, ty int) bool
	IsCaveSolidAt(tx, ty int) bool
	CanUseSonar() bool
	ActivateSonar(source gvec.Vec2, pulse SonarPulse)
	RemoveCaveNodeAt(tx, ty int)
	SpawnBubble(pos gvec.Vec2)
	SpawnDebris(pos gvec.Vec2, clr color.RGBA)
	TriggerScreenShake(duration int, intensity float64)
}
