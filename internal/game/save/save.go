package save

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jaredwarren/SubGame/internal/game/exploration"
	"github.com/jaredwarren/SubGame/internal/game/item"
	"github.com/jaredwarren/SubGame/internal/game/quest"
)

// storageBackend abstracts save persistence so desktop builds use files and
// WASM builds use the browser's localStorage (see storage_file.go / storage_js.go).
type storageBackend interface {
	read(name string) ([]byte, error)
	write(name string, data []byte) error
	remove(name string) error
	exists(name string) bool
}

const (
	DefaultSaveFileName = "save.json"
	NumSlots            = 3
)

// SlotInfo describes one save slot on disk.
type SlotInfo struct {
	Slot      int
	Occupied  bool
	Timestamp int64
	WorldSeed int64
}

type saveHeader struct {
	Timestamp int64 `json:"timestamp"`
	WorldSeed int64 `json:"worldSeed"`
}

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
	Power    float64             `json:"power"`
	MaxPower float64             `json:"maxPower"`
	Storage  item.SavedInventory `json:"storage"`
	Upgrades item.SavedInventory `json:"upgrades"`
}

// SavedDockedVehicle holds serialized state for a submersible docked in the Skiff.
type SavedDockedVehicle struct {
	BayIdx     int                 `json:"bayIdx"`
	ID         string              `json:"id"`
	Health     float64             `json:"health"`
	MaxHealth  float64             `json:"maxHealth"`
	Battery    float64             `json:"battery"`
	MaxBattery float64             `json:"maxBattery"`
	Cargo      item.SavedInventory `json:"cargo"`
	Upgrades   item.SavedInventory `json:"upgrades"`
}

// SavedVehicle holds serialized state for a single vehicle.
type SavedVehicle struct {
	ID         string               `json:"id,omitempty"`   // stable VehicleID (v2+)
	Type       string               `json:"type,omitempty"` // legacy display name (v1)
	PosX       float64              `json:"posX"`
	PosY       float64              `json:"posY"`
	Facing     float64              `json:"facing"`
	Health     float64              `json:"health"`
	MaxHealth  float64              `json:"maxHealth"`
	Battery    float64              `json:"battery"`
	MaxBattery float64              `json:"maxBattery"`
	Cargo      item.SavedInventory  `json:"cargo"`
	Upgrades   item.SavedInventory  `json:"upgrades"`
	Docked     []SavedDockedVehicle `json:"docked,omitempty"`
	IsActive   bool                 `json:"isActive"`
	Location   string               `json:"location"` // "overworld" or trenchKey
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
	Exploration          exploration.SavedExploration `json:"exploration"`
	UnlockedRecipes      []int                        `json:"unlockedRecipes"`
	UnlockedRecipeNames  []string                     `json:"unlockedRecipeNames,omitempty"`
	LostCargo            []SavedLostCargo             `json:"lostCargo,omitempty"`
	Quests               quest.QuestManagerState     `json:"quests,omitempty"`
}

// SavedLostCargo is a surface cargo beacon left on death.
type SavedLostCargo struct {
	PosX          float64             `json:"posX"`
	PosY          float64             `json:"posY"`
	LifetimeTicks int                 `json:"lifetimeTicks"`
	Cargo         item.SavedInventory `json:"cargo"`
}

// GetSlotPath returns the file path for a save slot (1–NumSlots).
func GetSlotPath(slot int) string {
	return fmt.Sprintf("save_%d.json", slot)
}

// GetSavePath returns the default save file path for slot 1.
func GetSavePath() string {
	return GetSlotPath(1)
}

// HasAnySaveFile reports whether any save slot (or legacy save.json) exists.
func HasAnySaveFile() bool {
	for slot := 1; slot <= NumSlots; slot++ {
		if slotExists(slot) {
			return true
		}
	}
	return store.exists(DefaultSaveFileName)
}

// HasSaveFile checks whether any save file exists on disk.
func HasSaveFile() bool {
	return HasAnySaveFile()
}

func slotExists(slot int) bool {
	return store.exists(GetSlotPath(slot))
}

func validSlot(slot int) bool {
	return slot >= 1 && slot <= NumSlots
}

// migrateLegacySave moves save.json into slot 1 when slot 1 is empty.
func migrateLegacySave() {
	legacyPath := DefaultSaveFileName
	if !store.exists(legacyPath) || store.exists(GetSlotPath(1)) {
		return
	}
	data, err := store.read(legacyPath)
	if err != nil {
		return
	}
	if store.write(GetSlotPath(1), data) == nil {
		_ = store.remove(legacyPath)
	}
}

// probeSlot returns metadata for one slot without loading the full save.
func probeSlot(slot int) SlotInfo {
	info := SlotInfo{Slot: slot}
	path := GetSlotPath(slot)
	if !store.exists(path) {
		return info
	}
	info.Occupied = true

	bytes, err := store.read(path)
	if err != nil {
		return info
	}
	var header saveHeader
	if err := json.Unmarshal(bytes, &header); err != nil {
		return info
	}
	info.Timestamp = header.Timestamp
	info.WorldSeed = header.WorldSeed
	return info
}

// ListSlots returns metadata for every save slot, migrating legacy saves first.
func ListSlots() []SlotInfo {
	migrateLegacySave()
	slots := make([]SlotInfo, NumSlots)
	for i := 0; i < NumSlots; i++ {
		slots[i] = probeSlot(i + 1)
	}
	return slots
}

// DeleteSlot removes the save file for a slot.
func DeleteSlot(slot int) error {
	if !validSlot(slot) {
		return fmt.Errorf("invalid save slot: %d", slot)
	}
	path := GetSlotPath(slot)
	if !store.exists(path) {
		return nil
	}
	return store.remove(path)
}

// SaveToFile serializes SaveData into a JSON file atomically.
func SaveToFile(filePath string, data *SaveData) error {
	if data.Version == 0 {
		data.Version = CurrentSaveVersion
	}
	if data.Timestamp == 0 {
		data.Timestamp = time.Now().Unix()
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return store.write(filePath, bytes)
}

// LoadFromFile reads and deserializes a SaveData JSON file.
func LoadFromFile(filePath string) (*SaveData, error) {
	bytes, err := store.read(filePath)
	if err != nil {
		return nil, err
	}
	var data SaveData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}
	if err := MigrateSaveData(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
