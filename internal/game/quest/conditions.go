package quest

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
	ItemName     string  // inventory / craft display name
	VehicleType  string  // HasVehicleInWorld key
	Amount       int     // inventory threshold (defaults to RequiredCount when 0)
	Distance     float64 // near-base threshold
	Depth        float64 // max depth threshold
	Prerequisite TaskID  // required sibling task
}

// taskConditions maps every default TaskID to its evaluation rule.
var taskConditions = map[TaskID]Condition{
	TaskTrainDive:        {Kind: CondInCave},
	TaskTrainTitanium:    {Kind: CondInventoryAtLeast, ItemName: "Titanium"},
	TaskTrainReturn:      {Kind: CondNearBaseOutOfCave, Distance: 120, Prerequisite: TaskTrainTitanium},
	TaskTrainSkiffCraft:  {Kind: CondCraftedOrVehicle, ItemName: "Skiff Kit", VehicleType: "skiff"},
	TaskTrainSkiffDeploy: {Kind: CondHasVehicle, VehicleType: "skiff"},
	TaskGearO2HC:         {Kind: CondCraftedOrOwned, ItemName: "High Capacity O2 Tank"},
	TaskGearFins:         {Kind: CondCraftedOrOwned, ItemName: "Propulsion Fins"},
	TaskGearScanner:      {Kind: CondCraftedOrOwned, ItemName: "Scanner Tool"},
	TaskVehScoutSub:      {Kind: CondCraftedOrVehicle, ItemName: "Scout Sub Kit", VehicleType: "scout_sub"},
	TaskVehHeavyMech:     {Kind: CondCraftedOrVehicle, ItemName: "Heavy Mech Kit", VehicleType: "heavy_mech"},
	TaskEscReachDepth:    {Kind: CondMaxDepth, Depth: 100},
	TaskEscAbyssalOre:    {Kind: CondInventoryAtLeast, ItemName: "Abyssal Ore"},
	TaskEscCraftRocket:   {Kind: CondCraftedOrOwned, ItemName: "Escape Rocket"},
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
		count := ctx.CountInventoryItem(cond.ItemName)
		if count > t.CurrentCount {
			t.CurrentCount = count
		}
		if t.CurrentCount >= need {
			t.CurrentCount = need
			t.Completed = true
		}
	case CondCraftedOrOwned:
		if ctx.HasCraftedItem(cond.ItemName) || ctx.CountInventoryItem(cond.ItemName) > 0 {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondHasVehicle:
		if ctx.HasVehicleInWorld(cond.VehicleType) {
			t.CurrentCount = 1
			t.Completed = true
		}
	case CondCraftedOrVehicle:
		if ctx.HasCraftedItem(cond.ItemName) || ctx.CountInventoryItem(cond.ItemName) > 0 || ctx.HasVehicleInWorld(cond.VehicleType) {
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
