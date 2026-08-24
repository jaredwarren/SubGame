package quest

import (
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/vehicle"
)

type mockQuestContext struct {
	inCave       bool
	trenchX      int
	trenchY      int
	distToBase   float64
	items        map[item.ItemID]int
	vehicles     map[vehicle.VehicleID]bool
	maxDepth     float64
	craftedItems map[item.ItemID]bool
}

func (m *mockQuestContext) IsPlayerInCave() bool {
	return m.inCave
}

func (m *mockQuestContext) PlayerTrenchCoords() (x, y int) {
	return m.trenchX, m.trenchY
}

func (m *mockQuestContext) PlayerDistanceToBase() float64 {
	return m.distToBase
}

func (m *mockQuestContext) CountInventoryItemID(id item.ItemID) int {
	if m.items == nil {
		return 0
	}
	return m.items[id]
}

func (m *mockQuestContext) HasVehicleInWorldID(id vehicle.VehicleID) bool {
	if m.vehicles == nil {
		return false
	}
	return m.vehicles[id]
}

func (m *mockQuestContext) MaxDepthReached() float64 {
	return m.maxDepth
}

func (m *mockQuestContext) HasCraftedItemID(id item.ItemID) bool {
	if m.craftedItems == nil {
		return false
	}
	return m.craftedItems[id]
}

func TestQuestManager_Progression(t *testing.T) {
	qm := NewQuestManager()
	if len(qm.Categories) == 0 {
		t.Fatal("expected default categories to be populated")
	}

	trainingQuest := qm.FindQuest("training_basics")
	if trainingQuest == nil {
		t.Fatal("expected training_basics quest to exist")
	}

	ctx := &mockQuestContext{
		items:        make(map[item.ItemID]int),
		vehicles:     make(map[vehicle.VehicleID]bool),
		craftedItems: make(map[item.ItemID]bool),
		distToBase:   300.0,
	}

	// 1. Initial check - nothing completed
	notifs := qm.CheckProgress(ctx)
	if len(notifs) != 0 {
		t.Fatalf("expected 0 notifications on empty state, got %d", len(notifs))
	}

	// 2. Dive into cave
	ctx.inCave = true
	notifs = qm.CheckProgress(ctx)
	if len(notifs) == 0 || !trainingQuest.GetTask("train_dive").Completed {
		t.Fatal("expected train_dive task to complete upon entering cave")
	}

	// 3. Mine 5 titanium -> partial progress notification
	ctx.items[item.IDTitanium] = 5
	notifs = qm.CheckProgress(ctx)
	if len(notifs) == 0 || trainingQuest.GetTask("train_titanium").Completed {
		t.Fatal("expected partial titanium progress notification, task should not be completed yet")
	}

	// 4. Mine 10 titanium -> task completes
	ctx.items[item.IDTitanium] = 10
	notifs = qm.CheckProgress(ctx)
	if !trainingQuest.GetTask("train_titanium").Completed {
		t.Fatal("expected train_titanium task to be completed with 10 items")
	}

	// 5. Surface and return to base
	ctx.inCave = false
	ctx.distToBase = 50.0
	notifs = qm.CheckProgress(ctx)
	if !trainingQuest.GetTask("train_return").Completed {
		t.Fatal("expected train_return task to be completed when close to base on overworld")
	}

	// 6. Craft Skiff Kit
	ctx.craftedItems[item.IDSkiffKit] = true
	notifs = qm.CheckProgress(ctx)
	if !trainingQuest.GetTask("train_skiff_craft").Completed {
		t.Fatal("expected train_skiff_craft to complete")
	}

	// 7. Deploy Skiff
	ctx.vehicles[vehicle.VehicleSkiff] = true
	notifs = qm.CheckProgress(ctx)
	if !trainingQuest.GetTask("train_skiff_deploy").Completed {
		t.Fatal("expected train_skiff_deploy to complete")
	}

	// Verify overall quest completed
	if !trainingQuest.Completed {
		t.Fatal("expected training_basics quest to be marked completed")
	}
}

func TestQuestManager_Serialization(t *testing.T) {
	qm1 := NewQuestManager()
	qm1.ToggleCategory("training") // collapse training

	trainingQuest := qm1.FindQuest("training_basics")
	trainingQuest.GetTask("train_dive").Completed = true
	trainingQuest.GetTask("train_titanium").CurrentCount = 7

	serialized := qm1.SerializeState()

	qm2 := NewQuestManager()
	qm2.DeserializeState(serialized)

	// Verify category collapsed
	for _, c := range qm2.Categories {
		if c.ID == "training" && !c.Collapsed {
			t.Fatal("expected training category to be collapsed in deserialized manager")
		}
	}

	restoredQuest := qm2.FindQuest("training_basics")
	if !restoredQuest.GetTask("train_dive").Completed {
		t.Fatal("expected train_dive task to remain completed after restore")
	}
	if restoredQuest.GetTask("train_titanium").CurrentCount != 7 {
		t.Fatalf("expected titanium count 7, got %d", restoredQuest.GetTask("train_titanium").CurrentCount)
	}
}
