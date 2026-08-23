package quest

// TaskID is a stable quest objective identity (also persisted in save JSON).
type TaskID string

const (
	TaskTrainDive        TaskID = "train_dive"
	TaskTrainTitanium    TaskID = "train_titanium"
	TaskTrainReturn      TaskID = "train_return"
	TaskTrainSkiffCraft  TaskID = "train_skiff_craft"
	TaskTrainSkiffDeploy TaskID = "train_skiff_deploy"
	TaskGearO2HC         TaskID = "gear_o2_hc"
	TaskGearFins         TaskID = "gear_fins"
	TaskGearScanner      TaskID = "gear_scanner"
	TaskVehScoutSub      TaskID = "veh_scout_sub"
	TaskVehHeavyMech     TaskID = "veh_heavy_mech"
	TaskEscReachDepth    TaskID = "esc_reach_depth"
	TaskEscAbyssalOre    TaskID = "esc_abyssal_ore"
	TaskEscCraftRocket   TaskID = "esc_craft_rocket"
)
