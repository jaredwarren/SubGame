package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/cave"
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
	return a.g.SoundWave.Timer
}

func (a *sonarAdapter) SoundWaveX() float64 {
	return a.g.SoundWave.X
}

func (a *sonarAdapter) SoundWaveY() float64 {
	return a.g.SoundWave.Y
}

func (a *sonarAdapter) SoundWaveRadius() float64 {
	return a.g.SoundWave.Radius
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

func (a *worldAdapter) IsShockKelpCave() bool {
	if a.g.caveState == nil || a.g.caveState.ActiveCave == nil {
		return false
	}
	return a.g.caveState.ActiveCave.GetCaveType() == cave.CaveShockKelp
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

func (a *entityRuntimeAdapter) FindClosestDecoy(pos gvec.Vec2, maxDist float64) (gvec.Vec2, bool) {
	if a.playerAdapter.g.caveState == nil {
		return gvec.Vec2{}, false
	}
	var closestPos gvec.Vec2
	closestDist := maxDist
	found := false

	for _, ent := range a.playerAdapter.g.caveState.Entities {
		if !ent.IsActive() {
			continue
		}
		if decoy, ok := ent.(*entity.SonicDecoy); ok {
			decoyCenter := gvec.Vec2{
				X: decoy.Pos.X + decoy.Dimensions.X/2.0,
				Y: decoy.Pos.Y + decoy.Dimensions.Y/2.0,
			}
			dist := math.Hypot(pos.X-decoyCenter.X, pos.Y-decoyCenter.Y)
			if dist <= closestDist {
				closestDist = dist
				closestPos = decoyCenter
				found = true
			}
		}
	}
	return closestPos, found
}

func (a *entityRuntimeAdapter) CheckDeterrentOcclusion(pos1, pos2 gvec.Vec2) bool {
	if a.playerAdapter.g.caveState == nil {
		return false
	}
	for _, ent := range a.playerAdapter.g.caveState.Entities {
		if !ent.IsActive() {
			continue
		}
		if cloud, ok := ent.(*entity.DeterrentCloud); ok {
			cloudCenter := gvec.Vec2{
				X: cloud.Pos.X + cloud.Dimensions.X/2.0,
				Y: cloud.Pos.Y + cloud.Dimensions.Y/2.0,
			}
			elapsed := float64(360 - cloud.LifeTimer)
			var radius float64
			if elapsed < 60 {
				radius = (elapsed / 60.0) * 96.0
			} else {
				radius = 96.0
			}

			A := pos1
			B := pos2
			C := cloudCenter

			v := gvec.Vec2{X: B.X - A.X, Y: B.Y - A.Y}
			w := gvec.Vec2{X: C.X - A.X, Y: C.Y - A.Y}

			vLenSq := v.X*v.X + v.Y*v.Y
			if vLenSq == 0 {
				dist := math.Hypot(A.X-C.X, A.Y-C.Y)
				if dist <= radius {
					return true
				}
				continue
			}

			t := (w.X*v.X + w.Y*v.Y) / vLenSq
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}

			closestPoint := gvec.Vec2{X: A.X + v.X*t, Y: A.Y + v.Y*t}
			dist := math.Hypot(closestPoint.X-C.X, closestPoint.Y-C.Y)
			if dist <= radius {
				return true
			}
		}
	}
	return false
}

func (a *entityRuntimeAdapter) CheckDeterrentSlowing(x, y, w, h float64) bool {
	if a.playerAdapter.g.caveState == nil {
		return false
	}
	for _, ent := range a.playerAdapter.g.caveState.Entities {
		if !ent.IsActive() {
			continue
		}
		if cloud, ok := ent.(*entity.DeterrentCloud); ok {
			cloudCenter := gvec.Vec2{
				X: cloud.Pos.X + cloud.Dimensions.X/2.0,
				Y: cloud.Pos.Y + cloud.Dimensions.Y/2.0,
			}
			elapsed := float64(360 - cloud.LifeTimer)
			var radius float64
			if elapsed < 60 {
				radius = (elapsed / 60.0) * 96.0
			} else {
				radius = 96.0
			}

			px := max(x, min(cloudCenter.X, x+w))
			py := max(y, min(cloudCenter.Y, y+h))

			dist := math.Hypot(px-cloudCenter.X, py-cloudCenter.Y)
			if dist <= radius {
				return true
			}
		}
	}
	return false
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
			g.player.CurrentOxygen = min(g.player.MaxOxygen, g.player.CurrentOxygen+c.Amount)
		case entity.TriggerSoundWaveCmd:
			g.SoundWave.Timer = 60
			g.SoundWave.Radius = 0.0
			g.SoundWave.X = c.Pos.X
			g.SoundWave.Y = c.Pos.Y
		case entity.SetPlayerSlowedCmd:
			g.playerSlowed = c.Slowed
		case entity.SetMineWarningCmd:
			g.MineWarning.Message = c.Message
			g.MineWarning.Timer = c.Duration
			g.MineWarning.Level = c.Level
		case entity.UpdateWeaverTrackingTimerCmd:
			g.WeaverTrackingTimer = max(g.WeaverTrackingTimer, c.Value)
		case entity.StunPlayerCmd:
			g.player.StunTimer = c.Duration
		case entity.TriggerShakeCmd:
			g.TriggerScreenShake(c.Duration, c.Intensity)
		case entity.DestroyDecoyCmd:
			if g.caveState != nil {
				closestDist := 999999.0
				var closestDecoy *entity.SonicDecoy
				for _, ent := range g.caveState.Entities {
					if decoy, ok := ent.(*entity.SonicDecoy); ok && decoy.IsActive() {
						decoyCenter := gvec.Vec2{
							X: decoy.Pos.X + decoy.Dimensions.X/2.0,
							Y: decoy.Pos.Y + decoy.Dimensions.Y/2.0,
						}
						dist := math.Hypot(c.Pos.X-decoyCenter.X, c.Pos.Y-decoyCenter.Y)
						if dist < closestDist {
							closestDist = dist
							closestDecoy = decoy
						}
					}
				}
				if closestDecoy != nil && closestDist < 80.0 {
					closestDecoy.SetActive(false)
				}
			}
		}
	}
	rt.cmds = rt.cmds[:0]
}
