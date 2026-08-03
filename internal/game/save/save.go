package save

import (
	"encoding/json"
	"os"
	"time"

	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
)

const DefaultSaveFileName = "save.json"

// SavedPlayer holds serialized player state.
type SavedPlayer struct {
	PosX       float64             `json:"posX"`
	PosY       float64             `json:"posY"`
	Facing     float64             `json:"facing"`
	Health     float64             `json:"health"`
	MaxHealth  float64             `json:"maxHealth"`
	Oxygen     float64             `json:"oxygen"`
	MaxOxygen  float64             `json:"maxOxygen"`
	Stamina    float64             `json:"stamina"`
	MaxStamina float64             `json:"maxStamina"`
	Energy     float64             `json:"energy"`
	MaxEnergy  float64             `json:"maxEnergy"`
	ActiveSlot int                 `json:"activeSlot"`
	Inventory  item.SavedInventory `json:"inventory"`
	Upgrades   item.SavedInventory `json:"upgrades"`
	Hotbar     item.SavedInventory `json:"hotbar"`
}

// SavedBaseStation holds serialized base station state.
type SavedBaseStation struct {
	PosX     float64             `json:"posX"`
	PosY     float64             `json:"posY"`
	Storage  item.SavedInventory `json:"storage"`
	Upgrades item.SavedInventory `json:"upgrades"`
}

// SavedVehicle holds serialized state for a single vehicle.
type SavedVehicle struct {
	Type       string              `json:"type"` // "Skiff", "ScoutSub", "HeavyMech"
	PosX       float64             `json:"posX"`
	PosY       float64             `json:"posY"`
	Facing     float64             `json:"facing"`
	Health     float64             `json:"health"`
	MaxHealth  float64             `json:"maxHealth"`
	Battery    float64             `json:"battery"`
	MaxBattery float64             `json:"maxBattery"`
	Cargo      item.SavedInventory `json:"cargo"`
	Upgrades   item.SavedInventory `json:"upgrades"`
	IsActive   bool                `json:"isActive"`
	Location   string              `json:"location"` // "overworld" or trenchKey
}

// SaveData is the root object serialized to JSON for game saves.
type SaveData struct {
	Version         int                         `json:"version"`
	Timestamp       int64                       `json:"timestamp"`
	WorldSeed       int64                       `json:"worldSeed"`
	TimeOfDay       float64                     `json:"timeOfDay"`
	Ticks           float64                     `json:"ticks"`
	TutorialActive  bool                        `json:"tutorialActive"`
	SceneState      string                      `json:"sceneState"` // "Overworld" or "Cave"
	LastOverworldX  float64                     `json:"lastOverworldX"`
	LastOverworldY  float64                     `json:"lastOverworldY"`
	ActiveTrenchX   int                         `json:"activeTrenchX"`
	ActiveTrenchY   int                         `json:"activeTrenchY"`
	ActiveTrenchKey string                      `json:"activeTrenchKey"`
	Player          SavedPlayer                 `json:"player"`
	BaseStation     SavedBaseStation            `json:"baseStation"`
	Vehicles        []SavedVehicle              `json:"vehicles"`
	Story           []string                    `json:"story"`
	Exploration     exploration.SavedExploration `json:"exploration"`
	UnlockedRecipes []int                       `json:"unlockedRecipes"`
}

// GetSavePath returns the default save file path.
func GetSavePath() string {
	return DefaultSaveFileName
}

// HasSaveFile checks whether the save file exists on disk.
func HasSaveFile() bool {
	_, err := os.Stat(GetSavePath())
	return err == nil
}

// SaveToFile serializes SaveData into a JSON file atomically.
func SaveToFile(filePath string, data *SaveData) error {
	if data.Version == 0 {
		data.Version = 1
	}
	if data.Timestamp == 0 {
		data.Timestamp = time.Now().Unix()
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, bytes, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, filePath)
}

// LoadFromFile reads and deserializes a SaveData JSON file.
func LoadFromFile(filePath string) (*SaveData, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var data SaveData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
