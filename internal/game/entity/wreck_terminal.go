package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/audio"
	"github.com/jaredwarren/SubGame/internal/game/camera"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// WreckTerminalContext defines the interface WreckTerminal needs from runtime.
type WreckTerminalContext interface {
	PlayerPos() gvec.Vec2
	PlayerDims() gvec.Vec2
}

// WreckTerminal is an interactive wall console found inside shipwrecks.
type WreckTerminal struct {
	BaseEntity
	ShipIndex      int
	LoreID         string
	Title          string
	IsRead         bool
	AnimTick       int
	InteractPrompt bool
}

// NewWreckTerminal creates a wall terminal configured for the given shipwreck tier.
func NewWreckTerminal(x, y float64, shipIndex int) *WreckTerminal {
	var loreID string
	var title string
	switch shipIndex {
	case 1:
		loreID = "wreck_transport_manifest"
		title = "Transport Cargo Manifest"
	case 2:
		loreID = "wreck_flagship_blackbox"
		title = "Flagship Black Box Telemetry"
	default:
		loreID = "wreck_research_log"
		title = "Triton-01 Science Telemetry"
	}

	return &WreckTerminal{
		BaseEntity: BaseEntity{
			Pos:        gvec.Vec2{X: x, Y: y},
			Dimensions: gvec.Vec2{X: 24, Y: 20},
			Active:     true,
		},
		ShipIndex: shipIndex,
		LoreID:    loreID,
		Title:     title,
	}
}

// CanInteract reports whether the player is close enough to interact with the terminal.
func (t *WreckTerminal) CanInteract(playerPos gvec.Vec2) bool {
	cx := t.Pos.X + t.Dimensions.X/2
	cy := t.Pos.Y + t.Dimensions.Y/2
	return math.Hypot(playerPos.X-cx, playerPos.Y-cy) <= 48.0
}

// Interact triggers terminal playback, audio, and unlocks lore.
func (t *WreckTerminal) Interact() (string, string) {
	if !t.IsRead {
		t.IsRead = true
		audio.Get().PlaySFX("sfx/terminal_read.wav")
	} else {
		audio.Get().PlaySFX("sfx/ui_confirm.wav")
	}
	return t.LoreID, t.Title
}

func (t *WreckTerminal) Update(gr Runtime) {
	t.AnimTick++
	if ctx, ok := gr.(WreckTerminalContext); ok {
		px := ctx.PlayerPos().X + ctx.PlayerDims().X/2
		py := ctx.PlayerPos().Y + ctx.PlayerDims().Y/2
		t.InteractPrompt = t.CanInteract(gvec.Vec2{X: px, Y: py})
	}
}

func (t *WreckTerminal) Draw(screen *ebiten.Image, camera *camera.Camera, timeOfDay float64) {
	sx := float32(t.Pos.X - camera.Pos.X)
	sy := float32(t.Pos.Y - camera.Pos.Y)
	w := float32(t.Dimensions.X)
	h := float32(t.Dimensions.Y)

	// Outer chassis / mounting bracket
	chassisColor := color.RGBA{45, 50, 60, 255}
	bracketColor := color.RGBA{70, 78, 92, 255}
	vector.FillRect(screen, sx, sy, w, h, chassisColor, false)
	vector.StrokeRect(screen, sx, sy, w, h, 1.2, bracketColor, false)

	// Mounting corner bolts
	boltColor := color.RGBA{130, 140, 155, 255}
	vector.FillCircle(screen, sx+2, sy+2, 1.0, boltColor, false)
	vector.FillCircle(screen, sx+w-2, sy+2, 1.0, boltColor, false)
	vector.FillCircle(screen, sx+2, sy+h-2, 1.0, boltColor, false)
	vector.FillCircle(screen, sx+w-2, sy+h-2, 1.0, boltColor, false)

	// CRT Screen Bezel
	scrX := sx + 3.0
	scrY := sy + 3.0
	scrW := w - 6.0
	scrH := h - 8.0

	var screenBg color.RGBA
	var phosphorColor color.RGBA
	var scanlineColor color.RGBA

	switch t.ShipIndex {
	case 1: // Transport: Amber diagnostics
		screenBg = color.RGBA{28, 20, 10, 255}
		phosphorColor = color.RGBA{240, 175, 45, 255}
		scanlineColor = color.RGBA{45, 32, 15, 255}
	case 2: // Flagship: Crimson alert
		screenBg = color.RGBA{32, 10, 10, 255}
		phosphorColor = color.RGBA{245, 60, 50, 255}
		scanlineColor = color.RGBA{50, 16, 16, 255}
	default: // Research: Cyan/green radar
		screenBg = color.RGBA{10, 26, 30, 255}
		phosphorColor = color.RGBA{50, 230, 200, 255}
		scanlineColor = color.RGBA{16, 42, 48, 255}
	}

	vector.FillRect(screen, scrX, scrY, scrW, scrH, screenBg, false)

	// Scanlines
	for y := scrY; y < scrY+scrH; y += 2.0 {
		vector.StrokeLine(screen, scrX, y, scrX+scrW, y, 0.8, scanlineColor, false)
	}

	// Dynamic animated screen contents
	cx := scrX + scrW/2.0
	cy := scrY + scrH/2.0

	switch t.ShipIndex {
	case 1:
		// Amber horizontal diagnostic bar graph
		barW := float32(math.Sin(float64(t.AnimTick)*0.08))*3.0 + (scrW - 4.0)*0.6
		vector.FillRect(screen, scrX+2.0, cy-2.0, barW, 2.0, phosphorColor, false)
		vector.FillRect(screen, scrX+2.0, cy+1.5, barW*0.75, 1.5, phosphorColor, false)
	case 2:
		// Pulsing red alert diamond / lock symbol
		pulseAlpha := float32(math.Sin(float64(t.AnimTick)*0.12))*0.4 + 0.6
		alertColor := color.RGBA{phosphorColor.R, phosphorColor.G, phosphorColor.B, uint8(pulseAlpha * 255)}
		vector.StrokeLine(screen, cx, cy-3.0, cx+3.5, cy, 1.2, alertColor, false)
		vector.StrokeLine(screen, cx+3.5, cy, cx, cy+3.0, 1.2, alertColor, false)
		vector.StrokeLine(screen, cx, cy+3.0, cx-3.5, cy, 1.2, alertColor, false)
		vector.StrokeLine(screen, cx-3.5, cy, cx, cy-3.0, 1.2, alertColor, false)
	default:
		// Cyan rotating radar sweep line
		sweepAng := float64(t.AnimTick) * 0.05
		rx := cx + float32(math.Cos(sweepAng))*4.5
		ry := cy + float32(math.Sin(sweepAng))*3.5
		vector.StrokeLine(screen, cx, cy, rx, ry, 1.0, phosphorColor, false)
		vector.StrokeCircle(screen, cx, cy, 4.0, 0.8, color.RGBA{phosphorColor.R, phosphorColor.G, phosphorColor.B, 120}, false)
	}

	// Status indicator LED at bottom right
	ledColor := color.RGBA{40, 220, 60, 255}
	if t.ShipIndex == 2 {
		ledColor = color.RGBA{240, 45, 30, 255}
	} else if t.IsRead {
		ledColor = color.RGBA{80, 180, 240, 255}
	}
	vector.FillCircle(screen, sx+w-4.0, sy+h-3.0, 1.2, ledColor, false)

	// Floating prompt if player is nearby
	if t.InteractPrompt {
		prompt := "[E] Read Terminal"
		if t.IsRead {
			prompt = "[E] Review Log"
		}
		promptW := float32(len(prompt)*6 + 10)
		promptX := cx - promptW/2.0
		promptY := sy - 14.0
		vector.FillRect(screen, promptX, promptY, promptW, 11.0, color.RGBA{14, 18, 25, 210}, false)
		vector.StrokeRect(screen, promptX, promptY, promptW, 11.0, 1.0, phosphorColor, false)
		ebitenutil.DebugPrintAt(screen, prompt, int(promptX)+4, int(promptY)-1)
	}
}
