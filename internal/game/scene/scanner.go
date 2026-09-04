package scene

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

const (
	// ScanMaxRange is how far away the player can target an entity with the scanner.
	ScanMaxRange = 140.0
	// ScanDurationFrames is how many frames are required to complete a scan (~1.5s at 60 FPS).
	ScanDurationFrames = 90.0
	// ScannerPulseCooldownMax is the cooldown for firing an exploration pulse (6s at 60 FPS).
	ScannerPulseCooldownMax = 360
	// ScannerPulseDurationMax is how long the pulse highlights targets through walls (3.5s at 60 FPS).
	ScannerPulseDurationMax = 210
	// ScannerPulseMaxRadius is the maximum reach of the pulse wave (~8 tiles).
	ScannerPulseMaxRadius = 220.0
)

// ScannableTarget represents an identified world entity, plant, or ore vein.
type ScannableTarget struct {
	Key         string
	DisplayName string
	Pos         gvec.Vec2
	Radius      float64
	Entity      entity.CaveEntity
	Node        resource.Resource
}

// ScannerState manages scanning progress, locked targets, and the exploration pulse.
type ScannerState struct {
	ActiveTarget  *ScannableTarget
	Progress      float64 // 0.0 to 1.0
	AudioTimer    int     // countdown to re-trigger scanner loop audio
	PulseCooldown int     // counts down to 0
	PulseTimer    int     // counts down from ScannerPulseDurationMax
	PulseOrigin   gvec.Vec2
	PulseRadius   float64
	ScannedKeys   map[string]bool
}

// NewScannerState instantiates a fresh ScannerState.
func NewScannerState() *ScannerState {
	return &ScannerState{
		ScannedKeys: make(map[string]bool),
	}
}

// NormalizeTargetKey normalizes creature debug names and node names to lore trigger targets.
func NormalizeTargetKey(raw string) (key, displayName string) {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch lower {
	case "thermoclinerammer", "thermocline_rammer", "thermocline rammer":
		return "thermocline_rammer", "Thermocline Rammer"
	case "electroweaver", "electro_weaver", "electro weaver", "electro-weaver":
		return "electro_weaver", "Electro-Weaver"
	case "voltaiclurker", "voltaic_lurker", "voltaic lurker":
		return "voltaic_lurker", "Voltaic Lurker"
	case "falsebulbsnare", "false_bulb_snare", "false-bulb snare", "false bulb snare":
		return "false_bulb_snare", "False-Bulb Snare"
	case "inksquid", "ink_squid", "ink squid":
		return "ink_squid", "Ink Squid"
	case "glowsquid", "glow_squid", "glow squid":
		return "ink_squid", "Glow Squid"
	case "sandviper", "sand_viper", "sand viper":
		return "sand_viper", "Sand Viper"
	case "shatterbulb", "shatter-bulb", "shatter_bulb":
		return "shatter-bulb", "Shatter-Bulb"
	case "passivefish", "raw fish", "cave fish":
		return "raw fish", "Cave Fish"
	case "passivecrab", "raw crab", "cave crab":
		return "raw crab", "Cave Crab"
	case "scraphermitcrab", "scrap hermit crab", "scraphermit":
		return "scraphermitcrab", "Scrap Hermit Crab"
	case "lanternfish":
		return "raw fish", "Lanternfish"
	case "wreckterminal", "wreck_terminal":
		return "wreck_research_log", "Derelict Terminal"
	case "lostcargo", "lost_cargo":
		return "wreck_transport_manifest", "Salvage Cargo Crate"
	case "brimstonesiphon", "brimstone_siphon":
		return "shatter-bulb", "Brimstone Siphon"
	case "shockkelp", "shock_kelp":
		return "shatter-bulb", "Shock Kelp"
	case "nervemat", "nerve_mat":
		return "shatter-bulb", "Nerve Mat"
	case "copper", "copper vein":
		return "copper", "Copper Vein"
	case "abyssal ore", "abyssal shard":
		return "abyssal ore", "Abyssal Ore Shard"
	case "titanium":
		return "copper", "Titanium Vein"
	case "quartz":
		return "copper", "Quartz Crystal"
	case "nickel":
		return "copper", "Nickel Ore"
	case "reinforced blast bulkhead", "reinforced bulkhead":
		return "wreck_flagship_blackbox", "Reinforced Blast Bulkhead"
	default:
		return lower, raw
	}
}

// GetBlueprintForTarget returns the craftable item result name unlocked by scanning targetKey.
func GetBlueprintForTarget(targetKey string) string {
	targetKey = strings.ToLower(strings.TrimSpace(targetKey))
	switch targetKey {
	case "thermocline_rammer", "thermoclinerammer":
		return "Decoy Launcher Module"
	case "voltaic_lurker", "voltaiclurker":
		return "Chemical Discharger Module"
	case "ink_squid", "inksquid", "glow_squid", "glowsquid":
		return "Chemical Deterrent"
	case "electro_weaver", "electroweaver":
		return "Sonar Amplifier"
	default:
		return ""
	}
}

// FindScannableAt locates a creature or node within range of the player and near the cursor.
func FindScannableAt(nodes []resource.Resource, entities []entity.CaveEntity, px, py, targetX, targetY, maxDist float64) *ScannableTarget {
	var best *ScannableTarget
	bestDistToCursor := 48.0 // cursor hit radius

	// 1. Check entities (creatures, plants, wrecks)
	for _, ent := range entities {
		if !ent.IsActive() {
			continue
		}
		pos := ent.GetPos()
		dims := ent.GetDimensions()
		cx := pos.X + dims.X/2.0
		cy := pos.Y + dims.Y/2.0

		// Must be in range of player
		distToPlayer := math.Hypot(px-cx, py-cy)
		if distToPlayer > maxDist {
			continue
		}

		distToCursor := math.Hypot(targetX-cx, targetY-cy)
		targetRadius := math.Max(dims.X, dims.Y) / 2.0
		if targetRadius < 16.0 {
			targetRadius = 16.0
		}

		if distToCursor <= targetRadius+20.0 && distToCursor < bestDistToCursor {
			bestDistToCursor = distToCursor
			k, dName := NormalizeTargetKey(ent.DebugName())
			best = &ScannableTarget{
				Key:         k,
				DisplayName: dName,
				Pos:         gvec.Vec2{X: cx, Y: cy},
				Radius:      targetRadius,
				Entity:      ent,
			}
		}
	}

	// 2. Check nodes if no entity was directly hit
	if best == nil {
		for _, node := range nodes {
			if node.GetHitsToMine() <= 0 {
				continue
			}
			tx, ty := node.GetTilePos()
			nx := float64(tx*config.TileSize + config.TileSize/2)
			ny := float64(ty*config.TileSize + config.TileSize/2)

			distToPlayer := math.Hypot(px-nx, py-ny)
			if distToPlayer > maxDist {
				continue
			}

			distToCursor := math.Hypot(targetX-nx, targetY-ny)
			if distToCursor <= float64(config.TileSize)+12.0 && distToCursor < bestDistToCursor {
				bestDistToCursor = distToCursor
				k, dName := NormalizeTargetKey(node.GetName())
				best = &ScannableTarget{
					Key:         k,
					DisplayName: dName,
					Pos:         gvec.Vec2{X: nx, Y: ny},
					Radius:      float64(config.TileSize) / 2.0,
					Node:        node,
				}
			}
		}
	}

	return best
}

// FindNearestScannable finds the closest in-range scannable entity or node (used for mobile USE button).
func FindNearestScannable(nodes []resource.Resource, entities []entity.CaveEntity, px, py float64, maxDist float64) *ScannableTarget {
	var best *ScannableTarget
	bestDist := maxDist

	// 1. Entities
	for _, ent := range entities {
		if !ent.IsActive() {
			continue
		}
		pos := ent.GetPos()
		dims := ent.GetDimensions()
		cx := pos.X + dims.X/2.0
		cy := pos.Y + dims.Y/2.0

		dist := math.Hypot(px-cx, py-cy)
		if dist <= bestDist {
			bestDist = dist
			k, dName := NormalizeTargetKey(ent.DebugName())
			r := math.Max(dims.X, dims.Y) / 2.0
			if r < 16.0 {
				r = 16.0
			}
			best = &ScannableTarget{
				Key:         k,
				DisplayName: dName,
				Pos:         gvec.Vec2{X: cx, Y: cy},
				Radius:      r,
				Entity:      ent,
			}
		}
	}

	// 2. Nodes
	for _, node := range nodes {
		if node.GetHitsToMine() <= 0 {
			continue
		}
		tx, ty := node.GetTilePos()
		nx := float64(tx*config.TileSize + config.TileSize/2)
		ny := float64(ty*config.TileSize + config.TileSize/2)

		dist := math.Hypot(px-nx, py-ny)
		if dist <= bestDist {
			bestDist = dist
			k, dName := NormalizeTargetKey(node.GetName())
			best = &ScannableTarget{
				Key:         k,
				DisplayName: dName,
				Pos:         gvec.Vec2{X: nx, Y: ny},
				Radius:      float64(config.TileSize) / 2.0,
				Node:        node,
			}
		}
	}

	return best
}

// DrawScanningHUD renders the holographic reticle and circular progress arc around the active target.
func (s *ScannerState) DrawScanningHUD(screen *ebiten.Image, cam *camera.Camera) {
	if s.ActiveTarget == nil {
		return
	}

	sx := float32(s.ActiveTarget.Pos.X - cam.Pos.X)
	sy := float32(s.ActiveTarget.Pos.Y - cam.Pos.Y)
	r := float32(s.ActiveTarget.Radius + 8.0)
	if r < 20.0 {
		r = 20.0
	}

	// Holographic reticle brackets (Cyan/Teal)
	reticleColor := color.RGBA{60, 220, 255, 230}
	bracketLen := float32(8.0)

	// Top-left bracket
	vector.StrokeLine(screen, sx-r, sy-r, sx-r+bracketLen, sy-r, 1.5, reticleColor, false)
	vector.StrokeLine(screen, sx-r, sy-r, sx-r, sy-r+bracketLen, 1.5, reticleColor, false)
	// Top-right bracket
	vector.StrokeLine(screen, sx+r, sy-r, sx+r-bracketLen, sy-r, 1.5, reticleColor, false)
	vector.StrokeLine(screen, sx+r, sy-r, sx+r, sy-r+bracketLen, 1.5, reticleColor, false)
	// Bottom-left bracket
	vector.StrokeLine(screen, sx-r, sy+r, sx-r+bracketLen, sy+r, 1.5, reticleColor, false)
	vector.StrokeLine(screen, sx-r, sy+r, sx-r, sy+r-bracketLen, 1.5, reticleColor, false)
	// Bottom-right bracket
	vector.StrokeLine(screen, sx+r, sy+r, sx+r-bracketLen, sy+r, 1.5, reticleColor, false)
	vector.StrokeLine(screen, sx+r, sy+r, sx+r, sy+r-bracketLen, 1.5, reticleColor, false)

	// Progress arc ring
	const segments = 32
	progSegments := int(float64(segments) * s.Progress)
	angleStep := (2.0 * math.Pi) / float64(segments)

	ringColorBg := color.RGBA{30, 80, 110, 100}
	vector.StrokeCircle(screen, sx, sy, r+4.0, 1.5, ringColorBg, false)

	arcColor := color.RGBA{90, 245, 255, 255}
	for i := 0; i < progSegments; i++ {
		a1 := -math.Pi/2.0 + float64(i)*angleStep
		a2 := a1 + angleStep
		x1 := sx + float32(math.Cos(a1))*(r+4.0)
		y1 := sy + float32(math.Sin(a1))*(r+4.0)
		x2 := sx + float32(math.Cos(a2))*(r+4.0)
		y2 := sy + float32(math.Sin(a2))*(r+4.0)
		vector.StrokeLine(screen, x1, y1, x2, y2, 2.2, arcColor, false)
	}

	// Text readout: [SCANNING: TARGET NAME XX%]
	percent := int(s.Progress * 100.0)
	statusText := fmt.Sprintf("SCANNING: %s %d%%", strings.ToUpper(s.ActiveTarget.DisplayName), percent)
	textW := float32(len(statusText) * 6)
	ebitenutil.DebugPrintAt(screen, statusText, int(sx-textW/2.0), int(sy+r+10.0))
}

// DrawExplorationPulse renders the expanding sonar shockwave and highlights visible through walls.
func (s *ScannerState) DrawExplorationPulse(screen *ebiten.Image, cam *camera.Camera, nodes []resource.Resource, entities []entity.CaveEntity) {
	if s.PulseTimer <= 0 {
		return
	}

	// Calculate pulse progress (0.0 at trigger to 1.0 when fully expanded)
	timeElapsed := float64(ScannerPulseDurationMax - s.PulseTimer)
	waveProgress := math.Min(1.0, timeElapsed/40.0)
	currentWaveRadius := float32(waveProgress * ScannerPulseMaxRadius)

	originSx := float32(s.PulseOrigin.X - cam.Pos.X)
	originSy := float32(s.PulseOrigin.Y - cam.Pos.Y)

	// 1. Draw expanding shockwave rings if still active
	if waveProgress < 1.0 {
		alpha := uint8(float64(220) * (1.0 - waveProgress))
		waveColor := color.RGBA{50, 220, 245, alpha}
		vector.StrokeCircle(screen, originSx, originSy, currentWaveRadius, 2.5, waveColor, false)
		if currentWaveRadius > 15.0 {
			vector.StrokeCircle(screen, originSx, originSy, currentWaveRadius-12.0, 1.0, color.RGBA{40, 180, 220, alpha / 2}, false)
		}
	}

	// 2. Highlight entities and nodes within pulse radius through walls
	pulseAlpha := uint8(180)
	if s.PulseTimer < 60 {
		pulseAlpha = uint8(float64(180) * (float64(s.PulseTimer) / 60.0))
	}
	glowColor := color.RGBA{40, 230, 255, pulseAlpha}

	for _, ent := range entities {
		if !ent.IsActive() {
			continue
		}
		pos := ent.GetPos()
		dims := ent.GetDimensions()
		cx := pos.X + dims.X/2.0
		cy := pos.Y + dims.Y/2.0

		if math.Hypot(cx-s.PulseOrigin.X, cy-s.PulseOrigin.Y) <= ScannerPulseMaxRadius {
			esx := float32(cx - cam.Pos.X)
			esy := float32(cy - cam.Pos.Y)
			er := float32(math.Max(dims.X, dims.Y)/2.0 + 4.0)
			if er < 14.0 {
				er = 14.0
			}

			vector.StrokeCircle(screen, esx, esy, er, 1.2, glowColor, false)
			_, dName := NormalizeTargetKey(ent.DebugName())
			ebitenutil.DebugPrintAt(screen, dName, int(esx-float32(len(dName)*3)), int(esy-er-12.0))
		}
	}

	for _, node := range nodes {
		if node.GetHitsToMine() <= 0 {
			continue
		}
		tx, ty := node.GetTilePos()
		nx := float64(tx*config.TileSize + config.TileSize/2)
		ny := float64(ty*config.TileSize + config.TileSize/2)

		if math.Hypot(nx-s.PulseOrigin.X, ny-s.PulseOrigin.Y) <= ScannerPulseMaxRadius {
			nsx := float32(nx - cam.Pos.X)
			nsy := float32(ny - cam.Pos.Y)
			nr := float32(config.TileSize/2.0 + 2.0)

			vector.StrokeRect(screen, nsx-nr, nsy-nr, nr*2.0, nr*2.0, 1.0, glowColor, false)
			nodeName := node.GetName()
			ebitenutil.DebugPrintAt(screen, nodeName, int(nsx-float32(len(nodeName)*3)), int(nsy-nr-12.0))
		}
	}
}
