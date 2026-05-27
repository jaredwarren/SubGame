package vehicle

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// InputSource provides the subset of input APIs vehicle logic needs.
type InputSource interface {
	Cursor() Vec2
	IsKeyJustPressed(k ebiten.Key) bool
	IsKeyPressed(k ebiten.Key) bool
}

// DrillableResource represents a cave node that can be mined by the heavy mech.
type DrillableResource interface {
	GetName() string
	GetMaxStack() int
	GetTilePos() (int, int)
	GetHitsToMine() int
	SetHitsToMine(hits int)
}

// Runtime exposes world/game state that vehicles need without importing package game.
type Runtime interface {
	TimeOfDay() float64
	IsActiveVehicle(v Vehicle) bool
	Input() InputSource
	PlayerScreenCenter() Vec2
	PlayerSlowed() bool
	IsOverworldSolidAt(tx, ty int) bool
	IsCaveSolidAt(tx, ty int) bool
	CanUseSonar() bool
	ActivateSonar(source Vec2)
	RemoveCaveNodeAt(tx, ty int)
}
