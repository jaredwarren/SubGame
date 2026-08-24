package quest

import (
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

// ConditionKind identifies a declarative quest progress check.
type ConditionKind int

const (
	CondInCave ConditionKind = iota
	CondInventoryAtLeast
	CondCraftedOrOwned // crafted OR inventory count > 0
	CondHasVehicle
	CondCraftedOrVehicle // crafted OR vehicle in world
	CondMaxDepth
	CondNearBaseOutOfCave // not in cave AND distance to base <= Distance
	CondPrerequisite      // another task on the same quest must be completed
)

// Condition is a data-driven objective check evaluated by CheckProgress.
type Condition struct {
	Kind         ConditionKind
	ItemID       item.ItemID     // inventory / craft check
	VehicleID    vehicle.VehicleID // HasVehicleInWorld check
	Amount       int             // inventory threshold (defaults to RequiredCount when 0)
	Distance     float64         // near-base threshold
	Depth        float64         // max depth threshold
	Prerequisite TaskID          // required sibling task
}

// taskConditions maps every default TaskID to its evaluation rule.
var taskConditions = map[TaskID]Condition{
	TaskTrainDive:        {Kind: CondInCave},
	TaskTrainTitanium:    {Kind: CondInventoryAtLeast, ItemID: item.IDTitanium},
	TaskTrainReturn:      {Kind: CondNearBaseOutOfCave, Distance: 120, Prerequisite: TaskTrainTitanium},
	TaskTrainSkiffCraft:  {Kind: CondCraftedOrVehicle, ItemID: item.IDSkiffKit, VehicleID: vehicle.VehicleSkiff},
	TaskTrainSkiffDeploy: {Kind: CondHasVehicle, VehicleID: vehicle.VehicleSkiff},
	TaskGearO2HC:         {Kind: CondCraftedOrOwned, ItemID: item.IDO2TankHC},
	TaskGearFins:         {Kind: CondCraftedOrOwned, ItemID: item.IDFins},
	TaskGearScanner:      {Kind: CondCraftedOrOwned, ItemID: item.IDScanner},
	TaskVehScoutSub:      {Kind: CondCraftedOrVehicle, ItemID: item.IDScoutSubKit, VehicleID: vehicle.VehicleScoutSub},
	TaskVehHeavyMech:     {Kind: CondCraftedOrVehicle, ItemID: item.IDHeavyMechKit, VehicleID: vehicle.VehicleHeavyMech},
	TaskEscReachDepth:    {Kind: CondMaxDepth, Depth: 100},
	TaskEscAbyssalOre:    {Kind: CondInventoryAtLeast, ItemID: item.IDAbyssalOre},
	TaskEscCraftRocket:   {Kind: CondCraftedOrOwned, ItemID: item.IDEscapeRocket},
}

func evaluateCondition(cond Condition, t *Task, q *Quest, ctx QuestContext) {
	switch cond.Kind {
	case CondInCave:
		if ctx.IsPlayerInCave() {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondInventoryAtLeast:
		need := cond.Amount
		if need <= 0 {
			need = t.RequiredCount
		}
		count := ctx.CountInventoryItemID(cond.ItemID)
		if count > t.CurrentCount {
			t.CurrentCount = count
		}
		if t.CurrentCount >= need {
			t.CurrentCount = need
			t.Completed = true
		}
	case CondCraftedOrOwned:
		if ctx.HasCraftedItemID(cond.ItemID) || ctx.CountInventoryItemID(cond.ItemID) > 0 {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondHasVehicle:
		if ctx.HasVehicleInWorldID(cond.VehicleID) {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondCraftedOrVehicle:
		if ctx.HasCraftedItemID(cond.ItemID) || ctx.CountInventoryItemID(cond.ItemID) > 0 || ctx.HasVehicleInWorldID(cond.VehicleID) {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondMaxDepth:
		if ctx.MaxDepthReached() >= cond.Depth {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondNearBaseOutOfCave:
		if cond.Prerequisite != "" {
			prereq := q.GetTask(string(cond.Prerequisite))
			if prereq == nil || !prereq.Completed {
				return
			}
		}
		if !ctx.IsPlayerInCave() && ctx.PlayerDistanceToBase() <= cond.Distance {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondPrerequisite:
		// handled only as a gate on other kinds
	}
}
