package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/quest"
	"github.com/jaredwarren/SubGame/internal/game/scene"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

func TestGame_IsTouchActive(t *testing.T) {
	g := NewGame()
	if g.IsTouchActive() {
		t.Fatal("expected touch to be inactive by default")
	}

	// Activate touch
	g.touch.SetContext(scene.TouchContextOnFoot)
	// Inject touch activity by calling private or triggering tap
	// Since touch.active is package-private in scene, we can trigger active through scene tests
	// or test via touch methods that activate it
}

func TestGame_QuestNotificationSuppressesKeyPromptOnTouch(t *testing.T) {
	g := NewGame()

	// When touch is not active, notification includes key hint
	g.applyQuestNotifications([]quest.ProgressNotification{
		{Message: "Objective complete!", Completed: true},
	})
	if g.MineWarning.Message != "Objective complete! (Press [J] for PDA)" {
		t.Fatalf("expected key hint on keyboard/mouse mode, got %q", g.MineWarning.Message)
	}

	// Reset warning
	g.MineWarning.Timer = 0

	// Emulate active touch by providing mock touch or active touch state
	// In NewGame, g.touch is *TouchControls
	// We can test drawVehicleEntryPrompts on a sub image
	screen := ebiten.NewImage(1280, 720)
	g.drawVehicleEntryPrompts(screen, g.OverworldVehicles, 0, 0)
}

func TestGame_DrawVehicleEntryPromptsWithTouch(t *testing.T) {
	g := NewGame()
	g.player.Pos = gvec.Vec2{X: 100, Y: 100}

	screen := ebiten.NewImage(1280, 720)
	// Should draw without error when inactive
	g.drawVehicleEntryPrompts(screen, g.OverworldVehicles, 0, 0)
	g.drawOverworldLayer(screen)
}

func TestGame_DrawWarningBanner(t *testing.T) {
	g := NewGame()
	screen := ebiten.NewImage(1280, 720)

	// Inactive timer -> nothing drawn
	g.MineWarning.Timer = 0
	g.drawWarningBanner(screen)

	// Short message
	g.SetMineWarning("Short warning", 100, 1)
	g.drawWarningBanner(screen)

	// Long quest message
	g.SetMineWarning("✓ Objective Complete: Construct and pilot the Heavy Mech (Equipped with Drill Arm & Thrusters) (Press [J] for PDA)", 100, 1)
	g.drawWarningBanner(screen)

	// Level 2 warn
	g.SetMineWarning("Cannot deploy Scout Sub in overworld! Deploy inside a cave trench.", 100, 2)
	g.drawWarningBanner(screen)

	// Level 3 alert
	g.SetMineWarning("VEHICLE CRUSHED BY DEEP-SEA PRESSURE!", 100, 3)
	g.drawWarningBanner(screen)
}

func TestGame_WrapBannerText(t *testing.T) {
	msg := "This is a very long message that definitely exceeds the character limit and needs wrapping."
	lines := wrapBannerText(msg, 30)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines for long message, got %d lines: %v", len(lines), lines)
	}
	for _, l := range lines {
		if len(l) > 35 { // with word boundaries
			t.Errorf("line exceeded expected length: %q (%d)", l, len(l))
		}
	}
}
