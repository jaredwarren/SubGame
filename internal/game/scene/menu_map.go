package scene

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/world"
)

type mapPOI struct {
	TX, TY int
	Type   world.TileType
}

const (
	mapDrawScale = 0.96 // 500→480 px
	mapPixelSize = 500
)

// ResetMapCache drops the cached chart image so the next open rebuilds from the tracker.
func (m *BaseMenuScene) ResetMapCache() {
	m.mapImage = nil
	m.mapPixels = nil
	m.mapPOIs = nil
	m.mapSeed = 0
}

func (m *BaseMenuScene) drawMapTab(g MenuContext, screen *ebiten.Image, panelX, panelY float32) {
	w := g.GetWorld()
	tracker := g.GetExploration()
	if w == nil || tracker == nil {
		ebitenutil.DebugPrintAt(screen, "Charting systems offline.", int(panelX)+40, int(panelY)+120)
		return
	}

	m.syncMapImage(w, tracker)

	const mapScreen = float32(mapPixelSize) * mapDrawScale // 480
	mapX := panelX + 20
	mapY := panelY + 90

	vector.FillRect(screen, mapX-2, mapY-2, mapScreen+4, mapScreen+4, color.RGBA{8, 10, 16, 255}, false)
	vector.StrokeRect(screen, mapX-2, mapY-2, mapScreen+4, mapScreen+4, 1.0, color.RGBA{68, 88, 120, 255}, false)

	if m.mapImage != nil {
		op := &ebiten.DrawImageOptions{}
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(mapDrawScale, mapDrawScale)
		op.GeoM.Translate(float64(mapX), float64(mapY))
		screen.DrawImage(m.mapImage, op)
	}

	m.drawMapIcons(g, screen, mapX, mapY, mapDrawScale)

	// Legend panel
	legX := panelX + 520
	legY := panelY + 90
	legW := float32(250)
	legH := float32(380)
	vector.FillRect(screen, legX, legY, legW, legH, color.RGBA{16, 22, 34, 255}, false)
	vector.StrokeRect(screen, legX, legY, legW, legH, 1.0, color.RGBA{48, 62, 85, 255}, false)

	drawColoredDebugText(screen, "OCEAN CHART", int(legX)+12, int(legY)+12, color.RGBA{220, 180, 50, 255})
	pct := tracker.ExploredFraction() * 100
	drawColoredDebugText(screen, fmt.Sprintf("Charted: %.1f%%", pct), int(legX)+12, int(legY)+32, color.RGBA{140, 200, 220, 255})

	y := int(legY) + 60
	legendRow := func(label string, drawIcon func(cx, cy float32)) {
		drawIcon(legX+18, float32(y)+6)
		drawColoredDebugText(screen, label, int(legX)+36, y, color.RGBA{200, 200, 200, 255})
		y += 22
	}

	legendRow("You (overworld)", func(cx, cy float32) {
		vector.FillCircle(screen, cx, cy, 4, color.RGBA{0, 220, 255, 255}, false)
	})
	legendRow("Life Pod (always)", func(cx, cy float32) {
		vector.FillCircle(screen, cx, cy, 4, color.RGBA{80, 220, 120, 255}, false)
		vector.StrokeCircle(screen, cx, cy, 6, 1.0, color.RGBA{40, 160, 80, 255}, false)
	})
	legendRow("Lost Cargo (always)", func(cx, cy float32) {
		entity.DrawLostCargoMapIcon(screen, cx, cy, g.GetTicks())
	})
	legendRow("Unvisited site (?)", func(cx, cy float32) {
		vector.FillCircle(screen, cx, cy, 3, color.RGBA{255, 255, 255, 255}, false)
		vector.StrokeCircle(screen, cx, cy, 3.5, 1.0, color.RGBA{20, 20, 20, 255}, false)
	})
	legendRow("Trench (visited)", func(cx, cy float32) {
		drawTrenchIcon(screen, cx, cy, color.RGBA{255, 140, 40, 255})
	})
	legendRow("Wreck (visited)", func(cx, cy float32) {
		drawWreckIcon(screen, cx, cy, color.RGBA{220, 70, 70, 255})
	})
	legendRow("Shock Kelp (visited)", func(cx, cy float32) {
		drawKelpIcon(screen, cx, cy, color.RGBA{230, 220, 60, 255})
	})
	legendRow("Thermo Vent (visited)", func(cx, cy float32) {
		drawThermoIcon(screen, cx, cy, color.RGBA{255, 160, 40, 255})
	})

	y += 10
	drawColoredDebugText(screen, "Nearest wreck always marked.", int(legX)+12, y, color.RGBA{120, 140, 160, 255})
	y += 16
	drawColoredDebugText(screen, "Dive a ? site to identify it.", int(legX)+12, y, color.RGBA{120, 140, 160, 255})
	y += 28
	drawColoredDebugText(screen, "Press [M] or [J] to close", int(legX)+12, y, color.RGBA{160, 170, 180, 255})
}

func (m *BaseMenuScene) syncMapImage(w *world.World, tracker *exploration.Tracker) {
	if m.mapImage == nil || m.mapSeed != w.Seed || len(m.mapPixels) != w.Width*w.Height*4 {
		m.rebuildMapImage(w, tracker)
		_ = tracker.Drain() // drop backlog accumulated before first open
		return
	}
	dirty := tracker.Drain()
	if len(dirty) == 0 {
		return
	}
	for _, idx := range dirty {
		m.writeMapPixel(idx, w, tracker)
	}
	m.mapImage.WritePixels(m.mapPixels)
}

func (m *BaseMenuScene) rebuildMapImage(w *world.World, tracker *exploration.Tracker) {
	m.mapSeed = w.Seed
	m.mapPixels = make([]byte, w.Width*w.Height*4)
	m.mapImage = ebiten.NewImage(w.Width, w.Height)
	m.scanMapPOIs(w)

	for ty := 0; ty < w.Height; ty++ {
		for tx := 0; tx < w.Width; tx++ {
			m.writeMapPixel(ty*w.Width+tx, w, tracker)
		}
	}
	m.mapImage.WritePixels(m.mapPixels)
}

func (m *BaseMenuScene) scanMapPOIs(w *world.World) {
	m.mapPOIs = m.mapPOIs[:0]
	for tx := 0; tx < w.Width; tx++ {
		for ty := 0; ty < w.Height; ty++ {
			tt := w.OverworldMap[tx][ty]
			switch tt {
			case world.TileTrench, world.TileWreckage, world.TileShockKelpCave, world.TileThermoCave:
				m.mapPOIs = append(m.mapPOIs, mapPOI{TX: tx, TY: ty, Type: tt})
			}
		}
	}
}

func (m *BaseMenuScene) writeMapPixel(idx int, w *world.World, tracker *exploration.Tracker) {
	if idx < 0 || idx*4+3 >= len(m.mapPixels) {
		return
	}
	tx := idx % w.Width
	ty := idx / w.Width
	off := idx * 4

	if !tracker.IsExplored(tx, ty) {
		m.mapPixels[off+0] = exploration.FogColor[0]
		m.mapPixels[off+1] = exploration.FogColor[1]
		m.mapPixels[off+2] = exploration.FogColor[2]
		m.mapPixels[off+3] = 255
		return
	}

	r, g, b := mapTileRGB(w, tx, ty)
	m.mapPixels[off+0] = r
	m.mapPixels[off+1] = g
	m.mapPixels[off+2] = b
	m.mapPixels[off+3] = 255
}

func mapTileRGB(w *world.World, tx, ty int) (r, g, b uint8) {
	tt := w.OverworldMap[tx][ty]
	landDist := w.LandDist[tx][ty]
	waterDist := w.WaterDist[tx][ty]
	offset := w.GetSmoothedWaterOffset(tx, ty)
	// Darken ~40% vs live overworld so icons pop on the chart.
	clr, _ := ComputeTileColorsWithOffset(tx, ty, tt, landDist, waterDist, w.Width, w.Height, 0.6, offset)

	info := world.GetTileInfo(tt)
	if info != nil && info.IsWater {
		// Special dive sites share biome water color; icons carry identity.
		return clr.R, clr.G, clr.B
	}
	if tt == world.TileLand {
		// Chart sand tone (plan: #C9B27C), shaded by inland distance.
		if waterDist <= 1 {
			return 0xC9, 0xB2, 0x7C
		}
		return uint8(float64(clr.R)), uint8(float64(clr.G)), uint8(float64(clr.B))
	}
	return clr.R, clr.G, clr.B
}

func (m *BaseMenuScene) drawMapIcons(g MenuContext, screen *ebiten.Image, mapX, mapY float32, scale float64) {
	tracker := g.GetExploration()
	w := g.GetWorld()
	if tracker == nil || w == nil {
		return
	}

	toScreen := func(tx, ty int) (float32, float32) {
		return mapX + float32(float64(tx)*scale), mapY + float32(float64(ty)*scale)
	}

	// Nearest wreck to the Life Pod — always marked (even through fog) as a breadcrumb.
	nearestWreckTX, nearestWreckTY := -1, -1
	if base := g.GetBaseStation(); base != nil {
		baseTX := tileAt(base.Pos.X+base.Size.X/2.0, config.TileSize)
		baseTY := tileAt(base.Pos.Y+base.Size.Y/2.0, config.TileSize)
		nearestWreckTX, nearestWreckTY = nearestWreckage(m.mapPOIs, baseTX, baseTY)
	}

	// Dive-site markers (explored only; nearest wreck handled separately so it can pierce fog)
	for _, poi := range m.mapPOIs {
		if poi.TX == nearestWreckTX && poi.TY == nearestWreckTY {
			continue
		}
		if !tracker.IsExplored(poi.TX, poi.TY) {
			continue
		}
		px, py := toScreen(poi.TX, poi.TY)
		if tracker.IsVisited(poi.TX, poi.TY) {
			drawVisitedPOIIcon(screen, px, py, poi.Type)
		} else {
			drawUnvisitedPOIMarker(screen, px, py)
		}
	}

	if nearestWreckTX >= 0 {
		px, py := toScreen(nearestWreckTX, nearestWreckTY)
		if tracker.IsVisited(nearestWreckTX, nearestWreckTY) {
			drawVisitedPOIIcon(screen, px, py, world.TileWreckage)
		} else {
			drawUnvisitedPOIMarker(screen, px, py)
			// Literal "?" so the breadcrumb reads clearly on fog.
			drawColoredDebugText(screen, "?", int(px)-3, int(py)-12, color.RGBA{255, 255, 255, 255})
		}
	}

	// Life Pod — always visible
	if base := g.GetBaseStation(); base != nil {
		tx := tileAt(base.Pos.X+base.Size.X/2.0, config.TileSize)
		ty := tileAt(base.Pos.Y+base.Size.Y/2.0, config.TileSize)
		px, py := toScreen(tx, ty)
		vector.FillCircle(screen, px, py, 4, color.RGBA{80, 220, 120, 255}, false)
		vector.StrokeCircle(screen, px, py, 6, 1.2, color.RGBA{40, 160, 80, 255}, false)
	}

	// Lost cargo crates — always visible through fog (recovery expedition breadcrumb).
	for _, b := range g.GetLostCargo() {
		if b == nil || !b.Active() {
			continue
		}
		tx := tileAt(b.Pos.X, config.TileSize)
		ty := tileAt(b.Pos.Y, config.TileSize)
		px, py := toScreen(tx, ty)
		entity.DrawLostCargoMapIcon(screen, px, py, g.GetTicks())
	}

	// From a cave: highlight the dive site. Otherwise show the player marker.
	fromCave := g.IsMenuOpenedAnywhere() && g.GetPDAPriorState() == StateCave
	if fromCave {
		dtx, dty := g.GetActiveTrenchCoords()
		if dtx >= 0 && dtx < w.Width && dty >= 0 && dty < w.Height {
			px, py := toScreen(dtx, dty)
			pulse := float32(3.0 + math.Sin(g.GetTicks()*0.12)*1.5)
			vector.StrokeCircle(screen, px, py, pulse+4, 1.5, color.RGBA{0, 240, 255, 200}, false)
			vector.FillCircle(screen, px, py, 3, color.RGBA{0, 220, 255, 255}, false)
		}
		return
	}

	p := g.GetPlayer()
	tx := tileAt(p.Pos.X+p.Width/2.0, config.TileSize)
	ty := tileAt(p.Pos.Y+p.Height/2.0, config.TileSize)
	if tx >= 0 && tx < w.Width && ty >= 0 && ty < w.Height {
		px, py := toScreen(tx, ty)
		pulse := float32(2.0 + (math.Sin(g.GetTicks()*0.1)+1.0)*1.0) // 2–4 px
		vector.FillCircle(screen, px, py, pulse, color.RGBA{0, 220, 255, 255}, false)
		vector.StrokeCircle(screen, px, py, pulse+1.5, 1.0, color.RGBA{200, 255, 255, 180}, false)
	}
}

func nearestWreckage(pois []mapPOI, fromTX, fromTY int) (tx, ty int) {
	bestDist := -1
	tx, ty = -1, -1
	for _, poi := range pois {
		if poi.Type != world.TileWreckage {
			continue
		}
		dx := poi.TX - fromTX
		dy := poi.TY - fromTY
		dist := dx*dx + dy*dy
		if bestDist < 0 || dist < bestDist {
			bestDist = dist
			tx, ty = poi.TX, poi.TY
		}
	}
	return tx, ty
}

func drawUnvisitedPOIMarker(screen *ebiten.Image, px, py float32) {
	vector.FillCircle(screen, px, py, 3, color.RGBA{255, 255, 255, 255}, false)
	vector.StrokeCircle(screen, px, py, 3.5, 1.0, color.RGBA{20, 20, 20, 220}, false)
}

func drawVisitedPOIIcon(screen *ebiten.Image, px, py float32, tt world.TileType) {
	switch tt {
	case world.TileTrench:
		drawTrenchIcon(screen, px, py, color.RGBA{255, 140, 40, 255})
	case world.TileWreckage:
		drawWreckIcon(screen, px, py, color.RGBA{220, 70, 70, 255})
	case world.TileShockKelpCave:
		drawKelpIcon(screen, px, py, color.RGBA{230, 220, 60, 255})
	case world.TileThermoCave:
		drawThermoIcon(screen, px, py, color.RGBA{255, 160, 40, 255})
	default:
		vector.FillCircle(screen, px, py, 3, color.RGBA{200, 200, 200, 255}, false)
	}
}

func drawTrenchIcon(screen *ebiten.Image, cx, cy float32, clr color.RGBA) {
	// Downward chevron
	vector.StrokeLine(screen, cx-4, cy-2, cx, cy+3, 1.5, clr, false)
	vector.StrokeLine(screen, cx+4, cy-2, cx, cy+3, 1.5, clr, false)
}

func drawWreckIcon(screen *ebiten.Image, cx, cy float32, clr color.RGBA) {
	vector.StrokeLine(screen, cx-4, cy-4, cx+4, cy+4, 1.5, clr, false)
	vector.StrokeLine(screen, cx+4, cy-4, cx-4, cy+4, 1.5, clr, false)
}

func drawKelpIcon(screen *ebiten.Image, cx, cy float32, clr color.RGBA) {
	vector.StrokeLine(screen, cx-3, cy+3, cx-1, cy-3, 1.2, clr, false)
	vector.StrokeLine(screen, cx, cy+3, cx+1, cy-4, 1.2, clr, false)
	vector.StrokeLine(screen, cx+3, cy+3, cx+2, cy-2, 1.2, clr, false)
}

func drawThermoIcon(screen *ebiten.Image, cx, cy float32, clr color.RGBA) {
	// Small upward triangle
	vector.StrokeLine(screen, cx, cy-4, cx-4, cy+3, 1.4, clr, false)
	vector.StrokeLine(screen, cx, cy-4, cx+4, cy+3, 1.4, clr, false)
	vector.StrokeLine(screen, cx-4, cy+3, cx+4, cy+3, 1.4, clr, false)
}
