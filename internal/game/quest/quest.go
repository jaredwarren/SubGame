package quest

import (
	"fmt"
	"strings"
)

// Task represents a single objective within a quest.
type Task struct {
	ID            string `json:"id"`
	Description   string `json:"description"`
	CurrentCount  int    `json:"current_count"`
	RequiredCount int    `json:"required_count"`
	Completed     bool   `json:"completed"`
}

// ProgressText returns a formatted progress label (e.g., "[✓] Submerge into a Trench" or "[ ] Harvest Titanium (4/10)").
func (t *Task) ProgressText() string {
	if t.Completed {
		if t.RequiredCount > 1 {
			return fmt.Sprintf("[✓] %s (%d/%d)", t.Description, t.RequiredCount, t.RequiredCount)
		}
		return fmt.Sprintf("[✓] %s", t.Description)
	}

	if t.RequiredCount > 1 {
		return fmt.Sprintf("[ ] %s (%d/%d)", t.Description, t.CurrentCount, t.RequiredCount)
	}
	return fmt.Sprintf("[ ] %s", t.Description)
}

// Quest represents a group of related objectives.
type Quest struct {
	ID          string  `json:"id"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Tasks       []*Task `json:"tasks"`
	Completed   bool    `json:"completed"`
}

// CompletedCount returns the number of completed tasks.
func (q *Quest) CompletedCount() int {
	count := 0
	for _, t := range q.Tasks {
		if t.Completed {
			count++
		}
	}
	return count
}

// IsAllTasksCompleted checks if every task is fulfilled.
func (q *Quest) IsAllTasksCompleted() bool {
	if len(q.Tasks) == 0 {
		return false
	}
	for _, t := range q.Tasks {
		if !t.Completed {
			return false
		}
	}
	return true
}

// QuestCategory groups quests under an accordion section in the PDA UI.
type QuestCategory struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Collapsed bool     `json:"collapsed"`
	Quests    []*Quest `json:"quests"`
}

// CompletedRatio returns completed vs total quests in the category.
func (qc *QuestCategory) CompletedRatio() (int, int) {
	completed := 0
	total := len(qc.Quests)
	for _, q := range qc.Quests {
		if q.Completed {
			completed++
		}
	}
	return completed, total
}

// QuestManager manages active quest categories, progression tracking, and save persistence.
type QuestManager struct {
	Categories []*QuestCategory
}

// NewQuestManager creates and returns a default QuestManager with predefined quest lines.
func NewQuestManager() *QuestManager {
	qm := &QuestManager{
		Categories: defaultQuestCategories(),
	}
	return qm
}

func defaultQuestCategories() []*QuestCategory {
	return []*QuestCategory{
		{
			ID:        "training",
			Title:     "TRAINING (TUTORIAL)",
			Collapsed: false,
			Quests: []*Quest{
				{
					ID:          "training_basics",
					Category:    "Training",
					Title:       "Sub-Surface Field Training",
					Description: "AetherCorp standard survival onboarding protocol. Complete initial dive, gather base construction materials, and deploy a surface skiff.",
					Tasks: []*Task{
						{
							ID:            string(TaskTrainDive),
							Description:   "Dive into an ocean trench (Press [E] on a trench tile)",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskTrainTitanium),
							Description:   "Harvest Titanium from subterranean cave walls",
							RequiredCount: 10,
						},
						{
							ID:            string(TaskTrainReturn),
							Description:   "Return to Life Pod 5 (Follow HUD Waypoint marker)",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskTrainSkiffCraft),
							Description:   "Fabricate Skiff Kit at the Life Pod Terminal",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskTrainSkiffDeploy),
							Description:   "Deploy Skiff on ocean surface (From Inventory [Tab])",
							RequiredCount: 1,
						},
					},
				},
			},
		},
		{
			ID:        "survival",
			Title:     "SURVIVAL & UPGRADES",
			Collapsed: false,
			Quests: []*Quest{
				{
					ID:          "tier1_gear",
					Category:    "Survival",
					Title:       "Personal Gear Expansion",
					Description: "Upgrade survival equipment at the Life Pod Fabricator to extend dive endurance and movement velocity.",
					Tasks: []*Task{
						{
							ID:            string(TaskGearO2HC),
							Description:   "Craft High Capacity O2 Tank (+60s capacity)",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskGearFins),
							Description:   "Craft Propulsion Fins (Boost cave swim speed)",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskGearScanner),
							Description:   "Craft Scanner Tool (Bio-scan wildlife & geology)",
							RequiredCount: 1,
						},
					},
				},
				{
					ID:          "vehicle_fleet",
					Category:    "Survival",
					Title:       "Submersible Fleet Construction",
					Description: "Fabricate specialized deep-sea exploration vehicles to withstand immense water pressure and break indestructible ore blocks.",
					Tasks: []*Task{
						{
							ID:            string(TaskVehScoutSub),
							Description:   "Construct and pilot the Scout Sub (Equipped with Sonar Ping [Q])",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskVehHeavyMech),
							Description:   "Construct and pilot the Heavy Mech (Equipped with Drill Arm & Thrusters)",
							RequiredCount: 1,
						},
					},
				},
			},
		},
		{
			ID:        "escape",
			Title:     "PROJECT ESCAPE",
			Collapsed: false,
			Quests: []*Quest{
				{
					ID:          "escape_rocket",
					Category:    "Escape",
					Title:       "AetherCorp Extraction Vehicle",
					Description: "Descend into the deepest abyssal trenches, harvest radioactive Abyssal Ore using the Heavy Mech Drill, and fabricate the Escape Rocket.",
					Tasks: []*Task{
						{
							ID:            string(TaskEscReachDepth),
							Description:   "Descend past depth 100 into the Abyssal Void",
							RequiredCount: 1,
						},
						{
							ID:            string(TaskEscAbyssalOre),
							Description:   "Harvest Abyssal Ore using the Heavy Mech Drill",
							RequiredCount: 10,
						},
						{
							ID:            string(TaskEscCraftRocket),
							Description:   "Fabricate and launch the Escape Rocket at Life Pod 5",
							RequiredCount: 1,
						},
					},
				},
			},
		},
	}
}

// QuestContext provides access to game state needed for evaluating quest progress.
type QuestContext interface {
	IsPlayerInCave() bool
	PlayerTrenchCoords() (x, y int)
	PlayerDistanceToBase() float64
	CountInventoryItem(name string) int
	HasVehicleInWorld(vType string) bool
	MaxDepthReached() float64
	HasCraftedItem(name string) bool
}

// ProgressNotification represents a message generated when a quest or task updates.
type ProgressNotification struct {
	Message   string
	Completed bool
}

// CheckProgress updates all quest objectives against current game state and returns any new notifications.
func (qm *QuestManager) CheckProgress(ctx QuestContext) []ProgressNotification {
	var notifications []ProgressNotification

	for _, cat := range qm.Categories {
		for _, q := range cat.Quests {
			for _, t := range q.Tasks {
				if t.Completed {
					continue
				}

				oldCompleted := t.Completed
				oldCount := t.CurrentCount

				switch TaskID(t.ID) {
				// Training Tasks
				case TaskTrainDive:
					if ctx.IsPlayerInCave() {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskTrainTitanium:
					titaniumCount := ctx.CountInventoryItem("Titanium")
					if titaniumCount > t.CurrentCount {
						t.CurrentCount = titaniumCount
					}
					if t.CurrentCount >= t.RequiredCount {
						t.CurrentCount = t.RequiredCount
						t.Completed = true
					}

				case TaskTrainReturn:
					// Must have completed titanium first
					if q.GetTask(string(TaskTrainTitanium)).Completed && !ctx.IsPlayerInCave() && ctx.PlayerDistanceToBase() <= 120.0 {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskTrainSkiffCraft:
					if ctx.HasCraftedItem("Skiff Kit") || ctx.CountInventoryItem("Skiff Kit") > 0 || ctx.HasVehicleInWorld("skiff") {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskTrainSkiffDeploy:
					if ctx.HasVehicleInWorld("skiff") {
						t.CurrentCount = 1
						t.Completed = true
					}

				// Survival Gear Tasks
				case TaskGearO2HC:
					if ctx.HasCraftedItem("High Capacity O2 Tank") || ctx.CountInventoryItem("High Capacity O2 Tank") > 0 {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskGearFins:
					if ctx.HasCraftedItem("Propulsion Fins") || ctx.CountInventoryItem("Propulsion Fins") > 0 {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskGearScanner:
					if ctx.HasCraftedItem("Scanner Tool") || ctx.CountInventoryItem("Scanner Tool") > 0 {
						t.CurrentCount = 1
						t.Completed = true
					}

				// Vehicle Fleet Tasks
				case TaskVehScoutSub:
					if ctx.HasVehicleInWorld("scout_sub") || ctx.HasCraftedItem("Scout Sub Kit") {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskVehHeavyMech:
					if ctx.HasVehicleInWorld("heavy_mech") || ctx.HasCraftedItem("Heavy Mech Kit") {
						t.CurrentCount = 1
						t.Completed = true
					}

				// Project Escape Tasks
				case TaskEscReachDepth:
					if ctx.MaxDepthReached() >= 100.0 {
						t.CurrentCount = 1
						t.Completed = true
					}

				case TaskEscAbyssalOre:
					abyssalCount := ctx.CountInventoryItem("Abyssal Ore")
					if abyssalCount > t.CurrentCount {
						t.CurrentCount = abyssalCount
					}
					if t.CurrentCount >= t.RequiredCount {
						t.CurrentCount = t.RequiredCount
						t.Completed = true
					}

				case TaskEscCraftRocket:
					if ctx.HasCraftedItem("Escape Rocket") {
						t.CurrentCount = 1
						t.Completed = true
					}
				}

				if !oldCompleted && t.Completed {
					notifications = append(notifications, ProgressNotification{
						Message:   fmt.Sprintf("✓ Objective Complete: %s", t.Description),
						Completed: true,
					})
				} else if t.RequiredCount > 1 && t.CurrentCount > oldCount && !t.Completed {
					notifications = append(notifications, ProgressNotification{
						Message:   fmt.Sprintf("Quest Progress: %s (%d/%d)", t.Description, t.CurrentCount, t.RequiredCount),
						Completed: false,
					})
				}
			}

			// Check quest overall completion
			if !q.Completed && q.IsAllTasksCompleted() {
				q.Completed = true
				notifications = append(notifications, ProgressNotification{
					Message:   fmt.Sprintf("★ QUEST COMPLETED: %s ★", q.Title),
					Completed: true,
				})
			}
		}
	}

	return notifications
}

// GetTask returns a pointer to the task with matching ID within the quest.
func (q *Quest) GetTask(taskID string) *Task {
	for _, t := range q.Tasks {
		if t.ID == taskID {
			return t
		}
	}
	return nil
}

// FindQuest searches across all categories for a quest by ID.
func (qm *QuestManager) FindQuest(questID string) *Quest {
	for _, cat := range qm.Categories {
		for _, q := range cat.Quests {
			if q.ID == questID {
				return q
			}
		}
	}
	return nil
}

// ToggleCategory toggles the collapsed state of a category by ID.
func (qm *QuestManager) ToggleCategory(categoryID string) {
	for _, cat := range qm.Categories {
		if strings.EqualFold(cat.ID, categoryID) {
			cat.Collapsed = !cat.Collapsed
			return
		}
	}
}

// AllQuests returns a flat slice of all quests across categories.
func (qm *QuestManager) AllQuests() []*Quest {
	var all []*Quest
	for _, cat := range qm.Categories {
		all = append(all, cat.Quests...)
	}
	return all
}

// -----------------------------------------------------------------
// Save / Load Serialization
// -----------------------------------------------------------------

// TaskSaveState represents serializable progress for a task.
type TaskSaveState struct {
	ID           string `json:"id"`
	CurrentCount int    `json:"current_count"`
	Completed    bool   `json:"completed"`
}

// QuestSaveState represents serializable progress for a quest.
type QuestSaveState struct {
	ID        string          `json:"id"`
	Completed bool            `json:"completed"`
	Tasks     []TaskSaveState `json:"tasks"`
}

// CategorySaveState represents serializable state for a category.
type CategorySaveState struct {
	ID        string `json:"id"`
	Collapsed bool   `json:"collapsed"`
}

// QuestManagerState holds all data needed to restore QuestManager.
type QuestManagerState struct {
	Categories []CategorySaveState `json:"categories"`
	Quests     []QuestSaveState     `json:"quests"`
}

// SerializeState converts the quest manager's state into a serializable struct.
func (qm *QuestManager) SerializeState() QuestManagerState {
	var state QuestManagerState

	for _, cat := range qm.Categories {
		state.Categories = append(state.Categories, CategorySaveState{
			ID:        cat.ID,
			Collapsed: cat.Collapsed,
		})

		for _, q := range cat.Quests {
			qs := QuestSaveState{
				ID:        q.ID,
				Completed: q.Completed,
			}
			for _, t := range q.Tasks {
				qs.Tasks = append(qs.Tasks, TaskSaveState{
					ID:           t.ID,
					CurrentCount: t.CurrentCount,
					Completed:    t.Completed,
				})
			}
			state.Quests = append(state.Quests, qs)
		}
	}

	return state
}

// DeserializeState restores the quest manager's state from serialized data.
func (qm *QuestManager) DeserializeState(state QuestManagerState) {
	catMap := make(map[string]bool)
	for _, c := range state.Categories {
		catMap[c.ID] = c.Collapsed
	}

	questMap := make(map[string]QuestSaveState)
	for _, q := range state.Quests {
		questMap[q.ID] = q
	}

	for _, cat := range qm.Categories {
		if collapsed, ok := catMap[cat.ID]; ok {
			cat.Collapsed = collapsed
		}

		for _, q := range cat.Quests {
			if qs, ok := questMap[q.ID]; ok {
				q.Completed = qs.Completed

				taskMap := make(map[string]TaskSaveState)
				for _, ts := range qs.Tasks {
					taskMap[ts.ID] = ts
				}

				for _, t := range q.Tasks {
					if ts, ok := taskMap[t.ID]; ok {
						t.CurrentCount = ts.CurrentCount
						t.Completed = ts.Completed
					}
				}
			}
		}
	}
}
