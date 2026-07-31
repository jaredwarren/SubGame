package story

import (
	"strings"
)

// StoryManager manages the narrative database, unlocking logic, and database state.
type StoryManager struct {
	entries []*LoreEntry
}

// NewStoryManager creates a StoryManager seeded with DefaultLoreEntries.
// Each call clones the default table so unlock state does not leak across games.
func NewStoryManager() *StoryManager {
	return &StoryManager{
		entries: cloneLoreEntries(DefaultLoreEntries),
	}
}

// cloneLoreEntries deep-copies lore entries so runtime Unlocked flags are independent.
func cloneLoreEntries(src []*LoreEntry) []*LoreEntry {
	out := make([]*LoreEntry, len(src))
	for i, e := range src {
		cp := *e
		if len(e.Paragraphs) > 0 {
			cp.Paragraphs = make([]Paragraph, len(e.Paragraphs))
			copy(cp.Paragraphs, e.Paragraphs)
		}
		cp.Unlocked = false
		out[i] = &cp
	}
	return out
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

// SerializeState returns a list of IDs of unlocked entries, suitable for save data.
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
	idMap := make(map[string]bool, len(unlockedIDs))
	for _, id := range unlockedIDs {
		idMap[id] = true
	}

	for _, entry := range sm.entries {
		entry.Unlocked = idMap[entry.ID]
	}
}
