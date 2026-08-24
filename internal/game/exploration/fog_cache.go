package exploration

import "math"

const fogDistUnset = float32(999)

// FogDistAt returns the cached distance in tile-space from (px,py) to the nearest
// explored tile. Sub-cell positions bilinearly interpolate the per-tile cache.
func (t *Tracker) FogDistAt(px, py float64) float64 {
	if t.fogDist == nil {
		t.RebuildFogCache()
	}
	tx := int(math.Floor(px))
	ty := int(math.Floor(py))
	fx := px - float64(tx)
	fy := py - float64(ty)

	d00 := t.fogDistTile(tx, ty)
	d10 := t.fogDistTile(tx+1, ty)
	d01 := t.fogDistTile(tx, ty+1)
	d11 := t.fogDistTile(tx+1, ty+1)

	top := d00*(1-fx) + d10*fx
	bot := d01*(1-fx) + d11*fx
	return top*(1-fy) + bot*fy
}

func (t *Tracker) fogDistTile(tx, ty int) float64 {
	if tx < 0 || ty < 0 || tx >= t.width || ty >= t.height {
		return float64(fogDistUnset)
	}
	if t.IsExplored(tx, ty) {
		return 0
	}
	if t.fogDist == nil {
		return float64(fogDistUnset)
	}
	return float64(t.fogDist[ty*t.width+tx])
}

// RebuildFogCache recomputes the full fog distance field (e.g. after deserialize).
func (t *Tracker) RebuildFogCache() {
	t.ensureFogDist()
	search := int(math.Ceil(FogFalloffTiles)) + 1
	for ty := 0; ty < t.height; ty++ {
		for tx := 0; tx < t.width; tx++ {
			idx := ty*t.width + tx
			if t.IsExplored(tx, ty) {
				t.fogDist[idx] = 0
				continue
			}
			t.fogDist[idx] = float32(t.localFogDist(tx, ty, search))
		}
	}
}

func (t *Tracker) ensureFogDist() {
	n := t.width * t.height
	if len(t.fogDist) == n {
		return
	}
	t.fogDist = make([]float32, n)
	for i := range t.fogDist {
		t.fogDist[i] = fogDistUnset
	}
}

func (t *Tracker) refreshFogDistNear(cx, cy int) {
	t.ensureFogDist()
	search := int(math.Ceil(FogFalloffTiles)) + 1
	for dy := -search; dy <= search; dy++ {
		for dx := -search; dx <= search; dx++ {
			tx, ty := cx+dx, cy+dy
			if tx < 0 || ty < 0 || tx >= t.width || ty >= t.height {
				continue
			}
			idx := ty*t.width + tx
			if t.IsExplored(tx, ty) {
				t.fogDist[idx] = 0
				continue
			}
			t.fogDist[idx] = float32(t.localFogDist(tx, ty, search))
		}
	}
}

func (t *Tracker) localFogDist(tx, ty, search int) float64 {
	px := float64(tx) + 0.5
	py := float64(ty) + 0.5
	best := math.MaxFloat64
	found := false

	for dy := -search; dy <= search; dy++ {
		for dx := -search; dx <= search; dx++ {
			ex, ey := tx+dx, ty+dy
			if !t.IsExplored(ex, ey) {
				continue
			}
			found = true
			cx := math.Min(math.Max(px, float64(ex)), float64(ex+1))
			cy := math.Min(math.Max(py, float64(ey)), float64(ey+1))
			d := math.Hypot(px-cx, py-cy)
			if d < best {
				best = d
			}
		}
	}
	if !found {
		return FogFalloffTiles + 1
	}
	return best
}
