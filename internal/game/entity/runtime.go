package entity

import (
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Runtime defines the read-only game/world state interface that entities need
// to update their logic.
type Runtime interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
	PlayerVel() gvec.Vec2
	PlayerFacing() float64
	IsPlayerSprinting() bool

	HasActiveVehicle() bool
	ActiveVehiclePos() gvec.Vec2
	ActiveVehicleDims() gvec.Vec2
	ActiveVehicleFacing() float64
	ActiveVehicleMoving() bool

	FlashlightOn() bool
	SoundWaveTimer() int
	SoundWaveX() float64
	SoundWaveY() float64
	SoundWaveRadius() float64

	SonarActive() bool
	TimeOfDay() float64

	Emit(cmd GameCommand)
}

// GameCommand is a sealed interface representing deferred mutations that entities
// request. They are applied at the end of the tick.
type GameCommand interface{ gameCommand() }

type DamagePlayerCmd struct {
	Amount float64
}

type DamageActiveVehicleCmd struct {
	Amount float64
}

type RestoreOxygenCmd struct {
	Amount float64
}

type TriggerSoundWaveCmd struct {
	Pos gvec.Vec2
}

type SetPlayerSlowedCmd struct {
	Slowed bool
}

type SetMineWarningCmd struct {
	Message  string
	Duration int
}

type UpdateWeaverTrackingTimerCmd struct {
	Value float64
}

type KnockbackPlayerCmd struct {
	Force gvec.Vec2
}

type KnockbackActiveVehicleCmd struct {
	Force gvec.Vec2
}

func (DamagePlayerCmd) gameCommand()             {}
func (DamageActiveVehicleCmd) gameCommand()      {}
func (RestoreOxygenCmd) gameCommand()            {}
func (TriggerSoundWaveCmd) gameCommand()         {}
func (SetPlayerSlowedCmd) gameCommand()          {}
func (SetMineWarningCmd) gameCommand()           {}
func (UpdateWeaverTrackingTimerCmd) gameCommand() {}
func (KnockbackPlayerCmd) gameCommand()          {}
func (KnockbackActiveVehicleCmd) gameCommand()   {}
