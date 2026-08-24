package cave

// FloraID identifies a flora archetype for weighted cave spawning.
type FloraID int

const (
	FloraKelp FloraID = iota
	FloraShockKelp
	FloraShatterBulb
	FloraCoral // floor-flora path historically spawned kelp for this id
	FloraNerveMat
	FloraCount
)

// FaunaID identifies a fauna archetype for weighted cave spawning.
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
	FaunaCount
)

// SpawnEntry is a weighted entry in a typed spawn table.
type SpawnEntry[T ~int] struct {
	Type   T
	Weight float64
}

// SelectWeightedEntry picks an item from a slice of SpawnEntry using roll [0, 1).
// Returns the zero value of T if entries is empty.
func SelectWeightedEntry[T ~int](entries []SpawnEntry[T], roll float64) T {
	var zero T
	if len(entries) == 0 {
		return zero
	}
	var total float64
	for _, e := range entries {
		total += e.Weight
	}
	if total <= 0 {
		return entries[0].Type
	}
	target := roll * total
	var current float64
	for _, e := range entries {
		current += e.Weight
		if target <= current {
			return e.Type
		}
	}
	return entries[len(entries)-1].Type
}
