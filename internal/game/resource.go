package game

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ResourceType defines the mineral type of a resource node.
type ResourceType int

const (
	ResourceTitanium ResourceType = iota
	ResourceCopper
	ResourceQuartz
	ResourceAbyssalOre
)

// String returns the user-facing name of the resource type.
func (r ResourceType) String() string {
	switch r {
	case ResourceTitanium:
		return "Titanium Node"
	case ResourceCopper:
		return "Copper Node"
	case ResourceQuartz:
		return "Quartz Crystal"
	case ResourceAbyssalOre:
		return "Abyssal Ore Block"
	default:
		return "Resource"
	}
}

// ItemType converts the resource type to its corresponding inventory item.
func (r ResourceType) ItemType() ItemType {
	switch r {
	case ResourceTitanium:
		return ItemTitanium
	case ResourceCopper:
		return ItemCopper
	case ResourceQuartz:
		return ItemQuartz
	case ResourceAbyssalOre:
		return ItemAbyssalOre
	default:
		return ItemNone
	}
}

// ResourceNode represents a mineable block embedded on a cave wall.
type ResourceNode struct {
	Type       ResourceType
	Tx, Ty     int // Tile coordinates
	HitsToMine int
}

// NewResourceNode creates a new ResourceNode instance.
func NewResourceNode(rType ResourceType, tx, ty int) *ResourceNode {
	return &ResourceNode{
		Type:       rType,
		Tx:         tx,
		Ty:         ty,
		HitsToMine: 3,
	}
}

// Draw renders the resource node in the cave viewport, drawing glowing crystal shapes.
func (n *ResourceNode) Draw(screen *ebiten.Image, camX, camY float64) {
	sx := float32(n.Tx*TileSize - int(camX))
	sy := float32(n.Ty*TileSize - int(camY))

	// Base backing block color
	vector.FillRect(screen, sx, sy, TileSize, TileSize, color.RGBA{25, 22, 30, 255}, false)
	vector.StrokeRect(screen, sx, sy, TileSize, TileSize, 0.5, color.RGBA{45, 40, 52, 255}, false)

	// Mineral color settings
	var mineralColor color.Color
	var coreColor color.Color
	switch n.Type {
	case ResourceTitanium:
		mineralColor = color.RGBA{160, 175, 185, 255} // Metallic silver
		coreColor = color.RGBA{220, 230, 240, 255}
	case ResourceCopper:
		mineralColor = color.RGBA{210, 110, 45, 255}  // Copper orange
		coreColor = color.RGBA{240, 160, 80, 255}
	case ResourceQuartz:
		mineralColor = color.RGBA{40, 210, 245, 200}  // Cyan bioluminescent quartz
		coreColor = color.RGBA{220, 250, 255, 255}
	case ResourceAbyssalOre:
		mineralColor = color.RGBA{140, 40, 210, 255}  // Glowing purple abyssal ore
		coreColor = color.RGBA{230, 180, 255, 255}
	}

	cx := sx + float32(TileSize)/2.0
	cy := sy + float32(TileSize)/2.0

	// Scale size based on hits left
	size := float32(14.0) * (float32(n.HitsToMine) / 3.0)
	if size < 4.0 {
		size = 4.0
	}

	// Draw raw minerals as overlapping angled rectangles (crystal facets)
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, mineralColor, false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, coreColor, false)

	// Additional crystal shard for quartz/abyssal to feel premium
	if n.Type == ResourceQuartz || n.Type == ResourceAbyssalOre {
		vector.FillRect(screen, cx-size/3.0, cy-size/1.2, size/1.5, size/1.5, coreColor, false)
	}
}

// GenerateResourceNodes scans the cave tile grid and generates mineral nodes on exposed wall surfaces.
func GenerateResourceNodes(grid [][]bool, seed int64) []ResourceNode {
	nodes := []ResourceNode{}
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

	for tx := 1; tx < gridW-1; tx++ {
		for ty := 1; ty < gridH-1; ty++ {
			// Place nodes inside solid rock walls
			if grid[tx][ty] {
				if isAdjacentToEmpty(tx, ty) {
					spawnRoll := r.Float64()
					var rType ResourceType
					var spawnChance = 0.05 // 5% chance per eligible tile face

					// Determine resource distribution based on depth
					if ty > 90 {
						// Abyssal depths: mostly Titanium/Quartz, but introduces Abyssal Ore
						roll := r.Float64()
						if roll < 0.18 {
							rType = ResourceAbyssalOre
							spawnChance = 0.035
						} else if roll < 0.45 {
							rType = ResourceQuartz
						} else if roll < 0.75 {
							rType = ResourceCopper
						} else {
							rType = ResourceTitanium
						}
					} else if ty > 45 {
						// Mid depths: Quartz and Copper common
						roll := r.Float64()
						if roll < 0.35 {
							rType = ResourceQuartz
						} else if roll < 0.70 {
							rType = ResourceCopper
						} else {
							rType = ResourceTitanium
						}
					} else {
						// Shallow depths: Titanium and Copper only
						roll := r.Float64()
						if roll < 0.28 {
							rType = ResourceCopper
						} else {
							rType = ResourceTitanium
						}
					}

					if spawnRoll < spawnChance {
						nodes = append(nodes, *NewResourceNode(rType, tx, ty))
					}
				}
			}
		}
	}

	return nodes
}
