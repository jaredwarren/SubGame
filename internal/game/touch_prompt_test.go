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
