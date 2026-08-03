package vehicle

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Vehicle defines the interface that all player-piloted vehicles must implement.
type Vehicle interface {
	Update(runtime Runtime)
	Draw(screen *ebiten.Image, camX, camY float64)
	GetPos() gvec.Vec2
	SetPos(pos gvec.Vec2)
	GetDimensions() gvec.Vec2
	GetHealth() float64
	GetMaxHealth() float64
	TakeDamage(amount float64)
	GetOxygen() float64
	GetDepthLimit() float64
	GetCargo() *item.Inventory
	GetUpgrades() *item.Inventory
	GetPerspective() string // "overworld" or "cave"
	GetName() string
	GetBattery() float64
	GetMaxBattery() float64
	RechargeBattery(amount float64)
	GetFacing() float64
	ApplyForce(force gvec.Vec2)
	GetKit() item.Item
}

// Deployable defines the interface for items that can deploy a vehicle.
type Deployable interface {
	item.Item
	Deploy(x, y float64) Vehicle
}

// TileSize matches the global tile size used for collision calculations.
const TileSize = 64

// solidAt checks if the bounding box at pos with dimensions dims is overlapping with solid tiles.
// It uses floor division to properly map negative coordinates, and subtracts an epsilon of 0.001
// from the maximum bounds to prevent flush-boundary probing errors.
func solidAt(query func(tx, ty int) bool, pos, dims gvec.Vec2) bool {
	x1, x2, y1, y2 := gvec.TileRange(pos, dims, TileSize)

	for tx := x1; tx <= x2; tx++ {
		for ty := y1; ty <= y2; ty++ {
			if query(tx, ty) {
				return true
			}
		}
	}
	return false
}

// NewVehicleByName creates a new Vehicle instance by its display or type name.
func NewVehicleByName(name string, x, y float64) Vehicle {
	switch name {
	case "Skiff", "Skiff Boat", "Surface Skiff", "The Skiff":
		return NewSkiff(x, y)
	case "Scout Submarine", "ScoutSub", "Scout Sub":
		return NewScoutSub(x, y)
	case "Heavy Mech", "HeavyMech", "Heavy Mech Walker":
		return NewHeavyMech(x, y)
	default:
		return NewSkiff(x, y)
	}
}

func init() {
	// Vehicle kits live outside the item registry; register them for inventory save/load.
	item.RegisterItemByName("Skiff Kit", func() item.Item { return &SkiffKit{} })
	item.RegisterItemByName("Scout Sub Kit", func() item.Item { return &ScoutSubKit{} })
	item.RegisterItemByName("Heavy Mech Kit", func() item.Item { return &HeavyMechKit{} })
}
