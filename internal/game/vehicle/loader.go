package vehicle

import (
	"log"

	"github.com/jaredwarren/SubGame/internal/assets"
)

// LoadAssets loads and processes all vehicle sheets at game startup.
func LoadAssets() {
	loadHeavyMechSheet()
	loadScoutSubSheet()
	loadSkiffSheet()
}

func loadHeavyMechSheet() {
	sheet, err := assets.LoadChromaKeyedImage("heavy_mech")
	if err != nil {
		log.Printf("Error: Failed to load heavy mech sheet: %v", err)
		return
	}
	heavyMechSheet = sheet
}

func loadScoutSubSheet() {
	sheet, err := assets.LoadChromaKeyedImage("scout_sub")
	if err != nil {
		log.Printf("Error: Failed to load scout sub sheet: %v", err)
		return
	}
	scoutSubSheet = sheet
}

func loadSkiffSheet() {
	sheet, err := assets.LoadChromaKeyedImage("skiff")
	if err != nil {
		log.Printf("Error: Failed to load skiff sheet: %v", err)
		return
	}
	skiffSheet = sheet
}
