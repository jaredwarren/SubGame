package cave

import (
	"math"
	"math/rand"

	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// BandedSpawn is a per-tile chance spawn constrained to a depth band and topology.
type BandedSpawn struct {
	MinTY, MaxTY int // MaxTY exclusive; use a large MaxTY for open-ended
	Chance       float64
	Fauna        FaunaID // set when spawning fauna; ignored if HasFlora
	Flora        FloraID
	HasFauna     bool
	HasFlora     bool
	NeedCeiling  bool
	NeedFloor    bool
	NeedOpen     bool
	NeedWall     bool
	MinSpacingPX float64 // skip if same fauna type already within this distance
	FloraAnchor  string  // "floor" (default), "left", "right"
}

// CountSpawn places Min–Max instances of a fauna type from matching candidate tiles.
type CountSpawn struct {
	Fauna          FaunaID
	Min, Max       int
	MinTY          int
	NeedOpen       bool
	NeedWall       bool
	MinTileSpacing float64 // tile units; 0 = no spacing constraint
}

// AnchoredFloraSpawn places flora on a specific solid face with a per-tile chance.
type AnchoredFloraSpawn struct {
	Anchor string // floor, left, right
	Chance float64
	Flora  FloraID
	MinH   float64
	MaxH   float64
}

// CaveSpec is the data-driven description of a cave type.
type CaveSpec struct {
	Type         CaveType
	Biome        *CaveBiomeSpec
	Music        string
	Ambient      [4]float32
	CoralBiome   int
	CoralChance  float64
	Banded       []BandedSpawn
	Counts       []CountSpawn
	AnchoredFlora []AnchoredFloraSpawn
}

// Spec returns the registry entry for t, or nil.
func Spec(t CaveType) *CaveSpec {
	return caveRegistry[t]
}

// MusicTrack returns the looping music path for a cave type.
func MusicTrack(t CaveType) string {
	if s := Spec(t); s != nil && s.Music != "" {
		return s.Music
	}
	return "music/cave_shallow.mp3"
}

// AmbientColor returns the default ambient tint for a cave type.
func AmbientColor(t CaveType) [4]float32 {
	if s := Spec(t); s != nil {
		return s.Ambient
	}
	return [4]float32{0.04, 0.06, 0.12, 0.85}
}

var caveRegistry = map[CaveType]*CaveSpec{
	CaveOrganicShallow: {
		Type:        CaveOrganicShallow,
		Biome:       ShallowReefBiome,
		Music:       "music/cave_shallow.mp3",
		Ambient:     [4]float32{0.04, 0.06, 0.12, 0.85},
		CoralBiome:  entity.CoralBiomeShallow,
		CoralChance: 0.10,
	},
	CaveOrganicTrench: {
		Type:        CaveOrganicTrench,
		Biome:       AbyssalBlueBiome,
		Music:       "music/cave_abyssal.mp3",
		Ambient:     [4]float32{0.01, 0.015, 0.03, 0.92},
		CoralBiome:  entity.CoralBiomeTrench,
		CoralChance: 0.22,
		Banded: []BandedSpawn{
			{MinTY: 4, MaxTY: 40, Chance: 0.10, HasFlora: true, Flora: FloraShatterBulb, NeedFloor: true},
			{MinTY: 4, MaxTY: 50, Chance: 0.05, HasFauna: true, Fauna: FaunaFalseBulbSnare, NeedCeiling: true},
			// Lanternfish: decreasing spawn rate with depth (spawn in open water or grotto tunnels)
			{MinTY: 4, MaxTY: 35, Chance: 0.012, HasFauna: true, Fauna: FaunaLanternfish, MinSpacingPX: 110},
			{MinTY: 35, MaxTY: 65, Chance: 0.006, HasFauna: true, Fauna: FaunaLanternfish, MinSpacingPX: 160},
			{MinTY: 65, MaxTY: 85, Chance: 0.002, HasFauna: true, Fauna: FaunaLanternfish, MinSpacingPX: 220},
			// InkSquid in mid depths (density cut in half with spacing)
			{MinTY: 20, MaxTY: 65, Chance: 0.012, HasFauna: true, Fauna: FaunaInkSquid, NeedOpen: true, MinSpacingPX: 220},
			// GlowSquid exclusively at deep abyssal depths
			{MinTY: 65, MaxTY: 9999, Chance: 0.015, HasFauna: true, Fauna: FaunaGlowSquid, NeedOpen: true, MinSpacingPX: 200},
			{MinTY: 40, MaxTY: 80, Chance: 0.06, HasFauna: true, Fauna: FaunaBrimstoneSiphon, NeedWall: true},
			{MinTY: 40, MaxTY: 80, Chance: 0.015, HasFauna: true, Fauna: FaunaThermoclineRammer, NeedOpen: true},
			{MinTY: 4, MaxTY: 80, Chance: 0.18, HasFlora: true, Flora: FloraShockKelp, NeedFloor: true},
			{MinTY: 70, MaxTY: 9999, Chance: 0.16, HasFlora: true, Flora: FloraNerveMat, NeedFloor: true},
			{MinTY: 70, MaxTY: 9999, Chance: 0.020, HasFauna: true, Fauna: FaunaElectroWeaver, NeedOpen: true, MinSpacingPX: 400},
		},
	},
	CaveShockKelp: {
		Type:        CaveShockKelp,
		Biome:       KelpForestBiome,
		Music:       "music/cave_kelp.mp3",
		Ambient:     [4]float32{0.03, 0.02, 0.06, 0.68},
		CoralBiome:  entity.CoralBiomeShock,
		CoralChance: 0.10,
		AnchoredFlora: []AnchoredFloraSpawn{
			{Anchor: "floor", Chance: 0.60, Flora: FloraShockKelp, MinH: 24, MaxH: 52},
			{Anchor: "left", Chance: 0.45, Flora: FloraShockKelp, MinH: 24, MaxH: 52},
			{Anchor: "right", Chance: 0.45, Flora: FloraShockKelp, MinH: 24, MaxH: 52},
		},
		Banded: []BandedSpawn{
			{MinTY: 1, MaxTY: 9999, Chance: 0.08, HasFauna: true, Fauna: FaunaVoltaicLurker, NeedWall: true},
		},
		Counts: []CountSpawn{
			{Fauna: FaunaElectroWeaver, Min: 0, Max: 2},
		},
	},
	CaveThermo: {
		Type:        CaveThermo,
		Biome:       ThermalBarrensBiome,
		Music:       "music/cave_volcanic.mp3",
		Ambient:     [4]float32{0.02, 0.01, 0.01, 0.95},
		CoralBiome:  entity.CoralBiomeThermo,
		CoralChance: 0.10,
		Counts: []CountSpawn{
			{Fauna: FaunaThermoclineRammer, Min: 1, Max: 2, NeedOpen: true},
			{Fauna: FaunaBrimstoneSiphon, Min: 4, Max: 6, NeedWall: true},
		},
	},
	CaveWreckage: {
		Type:        CaveWreckage,
		Biome:       nil,
		Music:       "music/cave_wreckage.mp3",
		Ambient:     [4]float32{0.01, 0.01, 0.03, 0.97},
		CoralBiome:  entity.CoralBiomeWreckage,
		CoralChance: 0.08,
		Banded: []BandedSpawn{
			{MinTY: 2, MaxTY: 9999, Chance: 0.05, HasFlora: true, Flora: FloraShatterBulb, NeedFloor: true},
		},
		Counts: []CountSpawn{
			{Fauna: FaunaElectroWeaver, Min: 1, Max: 2, MinTY: 80, NeedOpen: true, MinTileSpacing: 5},
		},
	},
	CaveVoid: {
		Type:    CaveVoid,
		Music:   "music/cave_abyssal.mp3",
		Ambient: [4]float32{0.01, 0.01, 0.03, 0.97},
	},
}

// GenerateEntitiesFromSpec runs deep-cave spawn tables (banded/count/anchored + coral).
// Shallow seabed keeps its own biome loop for historical seed stability.
func GenerateEntitiesFromSpec(spec *CaveSpec, grid [][]bool, seed int64) []entity.CaveEntity {
	return GenerateDeepEntitiesFromSpec(spec, grid, seed)
}

// GenerateDeepEntitiesFromSpec runs Banded/Count/Anchored tables plus coral.
func GenerateDeepEntitiesFromSpec(spec *CaveSpec, grid [][]bool, seed int64) []entity.CaveEntity {
	if spec == nil || grid == nil {
		return nil
	}
	r := rand.New(rand.NewSource(seed))
	var entities []entity.CaveEntity
	entities = append(entities, applyAnchoredFlora(spec.AnchoredFlora, grid, r)...)
	entities = append(entities, applyBandedSpawns(spec.Banded, grid, r)...)
	entities = append(entities, applyCoralOnly(grid, spec.CoralChance, spec.CoralBiome, r)...)
	entities = append(entities, applyCountSpawns(spec.Counts, grid, entities, r)...)
	return entities
}

// GenerateShallowBiomeEntities runs the historical shallow seabed tile spawn loop.
func GenerateShallowBiomeEntities(biome *CaveBiomeSpec, grid [][]bool, coralBiome int, r *rand.Rand) []entity.CaveEntity {
	return generateBiomeTileEntities(biome, grid, coralBiome, r)
}

func generateBiomeTileEntities(biome *CaveBiomeSpec, grid [][]bool, coralBiome int, r *rand.Rand) []entity.CaveEntity {
	rules := biome.SpawnRulesOrDefault()
	var entities []entity.CaveEntity
	gridW := len(grid)
	gridH := len(grid[0])
	for tx := 1; tx < gridW-1; tx++ {
		for ty := 2; ty < gridH-2; ty++ {
			if grid[tx][ty] {
				continue
			}
			hasFloor := ty < gridH-2 && grid[tx][ty+1]
			if hasFloor && r.Float64() < rules.ShatterBulbChance {
				height := 42.0 + r.Float64()*16.0
				if ent := SpawnFlora(FloraShatterBulb, tx, ty, height, r); ent != nil {
					entities = append(entities, ent)
				}
			}
			isOpenWater := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
			if isOpenWater {
				roll := r.Float64()
				if roll < rules.OpenWaterFishChance {
					if ent := SpawnFauna(FaunaPassiveFish, tx, ty, grid, r); ent != nil {
						entities = append(entities, ent)
					}
				} else if roll < rules.OpenWaterFishChance+0.006 {
					if ent := SpawnFauna(FaunaInkSquid, tx, ty, grid, r); ent != nil {
						entities = append(entities, ent)
					}
				}
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < rules.FaunaChance {
				faunaType := FaunaPassiveFish
				if len(biome.FaunaSpawns) > 0 {
					faunaType = SelectWeightedEntry(biome.FaunaSpawns, r.Float64())
				}
				if ent := SpawnFauna(faunaType, tx, ty, grid, r); ent != nil {
					entities = append(entities, ent)
				}
			}
			if ty < gridH-2 && grid[tx][ty+1] && r.Float64() < rules.FloraChance {
				height := 32.0 + r.Float64()*48.0
				floraType := FloraKelp
				if len(biome.FloraSpawns) > 0 {
					floraType = SelectWeightedEntry(biome.FloraSpawns, r.Float64())
				}
				if ent := SpawnFlora(floraType, tx, ty, height, r); ent != nil {
					entities = append(entities, ent)
				}
			} else if ty > 1 && ty < gridH-2 && !grid[tx][ty+1] && !grid[tx][ty-1] && (grid[tx-1][ty] || grid[tx+1][ty]) && r.Float64() < rules.FloraChance {
				if len(biome.FloraSpawns) > 0 {
					floraType := SelectWeightedEntry(biome.FloraSpawns, r.Float64())
					if floraType == FloraShatterBulb || floraType == FloraShockKelp {
						height := 32.0 + r.Float64()*48.0
						anchor := "left"
						if grid[tx-1][ty] && grid[tx+1][ty] {
							if r.Float64() < 0.5 {
								anchor = "right"
							}
						} else if grid[tx+1][ty] {
							anchor = "right"
						}
						if ent := SpawnFloraAnchored(floraType, tx, ty, height, anchor, r); ent != nil {
							entities = append(entities, ent)
						}
					}
				}
			}
			entities = MaybeSpawnCoral(entities, grid, tx, ty, rules.CoralChance, coralBiome, entity.CoralVariantCount(coralBiome), r)
		}
	}

	// For biomes featuring Ink Squid, ensure at least 1 InkSquid spawns in the cave
	hasSquidInBiome := false
	if biome != nil {
		for _, s := range biome.FaunaSpawns {
			if s.Type == FaunaInkSquid {
				hasSquidInBiome = true
				break
			}
		}
	}
	if hasSquidInBiome {
		squidCount := 0
		for _, ent := range entities {
			if _, ok := ent.(*entity.InkSquid); ok {
				squidCount++
			}
		}
		for squidCount < 1 {
			found := false
			for attempts := 0; attempts < 100; attempts++ {
				tx := 2 + r.Intn(gridW-4)
				ty := 2 + r.Intn(gridH-4)
				if !grid[tx][ty] && !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1] {
					if ent := SpawnFauna(FaunaInkSquid, tx, ty, grid, r); ent != nil {
						entities = append(entities, ent)
						squidCount++
						found = true
						break
					}
				}
			}
			if !found {
				break
			}
		}
	}

	return entities
}

func applyCoralOnly(grid [][]bool, chance float64, coralBiome int, r *rand.Rand) []entity.CaveEntity {
	var entities []entity.CaveEntity
	if chance <= 0 {
		return entities
	}
	gridW := len(grid)
	gridH := len(grid[0])
	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-2; ty++ {
			if grid[tx][ty] {
				continue
			}
			entities = MaybeSpawnCoral(entities, grid, tx, ty, chance, coralBiome, entity.CoralBiomeVariantCount, r)
		}
	}
	return entities
}

func applyAnchoredFlora(spawns []AnchoredFloraSpawn, grid [][]bool, r *rand.Rand) []entity.CaveEntity {
	var entities []entity.CaveEntity
	gridW := len(grid)
	gridH := len(grid[0])
	for _, sp := range spawns {
		for tx := 1; tx < gridW-1; tx++ {
			for ty := 1; ty < gridH-2; ty++ {
				if grid[tx][ty] {
					continue
				}
				anchor := sp.Anchor
				ok := false
				switch anchor {
				case "left":
					ok = grid[tx-1][ty]
				case "right":
					ok = grid[tx+1][ty]
				default:
					ok = grid[tx][ty+1]
					anchor = "floor"
				}
				if !ok || r.Float64() >= sp.Chance {
					continue
				}
				minH, maxH := sp.MinH, sp.MaxH
				if maxH <= minH {
					minH, maxH = 24, 52
				}
				height := minH + r.Float64()*(maxH-minH)
				if ent := SpawnFloraAnchored(sp.Flora, tx, ty, height, anchor, r); ent != nil {
					entities = append(entities, ent)
				}
			}
		}
	}
	return entities
}

func matchesFauna(ent entity.CaveEntity, id FaunaID) bool {
	switch id {
	case FaunaLanternfish:
		_, ok := ent.(*entity.Lanternfish)
		return ok
	case FaunaInkSquid:
		_, ok := ent.(*entity.InkSquid)
		return ok
	case FaunaGlowSquid:
		_, ok := ent.(*entity.GlowSquid)
		return ok
	case FaunaElectroWeaver:
		_, ok := ent.(*entity.ElectroWeaver)
		return ok
	case FaunaThermoclineRammer:
		_, ok := ent.(*entity.ThermoclineRammer)
		return ok
	case FaunaBrimstoneSiphon:
		_, ok := ent.(*entity.BrimstoneSiphon)
		return ok
	case FaunaFalseBulbSnare:
		_, ok := ent.(*entity.FalseBulbSnare)
		return ok
	case FaunaSandViper:
		_, ok := ent.(*entity.SandViper)
		return ok
	default:
		return false
	}
}

func applyBandedSpawns(spawns []BandedSpawn, grid [][]bool, r *rand.Rand) []entity.CaveEntity {
	var entities []entity.CaveEntity
	gridW := len(grid)
	gridH := len(grid[0])

	for _, sp := range spawns {
		minTY := max(0, sp.MinTY)
		maxTY := min(gridH-1, sp.MaxTY)
		for ty := minTY; ty <= maxTY; ty++ {
			for tx := 1; tx < gridW-1; tx++ {
				if grid[tx][ty] {
					continue
				}
				isOpen := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				hasCeiling := ty > 0 && grid[tx][ty-1]
				hasFloor := ty < gridH-1 && grid[tx][ty+1]
				hasWall := (tx > 0 && grid[tx-1][ty]) || (tx < gridW-1 && grid[tx+1][ty])

				if sp.NeedOpen && !isOpen {
					continue
				}
				if sp.NeedWall && !hasWall {
					continue
				}
				if sp.NeedCeiling && !hasCeiling {
					continue
				}
				if sp.NeedFloor && !hasFloor {
					continue
				}
				if r.Float64() >= sp.Chance {
					continue
				}
				if sp.MinSpacingPX > 0 && sp.HasFauna {
					px := float64(tx * config.TileSize)
					py := float64(ty * config.TileSize)
					minDistSq := sp.MinSpacingPX * sp.MinSpacingPX
					tooClose := false
					for _, ent := range entities {
						if !matchesFauna(ent, sp.Fauna) {
							continue
						}
						dx := ent.GetPos().X - px
						dy := ent.GetPos().Y - py
						if dx*dx+dy*dy < minDistSq {
							tooClose = true
							break
						}
					}
					if tooClose {
						continue
					}
				}
				if sp.HasFauna {
					if ent := SpawnFauna(sp.Fauna, tx, ty, grid, r); ent != nil {
						entities = append(entities, ent)
					}
				}
				if sp.HasFlora {
					height := 32.0 + r.Float64()*48.0
					anchor := sp.FloraAnchor
					if anchor == "" {
						anchor = "floor"
					}
					if ent := SpawnFloraAnchored(sp.Flora, tx, ty, height, anchor, r); ent != nil {
						entities = append(entities, ent)
					}
				}
			}
		}
	}
	return entities
}

func applyCountSpawns(spawns []CountSpawn, grid [][]bool, existing []entity.CaveEntity, r *rand.Rand) []entity.CaveEntity {
	var entities []entity.CaveEntity
	gridW := len(grid)
	gridH := len(grid[0])

	for _, sp := range spawns {
		type cand struct {
			tx, ty int
		}
		var candidates []cand
		for tx := 2; tx < gridW-2; tx++ {
			for ty := 2; ty < gridH-2; ty++ {
				if ty < sp.MinTY || grid[tx][ty] {
					continue
				}
				hasWall := grid[tx-1][ty] || grid[tx+1][ty] || grid[tx][ty-1] || grid[tx][ty+1]
				isOpen := !grid[tx-1][ty] && !grid[tx+1][ty] && !grid[tx][ty-1] && !grid[tx][ty+1]
				if sp.NeedWall && !hasWall {
					continue
				}
				if sp.NeedOpen && !isOpen {
					continue
				}
				candidates = append(candidates, cand{tx, ty})
			}
		}
		if len(candidates) == 0 {
			continue
		}
		n := sp.Min
		if sp.Max > sp.Min {
			n = sp.Min + r.Intn(sp.Max-sp.Min+1)
		}
		if n > len(candidates) {
			n = len(candidates)
		}
		r.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		placed := make([]gvec.Vec2, 0, n)
		for _, c := range candidates {
			if len(placed) >= n {
				break
			}
			if sp.MinTileSpacing > 0 {
				ok := true
				for _, p := range placed {
					if math.Hypot(float64(c.tx)-p.X, float64(c.ty)-p.Y) < sp.MinTileSpacing {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}
			}
			if ent := SpawnFauna(sp.Fauna, c.tx, c.ty, grid, r); ent != nil {
				entities = append(entities, ent)
				placed = append(placed, gvec.Vec2{X: float64(c.tx), Y: float64(c.ty)})
			}
		}
	}
	_ = existing
	return entities
}

// MineralSpawnEntries converts biome mineral weights for resource generation.
func (s *CaveSpec) MineralSpawnEntries() []resource.ResourceSpawnEntry {
	if s == nil || s.Biome == nil {
		return nil
	}
	out := make([]resource.ResourceSpawnEntry, len(s.Biome.MineralSpawns))
	for i, e := range s.Biome.MineralSpawns {
		out[i] = resource.ResourceSpawnEntry{Type: e.Type, Weight: e.Weight}
	}
	return out
}
