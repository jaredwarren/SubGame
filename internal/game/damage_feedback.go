package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
)

const damageFeedbackCooldownFrames = 15

func (g *Game) applyDamageFeedback(amount float64, kind entity.DamageKind, hitVehicle bool) {
	if g.GodMode || amount <= 0 {
		return
	}
	if g.damageFeedbackCooldown > 0 {
		return
	}
	g.damageFeedbackCooldown = damageFeedbackCooldownFrames

	intensity := math.Min(amount*0.22, 3.5)
	duration := int(math.Min(amount*1.0, 8))
	if duration < 3 {
		duration = 3
	}
	if kind == entity.DamageElectric {
		intensity = math.Max(intensity, 2.5)
		duration = max(duration, 6)
	}
	g.TriggerScreenShake(duration, intensity)

	flashDur := 10
	if kind == entity.DamageElectric {
		flashDur = 14
	}
	if flashDur > g.DamageFlash.Timer {
		g.DamageFlash.Timer = flashDur
		g.DamageFlash.Kind = kind
	}

	if hitVehicle {
		audio.Get().PlaySFXVaried("sfx/vehicle_alarm.wav", 0.45, 0.04)
	} else if kind == entity.DamageElectric {
		audio.Get().PlaySFXVaried("sfx/shock_kelp_zap.wav", 0.6, 0.05)
	} else {
		audio.Get().PlaySFXVaried("sfx/player_hurt.wav", 0.55, 0.04)
	}

	if kind != entity.DamageElectric {
		return
	}

	var px, py float64
	var hasPos bool
	if hitVehicle && g.ActiveVehicle != nil {
		pos := g.ActiveVehicle.GetPos()
		dims := g.ActiveVehicle.GetDimensions()
		px = pos.X + dims.X/2.0
		py = pos.Y + dims.Y/2.0
		hasPos = true
	} else if g.player != nil {
		px = g.player.Pos.X + g.player.Width/2.0
		py = g.player.Pos.Y + g.player.Height/2.0
		hasPos = true
	}
	if !hasPos {
		return
	}
	g.SpawnDebris(px, py, color.RGBA{120, 200, 230, 180})
}

func (g *Game) drawDamageFlash(screen *ebiten.Image) {
	if g.DamageFlash.Timer <= 0 {
		return
	}

	maxDur := 14.0
	if g.DamageFlash.Kind != entity.DamageElectric {
		maxDur = 10.0
	}
	t := float64(g.DamageFlash.Timer) / maxDur
	alpha := uint8(75 * t) // linear fade, lower peak

	w := float32(config.ScreenWidth)
	h := float32(config.ScreenHeight)
	border := float32(48)

	var edgeColor color.RGBA
	if g.DamageFlash.Kind == entity.DamageElectric {
		edgeColor = color.RGBA{100, 70, 140, alpha}
	} else {
		edgeColor = color.RGBA{160, 55, 60, alpha}
	}

	vector.FillRect(screen, 0, 0, w, border, edgeColor, false)
	vector.FillRect(screen, 0, h-border, w, border, edgeColor, false)
	vector.FillRect(screen, 0, 0, border, h, edgeColor, false)
	vector.FillRect(screen, w-border, 0, border, h, edgeColor, false)

	corner := edgeColor
	corner.A = uint8(float64(alpha) * 0.4)
	r := border * 1.6
	vector.FillCircle(screen, 0, 0, r, corner, false)
	vector.FillCircle(screen, w, 0, r, corner, false)
	vector.FillCircle(screen, 0, h, r, corner, false)
	vector.FillCircle(screen, w, h, r, corner, false)

	if g.DamageFlash.Kind == entity.DamageElectric && g.DamageFlash.Timer > 10 {
		tintAlpha := uint8(12 * float64(g.DamageFlash.Timer-10) / 4.0)
		vector.FillRect(screen, 0, 0, w, h, color.RGBA{80, 160, 190, tintAlpha}, false)
	}
}
