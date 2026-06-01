package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// Draw renders one frame: scene → particles → overlays → HUD → warnings.
func (g *Game) Draw(screen *ebiten.Image) {
	g.currentScene.Draw(g, screen)

	if g.currentState == StateOverworld || g.currentState == StateCave {
		particle.DrawParticles(screen, g.Particles, g.camera.Pos.X, g.camera.Pos.Y)
	}

	switch g.currentState {
	case StateOverworld:
		g.drawOverworldLayer(screen)
	case StateCave:
		g.drawCaveLayer(screen)
	}

	if g.currentState == StateOverworld || g.currentState == StateCave {
		g.hud.Draw(screen, g)
	}

	g.drawWarningBanner(screen)
}

// drawOverworldLayer renders vehicles, the base station, and interaction prompts.
func (g *Game) drawOverworldLayer(screen *ebiten.Image) {
	camX, camY := g.camera.Pos.X, g.camera.Pos.Y

	for _, v := range g.OverworldVehicles {
		v.Draw(screen, camX, camY)
	}

	g.baseStation.Draw(screen, g.camera)
	if g.baseStation.DistanceToPlayer(g.player) < 100.0 {
		sx := float32(g.baseStation.Pos.X-camX) + float32(g.baseStation.Size.X)/2.0 - 90
		sy := float32(g.baseStation.Pos.Y-camY) - 30
		vector.FillRect(screen, sx, sy, 180, 24, color.RGBA{0, 0, 0, 180}, false)
		ebitenutil.DebugPrintAt(screen, "Press [E] to Open Terminal", int(sx)+12, int(sy)+4)
	}

	if g.ActiveVehicle == nil {
		g.drawVehicleEntryPrompts(screen, g.OverworldVehicles, camX, camY)
	}
}

// drawCaveLayer renders cave vehicles, the sonar ring, and interaction prompts.
func (g *Game) drawCaveLayer(screen *ebiten.Image) {
	if g.caveState.IsScrollActive() {
		g.Sonar.Draw(screen, g.camera)
		return
	}

	camX, camY := g.camera.Pos.X, g.camera.Pos.Y
	caveVehicles := g.CaveVehicles[g.activeTrenchKey]

	for _, v := range caveVehicles {
		v.Draw(screen, camX, camY)
	}
	g.Sonar.Draw(screen, g.camera)

	if g.ActiveVehicle == nil {
		g.drawVehicleEntryPrompts(screen, caveVehicles, camX, camY)
	}
}

// drawVehicleEntryPrompts shows "Press [F] to Pilot" above any vehicle within boarding range.
func (g *Game) drawVehicleEntryPrompts(screen *ebiten.Image, vehicles []vehicle.Vehicle, camX, camY float64) {
	for _, v := range vehicles {
		vPos := v.GetPos()
		vDims := v.GetDimensions()
		dist := math.Hypot(
			vPos.X+vDims.X/2.0-g.player.Pos.X-g.player.Width/2.0,
			vPos.Y+vDims.Y/2.0-g.player.Pos.Y-g.player.Height/2.0,
		)
		if dist >= 60.0 {
			continue
		}
		sx := float32(vPos.X-camX) + float32(vDims.X)/2.0 - 75
		sy := float32(vPos.Y-camY) - 25
		vector.FillRect(screen, sx, sy, 150, 20, color.RGBA{0, 0, 0, 180}, false)
		ebitenutil.DebugPrintAt(screen, "Press [F] to Pilot", int(sx)+22, int(sy)+2)
	}
}

// drawWarningBanner renders the MineWarning text if its timer is active.
func (g *Game) drawWarningBanner(screen *ebiten.Image) {
	if g.MineWarningTimer <= 0 {
		return
	}
	wx := float32(config.ScreenWidth)/2.0 - 160
	wy := float32(config.ScreenHeight) / 4.0
	vector.FillRect(screen, wx, wy, 320, 30, color.RGBA{24, 6, 8, 220}, false)
	vector.StrokeRect(screen, wx, wy, 320, 30, 1.2, color.RGBA{235, 45, 45, 255}, false)
	ebitenutil.DebugPrintAt(screen, g.MineWarning, int(wx)+12, int(wy)+7)
}
