package item

// ItemID is a stable identity for inventory items. Display names may change;
// ItemID values in saves must not.
type ItemID string

const (
	IDTitanium           ItemID = "titanium"
	IDCopper             ItemID = "copper"
	IDQuartz             ItemID = "quartz"
	IDAbyssalOre         ItemID = "abyssal_ore"
	IDNickel             ItemID = "nickel"
	IDScrapMetal         ItemID = "scrap_metal"
	IDElectronicWaste    ItemID = "electronic_waste"
	IDRawFish            ItemID = "raw_fish"
	IDCookedFish         ItemID = "cooked_fish"
	IDRawCrab            ItemID = "raw_crab"
	IDCookedCrab         ItemID = "cooked_crab"
	IDO2TankHC           ItemID = "o2_tank_hc"
	IDO2TankUHC          ItemID = "o2_tank_uhc"
	IDFins               ItemID = "propulsion_fins"
	IDScanner            ItemID = "scanner_tool"
	IDFlashlight         ItemID = "flashlight"
	IDRepairTool         ItemID = "repair_tool"
	IDUpgradeSolar       ItemID = "solar_array"
	IDUpgradeSolarMKII   ItemID = "solar_array_mkii"
	IDUpgradeStorage     ItemID = "storage_vault"
	IDUpgradeStorageMKII ItemID = "storage_vault_mkii"
	IDDecoyLauncher      ItemID = "decoy_launcher"
	IDChemicalDischarger ItemID = "chemical_discharger"
	IDSonarAmplifier     ItemID = "sonar_amplifier"
	IDPowerCell          ItemID = "power_cell"
	IDThermalGenerator   ItemID = "thermal_generator"
	IDEscapeRocket       ItemID = "escape_rocket"
	IDSonicDecoy         ItemID = "sonic_decoy"
	IDChemicalDeterrent  ItemID = "chemical_deterrent"
	IDSkiffKit           ItemID = "skiff_kit"
	IDScoutSubKit        ItemID = "scout_sub_kit"
	IDHeavyMechKit       ItemID = "heavy_mech_kit"
)

// displayNameToID maps historical GetName() / save ItemName / Go type names to stable IDs.
var displayNameToID = map[string]ItemID{
	"Titanium":                    IDTitanium,
	"Copper":                      IDCopper,
	"Quartz":                      IDQuartz,
	"Abyssal Ore":                 IDAbyssalOre,
	"AbyssalOre":                  IDAbyssalOre,
	"Nickel":                      IDNickel,
	"Scrap Metal":                 IDScrapMetal,
	"ScrapMetal":                  IDScrapMetal,
	"Electronic Waste":            IDElectronicWaste,
	"ElectronicWaste":             IDElectronicWaste,
	"Raw Fish":                    IDRawFish,
	"RawFish":                     IDRawFish,
	"Cooked Fish":                 IDCookedFish,
	"CookedFish":                  IDCookedFish,
	"Raw Crab":                    IDRawCrab,
	"RawCrab":                     IDRawCrab,
	"Cooked Crab":                 IDCookedCrab,
	"CookedCrab":                  IDCookedCrab,
	"High Capacity O2 Tank":       IDO2TankHC,
	"O2TankHC":                    IDO2TankHC,
	"Ultra High Capacity O2 Tank": IDO2TankUHC,
	"O2TankUHC":                   IDO2TankUHC,
	"Propulsion Fins":             IDFins,
	"Fins":                        IDFins,
	"Scanner Tool":                IDScanner,
	"Scanner":                     IDScanner,
	"Flashlight":                  IDFlashlight,
	"Repair Tool":                 IDRepairTool,
	"RepairTool":                  IDRepairTool,
	"Solar Array Module":          IDUpgradeSolar,
	"UpgradeSolar":                IDUpgradeSolar,
	"Solar Array MKII Module":     IDUpgradeSolarMKII,
	"UpgradeSolarMKII":            IDUpgradeSolarMKII,
	"Storage Vault Module":        IDUpgradeStorage,
	"UpgradeStorage":              IDUpgradeStorage,
	"Storage Vault MKII Module":   IDUpgradeStorageMKII,
	"UpgradeStorageMKII":          IDUpgradeStorageMKII,
	"Decoy Launcher Module":       IDDecoyLauncher,
	"DecoyLauncher":               IDDecoyLauncher,
	"Chemical Discharger Module":  IDChemicalDischarger,
	"ChemicalDischarger":          IDChemicalDischarger,
	"Sonar Amplifier":             IDSonarAmplifier,
	"SonarAmplifier":              IDSonarAmplifier,
	"Power Cell":                  IDPowerCell,
	"PowerCell":                   IDPowerCell,
	"Thermal Generator":           IDThermalGenerator,
	"ThermalGenerator":            IDThermalGenerator,
	"Escape Rocket":               IDEscapeRocket,
	"EscapeRocket":                IDEscapeRocket,
	"Sonic Decoy":                 IDSonicDecoy,
	"SonicDecoy":                  IDSonicDecoy,
	"Chemical Deterrent":          IDChemicalDeterrent,
	"ChemicalDeterrent":           IDChemicalDeterrent,
	"Skiff Kit":                   IDSkiffKit,
	"Scout Sub Kit":               IDScoutSubKit,
	"Heavy Mech Kit":              IDHeavyMechKit,
}

// ItemIDFromName resolves a display name or historical type name to an ItemID.
func ItemIDFromName(name string) (ItemID, bool) {
	id, ok := displayNameToID[name]
	return id, ok
}
