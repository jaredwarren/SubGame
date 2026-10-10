package world

import (
	"math"

	"github.com/jaredwarren/SubGame/internal/pcghash"
	"github.com/jaredwarren/pcg/spinglass"
)

// Influence channel indices (adapter semantics — not inside spinglass).
const (
	InfluenceDanger   = 0 // abyssal pressure / open-ocean hostility
	InfluenceVitality = 1 // coastal productivity / fauna affinity
	influenceChannels = 2
)

// InfluenceScale is world tiles per coarse influence cell (keeps Minimize fast on 500² maps).
const InfluenceScale = 4

// Influence is a sidecar field from spinglass.Minimize. Geometry stays FBM/CA;
// this only paints coupled region labels + smooth channels for gameplay knobs.
type Influence struct {
	Scale  int
	Width  int // coarse grid
	Height int
	Labels [][]int8
	Danger [][]float32
	Vital  [][]float32
}

// BuildInfluenceMap runs spinglass on a downsampled grid using LandDist/WaterDist
// as biases. No-op unless pcghash.UseSpinGlass is set. Safe to call again to rebuild.
func (w *World) BuildInfluenceMap() {
	if !pcghash.UseSpinGlass {
		w.Influence = nil
		return
	}
	if w.Width <= 0 || w.Height <= 0 || w.LandDist == nil || w.WaterDist == nil {
		return
	}

	scale := InfluenceScale
	cw := (w.Width + scale - 1) / scale
	ch := (w.Height + scale - 1) / scale
	n := cw * ch
	cells := make([]spinglass.Cell, n)

	sample := func(cx, cy int) (tx, ty int, landD, waterD int, tile TileType) {
		tx = cx*scale + scale/2
		ty = cy*scale + scale/2
		if tx >= w.Width {
			tx = w.Width - 1
		}
		if ty >= w.Height {
			ty = w.Height - 1
		}
		return tx, ty, w.LandDist[tx][ty], w.WaterDist[tx][ty], w.OverworldMap[tx][ty]
	}

	// Seed unlocked labels as unset; pick two capitals after a pass.
	for i := range cells {
		cells[i].Label = -1
		cells[i].Channels = []float32{0, 0}
	}

	var coastI, abyssI = -1, -1
	bestCoast, bestAbyss := int(^uint(0)>>1), -1
	for cy := 0; cy < ch; cy++ {
		for cx := 0; cx < cw; cx++ {
			i := cy*cw + cx
			_, _, landD, _, tile := sample(cx, cy)
			if tile == TileLand {
				continue
			}
			if landD < bestCoast {
				bestCoast = landD
				coastI = i
			}
			if landD > bestAbyss {
				bestAbyss = landD
				abyssI = i
			}
		}
	}
	if coastI >= 0 {
		cells[coastI].Label = 0
		cells[coastI].Locked = true
		cells[coastI].Channels[InfluenceDanger] = 0.1
		cells[coastI].Channels[InfluenceVitality] = 0.9
	}
	if abyssI >= 0 && abyssI != coastI {
		cells[abyssI].Label = 1
		cells[abyssI].Locked = true
		cells[abyssI].Channels[InfluenceDanger] = 0.95
		cells[abyssI].Channels[InfluenceVitality] = 0.05
	}

	labelBias := func(cellIndex int, label int8) float64 {
		cx, cy := cellIndex%cw, cellIndex/cw
		_, _, landD, waterD, tile := sample(cx, cy)
		ld := float64(landD)
		wd := float64(waterD)
		switch label {
		case 0: // coastal / near-shore water
			if tile == TileLand {
				return 2
			}
			return 3.0 - ld*0.45 + math.Min(wd, 3)*0.1
		case 1: // open / abyssal
			if tile == TileLand {
				return -2
			}
			return ld*0.35 - 1.0
		case 2: // shelf / transition
			if tile == TileLand {
				return 0.5
			}
			return 2.0 - math.Abs(ld-5.0)*0.4
		default:
			return 0
		}
	}

	channelBias := func(cellIndex int, channel int) float64 {
		cx, cy := cellIndex%cw, cellIndex/cw
		_, _, landD, _, tile := sample(cx, cy)
		ld := float64(landD)
		switch channel {
		case InfluenceDanger:
			if tile == TileLand {
				return 0.05
			}
			return math.Min(1, ld/18.0)
		case InfluenceVitality:
			if tile == TileLand {
				return 0.2
			}
			return 1.0 / (1.0 + ld*0.55)
		default:
			return 0
		}
	}

	res, err := spinglass.Minimize(spinglass.Config{
		Width:          cw,
		Height:         ch,
		Seed:           uint64(w.Seed) ^ 0x51A55,
		NumLabels:      3,
		NumChannels:    influenceChannels,
		Cells:          cells,
		LabelBias:      labelBias,
		ChannelBias:    channelBias,
		NeighborWeight: 2.5,
		BorderPenalty:  0.5,
		ChannelSmooth:  0.4,
		Iterations:     35,
		Temperature0:   1.2,
		Temperature1:   0.05,
		Connectivity:   4,
	})
	if err != nil {
		w.Influence = nil
		return
	}

	inf := &Influence{
		Scale:  scale,
		Width:  cw,
		Height: ch,
		Labels: make([][]int8, cw),
		Danger: make([][]float32, cw),
		Vital:  make([][]float32, cw),
	}
	for x := 0; x < cw; x++ {
		inf.Labels[x] = make([]int8, ch)
		inf.Danger[x] = make([]float32, ch)
		inf.Vital[x] = make([]float32, ch)
		for y := 0; y < ch; y++ {
			c := res.Cells[y*cw+x]
			inf.Labels[x][y] = c.Label
			if len(c.Channels) > InfluenceDanger {
				inf.Danger[x][y] = clamp01(c.Channels[InfluenceDanger])
			}
			if len(c.Channels) > InfluenceVitality {
				inf.Vital[x][y] = clamp01(c.Channels[InfluenceVitality])
			}
		}
	}
	w.Influence = inf
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// InfluenceAt samples the sidecar at overworld tile (tx, ty).
// ok is false when Influence was not built.
func (w *World) InfluenceAt(tx, ty int) (label int8, danger, vitality float32, ok bool) {
	if w == nil || w.Influence == nil {
		return 0, 0, 0, false
	}
	inf := w.Influence
	if inf.Scale <= 0 || inf.Width <= 0 || inf.Height <= 0 {
		return 0, 0, 0, false
	}
	cx := tx / inf.Scale
	cy := ty / inf.Scale
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	if cx >= inf.Width {
		cx = inf.Width - 1
	}
	if cy >= inf.Height {
		cy = inf.Height - 1
	}
	return inf.Labels[cx][cy], inf.Danger[cx][cy], inf.Vital[cx][cy], true
}
