package story

import (
	"encoding/json"
	"strings"
)

// StoryManager manages the narrative database, unlocking logic, and database state.
type StoryManager struct {
	entries []*LoreEntry
}

// NewStoryManager creates and returns a new empty StoryManager.
func NewStoryManager() *StoryManager {
	return &StoryManager{
		entries: make([]*LoreEntry, 0),
	}
}

// Load parses the JSON configuration data containing the narrative entries.
func (sm *StoryManager) Load(jsonData []byte) error {
	var config struct {
		Entries []*LoreEntry `json:"entries"`
	}

	if err := json.Unmarshal(jsonData, &config); err != nil {
		return err
	}

	sm.entries = config.Entries
	return nil
}

// TriggerEvent checks if any locked story entry matches the trigger type and target.
// If a match is found, it is unlocked and returned (so a notification can be displayed).
func (sm *StoryManager) TriggerEvent(triggerType, target string) *LoreEntry {
	triggerType = strings.ToLower(strings.TrimSpace(triggerType))
	target = strings.ToLower(strings.TrimSpace(target))

	for _, entry := range sm.entries {
		if entry.Unlocked {
			continue
		}
		entryType := strings.ToLower(strings.TrimSpace(entry.TriggerType))
		entryTarget := strings.ToLower(strings.TrimSpace(entry.TriggerTarget))

		if entryType == triggerType && entryTarget == target {
			entry.Unlocked = true
			return entry
		}
	}
	return nil
}

// GetUnlockedEntries returns all entries that have been unlocked, ordered as in the config.
func (sm *StoryManager) GetUnlockedEntries() []*LoreEntry {
	var unlocked []*LoreEntry
	for _, entry := range sm.entries {
		if entry.Unlocked {
			unlocked = append(unlocked, entry)
		}
	}
	return unlocked
}

// GetEntries returns all loaded entries regardless of their unlock status.
func (sm *StoryManager) GetEntries() []*LoreEntry {
	return sm.entries
}

// SerializeState returns a list of IDs of unlocked entries, suitable for json saving.
func (sm *StoryManager) SerializeState() []string {
	var ids []string
	for _, entry := range sm.entries {
		if entry.Unlocked {
			ids = append(ids, entry.ID)
		}
	}
	return ids
}

// DeserializeState marks the entries in the list as unlocked, resetting others.
func (sm *StoryManager) DeserializeState(unlockedIDs []string) {
	idMap := make(map[string]bool)
	for _, id := range unlockedIDs {
		idMap[id] = true
	}

	for _, entry := range sm.entries {
		entry.Unlocked = idMap[entry.ID]
	}
}
