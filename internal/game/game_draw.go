package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/particle"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
	"github.com/jaredwarren/SubGame/internal/gvec"
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

	lightMult := GetOverworldLightMultiplier(g.TimeOfDay)
	g.baseStation.Draw(screen, g.camera, lightMult)
	if g.baseStation.DistanceToPlayer(g.player) < 100.0 {
		sx := float32(g.baseStation.Pos.X-camX) + float32(g.baseStation.Size.X)/2.0 - 90
		sy := float32(g.baseStation.Pos.Y-camY) - 30
		vector.FillRect(screen, sx, sy, 180, 24, color.RGBA{0, 0, 0, 180}, false)
		ebitenutil.DebugPrintAt(screen, "Press [E] to Open Terminal", int(sx)+12, int(sy)+4)
	}

	if g.ActiveVehicle == nil {
		g.drawVehicleEntryPrompts(screen, g.OverworldVehicles, camX, camY)
	}

	g.drawWaypointMarker(screen)
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
	if g.MineWarning.Timer <= 0 {
		return
	}

	borderColor := color.RGBA{R: 0, G: 191, B: 255, A: 255}
	bgColor := color.RGBA{R: 0, G: 16, B: 32, A: 220}
	glowColor := color.RGBA{R: 0, G: 191, B: 255, A: 60}

	switch g.MineWarning.Level {
	case 2: // warn/yellow
		borderColor = color.RGBA{R: 255, G: 215, B: 0, A: 255}
		bgColor = color.RGBA{R: 32, G: 24, B: 0, A: 220}
		glowColor = color.RGBA{R: 255, G: 215, B: 0, A: 60}
	case 3: // alert/red
		borderColor = color.RGBA{R: 235, G: 45, B: 45, A: 255}
		bgColor = color.RGBA{R: 24, G: 6, B: 8, A: 220}
		glowColor = color.RGBA{R: 235, G: 45, B: 45, A: 60}
	}

	wx := float32(config.ScreenWidth)/2.0 - 160
	wy := float32(config.ScreenHeight) / 4.0

	// Draw dark background
	vector.FillRect(screen, wx, wy, 320, 30, bgColor, false)
	// Draw glowing outer border
	vector.StrokeRect(screen, wx-1.5, wy-1.5, 323, 33, 2.5, glowColor, false)
	// Draw sharp inner border
	vector.StrokeRect(screen, wx, wy, 320, 30, 1.2, borderColor, false)
	// Print text
	ebitenutil.DebugPrintAt(screen, g.MineWarning.Message, int(wx)+12, int(wy)+7)
}

// drawWaypointMarker draws a wayfinding HUD element that always points back to the base station/lifepod.
func (g *Game) drawWaypointMarker(screen *ebiten.Image) {
	if g.currentState != StateOverworld {
		return
	}

	// 1. Calculate base center in world space
	baseCenter := gvec.Vec2{
		X: g.baseStation.Pos.X + g.baseStation.Size.X/2.0,
		Y: g.baseStation.Pos.Y + g.baseStation.Size.Y/2.0,
	}

	camX, camY := g.camera.Pos.X, g.camera.Pos.Y

	// 2. Base center in screen space
	screenX := baseCenter.X - camX
	screenY := baseCenter.Y - camY

	// 3. Define layout margins and limits
	margin := 40.0
	screenWidth := float64(config.ScreenWidth)
	screenHeight := float64(config.ScreenHeight)

	// 4. Calculate distance from player center to base center
	playerCenter := gvec.Vec2{
		X: g.player.Pos.X + g.player.Width/2.0,
		Y: g.player.Pos.Y + g.player.Height/2.0,
	}
	dist := math.Hypot(baseCenter.X-playerCenter.X, baseCenter.Y-playerCenter.Y)
	distMeters := int(dist / 16.0)

	// Format text string
	textStr := fmt.Sprintf("LIFEPOD (%dm)", distMeters)

	// 5. Check if lifepod is on screen
	isOnScreen := screenX >= margin && screenX <= screenWidth-margin &&
		screenY >= margin && screenY <= screenHeight-margin

	var markerX, markerY float64
	cyanColor := color.RGBA{R: 0, G: 200, B: 255, A: 255}
	cyanDimColor := color.RGBA{R: 0, G: 200, B: 255, A: 120}
	darkBgColor := color.RGBA{R: 0, G: 4, B: 12, A: 200}
	borderOutlineColor := color.RGBA{R: 0, G: 200, B: 255, A: 100}

	if isOnScreen {
		markerX = screenX
		markerY = screenY - 50.0 // Hover above the lifepod

		// Draw hovering waypoint pin (diamond + pointing stem)
		// Draw a small diamond
		vector.StrokeLine(screen, float32(markerX), float32(markerY-8), float32(markerX+6), float32(markerY), 1.2, cyanColor, false)
		vector.StrokeLine(screen, float32(markerX+6), float32(markerY), float32(markerX), float32(markerY+8), 1.2, cyanColor, false)
		vector.StrokeLine(screen, float32(markerX), float32(markerY+8), float32(markerX-6), float32(markerY), 1.2, cyanColor, false)
		vector.StrokeLine(screen, float32(markerX-6), float32(markerY), float32(markerX), float32(markerY-8), 1.2, cyanColor, false)

		// Fill a tiny center dot
		vector.FillCircle(screen, float32(markerX), float32(markerY), 2.5, cyanColor, false)

		// Draw text above the hovering pin
		textWidth := len(textStr) * 6
		textX := int(markerX) - textWidth/2
		textY := int(markerY) - 24

		// Draw text box background and border
		vector.FillRect(screen, float32(textX-4), float32(textY-2), float32(textWidth+8), 16, darkBgColor, false)
		vector.StrokeRect(screen, float32(textX-4), float32(textY-2), float32(textWidth+8), 16, 1.0, borderOutlineColor, false)
		ebitenutil.DebugPrintAt(screen, textStr, textX, textY)
	} else {
		// Pinned to edge. Calculate intersection with boundary box.
		center := gvec.Vec2{X: screenWidth / 2.0, Y: screenHeight / 2.0}
		dir := gvec.Vec2{X: screenX - center.X, Y: screenY - center.Y}

		minX := margin
		maxX := screenWidth - margin
		minY := margin
		maxY := screenHeight - margin

		t := 1.0
		var tCandidates []float64
		if dir.X < 0 {
			tCandidates = append(tCandidates, (minX-center.X)/dir.X)
		} else if dir.X > 0 {
			tCandidates = append(tCandidates, (maxX-center.X)/dir.X)
		}
		if dir.Y < 0 {
			tCandidates = append(tCandidates, (minY-center.Y)/dir.Y)
		} else if dir.Y > 0 {
			tCandidates = append(tCandidates, (maxY-center.Y)/dir.Y)
		}

		if len(tCandidates) > 0 {
			minT := tCandidates[0]
			for _, tc := range tCandidates {
				if tc < minT {
					minT = tc
				}
			}
			t = minT
		}

		markerX = center.X + dir.X*t
		markerY = center.Y + dir.Y*t

		// Draw a circular HUD badge
		vector.FillCircle(screen, float32(markerX), float32(markerY), 15, darkBgColor, false)
		vector.StrokeCircle(screen, float32(markerX), float32(markerY), 15, 1.2, cyanDimColor, false)

		// Draw pointing arrow inside the badge
		angle := math.Atan2(dir.Y, dir.X)
		tipX := markerX + math.Cos(angle)*8
		tipY := markerY + math.Sin(angle)*8
		leftX := markerX + math.Cos(angle+2.3)*6
		leftY := markerY + math.Sin(angle+2.3)*6
		rightX := markerX + math.Cos(angle-2.3)*6
		rightY := markerY + math.Sin(angle-2.3)*6

		vector.StrokeLine(screen, float32(tipX), float32(tipY), float32(leftX), float32(leftY), 1.5, cyanColor, false)
		vector.StrokeLine(screen, float32(tipX), float32(tipY), float32(rightX), float32(rightY), 1.5, cyanColor, false)
		vector.StrokeLine(screen, float32(leftX), float32(leftY), float32(rightX), float32(rightY), 1.5, cyanColor, false)

		// Dynamic text placement to prevent clipping
		textWidth := len(textStr) * 6
		var textX, textY int
		if markerX < center.X-150 {
			// Pinned left edge, draw text to the right
			textX = int(markerX) + 20
			textY = int(markerY) - 7
		} else if markerX > center.X+150 {
			// Pinned right edge, draw text to the left
			textX = int(markerX) - 20 - textWidth
			textY = int(markerY) - 7
		} else {
			// Pinned top/bottom edge
			if markerY < center.Y {
				// Top edge, draw text below
				textX = int(markerX) - textWidth/2
				textY = int(markerY) + 20
			} else {
				// Bottom edge, draw text above
				textX = int(markerX) - textWidth/2
				textY = int(markerY) - 30
			}
		}

		// Draw text box background and border
		vector.FillRect(screen, float32(textX-4), float32(textY-2), float32(textWidth+8), 16, darkBgColor, false)
		vector.StrokeRect(screen, float32(textX-4), float32(textY-2), float32(textWidth+8), 16, 1.0, borderOutlineColor, false)
		ebitenutil.DebugPrintAt(screen, textStr, textX, textY)
	}
}
