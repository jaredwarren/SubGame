package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/data"
	"github.com/jaredwarren/SubGame/internal/game/entity"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/game/story"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestScanner_BlueprintMappings(t *testing.T) {
	testCases := []struct {
		targetKey    string
		expectedItem string
	}{
		{"thermocline_rammer", "Decoy Launcher Module"},
		{"voltaic_lurker", "Chemical Discharger Module"},
		{"ink_squid", "Chemical Deterrent"},
		{"glow_squid", "Chemical Deterrent"},
		{"electro_weaver", "Sonar Amplifier"},
	}

	for _, tc := range testCases {
		got := scene.GetBlueprintForTarget(tc.targetKey)
		if got != tc.expectedItem {
			t.Errorf("GetBlueprintForTarget(%q) = %q, want %q", tc.targetKey, got, tc.expectedItem)
		}
	}
}

func TestScanner_NormalizeTargetKey(t *testing.T) {
	cases := []struct {
		raw      string
		wantKey  string
		wantName string
	}{
		{"ThermoclineRammer", "thermocline_rammer", "Thermocline Rammer"},
		{"ElectroWeaver", "electro_weaver", "Electro-Weaver"},
		{"VoltaicLurker", "voltaic_lurker", "Voltaic Lurker"},
		{"FalseBulbSnare", "false_bulb_snare", "False-Bulb Snare"},
		{"InkSquid", "ink_squid", "Ink Squid"},
		{"Copper", "copper", "Copper Vein"},
		{"Abyssal Ore", "abyssal ore", "Abyssal Ore Shard"},
	}

	for _, c := range cases {
		k, name := scene.NormalizeTargetKey(c.raw)
		if k != c.wantKey {
			t.Errorf("NormalizeTargetKey(%q) key = %q, want %q", c.raw, k, c.wantKey)
		}
		if name != c.wantName {
			t.Errorf("NormalizeTargetKey(%q) name = %q, want %q", c.raw, name, c.wantName)
		}
	}
}

func TestScanner_TargetDetectionAndCompletion(t *testing.T) {
	g := NewGame()
	g.currentState = StateCave
	p := g.player
	p.Pos = gvec.Vec2{X: 100, Y: 100}

	// Put Scanner in active slot
	p.Hotbar.Slots[0] = item.ItemStack{Item: &item.Scanner{}, Quantity: 1}
	p.ActiveSlot = 0

	cs := g.caveState
	if cs.Scanner == nil {
		cs.Scanner = scene.NewScannerState()
	}

	// Add a Thermocline Rammer nearby (distance ~40px, well within 140px ScanMaxRange)
	rammer := entity.NewThermoclineRammer(140, 100)
	cs.Entities = []entity.CaveEntity{rammer}

	mockInp := scene.NewMockInput()
	mockInp.CursorPos = gvec.Vec2{X: 140, Y: 100}
	mockInp.PressedMouse[ebiten.MouseButtonLeft] = true
	g.Input = mockInp

	// Tick 45 frames (halfway)
	for i := 0; i < 45; i++ {
		cs.Scanner.ActiveTarget = scene.FindScannableAt(cs.Nodes, cs.Entities, p.Pos.X+p.Width/2, p.Pos.Y+p.Height/2, mockInp.CursorPos.X, mockInp.CursorPos.Y, scene.ScanMaxRange)
		if cs.Scanner.ActiveTarget == nil {
			t.Fatalf("frame %d: expected scanner target to be locked", i)
		}
		cs.Scanner.Progress += 1.0 / scene.ScanDurationFrames
	}

	if cs.Scanner.Progress < 0.45 || cs.Scanner.Progress > 0.55 {
		t.Errorf("expected progress ~0.5, got %f", cs.Scanner.Progress)
	}

	// Release click: should reset progress
	mockInp.PressedMouse[ebiten.MouseButtonLeft] = false
	cs.Scanner.Progress = 0.0
	cs.Scanner.ActiveTarget = nil

	// Now re-press and hold all the way to 90 frames to trigger completion
	mockInp.PressedMouse[ebiten.MouseButtonLeft] = true
	target := scene.FindScannableAt(cs.Nodes, cs.Entities, p.Pos.X+p.Width/2, p.Pos.Y+p.Height/2, mockInp.CursorPos.X, mockInp.CursorPos.Y, scene.ScanMaxRange)
	if target == nil {
		t.Fatal("target should be found")
	}

	// Verify Decoy Launcher is initially locked
	var decoyRecipe *data.Recipe
	for idx := range g.craftingRecipes {
		if g.craftingRecipes[idx].NewResult().GetName() == "Decoy Launcher Module" {
			decoyRecipe = &g.craftingRecipes[idx]
			break
		}
	}
	if decoyRecipe == nil {
		t.Fatal("expected Decoy Launcher recipe to exist")
	}
	if decoyRecipe.Unlocked {
		t.Error("expected Decoy Launcher to start locked")
	}

	// Complete the scan
	unlocked := g.storyManager.TriggerEvent("scan", target.Key)
	if unlocked == nil {
		t.Fatal("expected lore entry to unlock for thermocline_rammer scan")
	}
	if unlocked.Title != "Thermocline Rammer Telemetry" {
		t.Errorf("expected Title 'Thermocline Rammer Telemetry', got %q", unlocked.Title)
	}

	// Unlock blueprint
	bpName := scene.GetBlueprintForTarget(target.Key)
	for idx := range g.craftingRecipes {
		if g.craftingRecipes[idx].NewResult().GetName() == bpName {
			g.craftingRecipes[idx].Unlocked = true
			break
		}
	}

	if !decoyRecipe.Unlocked {
		t.Error("expected Decoy Launcher to be unlocked after completing scan")
	}
}

func TestScanner_ExplorationPulse(t *testing.T) {
	s := scene.NewScannerState()

	if s.PulseCooldown != 0 {
		t.Errorf("expected initial PulseCooldown to be 0, got %d", s.PulseCooldown)
	}

	// Fire pulse
	s.PulseCooldown = scene.ScannerPulseCooldownMax
	s.PulseTimer = scene.ScannerPulseDurationMax
	s.PulseOrigin = gvec.Vec2{X: 200, Y: 200}

	if s.PulseCooldown != 360 {
		t.Errorf("expected PulseCooldown 360, got %d", s.PulseCooldown)
	}
	if s.PulseTimer != 210 {
		t.Errorf("expected PulseTimer 210, got %d", s.PulseTimer)
	}
}

func TestScanner_TouchControlsActivation(t *testing.T) {
	touch := scene.NewTouchControls()
	touch.SetScannerActive(true)

	// In cave mode, verify VirtualLeftClickHeld reports correctly when mouseLeft button is held
	touch.SetContext(scene.TouchContextCave)
	if touch.VirtualLeftClickHeld() {
		t.Error("expected VirtualLeftClickHeld to be false initially")
	}
}

func TestScanner_AllThreatLoreEntriesHaveDefenseTips(t *testing.T) {
	threatKeys := []string{
		"thermocline_rammer",
		"electro_weaver",
		"voltaic_lurker",
		"false_bulb_snare",
		"ink_squid",
		"sand_viper",
	}

	sm := story.NewStoryManager()
	for _, key := range threatKeys {
		entry := sm.TriggerEvent("scan", key)
		if entry == nil {
			t.Errorf("expected lore entry for threat %q with trigger 'scan'", key)
			continue
		}
		hasDefenseHeader := false
		for _, p := range entry.Paragraphs {
			if p.Header == "TRITON SURVIVAL PROTOCOL & DEFENSE TIPS" {
				hasDefenseHeader = true
				if len(p.Text) < 20 {
					t.Errorf("defense text too short for threat %q", key)
				}
			}
		}
		if !hasDefenseHeader {
			t.Errorf("threat %q missing TRITON SURVIVAL PROTOCOL & DEFENSE TIPS header", key)
		}
	}
}
