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

	"github.com/jaredwarren/SubGame/internal/game/cave"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/resource"
)

const caveTilePx = 6

type namedOpt struct {
	ID   string
	Name string
}

type CaveReport struct {
	TypeName       string
	Width          int
	Height         int
	OpenTiles      int
	OpenPercent    float64
	DeadEnds       int
	MaxBFSDepth    int
	O2Count        int
	DeepO2Count    int
	HasChasm       bool
	ChasmMinTX     int
	ChasmMaxTX     int
	ChasmTriggerTY int
	Entities       []nameCount
	Resources      []nameCount
	MapSrc         string
	ImageWidth     int
	ImageHeight    int
	Note           string
	Elapsed        string
	png            []byte
}

type caveKind struct {
	ID   string
	Name string
}

var caveKinds = []caveKind{
	{ID: "shallow", Name: "Shallow Seabed"},
	{ID: "trench_shallow", Name: "Trench (shallow layer)"},
	{ID: "trench", Name: "Organic Trench (deep)"},
	{ID: "wreckage", Name: "Wreckage Corridor"},
	{ID: "void", Name: "Void"},
	{ID: "shock_kelp_shallow", Name: "Shock Kelp (shallow layer)"},
	{ID: "shock_kelp", Name: "Shock Kelp (deep)"},
	{ID: "thermo", Name: "Thermo Cave"},
}

var caveBiomes = []namedOpt{
	{ID: "shallow_reef", Name: "Shallow Coral Reef"},
	{ID: "kelp_forest", Name: "Kelp Forest"},
	{ID: "thermal_barrens", Name: "Thermal Barrens"},
	{ID: "abyssal_blue", Name: "Abyssal Trench"},
}

var biomeByID = map[string]*cave.CaveBiomeSpec{
	"shallow_reef":    cave.ShallowReefBiome,
	"kelp_forest":     cave.KelpForestBiome,
	"thermal_barrens": cave.ThermalBarrensBiome,
	"abyssal_blue":    cave.AbyssalBlueBiome,
}

var (
	caveMu    sync.Mutex
	caveCache struct {
		key    string
		report *CaveReport
	}
)

func inspectCave(kindID, biomeID string, seed int64) (*CaveReport, error) {
	key := fmt.Sprintf("%s|%s|%d", kindID, biomeID, seed)
	caveMu.Lock()
	defer caveMu.Unlock()
	if caveCache.report != nil && caveCache.key == key {
		return caveCache.report, nil
	}

	start := time.Now()
	c, err := buildCave(kindID, biomeID, seed)
	if err != nil {
		return nil, err
	}

	grid := c.GetGrid()
	contentSeed := seed
	ents := c.GenerateEntities(contentSeed)
	nodes := c.GenerateResources(contentSeed)

	report := &CaveReport{
		TypeName:    caveKindName(kindID),
		Entities:    sortedCounts(countEntities(ents)),
		Resources:   sortedCounts(countResources(nodes)),
		O2Count:     countO2(ents, false),
		DeepO2Count: countO2(ents, true),
		Elapsed:     time.Since(start).Truncate(time.Millisecond).String(),
	}

	if ch, ok := c.(cave.ChasmProvider); ok && ch.HasFloorChasm() {
		minX, maxX, triggerY := ch.GetChasmBounds()
		ts := float64(config.TileSize)
		report.HasChasm = true
		report.ChasmMinTX = int(minX / ts)
		report.ChasmMaxTX = int(maxX / ts)
		report.ChasmTriggerTY = int(triggerY / ts)
	}

	if len(grid) == 0 {
		report.Note = "This cave type has no tile grid (endless void)."
		report.Elapsed = time.Since(start).Truncate(time.Millisecond).String()
		caveCache.key = key
		caveCache.report = report
		return report, nil
	}

	gw, gh := len(grid), len(grid[0])
	open, dead, depth := gridMetrics(grid)
	report.Width = gw
	report.Height = gh
	report.OpenTiles = open
	report.DeadEnds = dead
	report.MaxBFSDepth = depth
	if total := gw * gh; total > 0 {
		report.OpenPercent = 100 * float64(open) / float64(total)
	}
	report.png = renderCavePNG(grid, biomeSpec(biomeID), ents, nodes, report)
	report.ImageWidth = gw * caveTilePx
	report.ImageHeight = gh * caveTilePx
	report.MapSrc = fmt.Sprintf("/caves/map.png?type=%s&biome=%s&seed=%d", kindID, biomeID, seed)
	report.Elapsed = time.Since(start).Truncate(time.Millisecond).String()

	caveCache.key = key
	caveCache.report = report
	return report, nil
}

func caveKindName(id string) string {
	for _, k := range caveKinds {
		if k.ID == id {
			return k.Name
		}
	}
	return id
}

func biomeSpec(id string) *cave.CaveBiomeSpec {
	if s, ok := biomeByID[id]; ok && s != nil {
		return s
	}
	return cave.DefaultShallowReefBiome
}

func buildCave(kindID, biomeID string, seed int64) (cave.Cave, error) {
	r := rand.New(rand.NewSource(seed))
	dist, left, right := 8.0, true, true
	spec := biomeSpec(biomeID)

	switch kindID {
	case "shallow":
		grid := cave.GenerateShallowSeabedGrid(r, dist, left, right)
		return cave.NewShallowSeabedCaveWithBiome(grid, spec), nil
	case "trench_shallow":
		grid := cave.GenerateTrenchShallowGrid(r, dist, left, right)
		return cave.NewTrenchShallowCave(grid), nil
	case "trench":
		grid := cave.GenerateOrganicTrenchGrid(r)
		return cave.NewOrganicTrenchCave(grid), nil
	case "wreckage":
		grid := cave.GenerateWreckageGrid(r)
		return cave.NewWreckageCorridorCave(grid, 0), nil
	case "void":
		return cave.NewVoidCave(), nil
	case "shock_kelp_shallow":
		grid := cave.GenerateShockKelpShallowGrid(r, dist, left, right)
		return cave.NewShockKelpShallowCave(grid), nil
	case "shock_kelp":
		grid := cave.GenerateShockKelpCaveGrid(r)
		return cave.NewShockKelpCave(grid), nil
	case "thermo":
		grid := cave.GenerateThermoCaveGrid(r)
		return cave.NewThermoCave(grid), nil
	default:
		return nil, fmt.Errorf("unknown cave type %q", kindID)
	}
}

func countEntities(ents []entity.CaveEntity) map[string]int {
	m := map[string]int{}
	for _, e := range ents {
		bump(m, entityKind(e))
	}
	return m
}

func countResources(nodes []resource.Resource) map[string]int {
	m := map[string]int{}
	for _, n := range nodes {
		bump(m, n.GetName())
	}
	return m
}

func countO2(ents []entity.CaveEntity, deepOnly bool) int {
	n := 0
	for _, e := range ents {
		if _, ok := e.(*entity.ShatterBulb); !ok {
			continue
		}
		ty := int(e.GetPos().Y) / config.TileSize
		if deepOnly && ty < 40 {
			continue
		}
		n++
	}
	return n
}

func entityKind(e entity.CaveEntity) string {
	switch e.(type) {
	case *entity.ShatterBulb:
		return "ShatterBulb"
	case *entity.SandViper:
		return "SandViper"
	case *entity.FalseBulbSnare:
		return "FalseBulbSnare"
	case *entity.ThermoclineRammer:
		return "ThermoclineRammer"
	case *entity.BrimstoneSiphon:
		return "BrimstoneSiphon"
	case *entity.ShockKelp:
		return "ShockKelp"
	case *entity.ElectroWeaver:
		return "ElectroWeaver"
	case *entity.VoltaicLurker:
		return "VoltaicLurker"
	case *entity.PassiveFish:
		return "PassiveFish"
	case *entity.PassiveCrab:
		return "PassiveCrab"
	case *entity.Kelp:
		return "Kelp"
	case *entity.Coral:
		return "Coral"
	case *entity.NerveMat:
		return "NerveMat"
	default:
		return fmt.Sprintf("%T", e)
	}
}

func entityColor(e entity.CaveEntity) color.RGBA {
	switch e.(type) {
	case *entity.ShatterBulb:
		return color.RGBA{60, 255, 210, 255}
	case *entity.SandViper, *entity.FalseBulbSnare, *entity.ThermoclineRammer,
		*entity.BrimstoneSiphon, *entity.ElectroWeaver, *entity.VoltaicLurker:
		return color.RGBA{240, 160, 32, 255}
	case *entity.PassiveFish, *entity.PassiveCrab:
		return color.RGBA{136, 204, 255, 255}
	case *entity.Kelp, *entity.Coral, *entity.ShockKelp, *entity.NerveMat:
		return color.RGBA{106, 223, 60, 255}
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}

func resourceColor(name string) color.RGBA {
	switch name {
	case "Titanium":
		return color.RGBA{200, 200, 208, 255}
	case "Copper":
		return color.RGBA{210, 110, 45, 255}
	case "Quartz":
		return color.RGBA{40, 210, 240, 255}
	case "Abyssal Ore":
		return color.RGBA{140, 40, 210, 255}
	case "Nickel":
		return color.RGBA{150, 165, 140, 255}
	case "Scrap Metal":
		return color.RGBA{140, 110, 95, 255}
	case "Electronic Waste":
		return color.RGBA{70, 130, 90, 255}
	default:
		return color.RGBA{0, 180, 255, 255}
	}
}

func gridMetrics(grid [][]bool) (open, deadEnds, maxDepth int) {
	w, h := len(grid), len(grid[0])
	type pos struct{ x, y int }
	openAt := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < w && y < h && !grid[x][y]
	}
	neighbors := func(x, y int) int {
		n := 0
		if openAt(x-1, y) {
			n++
		}
		if openAt(x+1, y) {
			n++
		}
		if openAt(x, y-1) {
			n++
		}
		if openAt(x, y+1) {
			n++
		}
		return n
	}

	sx, sy := w/2, 0
	found := false
	for y := 0; y < h && !found; y++ {
		for dx := 0; dx < w; dx++ {
			x := (w/2 + dx) % w
			if openAt(x, y) {
				sx, sy = x, y
				found = true
				break
			}
		}
	}

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if !grid[x][y] {
				open++
				if neighbors(x, y) == 1 {
					deadEnds++
				}
			}
		}
	}

	dist := make([][]int, w)
	for x := 0; x < w; x++ {
		dist[x] = make([]int, h)
		for y := 0; y < h; y++ {
			dist[x][y] = -1
		}
	}
	if !found {
		return open, deadEnds, 0
	}
	q := []pos{{sx, sy}}
	dist[sx][sy] = 0
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if dist[cur.x][cur.y] > maxDepth {
			maxDepth = dist[cur.x][cur.y]
		}
		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if !openAt(nx, ny) || dist[nx][ny] >= 0 {
				continue
			}
			dist[nx][ny] = dist[cur.x][cur.y] + 1
			q = append(q, pos{nx, ny})
		}
	}
	return open, deadEnds, maxDepth
}

func renderCavePNG(grid [][]bool, spec *cave.CaveBiomeSpec, ents []entity.CaveEntity, nodes []resource.Resource, report *CaveReport) []byte {
	if spec == nil {
		spec = cave.DefaultShallowReefBiome
	}
	w, h := len(grid), len(grid[0])
	img := image.NewRGBA(image.Rect(0, 0, w*caveTilePx, h*caveTilePx))
	rock := spec.CaveRockColor
	openCol := spec.CaveAmbientTint
	if openCol.A == 0 {
		openCol = color.RGBA{10, 40, 70, 255}
	}

	fillTile := func(tx, ty int, c color.RGBA) {
		x0, y0 := tx*caveTilePx, ty*caveTilePx
		for py := 0; py < caveTilePx; py++ {
			for px := 0; px < caveTilePx; px++ {
				img.SetRGBA(x0+px, y0+py, c)
			}
		}
	}

	for tx := 0; tx < w; tx++ {
		for ty := 0; ty < h; ty++ {
			if grid[tx][ty] {
				fillTile(tx, ty, rock)
			} else {
				fillTile(tx, ty, openCol)
			}
		}
	}

	if report.HasChasm {
		chasm := color.RGBA{255, 224, 96, 255}
		for tx := report.ChasmMinTX; tx < report.ChasmMaxTX && tx < w; tx++ {
			if tx < 0 {
				continue
			}
			ty := report.ChasmTriggerTY
			if ty >= 0 && ty < h {
				fillTile(tx, ty, chasm)
			}
		}
	}

	mark := func(tx, ty int, c color.RGBA) {
		if tx < 0 || ty < 0 || tx >= w || ty >= h {
			return
		}
		cx := tx*caveTilePx + caveTilePx/2
		cy := ty*caveTilePx + caveTilePx/2
		stamp(img, cx, cy, caveTilePx/3, c)
	}

	for _, n := range nodes {
		tx, ty := n.GetTilePos()
		mark(tx, ty, resourceColor(n.GetName()))
	}
	for _, e := range ents {
		tx := int(e.GetPos().X) / config.TileSize
		ty := int(e.GetPos().Y) / config.TileSize
		mark(tx, ty, entityColor(e))
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
