package exploration

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jaredwarren/SubGame/internal/world"
)

// RevealRadius is the tile radius charted around the player each time they
// enter a new overworld tile. Kept smaller than half the screen (~10×6 tiles)
// so a ring of fog stays visible at the viewport edges while swimming.
const RevealRadius = 5

// FogFalloffTiles is how far (in tiles) the fog fades from clear to fully opaque
// past the explored frontier.
const FogFalloffTiles = 3.0

// FogColor is the near-black overlay used for unexplored overworld tiles.
var FogColor = [4]uint8{0x0A, 0x0E, 0x14, 0xFF}

// SavedExploration is the serialize-ready snapshot of exploration state.
type SavedExploration struct {
	Explored   string   // base64(zlib(bitset bytes))
	Visited    []string // "tx_ty:tileType"
	Discovered []string // "tx_ty" for POIs marked on chart
}

// Tracker records which overworld tiles have been charted and which dive
// sites have been visited. It has no dependency on scene/game packages.
type Tracker struct {
	width, height int
	explored      []uint64
	visited       map[string]world.TileType
	discovered    map[string]bool
	newlyRevealed []int
	overflowed    bool
	exploredCount int
	lastTX        int
	lastTY        int
	hasLast       bool
	fogDist       []float32
}

const maxNewlyRevealed = 8192

// NewTracker creates an empty exploration tracker for a w×h tile world.
func NewTracker(w, h int) *Tracker {
	bits := w * h
	words := (bits + 63) / 64
	return &Tracker{
		width:      w,
		height:     h,
		explored:   make([]uint64, words),
		visited:    make(map[string]world.TileType),
		discovered: make(map[string]bool),
		lastTX:     -1,
		lastTY:     -1,
	}
}

// Reveal marks a filled circle of tiles as explored.
// Bounds-checked; no-ops if (cx,cy) matches the last revealed center.
func (t *Tracker) Reveal(cx, cy, radius int) {
	if t.hasLast && cx == t.lastTX && cy == t.lastTY {
		return
	}
	t.lastTX = cx
	t.lastTY = cy
	t.hasLast = true

	r2 := radius * radius
	minX := cx - radius
	maxX := cx + radius
	minY := cy - radius
	maxY := cy + radius

	revealed := false
	for ty := minY; ty <= maxY; ty++ {
		if ty < 0 || ty >= t.height {
			continue
		}
		dy := ty - cy
		for tx := minX; tx <= maxX; tx++ {
			if tx < 0 || tx >= t.width {
				continue
			}
			dx := tx - cx
			if dx*dx+dy*dy > r2 {
				continue
			}
			if t.markExplored(tx, ty) {
				revealed = true
			}
		}
	}
	if revealed {
		t.refreshFogDistNear(cx, cy, radius)
	}
}

func (t *Tracker) markExplored(tx, ty int) bool {
	idx := ty*t.width + tx
	word := idx / 64
	bit := uint64(1) << uint(idx%64)
	if t.explored[word]&bit != 0 {
		return false
	}
	t.explored[word] |= bit
	t.exploredCount++
	if len(t.newlyRevealed) < maxNewlyRevealed {
		t.newlyRevealed = append(t.newlyRevealed, idx)
	} else {
		t.overflowed = true
	}
	return true
}

// IsExplored reports whether tile (tx,ty) has been charted.
// Out-of-bounds tiles are never explored.
func (t *Tracker) IsExplored(tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= t.width || ty >= t.height {
		return false
	}
	idx := ty*t.width + tx
	return t.explored[idx/64]&(uint64(1)<<uint(idx%64)) != 0
}

// ExploredFraction returns the fraction of world tiles that have been charted.
func (t *Tracker) ExploredFraction() float64 {
	total := t.width * t.height
	if total == 0 {
		return 0
	}
	return float64(t.exploredCount) / float64(total)
}

// MarkVisited records that the player has dived the site at (tx,ty).
func (t *Tracker) MarkVisited(tx, ty int, tt world.TileType) {
	if tx < 0 || ty < 0 || tx >= t.width || ty >= t.height {
		return
	}
	t.visited[fmt.Sprintf("%d_%d", tx, ty)] = tt
}

// IsVisited reports whether the player has dived the site at (tx,ty).
func (t *Tracker) IsVisited(tx, ty int) bool {
	_, ok := t.visited[fmt.Sprintf("%d_%d", tx, ty)]
	return ok
}

// VisitedTile returns the tile type recorded at dive time, if visited.
func (t *Tracker) VisitedTile(tx, ty int) (world.TileType, bool) {
	tt, ok := t.visited[fmt.Sprintf("%d_%d", tx, ty)]
	return tt, ok
}

// Drain returns and clears the list of tile indices revealed since the last Drain.
func (t *Tracker) Drain() []int {
	out := t.newlyRevealed
	t.newlyRevealed = nil
	t.overflowed = false
	return out
}

// Overflowed reports whether newly revealed updates exceeded the backlog capacity.
func (t *Tracker) Overflowed() bool {
	return t.overflowed
}

// Width returns the tracker width in tiles.
func (t *Tracker) Width() int { return t.width }

// Height returns the tracker height in tiles.
func (t *Tracker) Height() int { return t.height }

// IndexToTile converts a flat tile index to (tx, ty).
func (t *Tracker) IndexToTile(idx int) (tx, ty int) {
	return idx % t.width, idx / t.width
}

// MarkPOIDiscovered records that an overworld POI at (tx,ty) has been detected by sonar.
func (t *Tracker) MarkPOIDiscovered(tx, ty int) {
	if tx < 0 || ty < 0 || tx >= t.width || ty >= t.height {
		return
	}
	if t.discovered == nil {
		t.discovered = make(map[string]bool)
	}
	t.discovered[fmt.Sprintf("%d_%d", tx, ty)] = true
}

// IsPOIDiscovered reports whether an overworld POI at (tx,ty) has been detected by sonar.
func (t *Tracker) IsPOIDiscovered(tx, ty int) bool {
	if t.discovered == nil {
		return false
	}
	return t.discovered[fmt.Sprintf("%d_%d", tx, ty)]
}

// SerializeState encodes explored bits and visited sites for a future save system.
func (t *Tracker) SerializeState() SavedExploration {
	raw := make([]byte, len(t.explored)*8)
	for i, w := range t.explored {
		for b := 0; b < 8; b++ {
			raw[i*8+b] = byte(w >> (8 * b))
		}
	}
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	_, _ = zw.Write(raw)
	_ = zw.Close()

	visited := make([]string, 0, len(t.visited))
	for key, tt := range t.visited {
		visited = append(visited, fmt.Sprintf("%s:%d", key, int(tt)))
	}
	discovered := make([]string, 0, len(t.discovered))
	for key := range t.discovered {
		discovered = append(discovered, key)
	}
	return SavedExploration{
		Explored:   base64.StdEncoding.EncodeToString(buf.Bytes()),
		Visited:    visited,
		Discovered: discovered,
	}
}

// DeserializeState restores exploration state from a SavedExploration snapshot.
// After deserialize, Drain is cleared and last-center cache is reset so the next
// Reveal always runs. Callers that cache a map image should rebuild it.
func (t *Tracker) DeserializeState(s SavedExploration) {
	t.explored = make([]uint64, (t.width*t.height+63)/64)
	t.exploredCount = 0
	t.newlyRevealed = nil
	t.overflowed = false
	t.visited = make(map[string]world.TileType)
	t.hasLast = false
	t.lastTX, t.lastTY = -1, -1
	t.fogDist = nil

	if s.Explored != "" {
		compressed, err := base64.StdEncoding.DecodeString(s.Explored)
		if err == nil {
			zr, err := zlib.NewReader(bytes.NewReader(compressed))
			if err == nil {
				raw, err := io.ReadAll(zr)
				_ = zr.Close()
				if err == nil {
					words := len(raw) / 8
					if words > len(t.explored) {
						words = len(t.explored)
					}
					for i := 0; i < words; i++ {
						var w uint64
						for b := 0; b < 8; b++ {
							w |= uint64(raw[i*8+b]) << (8 * b)
						}
						t.explored[i] = w
						t.exploredCount += bitsSet(w)
					}
					// Clamp count for a possible partial last word.
					maxBits := t.width * t.height
					if t.exploredCount > maxBits {
						t.exploredCount = maxBits
					}
				}
			}
		}
	}

	for _, entry := range s.Visited {
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			continue
		}
		ttVal, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		var tx, ty int
		if _, err := fmt.Sscanf(parts[0], "%d_%d", &tx, &ty); err != nil {
			continue
		}
		t.MarkVisited(tx, ty, world.TileType(ttVal))
	}

	t.discovered = make(map[string]bool)
	for _, entry := range s.Discovered {
		t.discovered[entry] = true
	}
}

func bitsSet(w uint64) int {
	n := 0
	for w != 0 {
		n++
		w &= w - 1
	}
	return n
}
