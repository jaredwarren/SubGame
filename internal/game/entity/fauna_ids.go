package entity

// FaunaID identifies a fauna archetype for weighted cave spawning and balance lookup.
type FaunaID int

const (
	FaunaPassiveFish FaunaID = iota
	FaunaPassiveCrab
	FaunaSandViper
	FaunaFalseBulbSnare
	FaunaThermoclineRammer
	FaunaElectroWeaver
	FaunaVoltaicLurker
	FaunaBrimstoneSiphon
	FaunaInkSquid
	FaunaCount
)
