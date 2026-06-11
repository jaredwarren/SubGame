package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type playerAdapter struct {
	g *Game
}

func (a *playerAdapter) PlayerPos() gvec.Vec2 {
	return a.g.player.Pos
}

func (a *playerAdapter) PlayerDims() gvec.Vec2 {
	return gvec.Vec2{X: a.g.player.Width, Y: a.g.player.Height}
}

func (a *playerAdapter) PlayerVel() gvec.Vec2 {
	return a.g.player.Vel
}

func (a *playerAdapter) PlayerFacing() float64 {
	return a.g.player.Facing
}

func (a *playerAdapter) IsPlayerSprinting() bool {
	return a.g.Input.IsKeyPressed(ebiten.KeyShift)
}

type vehicleAdapter struct {
	g *Game
}

func (a *vehicleAdapter) HasActiveVehicle() bool {
	return a.g.ActiveVehicle != nil
}

func (a *vehicleAdapter) ActiveVehiclePos() gvec.Vec2 {
	if a.g.ActiveVehicle != nil {
		return a.g.ActiveVehicle.GetPos()
	}
	return gvec.Vec2{}
}

func (a *vehicleAdapter) ActiveVehicleDims() gvec.Vec2 {
	if a.g.ActiveVehicle != nil {
		return a.g.ActiveVehicle.GetDimensions()
	}
	return gvec.Vec2{}
}

func (a *vehicleAdapter) ActiveVehicleFacing() float64 {
	if a.g.ActiveVehicle != nil {
		return a.g.ActiveVehicle.GetFacing()
	}
	return 0.0
}

func (a *vehicleAdapter) ActiveVehicleMoving() bool {
	if a.g.ActiveVehicle != nil {
		return a.g.Input.IsKeyPressed(ebiten.KeyW) ||
			a.g.Input.IsKeyPressed(ebiten.KeyA) ||
			a.g.Input.IsKeyPressed(ebiten.KeyS) ||
			a.g.Input.IsKeyPressed(ebiten.KeyD) ||
			a.g.Input.IsKeyPressed(ebiten.KeySpace)
	}
	return false
}

type sonarAdapter struct {
	g *Game
}

func (a *sonarAdapter) FlashlightOn() bool {
	return a.g.FlashlightOn
}

func (a *sonarAdapter) SoundWaveTimer() int {
	return a.g.SoundWaveTimer
}

func (a *sonarAdapter) SoundWaveX() float64 {
	return a.g.SoundWaveX
}

func (a *sonarAdapter) SoundWaveY() float64 {
	return a.g.SoundWaveY
}

func (a *sonarAdapter) SoundWaveRadius() float64 {
	return a.g.SoundWaveRadius
}

func (a *sonarAdapter) SonarActive() bool {
	return a.g.Sonar.Timer > 0
}

func (a *sonarAdapter) TimeOfDay() float64 {
	return a.g.TimeOfDay
}

type worldAdapter struct {
	g *Game
}

func (a *worldAdapter) IsSolid(x, y, w, h float64) bool {
	if a.g.caveState == nil {
		return false
	}
	return a.g.caveState.IsSolid(a.g, x, y, w, h)
}

// entityRuntimeAdapter satisfies entity.Runtime, reading from *Game.
// Mutations are requested via Emit and collected in a slice, then processed
// safely by drainEntityCommands.
type entityRuntimeAdapter struct {
	playerAdapter
	vehicleAdapter
	sonarAdapter
	worldAdapter
	cmds []entity.GameCommand
}

func (a *entityRuntimeAdapter) Emit(cmd entity.GameCommand) {
	a.cmds = append(a.cmds, cmd)
}

// drainEntityCommands applies all entity mutation commands collected during the tick.
func (g *Game) drainEntityCommands(rt *entityRuntimeAdapter) {
	for _, cmd := range rt.cmds {
		switch c := cmd.(type) {
		case entity.DamagePlayerCmd:
			g.player.CurrentHealth -= c.Amount
		case entity.KnockbackPlayerCmd:
			g.player.Vel = g.player.Vel.Add(c.Force)
		case entity.DamageActiveVehicleCmd:
			if g.ActiveVehicle != nil {
				g.ActiveVehicle.TakeDamage(c.Amount)
			}
		case entity.KnockbackActiveVehicleCmd:
			if g.ActiveVehicle != nil {
				g.ActiveVehicle.ApplyForce(c.Force)
			}
		case entity.RestoreOxygenCmd:
			g.player.CurrentOxygen = math.Min(g.player.MaxOxygen, g.player.CurrentOxygen+c.Amount)
		case entity.TriggerSoundWaveCmd:
			g.SoundWaveTimer = 60
			g.SoundWaveRadius = 0.0
			g.SoundWaveX = c.Pos.X
			g.SoundWaveY = c.Pos.Y
		case entity.SetPlayerSlowedCmd:
			g.playerSlowed = c.Slowed
		case entity.SetMineWarningCmd:
			g.MineWarning = c.Message
			g.MineWarningTimer = c.Duration
			g.MineWarningLevel = c.Level
		case entity.UpdateWeaverTrackingTimerCmd:
			g.WeaverTrackingTimer = math.Max(g.WeaverTrackingTimer, c.Value)
		}
	}
	rt.cmds = rt.cmds[:0]
}
