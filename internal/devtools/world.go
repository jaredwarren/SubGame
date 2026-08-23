package devtools

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"sync"
	"time"

	"github.com/jaredwarren/SubGame/internal/world"
)

type poi struct {
	Kind  string
	Biome string
	X     int
	Y     int
}

type WorldReport struct {
	Seed        int64
	Width       int
	Height      int
	LifepodX    int
	LifepodY    int
	LandTiles   int
	WaterTiles  int
	BiomeCounts []nameCount
	POIs        []poi
	Minerals    []nameCount
	MineralNote string
	MapSrc      string
	Elapsed     string
	png         []byte
}

var (
	worldMu    sync.Mutex
	worldCache struct {
		seed   int64
		report *WorldReport
	}
)

var tileKindNames = map[world.TileType]string{
	world.TileTrench:        "Trench",
	world.TileShockKelpCave: "Shock Kelp",
	world.TileThermoCave:    "Thermo",
	world.TileWreckage:      "Wreckage",
}

func inspectWorld(seed int64) *WorldReport {
	worldMu.Lock()
	defer worldMu.Unlock()
	if worldCache.report != nil && worldCache.seed == seed {
		return worldCache.report
	}

	start := time.Now()
	w := world.NewWorld(seed)
	lx, ly := w.FindLifepodSpawn()

	biomeCounts := map[string]int{}
	land, water := 0, 0
	var pois []poi
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			tt := w.OverworldMap[x][y]
			spec := world.GetBiomeInfo(w.BiomeMap[x][y])
			biomeName := string(w.BiomeMap[x][y])
			if spec != nil {
				biomeName = spec.Name
			}
			biomeCounts[biomeName]++
			if tt == world.TileLand {
				land++
				continue
			}
			water++
			if kind, ok := tileKindNames[tt]; ok {
				pois = append(pois, poi{Kind: kind, Biome: biomeName, X: x, Y: y})
			}
		}
	}

	minerals, nSpecial, nShallow := sampleWorldMinerals(w, seed)
	pngBytes := renderWorldPNG(w, lx, ly)

	report := &WorldReport{
		Seed:        seed,
		Width:       w.Width,
		Height:      w.Height,
		LifepodX:    lx,
		LifepodY:    ly,
		LandTiles:   land,
		WaterTiles:  water,
		BiomeCounts: sortedCounts(biomeCounts),
		POIs:        pois,
		Minerals:    sortedCounts(minerals),
		MineralNote: fmt.Sprintf("All %d special dive sites plus %d shallow water samples.", nSpecial, nShallow),
		MapSrc:      fmt.Sprintf("/world/map.png?seed=%d", seed),
		Elapsed:     time.Since(start).Truncate(time.Millisecond).String(),
		png:         pngBytes,
	}
	worldCache.seed = seed
	worldCache.report = report
	return report
}

func sampleWorldMinerals(w *world.World, seed int64) (map[string]int, int, int) {
	counts := map[string]int{}
	special := 0
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			tt := w.OverworldMap[x][y]
			if _, ok := tileKindNames[tt]; !ok {
				continue
			}
			special++
			tallyCaveMinerals(counts, w, x, y, tt)
		}
	}

	r := rand.New(rand.NewSource(seed + 91))
	perBiome := 2
	sampled := 0
	biomeIDs := []world.BiomeID{
		world.BiomeShallowReef, world.BiomeKelpForest,
		world.BiomeThermalBarrens, world.BiomeAbyssalBlue,
	}
	for _, bID := range biomeIDs {
		picked := 0
		attempts := 0
		for picked < perBiome && attempts < 4000 {
			attempts++
			x := r.Intn(w.Width)
			y := r.Intn(w.Height)
			if w.OverworldMap[x][y] != world.TileWater || w.BiomeMap[x][y] != bID {
				continue
			}
			tallyCaveMinerals(counts, w, x, y, world.TileWater)
			picked++
			sampled++
		}
	}
	return counts, special, sampled
}

func tallyCaveMinerals(counts map[string]int, w *world.World, tx, ty int, tt world.TileType) {
	info := world.GetTileInfo(tt)
	if info == nil {
		return
	}
	grid := w.GetCave(tx, ty)
	if info.CaveFactory != nil && grid != nil {
		c := info.CaveFactory(grid, w, tx, ty)
		if c != nil {
			for _, res := range c.GenerateResources(int64(tx*97 + ty*41)) {
				bump(counts, res.GetName())
			}
		}
	}
	if info.Subterranean != nil && info.Subterranean.DeepFactory != nil {
		deep := w.GetSubterraneanCave(tx, ty)
		if deep != nil {
			c := info.Subterranean.DeepFactory(deep, w, tx, ty)
			if c != nil {
				for _, res := range c.GenerateResources(int64(tx*97 + ty*41 + 5555)) {
					bump(counts, res.GetName())
				}
			}
		}
	}
}

func renderWorldPNG(w *world.World, lx, ly int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w.Width, w.Height))
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			img.SetRGBA(x, y, worldTileColor(w, x, y))
		}
	}
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			tt := w.OverworldMap[x][y]
			if _, ok := tileKindNames[tt]; ok {
				stamp(img, x, y, 2, worldTileColor(w, x, y))
			}
		}
	}
	stamp(img, lx, ly, 3, color.RGBA{255, 255, 255, 255})
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func worldTileColor(w *world.World, x, y int) color.RGBA {
	switch w.OverworldMap[x][y] {
	case world.TileLand:
		return color.RGBA{72, 110, 58, 255}
	case world.TileTrench:
		return color.RGBA{255, 80, 220, 255}
	case world.TileShockKelpCave:
		return color.RGBA{40, 255, 200, 255}
	case world.TileThermoCave:
		return color.RGBA{255, 120, 40, 255}
	case world.TileWreckage:
		return color.RGBA{255, 220, 60, 255}
	default:
		switch w.BiomeMap[x][y] {
		case world.BiomeKelpForest:
			return color.RGBA{20, 90, 70, 255}
		case world.BiomeThermalBarrens:
			return color.RGBA{90, 50, 40, 255}
		case world.BiomeAbyssalBlue:
			return color.RGBA{20, 25, 80, 255}
		default:
			return color.RGBA{30, 110, 160, 255}
		}
	}
}

func stamp(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	b := img.Bounds()
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			x, y := cx+dx, cy+dy
			if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
				continue
			}
			img.SetRGBA(x, y, c)
		}
	}
}
