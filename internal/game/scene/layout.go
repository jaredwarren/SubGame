package scene

import (
	"image"
)

// LayoutDescriptor holds parameters for rendering and click hit-testing of an inventory slot grid.
type LayoutDescriptor struct {
	PanelW  float64
	PanelH  float64
	Cols    int
	Rows    int
	SlotSz  float64
	Gap     float64
	StartX  float64
	StartY  float64
}

// SlotRect returns the slot boundaries in screen coordinates relative to the panel's origin.
func (d LayoutDescriptor) SlotRect(panelX, panelY float64, idx int) image.Rectangle {
	r := idx / d.Cols
	c := idx % d.Cols
	sx := panelX + d.StartX + float64(c)*(d.SlotSz+d.Gap)
	sy := panelY + d.StartY + float64(r)*(d.SlotSz+d.Gap)
	return image.Rect(int(sx), int(sy), int(sx+d.SlotSz), int(sy+d.SlotSz))
}

// InSlot checks if a coordinate is within a slot's boundaries.
func (d LayoutDescriptor) InSlot(panelX, panelY float64, idx int, mx, my int) bool {
	rect := d.SlotRect(panelX, panelY, idx)
	return mx >= rect.Min.X && mx < rect.Max.X && my >= rect.Min.Y && my < rect.Max.Y
}

// HoveredSlot returns the index of the slot containing the coordinate, or -1 if none is.
func (d LayoutDescriptor) HoveredSlot(panelX, panelY float64, numSlots int, mx, my int) int {
	for i := 0; i < numSlots; i++ {
		if d.InSlot(panelX, panelY, i, mx, my) {
			return i
		}
	}
	return -1
}

// Precalculated Layout Configurations

// SoloInventoryGridDescriptor defines the layout for the player's primary grid inventory panel (600x420).
var SoloInventoryGridDescriptor = LayoutDescriptor{
	PanelW:  600,
	PanelH:  420,
	Cols:    8,
	Rows:    3,
	SlotSz:  56,
	Gap:     10,
	StartX:  41,
	StartY:  60,
}

// SoloGearGridDescriptor defines the layout for the player's equipped gear inventory panel.
var SoloGearGridDescriptor = LayoutDescriptor{
	PanelW:  600,
	PanelH:  420,
	Cols:    4,
	Rows:    1,
	SlotSz:  56,
	Gap:     10,
	StartX:  173,
	StartY:  285,
}

// VehiclePlayerInvLayout defines the layout for the player's inventory list in the split vehicle view.
var VehiclePlayerInvLayout = LayoutDescriptor{
	PanelW:  960,
	PanelH:  360,
	Cols:    8,
	Rows:    3,
	SlotSz:  48,
	Gap:     8,
	StartX:  30,
	StartY:  60,
}

// GetVehicleCargoLayout returns the layout configuration for a vehicle's cargo based on its slot size.
func GetVehicleCargoLayout(numSlots int) LayoutDescriptor {
	cols, rows := 4, 2
	switch numSlots {
	case 24:
		cols, rows = 8, 3
	case 12:
		cols, rows = 6, 2
	}
	return LayoutDescriptor{
		PanelW:  960,
		PanelH:  360,
		Cols:    cols,
		Rows:    rows,
		SlotSz:  48,
		Gap:     8,
		StartX:  510,
		StartY:  60,
	}
}

// GetVehicleUpgradeLayout returns the layout configuration for a vehicle's upgrade slots.
func GetVehicleUpgradeLayout(numSlots int) LayoutDescriptor {
	return LayoutDescriptor{
		PanelW:  960,
		PanelH:  360,
		Cols:    numSlots,
		Rows:    1,
		SlotSz:  48,
		Gap:     8,
		StartX:  510,
		StartY:  244,
	}
}

// BaseMenuPlayerInvLayout defines the layout for the player's inventory under Base Menu Overview tab.
var BaseMenuPlayerInvLayout = LayoutDescriptor{
	PanelW:  800,
	PanelH:  500,
	Cols:    8,
	Rows:    3,
	SlotSz:  40,
	Gap:     6,
	StartX:  45,
	StartY:  140,
}

// BaseMenuInstalledModulesLayout defines the layout for the base's installed upgrade modules.
var BaseMenuInstalledModulesLayout = LayoutDescriptor{
	PanelW:  800,
	PanelH:  500,
	Cols:    4,
	Rows:    1,
	SlotSz:  40,
	Gap:     6,
	StartX:  445,
	StartY:  140,
}

// BaseVaultPlayerInvLayout defines the player's inventory layout under Base Menu Vault tab.
var BaseVaultPlayerInvLayout = LayoutDescriptor{
	PanelW:  800,
	PanelH:  500,
	Cols:    8,
	Rows:    3,
	SlotSz:  40,
	Gap:     6,
	StartX:  30,
	StartY:  110,
}

// GetBaseVaultStorageLayout returns the vault's storage slot layout configuration.
func GetBaseVaultStorageLayout(numSlots int) LayoutDescriptor {
	cols := 8
	rows := numSlots / cols
	return LayoutDescriptor{
		PanelW:  800,
		PanelH:  500,
		Cols:    cols,
		Rows:    rows,
		SlotSz:  40,
		Gap:     6,
		StartX:  430,
		StartY:  110,
	}
}
