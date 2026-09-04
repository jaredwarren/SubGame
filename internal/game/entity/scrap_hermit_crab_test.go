package entity

import (
	"math"
	"testing"

	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockCrabCtx struct {
	pPos    gvec.Vec2
	pDims   gvec.Vec2
	flashOn bool
	pFacing float64
	solids  map[[2]int]bool
}

func (m *mockCrabCtx) PlayerPos() gvec.Vec2    { return m.pPos }
func (m *mockCrabCtx) PlayerDims() gvec.Vec2   { return m.pDims }
func (m *mockCrabCtx) FlashlightOn() bool      { return m.flashOn }
func (m *mockCrabCtx) PlayerFacing() float64   { return m.pFacing }
func (m *mockCrabCtx) IsSolid(x, y, w, h float64) bool {
	// Simple floor collision: y >= 100 is solid floor
	if y+h >= 100 {
		return true
	}
	return false
}

func TestScrapHermitCrab_CatchAndHarvest(t *testing.T) {
	crabCan := NewScrapHermitCrabWithShell(50, 50, ShellTinCan)
	if _, ok := crabCan.GetHarvestedItem().(*item.RawCrab); !ok {
		t.Fatalf("expected RawCrab, got %T", crabCan.GetHarvestedItem())
	}
	if _, ok := crabCan.GetBonusHarvestItem().(*item.ScrapMetal); !ok {
		t.Fatalf("expected ScrapMetal from tin can shell, got %T", crabCan.GetBonusHarvestItem())
	}

	crabGear := NewScrapHermitCrabWithShell(50, 50, ShellCogGear)
	if _, ok := crabGear.GetBonusHarvestItem().(*item.ElectronicWaste); !ok {
		t.Fatalf("expected ElectronicWaste from cog gear shell, got %T", crabGear.GetBonusHarvestItem())
	}

	// Catch distance check
	closePos := gvec.Vec2{X: 55, Y: 55}
	if !crabCan.CanCatch(closePos) {
		t.Errorf("expected crab to be catchable at close distance")
	}

	farPos := gvec.Vec2{X: 300, Y: 300}
	if crabCan.CanCatch(farPos) {
		t.Errorf("expected crab NOT to be catchable at distance 300")
	}
}

func TestScrapHermitCrab_RetreatAndArmor(t *testing.T) {
	crab := NewScrapHermitCrabWithShell(100, 80, ShellPipeElbow)
	ctx := &mockCrabCtx{
		pPos:  gvec.Vec2{X: 300, Y: 80},
		pDims: gvec.Vec2{X: 16, Y: 24},
	}

	// Initially out of shell
	crab.update(ctx)
	if crab.InShell {
		t.Fatalf("crab should not be in shell when player is far away")
	}
	if crab.IsArmored() {
		t.Fatalf("crab should not be armored when out of shell")
	}

	// Move player within threat range (<80px)
	ctx.pPos = gvec.Vec2{X: 120, Y: 80}
	crab.update(ctx)
	if !crab.InShell {
		t.Fatalf("crab should have retreated into shell when player got close")
	}
	if !crab.IsArmored() {
		t.Fatalf("crab should be armored while tucked in shell")
	}
	if crab.Vel.X != 0 {
		t.Fatalf("crab velocity should freeze when retreating into shell, got %v", crab.Vel.X)
	}

	// Back player away; crab remains tucked until ShellTimer expires
	ctx.pPos = gvec.Vec2{X: 500, Y: 80}
	initialTimer := crab.ShellTimer
	crab.update(ctx)
	if crab.ShellTimer >= initialTimer {
		t.Fatalf("shell timer should count down once threat leaves")
	}

	// Advance past timer
	for crab.ShellTimer > 0 {
		crab.update(ctx)
	}
	if crab.InShell {
		t.Fatalf("crab should emerge from shell after timer expires")
	}
}

func TestScrapHermitCrab_FlashlightThreat(t *testing.T) {
	crab := NewScrapHermitCrabWithShell(100, 80, ShellTinCan)
	ctx := &mockCrabCtx{
		pPos:    gvec.Vec2{X: 250, Y: 80}, // outside threat range (150px away) but inside LightRange (280)
		pDims:   gvec.Vec2{X: 16, Y: 24},
		flashOn: true,
		pFacing: math.Pi, // facing left toward crab at X=100
	}

	crab.update(ctx)
	if !crab.InShell {
		t.Fatalf("flashlight beam should startle crab into shell")
	}
}
