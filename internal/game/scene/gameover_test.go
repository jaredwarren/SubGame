package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

type mockGameOverContext struct {
	input            *MockInput
	deathReason      string
	respawnCalled    bool
	titleCalled      bool
	currentState     State
}

func newMockGameOverContext() *mockGameOverContext {
	return &mockGameOverContext{
		input:        NewMockInput(),
		deathReason:  "Suffocated in cave trench",
		currentState: StateGameOver,
	}
}

func (m *mockGameOverContext) GetInput() InputSource    { return m.input }
func (m *mockGameOverContext) GetDeathReason() string   { return m.deathReason }
func (m *mockGameOverContext) Respawn()                 { m.respawnCalled = true }
func (m *mockGameOverContext) TransitionToTitle()       { m.titleCalled = true }
func (m *mockGameOverContext) SetCurrentState(s State)  { m.currentState = s }

func TestGameOverScene_CenteredLayout(t *testing.T) {
	s := NewGameOverScene()
	s.layout()

	// Panel should be centered horizontally and vertically
	expectedPanelX := float64(config.ScreenWidth-int(s.panelW)) / 2.0
	expectedPanelY := float64(config.ScreenHeight-int(s.panelH)) / 2.0

	if s.panelX != expectedPanelX || s.panelY != expectedPanelY {
		t.Fatalf("expected panel at (%f, %f), got (%f, %f)",
			expectedPanelX, expectedPanelY, s.panelX, s.panelY)
	}

	// Buttons should be inside the panel
	if s.respawnBtnX < s.panelX || s.respawnBtnX+s.respawnBtnW > s.panelX+s.panelW {
		t.Fatalf("respawn button out of panel bounds horizontally")
	}
	if s.respawnBtnY < s.panelY || s.respawnBtnY+s.respawnBtnH > s.panelY+s.panelH {
		t.Fatalf("respawn button out of panel bounds vertically")
	}

	if s.titleBtnX < s.panelX || s.titleBtnX+s.titleBtnW > s.panelX+s.panelW {
		t.Fatalf("title button out of panel bounds horizontally")
	}
	if s.titleBtnY < s.panelY || s.titleBtnY+s.titleBtnH > s.panelY+s.panelH {
		t.Fatalf("title button out of panel bounds vertically")
	}
}

func TestGameOverScene_KeyInput(t *testing.T) {
	// Test ENTER triggers Respawn
	{
		s := NewGameOverScene()
		ctx := newMockGameOverContext()
		ctx.input.JustPressedKeys[ebiten.KeyEnter] = true

		if err := s.update(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ctx.respawnCalled {
			t.Fatalf("expected ENTER to trigger Respawn")
		}
	}

	// Test SPACE triggers Respawn
	{
		s := NewGameOverScene()
		ctx := newMockGameOverContext()
		ctx.input.JustPressedKeys[ebiten.KeySpace] = true

		if err := s.update(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ctx.respawnCalled {
			t.Fatalf("expected SPACE to trigger Respawn")
		}
	}

	// Test ESCAPE triggers TransitionToTitle
	{
		s := NewGameOverScene()
		ctx := newMockGameOverContext()
		ctx.input.JustPressedKeys[ebiten.KeyEscape] = true

		if err := s.update(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ctx.titleCalled {
			t.Fatalf("expected ESCAPE to trigger TransitionToTitle")
		}
	}
}

func TestGameOverScene_MouseClicks(t *testing.T) {
	// Click Respawn Button
	{
		s := NewGameOverScene()
		ctx := newMockGameOverContext()
		ctx.input.JustPressedMouse[ebiten.MouseButtonLeft] = true
		ctx.input.CursorPos = gvec.Vec2{
			X: s.respawnBtnX + s.respawnBtnW/2.0,
			Y: s.respawnBtnY + s.respawnBtnH/2.0,
		}

		if err := s.update(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ctx.respawnCalled {
			t.Fatalf("expected mouse click on Respawn button to call Respawn")
		}
	}

	// Click Title Button
	{
		s := NewGameOverScene()
		ctx := newMockGameOverContext()
		ctx.input.JustPressedMouse[ebiten.MouseButtonLeft] = true
		ctx.input.CursorPos = gvec.Vec2{
			X: s.titleBtnX + s.titleBtnW/2.0,
			Y: s.titleBtnY + s.titleBtnH/2.0,
		}

		if err := s.update(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ctx.titleCalled {
			t.Fatalf("expected mouse click on Title button to call TransitionToTitle")
		}
	}
}

func TestGameOverScene_DrawNoPanic(t *testing.T) {
	s := NewGameOverScene()
	ctx := newMockGameOverContext()

	screen := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	s.draw(ctx, screen)
}

func TestGameOverScene_TouchHelper(t *testing.T) {
	s := NewGameOverScene()
	s.layout()

	// Inside respawn button
	if !s.inRespawnBtn(s.respawnBtnX+10, s.respawnBtnY+10) {
		t.Fatalf("expected inside respawn button")
	}
	// Outside respawn button
	if s.inRespawnBtn(s.respawnBtnX-5, s.respawnBtnY) {
		t.Fatalf("expected outside respawn button")
	}

	// Inside title button
	if !s.inTitleBtn(s.titleBtnX+10, s.titleBtnY+10) {
		t.Fatalf("expected inside title button")
	}
	// Outside title button
	if s.inTitleBtn(s.titleBtnX-5, s.titleBtnY) {
		t.Fatalf("expected outside title button")
	}
}
