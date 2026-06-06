package vehicle

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/gvec"
)

// Vehicle defines the interface that all player-piloted vehicles must implement.
type Vehicle interface {
	Update(runtime Runtime)
	Draw(screen *ebiten.Image, camX, camY float64)
	GetPos() gvec.Vec2
	SetPos(pos gvec.Vec2)
	GetDimensions() gvec.Vec2
	GetHealth() float64
	GetMaxHealth() float64
	TakeDamage(amount float64)
	GetOxygen() float64
	GetDepthLimit() float64
	GetCargo() *item.Inventory
	GetUpgrades() *item.Inventory
	GetPerspective() string // "overworld" or "cave"
	GetName() string
	GetBattery() float64
	GetMaxBattery() float64
	RechargeBattery(amount float64)
	GetFacing() float64
	ApplyForce(force gvec.Vec2)
}

// TileSize matches the global tile size used for collision calculations.
const TileSize = 64

// emptyImage is a 1x1 white image reused by drawFilledTriangle.
var emptyImage = ebiten.NewImage(3, 3)

func init() {
	emptyImage.Fill(color.White)
}

// Pre-allocated buffers for drawing triangles — avoids per-frame heap allocation.
var (
	triangleVertices = make([]ebiten.Vertex, 3)
	triangleIndices  = []uint16{0, 1, 2}
)

// drawFilledTriangle fills a 2D triangle using Ebitengine DrawTriangles.
func drawFilledTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float32, clr color.Color) {
	r, g, b, a := clr.RGBA()
	rf := float32(r) / 0xffff
	gf := float32(g) / 0xffff
	bf := float32(b) / 0xffff
	af := float32(a) / 0xffff

	triangleVertices[0] = ebiten.Vertex{DstX: x1, DstY: y1, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af}
	triangleVertices[1] = ebiten.Vertex{DstX: x2, DstY: y2, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af}
	triangleVertices[2] = ebiten.Vertex{DstX: x3, DstY: y3, SrcX: 1, SrcY: 1, ColorR: rf, ColorG: gf, ColorB: bf, ColorA: af}

	screen.DrawTriangles(triangleVertices, triangleIndices, emptyImage, nil)
}
