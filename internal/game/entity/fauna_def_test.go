package entity

import "testing"

func TestFaunaDefRegistry(t *testing.T) {
	if len(faunaRegistry) != int(FaunaCount) {
		t.Fatalf("expected %d fauna defs, got %d", FaunaCount, len(faunaRegistry))
	}
	for id := FaunaID(0); id < FaunaCount; id++ {
		def := FaunaDefFor(id)
		if def == nil {
			t.Fatalf("missing FaunaDef for id %d", id)
		}
		if def.ID != id {
			t.Fatalf("def.ID mismatch: got %d want %d", def.ID, id)
		}
		if def.Behavior < BehaviorPassiveFish || def.Behavior > BehaviorInkSquid {
			t.Fatalf("invalid BehaviorID %d for fauna %d", def.Behavior, id)
		}
	}
}

func TestFaunaLegacyArchetypesMatchRegistry(t *testing.T) {
	cases := []struct {
		id   FaunaID
		arch **FaunaDef
	}{
		{FaunaPassiveFish, &PassiveFishArchetype},
		{FaunaPassiveCrab, &PassiveCrabArchetype},
		{FaunaSandViper, &SandViperArchetype},
		{FaunaFalseBulbSnare, &FalseBulbSnareArchetype},
		{FaunaThermoclineRammer, &ThermoclineRammerArchetype},
		{FaunaElectroWeaver, &ElectroWeaverArchetype},
		{FaunaVoltaicLurker, &VoltaicLurkerArchetype},
		{FaunaBrimstoneSiphon, &BrimstoneSiphonArchetype},
		{FaunaInkSquid, &InkSquidArchetype},
	}
	for _, tc := range cases {
		if *tc.arch != FaunaDefFor(tc.id) {
			t.Fatalf("legacy archetype pointer mismatch for fauna %d", tc.id)
		}
	}
}

func TestFalseBulbSnareUsesDefDims(t *testing.T) {
	def := FaunaDefFor(FaunaFalseBulbSnare)
	ent := NewFalseBulbSnare(10, 20)
	if ent.Dimensions != def.Dims {
		t.Fatalf("dims mismatch: got %+v want %+v", ent.Dimensions, def.Dims)
	}
}

func TestShatterBulbRestoreOxygenFromDef(t *testing.T) {
	if ShatterBulbArchetype.RestoreOxygen != 20 {
		t.Fatalf("expected RestoreOxygen 20, got %v", ShatterBulbArchetype.RestoreOxygen)
	}
}
