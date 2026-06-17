package item

type DecoyLauncher struct {
	VehicleUpgradeNode[DecoyLauncher]
}
type ChemicalDischarger struct {
	VehicleUpgradeNode[ChemicalDischarger]
}
type SonarAmplifier struct {
	VehicleUpgradeNode[SonarAmplifier]
}
type PowerCell struct{ BaseItem[PowerCell] }
type ThermalGenerator struct {
	VehicleUpgradeNode[ThermalGenerator]
}
