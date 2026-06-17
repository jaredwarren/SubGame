package item

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// UpgradeSolar represents the solar array base upgrade module.
type UpgradeSolar struct{}

func (u *UpgradeSolar) GetName() string       { return "Solar Array Module" }
func (u *UpgradeSolar) GetMaxStack() int      { return 1 }
func (u *UpgradeSolar) GetColor() color.Color { return color.RGBA{220, 200, 30, 255} }
func (u *UpgradeSolar) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, color.RGBA{255, 255, 255, 128}, false)
}
func (u *UpgradeSolar) GetModuleType() BaseModule { return ModuleSolar }
func (u *UpgradeSolar) GetStorageSlots() int      { return 0 }
func (u *UpgradeSolar) GetSolarRecharge() float64 { return 0.08 }

// UpgradeSolarMKII represents the tier 2 solar array base upgrade module.
type UpgradeSolarMKII struct{}

func (u *UpgradeSolarMKII) GetName() string       { return "Solar Array MKII Module" }
func (u *UpgradeSolarMKII) GetMaxStack() int      { return 1 }
func (u *UpgradeSolarMKII) GetColor() color.Color { return color.RGBA{240, 220, 50, 255} }
func (u *UpgradeSolarMKII) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 2.0, color.RGBA{255, 255, 255, 200}, false)
}
func (u *UpgradeSolarMKII) GetModuleType() BaseModule { return ModuleSolarMKII }
func (u *UpgradeSolarMKII) GetStorageSlots() int      { return 0 }
func (u *UpgradeSolarMKII) GetSolarRecharge() float64 { return 0.20 }

// UpgradeStorage represents the storage vault base upgrade module.
type UpgradeStorage struct{}

func (u *UpgradeStorage) GetName() string       { return "Storage Vault Module" }
func (u *UpgradeStorage) GetMaxStack() int      { return 1 }
func (u *UpgradeStorage) GetColor() color.Color { return color.RGBA{130, 150, 180, 255} }
func (u *UpgradeStorage) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 1.0, color.RGBA{255, 255, 255, 128}, false)
}
func (u *UpgradeStorage) GetModuleType() BaseModule { return ModuleStorage }
func (u *UpgradeStorage) GetStorageSlots() int      { return 24 }
func (u *UpgradeStorage) GetSolarRecharge() float64 { return 0.0 }

// UpgradeStorageMKII represents the tier 2 storage vault base upgrade module.
type UpgradeStorageMKII struct{}

func (u *UpgradeStorageMKII) GetName() string       { return "Storage Vault MKII Module" }
func (u *UpgradeStorageMKII) GetMaxStack() int      { return 1 }
func (u *UpgradeStorageMKII) GetColor() color.Color { return color.RGBA{150, 180, 220, 255} }
func (u *UpgradeStorageMKII) DrawIcon(screen *ebiten.Image, cx, cy, size float32) {
	if drawItemIconSprite(screen, u.GetName(), cx, cy, size) {
		return
	}
	vector.FillRect(screen, cx-size/2.0, cy-size/2.0, size, size, u.GetColor(), false)
	vector.StrokeRect(screen, cx-size/2.0, cy-size/2.0, size, size, 2.0, color.RGBA{255, 255, 255, 200}, false)
}
func (u *UpgradeStorageMKII) GetModuleType() BaseModule { return ModuleStorageMKII }
func (u *UpgradeStorageMKII) GetStorageSlots() int      { return 48 }
func (u *UpgradeStorageMKII) GetSolarRecharge() float64 { return 0.0 }
