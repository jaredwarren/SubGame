package entity

import (
	"image/color"
	"testing"

	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestCaveEntityMapMeta(t *testing.T) {
	cases := []struct {
		e        CaveEntity
		name     string
		color    color.RGBA
		provides bool
	}{
		{NewShatterBulb(0, 0), "ShatterBulb", mapColorOxygen, true},
		{NewSandViper(0, 0), "SandViper", mapColorPredator, false},
		{NewPassiveFish(0, 0, true, 0), "PassiveFish", mapColorPassive, false},
		{&Kelp{}, "Kelp", mapColorFlora, false},
		{NewSonicDecoy(0, 0, gvec.Vec2{}), "SonicDecoy", mapColorEffect, false},
	}
	for _, tc := range cases {
		if got := tc.e.DebugName(); got != tc.name {
			t.Errorf("%T DebugName = %q, want %q", tc.e, got, tc.name)
		}
		if got := tc.e.MapColor(); got != tc.color {
			t.Errorf("%T MapColor = %v, want %v", tc.e, got, tc.color)
		}
		if got := tc.e.ProvidesOxygen(); got != tc.provides {
			t.Errorf("%T ProvidesOxygen = %v, want %v", tc.e, got, tc.provides)
		}
	}
}
