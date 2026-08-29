package entity

import "image/color"

// Overview / debug map marker colors by entity role.
var (
	mapColorOxygen   = color.RGBA{60, 255, 210, 255}
	mapColorPredator = color.RGBA{240, 160, 32, 255}
	mapColorPassive  = color.RGBA{136, 204, 255, 255}
	mapColorFlora    = color.RGBA{106, 223, 60, 255}
	mapColorEffect   = color.RGBA{255, 255, 255, 255}
)

// ProvidesOxygen reports whether this entity restores oxygen when interacted with.
// Default for all BaseEntity-backed types; ShatterBulb overrides to true.
func (b *BaseEntity) ProvidesOxygen() bool { return false }

func (s *ShatterBulb) DebugName() string     { return "ShatterBulb" }
func (s *ShatterBulb) MapColor() color.RGBA  { return mapColorOxygen }
func (s *ShatterBulb) ProvidesOxygen() bool  { return true }

func (e *SandViper) DebugName() string    { return "SandViper" }
func (e *SandViper) MapColor() color.RGBA { return mapColorPredator }

func (e *FalseBulbSnare) DebugName() string    { return "FalseBulbSnare" }
func (e *FalseBulbSnare) MapColor() color.RGBA { return mapColorPredator }

func (e *ThermoclineRammer) DebugName() string    { return "ThermoclineRammer" }
func (e *ThermoclineRammer) MapColor() color.RGBA { return mapColorPredator }

func (e *BrimstoneSiphon) DebugName() string    { return "BrimstoneSiphon" }
func (e *BrimstoneSiphon) MapColor() color.RGBA { return mapColorPredator }

func (e *ElectroWeaver) DebugName() string    { return "ElectroWeaver" }
func (e *ElectroWeaver) MapColor() color.RGBA { return mapColorPredator }

func (e *VoltaicLurker) DebugName() string    { return "VoltaicLurker" }
func (e *VoltaicLurker) MapColor() color.RGBA { return mapColorPredator }

func (e *PassiveFish) DebugName() string    { return "PassiveFish" }
func (e *PassiveFish) MapColor() color.RGBA { return mapColorPassive }

func (e *PassiveCrab) DebugName() string    { return "PassiveCrab" }
func (e *PassiveCrab) MapColor() color.RGBA { return mapColorPassive }

func (e *Kelp) DebugName() string    { return "Kelp" }
func (e *Kelp) MapColor() color.RGBA { return mapColorFlora }

func (e *Coral) DebugName() string    { return "Coral" }
func (e *Coral) MapColor() color.RGBA { return mapColorFlora }

func (e *ShockKelp) DebugName() string    { return "ShockKelp" }
func (e *ShockKelp) MapColor() color.RGBA { return mapColorFlora }

func (e *NerveMat) DebugName() string    { return "NerveMat" }
func (e *NerveMat) MapColor() color.RGBA { return mapColorFlora }

func (e *SonicDecoy) DebugName() string    { return "SonicDecoy" }
func (e *SonicDecoy) MapColor() color.RGBA { return mapColorEffect }

func (e *DeterrentCloud) DebugName() string    { return "DeterrentCloud" }
func (e *DeterrentCloud) MapColor() color.RGBA { return mapColorEffect }

func (e *InkSquid) DebugName() string    { return "InkSquid" }
func (e *InkSquid) MapColor() color.RGBA { return mapColorPassive }

func (e *InkCloud) DebugName() string    { return "InkCloud" }
func (e *InkCloud) MapColor() color.RGBA { return mapColorEffect }

