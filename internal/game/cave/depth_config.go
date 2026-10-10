package cave

// CaveDepthConfig defines adjustable configuration for 2.5D depth layers
// and sub-ground tunnel networks.
type CaveDepthConfig struct {
	// TunnelOccurrenceChance controls the probability (0.0 to 1.0) of generating
	// sub-ground tunnel networks and grottoes in natural caves.
	// 0.0 = completely flat solid ground without tunnels (legacy style)
	// 1.0 = every natural cave has sub-ground tunnels
	TunnelOccurrenceChance float64

	// MaxTunnelsPerCave controls the maximum number of distinct sub-ground
	// tunnel routes/loops per cave.
	MaxTunnelsPerCave int

	// TunnelRadius controls the carving radius (in tiles) for sub-ground tunnels.
	// A radius of 2 produces corridors 3–5 tiles wide.
	TunnelRadius int

	// ForegroundDensity controls the frequency of foreground framing arches
	// and crossing pillars/stalactites.
	ForegroundDensity float64

	// DeepGrottoChance controls the probability that a sub-ground tunnel branches
	// into a wide, high-reward subterranean grotto chamber.
	DeepGrottoChance float64
}

// DefaultCaveDepthConfig represents the balanced default for the 3-layer depth system.
var DefaultCaveDepthConfig = CaveDepthConfig{
	TunnelOccurrenceChance: 0.80, // High natural occurrence
	MaxTunnelsPerCave:      2,
	TunnelRadius:           2,
	ForegroundDensity:      0.05,
	DeepGrottoChance:       0.70,
}

// ActiveCaveDepthConfig is the active, mutable configuration used during cave generation.
// It can be adjusted at runtime, via debug options, or balance configuration.
var ActiveCaveDepthConfig = DefaultCaveDepthConfig

// SetTunnelOccurrenceChance updates TunnelOccurrenceChance on ActiveCaveDepthConfig,
// clamping the value to [0.0, 1.0].
func SetTunnelOccurrenceChance(chance float64) {
	if chance < 0.0 {
		chance = 0.0
	} else if chance > 1.0 {
		chance = 1.0
	}
	ActiveCaveDepthConfig.TunnelOccurrenceChance = chance
}

// GetTunnelOccurrenceChance returns the active sub-ground tunnel occurrence rate.
func GetTunnelOccurrenceChance() float64 {
	return ActiveCaveDepthConfig.TunnelOccurrenceChance
}

// ResetCaveDepthConfig restores ActiveCaveDepthConfig to DefaultCaveDepthConfig.
func ResetCaveDepthConfig() {
	ActiveCaveDepthConfig = DefaultCaveDepthConfig
}
