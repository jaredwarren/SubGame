package scene

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jaredwarren/SubGame/internal/game/config"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

const (
	ToastMaxDuration = 180 // 3.0 seconds at 60 FPS
	ToastSlideFrames = 12  // 0.2 seconds slide-in
	ToastFadeFrames  = 30  // 0.5 seconds fade-out
	ToastMaxActive   = 5   // Max active toasts on screen
	ToastHeight      = 34.0
	ToastGap         = 8.0
	ToastMarginTop   = 20.0
	ToastMarginRight = 20.0
)

// ItemToast represents a single active item pickup popup.
type ItemToast struct {
	Item     item.Item
	Quantity int
	Timer    int // counts down from ToastMaxDuration to 0
}

// ToastManager manages active item pickup toasts, updating their lifetimes
// and rendering them in the top-right corner.
type ToastManager struct {
	toasts    []*ItemToast
	offscreen *ebiten.Image
}

// NewToastManager creates a new ToastManager.
func NewToastManager() *ToastManager {
	return &ToastManager{
		toasts:    make([]*ItemToast, 0, ToastMaxActive),
		offscreen: ebiten.NewImage(360, 48),
	}
}

// Add adds an item pickup toast or aggregates quantity if already displayed.
func (m *ToastManager) Add(it item.Item, qty int) {
	if it == nil || qty <= 0 {
		return
	}

	itemID := it.GetID()

	// Check if already in active toasts
	for i, t := range m.toasts {
		if t.Item != nil && t.Item.GetID() == itemID {
			t.Quantity += qty
			t.Timer = ToastMaxDuration // Refresh duration
			// Move to the front (top of list) so player sees recent pickup
			if i > 0 {
				m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
				m.toasts = append([]*ItemToast{t}, m.toasts...)
			}
			return
		}
	}

	// Create new toast
	newToast := &ItemToast{
		Item:     it,
		Quantity: qty,
		Timer:    ToastMaxDuration,
	}

	m.toasts = append([]*ItemToast{newToast}, m.toasts...)
	if len(m.toasts) > ToastMaxActive {
		m.toasts = m.toasts[:ToastMaxActive]
	}
}

// Update decrements timers and removes expired toasts.
func (m *ToastManager) Update() {
	active := m.toasts[:0]
	for _, t := range m.toasts {
		t.Timer--
		if t.Timer > 0 {
			active = append(active, t)
		}
	}
	m.toasts = active
}

// GetToasts returns the current slice of active toasts (useful for testing).
func (m *ToastManager) GetToasts() []*ItemToast {
	return m.toasts
}

// Clear removes all active toasts.
func (m *ToastManager) Clear() {
	m.toasts = m.toasts[:0]
}

// Draw renders the active toasts in the top-right corner of the screen.
func (m *ToastManager) Draw(screen *ebiten.Image) {
	if len(m.toasts) == 0 {
		return
	}

	for i, t := range m.toasts {
		if t.Item == nil {
			continue
		}

		qtyStr := fmt.Sprintf("+%d", t.Quantity)
		nameStr := t.Item.GetName()
		fullText := qtyStr + " " + nameStr
		textLen := len(fullText)

		toastW := float32(44 + textLen*6 + 16)
		if toastW < 200 {
			toastW = 200
		}
		if toastW > 350 {
			toastW = 350
		}
		toastH := float32(ToastHeight)

		// Target position (resting position)
		targetX := float32(config.ScreenWidth) - toastW - ToastMarginRight
		targetY := float32(ToastMarginTop) + float32(i)*(toastH+ToastGap)

		// Slide-in calculation
		age := ToastMaxDuration - t.Timer
		var slideOffset float32
		if age < ToastSlideFrames {
			progress := float32(age) / float32(ToastSlideFrames)
			// Cubic ease-out
			inv := 1.0 - progress
			ease := 1.0 - (inv * inv * inv)
			slideOffset = (1.0 - ease) * 100.0
		}
		drawX := targetX + slideOffset

		// Fade-out alpha calculation
		alpha := float32(1.0)
		if t.Timer < ToastFadeFrames {
			alpha = float32(t.Timer) / float32(ToastFadeFrames)
		}

		// Ensure offscreen buffer is large enough
		iw, ih := int(toastW), int(toastH)
		if m.offscreen == nil || m.offscreen.Bounds().Dx() < iw || m.offscreen.Bounds().Dy() < ih {
			m.offscreen = ebiten.NewImage(iw+60, ih+20)
		}

		subRect := image.Rect(0, 0, iw, ih)
		toastBuf := m.offscreen.SubImage(subRect).(*ebiten.Image)
		toastBuf.Clear()

		// Background dark panel
		vector.FillRect(toastBuf, 0, 0, toastW, toastH, color.RGBA{14, 20, 32, 225}, false)

		// Inner tech border
		vector.StrokeRect(toastBuf, 0, 0, toastW, toastH, 1.2, color.RGBA{45, 110, 165, 200}, false)

		// Glowing cyan left accent bar
		vector.FillRect(toastBuf, 0, 0, 3.5, toastH, color.RGBA{0, 230, 255, 255}, false)

		// Item icon on the left
		iconCX := float32(22)
		iconCY := toastH / 2.0
		t.Item.DrawIcon(toastBuf, iconCX, iconCY, 11)

		// Text positioning
		textX := 40
		textY := int(toastH-16)/2 + 1

		// Quantity in bright neon cyan (#00E6FF)
		drawColoredDebugText(toastBuf, qtyStr, textX, textY, color.RGBA{0, 230, 255, 255})

		// Name in crisp white (#F5F8FF)
		nameX := textX + (len(qtyStr)+1)*6
		drawColoredDebugText(toastBuf, nameStr, nameX, textY, color.RGBA{245, 248, 255, 255})

		// Blit to screen with alpha
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(drawX), float64(targetY))
		op.ColorScale.ScaleAlpha(alpha)

		screen.DrawImage(toastBuf, op)
	}
}
