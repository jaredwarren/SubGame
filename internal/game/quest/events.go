package quest

import (
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// EventKind identifies what changed in the game so quests can re-evaluate
// only the tasks that care about that change.
type EventKind int

const (
	// EventResync re-evaluates every incomplete task (load / debug / tests).
	EventResync EventKind = iota
	EventCaveEnter
	EventCaveExit
	EventInventory // ItemID filters which inventory conditions apply; empty = all
	EventCrafted   // ItemID of crafted result
	EventVehicle   // VehicleID of deployed craft
	EventDepth     // current / max depth sample
	EventNearBase  // player position relative to base may have changed
)

// ProgressEvent is a game-state change that may advance quest objectives.
type ProgressEvent struct {
	Kind      EventKind
	ItemID    item.ItemID
	VehicleID vehicle.VehicleID
	Depth     float64
}

// conditionMatchesEvent reports whether cond should be re-checked for ev.
func conditionMatchesEvent(cond Condition, ev ProgressEvent) bool {
	switch ev.Kind {
	case EventResync:
		return true
	case EventCaveEnter:
		return cond.Kind == CondInCave
	case EventCaveExit:
		return cond.Kind == CondNearBaseOutOfCave
	case EventInventory:
		switch cond.Kind {
		case CondInventoryAtLeast, CondCraftedOrOwned, CondCraftedOrVehicle:
			return ev.ItemID == "" || cond.ItemID == ev.ItemID
		}
	case EventCrafted:
		switch cond.Kind {
		case CondCraftedOrOwned, CondCraftedOrVehicle:
			return ev.ItemID == "" || cond.ItemID == ev.ItemID
		}
	case EventVehicle:
		switch cond.Kind {
		case CondHasVehicle, CondCraftedOrVehicle:
			return ev.VehicleID == "" || cond.VehicleID == ev.VehicleID
		}
	case EventDepth:
		return cond.Kind == CondMaxDepth
	case EventNearBase:
		return cond.Kind == CondNearBaseOutOfCave
	}
	return false
}
