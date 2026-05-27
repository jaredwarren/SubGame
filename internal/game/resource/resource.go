package resource

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

const TileSize = 64

// Resource defines the interface that all mineable nodes/objects must implement.
type Resource interface {
	item.Item
	GetTilePos() (int, int)
	GetHitsToMine() int
	SetHitsToMine(hits int)
	RequiresMech() bool
	Draw(screen *ebiten.Image, camX, camY float64)
}

// BaseResourceNode holds the shared state for all resource node types.
type BaseResourceNode struct {
	Tx, Ty     int // Tile coordinates
	HitsToMine int
}

func (b *BaseResourceNode) GetTilePos() (int, int) {
	return b.Tx, b.Ty
}

func (b *BaseResourceNode) GetHitsToMine() int {
	return b.HitsToMine
}

func (b *BaseResourceNode) SetHitsToMine(hits int) {
	b.HitsToMine = hits
}

// drawNodeBase renders the shared backing block behind all resource nodes.
func drawNodeBase(screen *ebiten.Image, tx, ty int, camX, camY float64) (float32, float32) {
	sx := float32(tx*TileSize - int(camX))
	sy := float32(ty*TileSize - int(camY))
	vector.FillRect(screen, sx, sy, TileSize, TileSize, color.RGBA{25, 22, 30, 255}, false)
	vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, color.RGBA{45, 40, 52, 255}, false)
	return sx, sy
}

// drawMineral renders the mineral crystal at the center of a node.
func drawMineral(screen *ebiten.Image, sx, sy float32, hitsToMine int, mineralColor, coreColor color.Color, hasExtraShard bool) {
	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0

	// Scale size based on hits left
	size := float32(14.0) * (float32(hitsToMine) / 3.0)
	if size < 4.0 {
		size = 4.0
	}

	// Draw raw minerals as overlapping angled rectangles (crystal facets)
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, mineralColor, false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, coreColor, false)

	// Additional crystal shard for premium minerals
	if hasExtraShard {
		vector.FillRect(screen, cx-size/3.0, cy-size/1.2, size/1.5, size/1.5, coreColor, false)
	}
}

// ---------------------------------------------------------
// TitaniumNode
// ---------------------------------------------------------

type TitaniumNode struct {
	BaseResourceNode
}

func (n *TitaniumNode) GetName() string    { return "Titanium" }
func (n *TitaniumNode) GetMaxStack() int   { return 10 }
func (n *TitaniumNode) RequiresMech() bool { return false }

func (n *TitaniumNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{160, 175, 185, 255} // Metallic silver
	coreColor := color.RGBA{220, 230, 240, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, false)
}

// ---------------------------------------------------------
// CopperNode
// ---------------------------------------------------------

type CopperNode struct {
	BaseResourceNode
}

func (n *CopperNode) GetName() string    { return "Copper" }
func (n *CopperNode) GetMaxStack() int   { return 10 }
func (n *CopperNode) RequiresMech() bool { return false }

func (n *CopperNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{210, 110, 45, 255} // Copper orange
	coreColor := color.RGBA{240, 160, 80, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, false)
}

// ---------------------------------------------------------
// QuartzNode
// ---------------------------------------------------------

type QuartzNode struct {
	BaseResourceNode
}

func (n *QuartzNode) GetName() string    { return "Quartz" }
func (n *QuartzNode) GetMaxStack() int   { return 10 }
func (n *QuartzNode) RequiresMech() bool { return false }

func (n *QuartzNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{40, 210, 245, 200} // Cyan bioluminescent quartz
	coreColor := color.RGBA{220, 250, 255, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, true)
}

// ---------------------------------------------------------
// AbyssalOreNode
// ---------------------------------------------------------

type AbyssalOreNode struct {
	BaseResourceNode
}

func (n *AbyssalOreNode) GetName() string    { return "Abyssal Ore" }
func (n *AbyssalOreNode) GetMaxStack() int   { return 10 }
func (n *AbyssalOreNode) RequiresMech() bool { return true }

func (n *AbyssalOreNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx, sy := drawNodeBase(screen, n.Tx, n.Ty, camX, camY)
	mineralColor := color.RGBA{140, 40, 210, 255} // Glowing purple abyssal ore
	coreColor := color.RGBA{230, 180, 255, 255}
	drawMineral(screen, sx, sy, n.HitsToMine, mineralColor, coreColor, true)
}

// ---------------------------------------------------------
// Constructor helpers
// ---------------------------------------------------------

// NewTitaniumNode creates a new titanium resource node.
func NewTitaniumNode(tx, ty int) *TitaniumNode {
	return &TitaniumNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// NewCopperNode creates a new copper resource node.
func NewCopperNode(tx, ty int) *CopperNode {
	return &CopperNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// NewQuartzNode creates a new quartz resource node.
func NewQuartzNode(tx, ty int) *QuartzNode {
	return &QuartzNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// NewAbyssalOreNode creates a new abyssal ore resource node.
func NewAbyssalOreNode(tx, ty int) *AbyssalOreNode {
	return &AbyssalOreNode{BaseResourceNode{Tx: tx, Ty: ty, HitsToMine: 3}}
}

// ---------------------------------------------------------
// World Generation
// ---------------------------------------------------------

// GenerateResourceNodes scans the cave tile grid and generates mineral nodes on exposed wall surfaces.
func GenerateResourceNodes(grid [][]bool, seed int64) []Resource {
	nodes := []Resource{}
	if grid == nil {
		return nodes
	}
	gridW := len(grid)
	gridH := len(grid[0])

	r := rand.New(rand.NewSource(seed))

	// Helper to check if a tile is adjacent to empty water (open path)
	isAdjacentToEmpty := func(tx, ty int) bool {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := tx+dx, ty+dy
				if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH {
					if !grid[nx][ny] {
						return true
					}
				}
			}
		}
		return false
	}

	// newNode creates the appropriate concrete resource node for a given depth.
	type nodeKind int
	const (
		kindTitanium nodeKind = iota
		kindCopper
		kindQuartz
		kindAbyssalOre
	)

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			// Place nodes inside solid rock walls
			if grid[tx][ty] {
				if isAdjacentToEmpty(tx, ty) {
					spawnRoll := r.Float64()
					var kind nodeKind
					var spawnChance = 0.05 // 5% chance per eligible tile face

					// Determine resource distribution based on depth
					if ty > 90 {
						// Abyssal depths: mostly Titanium/Quartz, but introduces Abyssal Ore
						roll := r.Float64()
						if roll < 0.18 {
							kind = kindAbyssalOre
							spawnChance = 0.035
						} else if roll < 0.45 {
							kind = kindQuartz
						} else if roll < 0.75 {
							kind = kindCopper
						} else {
							kind = kindTitanium
						}
					} else if ty > 45 {
						// Mid depths: Quartz and Copper common
						roll := r.Float64()
						if roll < 0.35 {
							kind = kindQuartz
						} else if roll < 0.70 {
							kind = kindCopper
						} else {
							kind = kindTitanium
						}
					} else {
						// Shallow depths: Titanium and Copper only
						roll := r.Float64()
						if roll < 0.28 {
							kind = kindCopper
						} else {
							kind = kindTitanium
						}
					}

					if spawnRoll < spawnChance {
						switch kind {
						case kindTitanium:
							nodes = append(nodes, NewTitaniumNode(tx, ty))
						case kindCopper:
							nodes = append(nodes, NewCopperNode(tx, ty))
						case kindQuartz:
							nodes = append(nodes, NewQuartzNode(tx, ty))
						case kindAbyssalOre:
							nodes = append(nodes, NewAbyssalOreNode(tx, ty))
						}
					}
				}
			}
		}
	}

	return nodes
}
