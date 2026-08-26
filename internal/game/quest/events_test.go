package quest

import "testing"

func TestConditionMatchesEvent(t *testing.T) {
	inv := Condition{Kind: CondInventoryAtLeast, ItemID: "titanium"}
	if !conditionMatchesEvent(inv, ProgressEvent{Kind: EventInventory, ItemID: "titanium"}) {
		t.Fatal("inventory event should match same ItemID")
	}
	if conditionMatchesEvent(inv, ProgressEvent{Kind: EventInventory, ItemID: "copper"}) {
		t.Fatal("inventory event should not match different ItemID")
	}
	if !conditionMatchesEvent(inv, ProgressEvent{Kind: EventResync}) {
		t.Fatal("resync should match every condition")
	}
	if conditionMatchesEvent(inv, ProgressEvent{Kind: EventCaveEnter}) {
		t.Fatal("cave enter should not match inventory condition")
	}

	dive := Condition{Kind: CondInCave}
	if !conditionMatchesEvent(dive, ProgressEvent{Kind: EventCaveEnter}) {
		t.Fatal("cave enter should match CondInCave")
	}

	near := Condition{Kind: CondNearBaseOutOfCave}
	if !conditionMatchesEvent(near, ProgressEvent{Kind: EventCaveExit}) {
		t.Fatal("cave exit should match near-base")
	}
	if !conditionMatchesEvent(near, ProgressEvent{Kind: EventNearBase}) {
		t.Fatal("near-base event should match near-base condition")
	}
}
