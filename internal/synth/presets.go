package synth

// PresetGenerator is a function that synthesizes an audio Buffer.
type PresetGenerator func(seed int64) *Buffer

// SoundCatalog maps relative file paths to their procedural generators.
var SoundCatalog = map[string]PresetGenerator{
	// -----------------------------------------------------------------
	// 4.1 Player Movement & Survival SFX
	// -----------------------------------------------------------------
	"sfx/splash.wav": func(seed int64) *Buffer {
		// Low-pass noise whoosh + resonant sine pop
		layer1 := SoundLayer{
			EnableNoise:   true,
			NoiseType:     WaveNoiseBrown,
			NoiseVolume:   1.0,
			Envelope:      AHDSR{AttackTime: 0.01, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter:  FilterLowPass,
			FilterCutoff:  1200,
			FilterEndCut:  200,
			FilterQ:       1.5,
			Overdrive:     0.2,
			ReverbRoom:    0.5,
			ReverbMix:     0.25,
		}.Generate(0.5, seed)

		layer2 := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        180,
			EndFreq1:     45,
			FreqCurve1:   CurveExponential,
			Volume1:      0.8,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.2, SustainLvl: 0.0, ReleaseTime: 0.05, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 300,
			FilterQ:      2.0,
		}.Generate(0.3, seed+1)

		return MixBuffers(layer1, layer2)
	},

	"sfx/splash_exit.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.9,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.08, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1800,
			FilterEndCut: 600,
			FilterQ:      2.5,
			ReverbRoom:   0.4,
			ReverbMix:    0.2,
		}.Generate(0.35, seed)
	},

	"sfx/swim_stroke.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.75,
			Wave1:        WaveSine,
			Freq1:        90,
			EndFreq1:     40,
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.08, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 450,
			FilterEndCut: 150,
			FilterQ:      1.2,
		}.Generate(0.45, seed)
	},

	"sfx/swim_sprint.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoisePink,
			NoiseVolume:    0.85,
			Envelope:       AHDSR{AttackTime: 0.05, DecayTime: 0.35, SustainLvl: 0.4, ReleaseTime: 0.15, Curve: CurveLinear},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   700,
			FilterQ:        2.0,
			FilterLFOFreq:  8.0,
			FilterLFODepth: 300.0,
		}.Generate(0.6, seed)
	},

	"sfx/heavy_breathing_loop.wav": func(seed int64) *Buffer {
		// Inhale
		inhale := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.65,
			Envelope:     AHDSR{AttackTime: 0.25, DecayTime: 0.45, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 650,
			FilterEndCut: 1100,
			FilterQ:      3.0,
		}.Generate(0.8, seed)

		// Exhale
		exhale := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.8,
			Envelope:     AHDSR{AttackTime: 0.1, DecayTime: 0.6, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 500,
			FilterEndCut: 200,
			FilterQ:      1.5,
		}.Generate(0.8, seed+1)

		buf := NewMonoBuffer(1.8)
		copy(buf.SamplesLeft[0:], inhale.SamplesLeft)
		copy(buf.SamplesLeft[int(0.9*float64(SampleRate)):], exhale.SamplesLeft)
		buf.Normalize(0.9)
		return buf
	},

	"sfx/heartbeat_loop.wav": func(seed int64) *Buffer {
		// "Lub-dub" double pulse
		lub := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        75,
			EndFreq1:     35,
			Volume1:      1.0,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 120,
			FilterQ:      1.8,
			Overdrive:    0.3,
		}.Generate(0.2, seed)

		dub := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        65,
			EndFreq1:     30,
			Volume1:      0.8,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.14, SustainLvl: 0.0, ReleaseTime: 0.05, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 100,
			FilterQ:      1.8,
			Overdrive:    0.2,
		}.Generate(0.25, seed+1)

		buf := NewMonoBuffer(0.9)
		copy(buf.SamplesLeft[0:], lub.SamplesLeft)
		copy(buf.SamplesLeft[int(0.18*float64(SampleRate)):], dub.SamplesLeft)
		buf.Normalize(0.95)
		return buf
	},

	"sfx/player_hurt.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        160,
			EndFreq1:     50,
			FreqCurve1:   CurveExponential,
			Volume1:      0.7,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.6,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.22, SustainLvl: 0.0, ReleaseTime: 0.08, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 600,
			FilterEndCut: 150,
			FilterQ:      2.0,
			Overdrive:    0.5,
			BitDepth:     8,
			Downsample:   2,
		}.Generate(0.3, seed)
	},

	"sfx/player_drown.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.8,
			Wave1:        WaveSine,
			Freq1:        110,
			EndFreq1:     30,
			Volume1:      0.5,
			Envelope:     AHDSR{AttackTime: 0.05, DecayTime: 0.9, SustainLvl: 0.0, ReleaseTime: 0.2, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 350,
			FilterEndCut: 60,
			FilterQ:      2.0,
			ReverbRoom:   0.6,
			ReverbMix:    0.3,
		}.Generate(1.2, seed)
	},

	"sfx/suit_breach.wav": func(seed int64) *Buffer {
		alarm := SoundLayer{
			Wave1:      WaveSquare,
			Freq1:      1400,
			Duty1:      0.5,
			Volume1:    0.6,
			Envelope:   AHDSR{AttackTime: 0.01, DecayTime: 0.1, SustainLvl: 0.0, ReleaseTime: 0.05},
		}.Generate(0.15, seed)

		hiss := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.7,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.5, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterHighPass,
			FilterCutoff: 2500,
			FilterQ:      1.5,
		}.Generate(0.6, seed+1)

		return MixBuffers(alarm, hiss)
	},

	"sfx/o2_refill.wav": func(seed int64) *Buffer {
		gas := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.85,
			Envelope:     AHDSR{AttackTime: 0.08, DecayTime: 0.5, SustainLvl: 0.0, ReleaseTime: 0.15},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1400,
			FilterEndCut: 2800,
			FilterQ:      3.5,
		}.Generate(0.7, seed)

		chime := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        587.33, // D5
			EndFreq1:     880.0,  // A5
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.05, DecayTime: 0.4, SustainLvl: 0.0, ReleaseTime: 0.2},
			ReverbRoom:   0.5,
			ReverbMix:    0.25,
		}.Generate(0.7, seed+1)

		return MixBuffers(gas, chime)
	},

	"sfx/item_pickup.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        440,
			EndFreq1:     880,
			FreqCurve1:   CurveExponential,
			Volume1:      0.8,
			EnableOsc2:   true,
			Wave2:        WaveTriangle,
			Freq2:        1320,
			Volume2:      0.4,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05, Curve: CurveExponential},
			ReverbRoom:   0.3,
			ReverbMix:    0.2,
		}.Generate(0.18, seed)
	},

	// -----------------------------------------------------------------
	// 4.2 Tools, Equipment & Usables SFX
	// -----------------------------------------------------------------
	"sfx/mining_hit.wav": func(seed int64) *Buffer {
		strike := SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        1200,
			EndFreq1:     300,
			Duty1:        0.25,
			Volume1:      0.8,
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.4,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.08, SustainLvl: 0.0, ReleaseTime: 0.04, Curve: CurveExponential},
			BitDepth:     8,
			Downsample:   1,
		}.Generate(0.15, seed)

		ring := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        880,
			Volume1:      0.6,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.2, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 880,
			FilterQ:      8.0,
			ReverbRoom:   0.4,
			ReverbMix:    0.3,
		}.Generate(0.3, seed+1)

		return MixBuffers(strike, ring)
	},

	"sfx/dig_crunch.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.9,
			Wave1:        WaveTriangle,
			Freq1:        150,
			EndFreq1:     40,
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.18, SustainLvl: 0.0, ReleaseTime: 0.06, Curve: CurveExponential},
			EnableFilter: FilterLowPass,
			FilterCutoff: 800,
			FilterEndCut: 200,
			FilterQ:      1.8,
			Overdrive:    0.3,
			BitDepth:     6,
			Downsample:   2,
		}.Generate(0.25, seed)
	},

	"sfx/ore_break.wav": func(seed int64) *Buffer {
		shatter := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.9,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 2400,
			FilterEndCut: 400,
			FilterQ:      3.0,
			Overdrive:    0.4,
		}.Generate(0.45, seed)

		thud := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        220,
			EndFreq1:     40,
			Volume1:      0.9,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 200,
			FilterQ:      2.0,
		}.Generate(0.35, seed+1)

		return MixBuffers(shatter, thud)
	},

	"sfx/scanner_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSine,
			Freq1:          950,
			Volume1:        0.7,
			EnableOsc2:     true,
			Wave2:          WaveTriangle,
			Freq2:          1900,
			Volume2:        0.3,
			VibratoFreq:    12.0,
			VibratoDepth:   150.0,
			Envelope:       AHDSR{AttackTime: 0.05, DecayTime: 0.4, SustainLvl: 0.6, ReleaseTime: 0.05, Curve: CurveLinear},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   1400,
			FilterQ:        5.0,
			FilterLFOFreq:  4.0,
			FilterLFODepth: 400.0,
		}.Generate(0.8, seed)
	},

	"sfx/scanner_complete.wav": func(seed int64) *Buffer {
		// Trill: C6 -> E6 -> G6
		t1 := SoundLayer{
			Wave1:      WaveSine,
			Freq1:      1046.50, // C6
			Volume1:    0.7,
			Envelope:   AHDSR{AttackTime: 0.005, DecayTime: 0.1, SustainLvl: 0.0, ReleaseTime: 0.05},
			ReverbRoom: 0.4,
			ReverbMix:  0.2,
		}.Generate(0.12, seed)

		t2 := SoundLayer{
			Wave1:      WaveSine,
			Freq1:      1318.51, // E6
			Volume1:    0.8,
			Envelope:   AHDSR{AttackTime: 0.005, DecayTime: 0.1, SustainLvl: 0.0, ReleaseTime: 0.05},
			ReverbRoom: 0.4,
			ReverbMix:  0.2,
		}.Generate(0.12, seed+1)

		t3 := SoundLayer{
			Wave1:      WaveSine,
			Freq1:      1567.98, // G6
			Volume1:    0.9,
			Envelope:   AHDSR{AttackTime: 0.005, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.15},
			ReverbRoom: 0.6,
			ReverbMix:  0.35,
		}.Generate(0.45, seed+2)

		buf := NewMonoBuffer(0.65)
		copy(buf.SamplesLeft[0:], t1.SamplesLeft)
		copy(buf.SamplesLeft[int(0.09*float64(SampleRate)):], t2.SamplesLeft)
		copy(buf.SamplesLeft[int(0.18*float64(SampleRate)):], t3.SamplesLeft)
		buf.Normalize(0.95)
		return buf
	},

	"sfx/flashlight_toggle.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        600,
			EndFreq1:     120,
			Duty1:        0.3,
			Volume1:      0.8,
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.5,
			Envelope:     AHDSR{AttackTime: 0.001, DecayTime: 0.04, SustainLvl: 0.0, ReleaseTime: 0.02},
			EnableFilter: FilterHighPass,
			FilterCutoff: 500,
			FilterQ:      2.0,
			BitDepth:     8,
			Downsample:   1,
		}.Generate(0.08, seed)
	},

	"sfx/repair_tool_loop.wav": func(seed int64) *Buffer {
		arc := SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        120,
			Volume1:      0.6,
			Overdrive:    0.8,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.4, SustainLvl: 0.6, ReleaseTime: 0.05},
			EnableFilter: FilterBandPass,
			FilterCutoff: 800,
			FilterQ:      4.0,
		}.Generate(0.6, seed)

		spark := SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoiseWhite,
			NoiseVolume:    0.7,
			Envelope:       AHDSR{AttackTime: 0.01, DecayTime: 0.4, SustainLvl: 0.5, ReleaseTime: 0.05},
			EnableFilter:   FilterHighPass,
			FilterCutoff:   3000,
			FilterQ:        2.0,
			FilterLFOFreq:  15.0,
			FilterLFODepth: 1000.0,
		}.Generate(0.6, seed+1)

		return MixBuffers(arc, spark)
	},

	"sfx/repair_tool_complete.wav": func(seed int64) *Buffer {
		c1 := SoundLayer{
			Wave1:      WaveTriangle,
			Freq1:      523.25, // C5
			Volume1:    0.8,
			Envelope:   AHDSR{AttackTime: 0.005, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05},
		}.Generate(0.15, seed)

		c2 := SoundLayer{
			Wave1:      WaveTriangle,
			Freq1:      783.99, // G5
			Volume1:    0.9,
			Envelope:   AHDSR{AttackTime: 0.005, DecayTime: 0.3, SustainLvl: 0.0, ReleaseTime: 0.1},
			ReverbRoom: 0.4,
			ReverbMix:  0.25,
		}.Generate(0.35, seed+1)

		buf := NewMonoBuffer(0.45)
		copy(buf.SamplesLeft[0:], c1.SamplesLeft)
		copy(buf.SamplesLeft[int(0.1*float64(SampleRate)):], c2.SamplesLeft)
		buf.Normalize(0.95)
		return buf
	},

	"sfx/decoy_launch.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.9,
			Wave1:        WaveSine,
			Freq1:        300,
			EndFreq1:     80,
			Volume1:      0.5,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 1800,
			FilterEndCut: 300,
			FilterQ:      2.0,
			Overdrive:    0.2,
		}.Generate(0.35, seed)
	},

	"sfx/decoy_pulse.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        660,
			EndFreq1:     440,
			Volume1:      0.9,
			EnableOsc2:   true,
			Wave2:        WaveTriangle,
			Freq2:        1320,
			Volume2:      0.3,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.3, SustainLvl: 0.0, ReleaseTime: 0.15},
			ReverbRoom:   0.7,
			ReverbMix:    0.4,
		}.Generate(0.5, seed)
	},

	"sfx/deterrent_disperse.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoisePink,
			NoiseVolume:    0.85,
			Envelope:       AHDSR{AttackTime: 0.04, DecayTime: 0.6, SustainLvl: 0.0, ReleaseTime: 0.15},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   1200,
			FilterEndCut:   400,
			FilterQ:        2.5,
			FilterLFOFreq:  6.0,
			FilterLFODepth: 250.0,
		}.Generate(0.75, seed)
	},

	// -----------------------------------------------------------------
	// 4.3 Vehicles & Machinery SFX
	// -----------------------------------------------------------------
	"sfx/skiff_engine_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSawtooth,
			Freq1:          85,
			Volume1:        0.7,
			EnableOsc2:     true,
			Wave2:          WaveSquare,
			Freq2:          42.5,
			Duty2:          0.3,
			Volume2:        0.5,
			EnableNoise:    true,
			NoiseType:      WaveNoiseBrown,
			NoiseVolume:    0.4,
			Envelope:       AHDSR{AttackTime: 0.1, DecayTime: 0.4, SustainLvl: 0.7, ReleaseTime: 0.1},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   450,
			FilterQ:        1.8,
			VibratoFreq:    14.0,
			VibratoDepth:   4.0,
			Overdrive:      0.3,
		}.Generate(0.8, seed)
	},

	"sfx/skiff_solar_charge.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSine,
			Freq1:          3520, // A7
			Volume1:        0.4,
			VibratoFreq:    8.0,
			VibratoDepth:   20.0,
			Envelope:       AHDSR{AttackTime: 0.2, DecayTime: 0.4, SustainLvl: 0.5, ReleaseTime: 0.2},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   3520,
			FilterQ:        6.0,
		}.Generate(0.9, seed)
	},

	"sfx/sub_engine_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        60,
			Volume1:      0.8,
			EnableOsc2:   true,
			Wave2:        WaveTriangle,
			Freq2:        120,
			Volume2:      0.4,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.5,
			Envelope:     AHDSR{AttackTime: 0.1, DecayTime: 0.5, SustainLvl: 0.6, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 250,
			FilterQ:      1.5,
		}.Generate(0.8, seed)
	},

	"sfx/sub_sonar_ping.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        1046.50, // C6
			EndFreq1:     1030.0,
			Volume1:      0.95,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.6, SustainLvl: 0.0, ReleaseTime: 0.3, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1046.50,
			FilterQ:      12.0,
			ReverbRoom:   0.85,
			ReverbMix:    0.45,
		}.Generate(1.0, seed)
	},

	"sfx/sub_sonar_echo.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        1046.50,
			Volume1:      0.5,
			Envelope:     AHDSR{AttackTime: 0.03, DecayTime: 0.4, SustainLvl: 0.0, ReleaseTime: 0.2, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1046.50,
			FilterQ:      8.0,
			ReverbRoom:   0.9,
			ReverbMix:    0.6,
		}.Generate(0.8, seed)
	},

	"sfx/sub_hull_creak.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSawtooth,
			Freq1:          180,
			EndFreq1:       75,
			Volume1:        0.7,
			VibratoFreq:    5.0,
			VibratoDepth:   30.0,
			Envelope:       AHDSR{AttackTime: 0.1, DecayTime: 0.7, SustainLvl: 0.0, ReleaseTime: 0.2, Curve: CurveLinear},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   400,
			FilterEndCut:   150,
			FilterQ:        5.0,
			Overdrive:      0.4,
			ReverbRoom:     0.7,
			ReverbMix:      0.35,
		}.Generate(1.1, seed)
	},

	"sfx/mech_step.wav": func(seed int64) *Buffer {
		thud := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        110,
			EndFreq1:     30,
			FreqCurve1:   CurveExponential,
			Volume1:      0.9,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.6,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.08},
			EnableFilter: FilterLowPass,
			FilterCutoff: 300,
			FilterQ:      2.0,
			Overdrive:    0.4,
		}.Generate(0.35, seed)

		servo := SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        440,
			EndFreq1:     220,
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.15, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterBandPass,
			FilterCutoff: 600,
			FilterQ:      4.0,
		}.Generate(0.25, seed+1)

		return MixBuffers(thud, servo)
	},

	"sfx/mech_drill_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSawtooth,
			Freq1:          90,
			Volume1:        0.7,
			EnableOsc2:     true,
			Wave2:          WaveSquare,
			Freq2:          180,
			Duty2:          0.2,
			Volume2:        0.5,
			EnableNoise:    true,
			NoiseType:      WaveNoiseWhite,
			NoiseVolume:    0.4,
			VibratoFreq:    20.0,
			VibratoDepth:   15.0,
			Envelope:       AHDSR{AttackTime: 0.05, DecayTime: 0.4, SustainLvl: 0.7, ReleaseTime: 0.1},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   900,
			FilterQ:        2.5,
			Overdrive:      0.6,
		}.Generate(0.6, seed)
	},

	"sfx/mech_thruster_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoiseBrown,
			NoiseVolume:    0.9,
			Wave1:          WaveSine,
			Freq1:          55,
			Volume1:        0.5,
			Envelope:       AHDSR{AttackTime: 0.05, DecayTime: 0.4, SustainLvl: 0.7, ReleaseTime: 0.1},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   700,
			FilterQ:        1.6,
			FilterLFOFreq:  12.0,
			FilterLFODepth: 150.0,
		}.Generate(0.6, seed)
	},

	"sfx/mech_impact.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        140,
			EndFreq1:     25,
			FreqCurve1:   CurveExponential,
			Volume1:      1.0,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.8,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.4, SustainLvl: 0.0, ReleaseTime: 0.15},
			EnableFilter: FilterLowPass,
			FilterCutoff: 350,
			FilterEndCut: 60,
			FilterQ:      2.5,
			Overdrive:    0.5,
			ReverbRoom:   0.6,
			ReverbMix:    0.3,
		}.Generate(0.6, seed)
	},

	"sfx/vehicle_enter.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.8,
			Wave1:        WaveSine,
			Freq1:        200,
			EndFreq1:     500,
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 1500,
			FilterEndCut: 400,
			FilterQ:      2.0,
		}.Generate(0.35, seed)
	},

	"sfx/vehicle_exit.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.8,
			Wave1:        WaveSine,
			Freq1:        500,
			EndFreq1:     150,
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.22, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 1200,
			FilterEndCut: 300,
			FilterQ:      2.0,
		}.Generate(0.3, seed)
	},

	"sfx/vehicle_alarm.wav": func(seed int64) *Buffer {
		b1 := SoundLayer{
			Wave1:    WaveSquare,
			Freq1:    880,
			Duty1:    0.5,
			Volume1:  0.7,
			Envelope: AHDSR{AttackTime: 0.01, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.04},
		}.Generate(0.15, seed)

		b2 := SoundLayer{
			Wave1:    WaveSquare,
			Freq1:    659.25, // E5
			Duty1:    0.5,
			Volume1:  0.7,
			Envelope: AHDSR{AttackTime: 0.01, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.04},
		}.Generate(0.15, seed+1)

		buf := NewMonoBuffer(0.4)
		copy(buf.SamplesLeft[0:], b1.SamplesLeft)
		copy(buf.SamplesLeft[int(0.15*float64(SampleRate)):], b2.SamplesLeft)
		buf.Normalize(0.9)
		return buf
	},

	// -----------------------------------------------------------------
	// 4.4 Marine Fauna, Flora & Hazards SFX
	// -----------------------------------------------------------------
	"sfx/rammer_roar.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSawtooth,
			Freq1:          90,
			EndFreq1:       45,
			Volume1:        0.85,
			EnableNoise:    true,
			NoiseType:      WaveNoiseBrown,
			NoiseVolume:    0.6,
			VibratoFreq:    8.0,
			VibratoDepth:   25.0,
			Envelope:       AHDSR{AttackTime: 0.12, DecayTime: 0.7, SustainLvl: 0.0, ReleaseTime: 0.2},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   450,
			FilterEndCut:   120,
			FilterQ:        3.5,
			Overdrive:      0.6,
			ReverbRoom:     0.7,
			ReverbMix:      0.35,
		}.Generate(1.1, seed)
	},

	"sfx/rammer_impact.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        180,
			EndFreq1:     35,
			FreqCurve1:   CurveExponential,
			Volume1:      1.0,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.9,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 500,
			FilterEndCut: 80,
			FilterQ:      2.5,
			Overdrive:    0.6,
			ReverbRoom:   0.6,
			ReverbMix:    0.3,
		}.Generate(0.55, seed)
	},

	"sfx/weaver_charge.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSawtooth,
			Freq1:          180,
			EndFreq1:       1200,
			FreqCurve1:     CurveExponential,
			Volume1:        0.75,
			EnableNoise:    true,
			NoiseType:      WaveNoiseWhite,
			NoiseVolume:    0.4,
			VibratoFreq:    30.0,
			VibratoDepth:   40.0,
			Envelope:       AHDSR{AttackTime: 0.2, DecayTime: 0.8, SustainLvl: 0.0, ReleaseTime: 0.2},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   400,
			FilterEndCut:   2800,
			FilterQ:        5.0,
			Overdrive:      0.4,
		}.Generate(1.2, seed)
	},

	"sfx/weaver_shock.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        80,
			Volume1:      0.8,
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.95,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.1, Curve: CurveExponential},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1800,
			FilterEndCut: 300,
			FilterQ:      4.0,
			Overdrive:    0.8,
			BitDepth:     6,
			Downsample:   2,
			ReverbRoom:   0.6,
			ReverbMix:    0.35,
		}.Generate(0.5, seed)
	},

	"sfx/lurker_ambush.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        110,
			EndFreq1:     400,
			Volume1:      0.8,
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.7,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.3, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 800,
			FilterQ:      3.0,
			Overdrive:    0.5,
		}.Generate(0.45, seed)
	},

	"sfx/snare_lure_hum.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSine,
			Freq1:          784, // G5
			Volume1:        0.7,
			EnableOsc2:     true,
			Wave2:          WaveSine,
			Freq2:          1174.66, // D6
			Volume2:        0.4,
			VibratoFreq:    4.0,
			VibratoDepth:   30.0,
			Envelope:       AHDSR{AttackTime: 0.2, DecayTime: 0.5, SustainLvl: 0.4, ReleaseTime: 0.2},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   900,
			FilterQ:        4.0,
			ReverbRoom:     0.7,
			ReverbMix:      0.4,
		}.Generate(0.9, seed)
	},

	"sfx/snare_snap.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        350,
			EndFreq1:     50,
			FreqCurve1:   CurveExponential,
			Volume1:      0.9,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.8,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.18, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 900,
			FilterEndCut: 150,
			FilterQ:      2.5,
			Overdrive:    0.5,
		}.Generate(0.25, seed)
	},

	"sfx/viper_burrow.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoisePink,
			NoiseVolume:    0.8,
			Envelope:       AHDSR{AttackTime: 0.08, DecayTime: 0.4, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   750,
			FilterQ:        3.0,
			FilterLFOFreq:  8.0,
			FilterLFODepth: 200.0,
		}.Generate(0.55, seed)
	},

	"sfx/viper_strike.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.9,
			Wave1:        WaveSawtooth,
			Freq1:        280,
			EndFreq1:     70,
			Volume1:      0.6,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.22, SustainLvl: 0.0, ReleaseTime: 0.08},
			EnableFilter: FilterBandPass,
			FilterCutoff: 2200,
			FilterEndCut: 500,
			FilterQ:      3.0,
			Overdrive:    0.4,
		}.Generate(0.35, seed)
	},

	"sfx/shatter_bulb_pop.wav": func(seed int64) *Buffer {
		pop := SoundLayer{
			Wave1:        WaveSine,
			Freq1:        600,
			EndFreq1:     80,
			FreqCurve1:   CurveExponential,
			Volume1:      0.9,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.04},
			EnableFilter: FilterLowPass,
			FilterCutoff: 700,
			FilterQ:      3.0,
		}.Generate(0.18, seed)

		gasWave := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.7,
			Wave1:        WaveSine,
			Freq1:        3200,
			EndFreq1:     1200,
			Volume1:      0.5,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.4, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterHighPass,
			FilterCutoff: 1500,
			FilterQ:      2.0,
			ReverbRoom:   0.6,
			ReverbMix:    0.3,
		}.Generate(0.5, seed+1)

		return MixBuffers(pop, gasWave)
	},

	"sfx/shock_kelp_hum.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        60,
			Volume1:      0.6,
			EnableOsc2:   true,
			Wave2:        WaveSine,
			Freq2:        180,
			Volume2:      0.4,
			Envelope:     AHDSR{AttackTime: 0.1, DecayTime: 0.5, SustainLvl: 0.6, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 300,
			FilterQ:      3.0,
		}.Generate(0.7, seed)
	},

	"sfx/shock_kelp_zap.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.9,
			Wave1:        WaveSquare,
			Freq1:        350,
			Duty1:        0.1,
			Volume1:      0.6,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.15, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterHighPass,
			FilterCutoff: 1800,
			FilterQ:      3.0,
			BitDepth:     6,
			Downsample:   2,
		}.Generate(0.22, seed)
	},

	"sfx/thermal_vent_hiss.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoisePink,
			NoiseVolume:    0.85,
			Envelope:       AHDSR{AttackTime: 0.1, DecayTime: 0.5, SustainLvl: 0.7, ReleaseTime: 0.1},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   1100,
			FilterQ:        2.0,
			FilterLFOFreq:  5.0,
			FilterLFODepth: 200.0,
		}.Generate(0.7, seed)
	},

	"sfx/vent_bubble_rumble.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.9,
			Wave1:        WaveSine,
			Freq1:        45,
			EndFreq1:     30,
			Volume1:      0.6,
			Envelope:     AHDSR{AttackTime: 0.1, DecayTime: 0.6, SustainLvl: 0.6, ReleaseTime: 0.2},
			EnableFilter: FilterLowPass,
			FilterCutoff: 160,
			FilterQ:      2.5,
			Overdrive:    0.4,
		}.Generate(0.9, seed)
	},

	"sfx/nerve_mat_sting.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        440,
			EndFreq1:     120,
			Volume1:      0.7,
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.5,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.25, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1200,
			FilterQ:      4.0,
			Overdrive:    0.3,
		}.Generate(0.32, seed)
	},

	"sfx/whirlpool_roar_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoiseBrown,
			NoiseVolume:    0.95,
			Wave1:          WaveSine,
			Freq1:          40,
			Volume1:        0.5,
			Envelope:       AHDSR{AttackTime: 0.1, DecayTime: 0.6, SustainLvl: 0.8, ReleaseTime: 0.2},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   350,
			FilterQ:        2.0,
			FilterLFOFreq:  3.0,
			FilterLFODepth: 120.0,
			Overdrive:      0.3,
		}.Generate(1.0, seed)
	},

	"sfx/fish_swim_flutter.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.6,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.04},
			EnableFilter: FilterBandPass,
			FilterCutoff: 900,
			FilterEndCut: 300,
			FilterQ:      2.5,
		}.Generate(0.18, seed)
	},

	"sfx/crab_skitter.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveTriangle,
			Freq1:        800,
			EndFreq1:     200,
			Volume1:      0.7,
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.4,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.05, SustainLvl: 0.0, ReleaseTime: 0.02},
			BitDepth:     6,
			Downsample:   2,
		}.Generate(0.08, seed)
	},

	"sfx/cargo_unlatch.wav": func(seed int64) *Buffer {
		latch := SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        220,
			EndFreq1:     60,
			Volume1:      0.7,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.1, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 800,
			FilterQ:      2.0,
		}.Generate(0.18, seed)

		chime := SoundLayer{
			Wave1:      WaveSine,
			Freq1:      880, // A5
			EndFreq1:   1174.66, // D6
			Volume1:    0.7,
			Envelope:   AHDSR{AttackTime: 0.02, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.15},
			ReverbRoom: 0.5,
			ReverbMix:  0.3,
		}.Generate(0.5, seed+1)

		buf := NewMonoBuffer(0.55)
		copy(buf.SamplesLeft[0:], latch.SamplesLeft)
		copy(buf.SamplesLeft[int(0.08*float64(SampleRate)):], chime.SamplesLeft)
		buf.Normalize(0.95)
		return buf
	},

	// -----------------------------------------------------------------
	// 4.5 Base Building, Fabricator Crafting & Energy Systems SFX
	// -----------------------------------------------------------------
	"sfx/fabricator_craft_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:          WaveSawtooth,
			Freq1:          330,
			Volume1:        0.6,
			VibratoFreq:    16.0,
			VibratoDepth:   40.0,
			EnableNoise:    true,
			NoiseType:      WaveNoiseWhite,
			NoiseVolume:    0.3,
			Envelope:       AHDSR{AttackTime: 0.05, DecayTime: 0.4, SustainLvl: 0.7, ReleaseTime: 0.1},
			EnableFilter:   FilterBandPass,
			FilterCutoff:   1200,
			FilterQ:        4.0,
			FilterLFOFreq:  6.0,
			FilterLFODepth: 300.0,
		}.Generate(0.8, seed)
	},

	"sfx/fabricator_success.wav": func(seed int64) *Buffer {
		n1 := SoundLayer{
			Wave1:    WaveTriangle,
			Freq1:    523.25, // C5
			Volume1:  0.8,
			Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05},
		}.Generate(0.15, seed)

		n2 := SoundLayer{
			Wave1:    WaveTriangle,
			Freq1:    659.25, // E5
			Volume1:  0.8,
			Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05},
		}.Generate(0.15, seed+1)

		n3 := SoundLayer{
			Wave1:      WaveTriangle,
			Freq1:      1046.50, // C6
			Volume1:    0.9,
			Envelope:   AHDSR{AttackTime: 0.005, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.15},
			ReverbRoom: 0.5,
			ReverbMix:  0.3,
		}.Generate(0.45, seed+2)

		buf := NewMonoBuffer(0.6)
		copy(buf.SamplesLeft[0:], n1.SamplesLeft)
		copy(buf.SamplesLeft[int(0.09*float64(SampleRate)):], n2.SamplesLeft)
		copy(buf.SamplesLeft[int(0.18*float64(SampleRate)):], n3.SamplesLeft)
		buf.Normalize(0.95)
		return buf
	},

	"sfx/base_build.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        180,
			EndFreq1:     60,
			Duty1:        0.3,
			Volume1:      0.8,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.6,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.3, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 600,
			FilterEndCut: 150,
			FilterQ:      2.0,
			ReverbRoom:   0.5,
			ReverbMix:    0.25,
		}.Generate(0.45, seed)
	},

	"sfx/base_deconstruct.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        60,
			EndFreq1:     280,
			Volume1:      0.7,
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.5,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.28, SustainLvl: 0.0, ReleaseTime: 0.08},
			EnableFilter: FilterBandPass,
			FilterCutoff: 800,
			FilterQ:      3.0,
		}.Generate(0.4, seed)
	},

	"sfx/airlock_cycle.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoisePink,
			NoiseVolume:    0.9,
			Wave1:          WaveSine,
			Freq1:          90,
			Volume1:        0.5,
			Envelope:       AHDSR{AttackTime: 0.1, DecayTime: 0.8, SustainLvl: 0.0, ReleaseTime: 0.2},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   1200,
			FilterEndCut:   200,
			FilterQ:        2.0,
			FilterLFOFreq:  4.0,
			FilterLFODepth: 150.0,
		}.Generate(1.1, seed)
	},

	"sfx/power_online.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:      WaveSine,
			Freq1:      120,
			EndFreq1:   880,
			FreqCurve1: CurveExponential,
			Volume1:    0.8,
			Envelope:   AHDSR{AttackTime: 0.05, DecayTime: 0.4, SustainLvl: 0.0, ReleaseTime: 0.15},
			ReverbRoom: 0.5,
			ReverbMix:  0.3,
		}.Generate(0.6, seed)
	},

	"sfx/generator_hum_loop.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        60,
			Volume1:      0.8,
			EnableOsc2:   true,
			Wave2:        WaveTriangle,
			Freq2:        120,
			Volume2:      0.4,
			Envelope:     AHDSR{AttackTime: 0.1, DecayTime: 0.5, SustainLvl: 0.7, ReleaseTime: 0.1},
			EnableFilter: FilterLowPass,
			FilterCutoff: 200,
			FilterQ:      2.0,
		}.Generate(0.7, seed)
	},

	"sfx/base_power_down.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        440,
			EndFreq1:     40,
			FreqCurve1:   CurveExponential,
			Volume1:      0.8,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.7, SustainLvl: 0.0, ReleaseTime: 0.2},
			EnableFilter: FilterLowPass,
			FilterCutoff: 600,
			FilterEndCut: 60,
			FilterQ:      3.0,
		}.Generate(0.95, seed)
	},

	"sfx/power_alarm.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:    WaveSquare,
			Freq1:    440,
			Duty1:    0.5,
			Volume1:  0.7,
			Envelope: AHDSR{AttackTime: 0.01, DecayTime: 0.15, SustainLvl: 0.0, ReleaseTime: 0.05},
		}.Generate(0.25, seed)
	},

	"sfx/storage_open.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        300,
			EndFreq1:     600,
			Duty1:        0.3,
			Volume1:      0.6,
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.4,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.15, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 1000,
			FilterQ:      1.5,
		}.Generate(0.22, seed)
	},

	"sfx/storage_close.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSquare,
			Freq1:        500,
			EndFreq1:     150,
			Duty1:        0.3,
			Volume1:      0.7,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.5,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.18, SustainLvl: 0.0, ReleaseTime: 0.05},
			EnableFilter: FilterLowPass,
			FilterCutoff: 800,
			FilterQ:      1.5,
		}.Generate(0.25, seed)
	},

	"sfx/eat_crunch.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.8,
			Wave1:        WaveTriangle,
			Freq1:        200,
			EndFreq1:     70,
			Volume1:      0.4,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.1, SustainLvl: 0.0, ReleaseTime: 0.04},
			EnableFilter: FilterBandPass,
			FilterCutoff: 900,
			FilterQ:      3.0,
			BitDepth:     8,
			Downsample:   1,
		}.Generate(0.16, seed)
	},

	"sfx/medkit_apply.wav": func(seed int64) *Buffer {
		spray := SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.8,
			Envelope:     AHDSR{AttackTime: 0.02, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterBandPass,
			FilterCutoff: 2200,
			FilterEndCut: 1000,
			FilterQ:      2.5,
		}.Generate(0.45, seed)

		heal := SoundLayer{
			Wave1:      WaveSine,
			Freq1:      440,
			EndFreq1:   659.25,
			Volume1:    0.6,
			Envelope:   AHDSR{AttackTime: 0.05, DecayTime: 0.35, SustainLvl: 0.0, ReleaseTime: 0.15},
			ReverbRoom: 0.4,
			ReverbMix:  0.25,
		}.Generate(0.5, seed+1)

		return MixBuffers(spray, heal)
	},

	// -----------------------------------------------------------------
	// 4.6 User Interface, HUD, Map & PDA SFX
	// -----------------------------------------------------------------
	"sfx/ui_hover.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:      WaveTriangle,
			Freq1:      1200,
			Volume1:    0.5,
			Envelope:   AHDSR{AttackTime: 0.002, DecayTime: 0.025, SustainLvl: 0.0, ReleaseTime: 0.01},
			BitDepth:   8,
			Downsample: 1,
		}.Generate(0.04, seed)
	},

	"sfx/ui_confirm.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        880,
			EndFreq1:     1320,
			FreqCurve1:   CurveExponential,
			Volume1:      0.75,
			Envelope:     AHDSR{AttackTime: 0.002, DecayTime: 0.08, SustainLvl: 0.0, ReleaseTime: 0.03},
			ReverbRoom:   0.3,
			ReverbMix:    0.15,
		}.Generate(0.12, seed)
	},

	"sfx/ui_cancel.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:      WaveSine,
			Freq1:      440,
			EndFreq1:   220,
			FreqCurve1: CurveExponential,
			Volume1:    0.7,
			Envelope:   AHDSR{AttackTime: 0.002, DecayTime: 0.09, SustainLvl: 0.0, ReleaseTime: 0.03},
		}.Generate(0.13, seed)
	},

	"sfx/ui_error.wav": func(seed int64) *Buffer {
		b1 := SoundLayer{
			Wave1:    WaveSquare,
			Freq1:    150,
			Duty1:    0.5,
			Volume1:  0.7,
			Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.07, SustainLvl: 0.0, ReleaseTime: 0.02},
		}.Generate(0.09, seed)

		b2 := SoundLayer{
			Wave1:    WaveSquare,
			Freq1:    130,
			Duty1:    0.5,
			Volume1:  0.7,
			Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.09, SustainLvl: 0.0, ReleaseTime: 0.03},
		}.Generate(0.12, seed+1)

		buf := NewMonoBuffer(0.24)
		copy(buf.SamplesLeft[0:], b1.SamplesLeft)
		copy(buf.SamplesLeft[int(0.1*float64(SampleRate)):], b2.SamplesLeft)
		buf.Normalize(0.9)
		return buf
	},

	"sfx/inventory_open.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        350,
			EndFreq1:     700,
			Volume1:      0.6,
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.3,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.04},
			EnableFilter: FilterLowPass,
			FilterCutoff: 1500,
			FilterQ:      1.5,
		}.Generate(0.18, seed)
	},

	"sfx/inventory_close.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        700,
			EndFreq1:     280,
			Volume1:      0.6,
			EnableNoise:  true,
			NoiseType:    WaveNoisePink,
			NoiseVolume:  0.3,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.11, SustainLvl: 0.0, ReleaseTime: 0.04},
			EnableFilter: FilterLowPass,
			FilterCutoff: 1200,
			FilterQ:      1.5,
		}.Generate(0.16, seed)
	},

	"sfx/hotbar_switch.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:      WaveSquare,
			Freq1:      800,
			EndFreq1:   1200,
			Duty1:      0.2,
			Volume1:    0.5,
			Envelope:   AHDSR{AttackTime: 0.001, DecayTime: 0.03, SustainLvl: 0.0, ReleaseTime: 0.01},
			BitDepth:   8,
			Downsample: 1,
		}.Generate(0.045, seed)
	},

	"sfx/map_open.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:      WaveSine,
			Freq1:      440,
			EndFreq1:   880,
			Volume1:    0.6,
			Envelope:   AHDSR{AttackTime: 0.02, DecayTime: 0.2, SustainLvl: 0.0, ReleaseTime: 0.08},
			ReverbRoom: 0.4,
			ReverbMix:  0.25,
		}.Generate(0.3, seed)
	},

	"sfx/map_ping.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSine,
			Freq1:        1320,
			Volume1:      0.8,
			Envelope:     AHDSR{AttackTime: 0.005, DecayTime: 0.3, SustainLvl: 0.0, ReleaseTime: 0.1},
			EnableFilter: FilterBandPass,
			FilterCutoff: 1320,
			FilterQ:      8.0,
			ReverbRoom:   0.5,
			ReverbMix:    0.35,
		}.Generate(0.42, seed)
	},

	"sfx/pda_typewriter_tick.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseWhite,
			NoiseVolume:  0.5,
			Wave1:        WaveTriangle,
			Freq1:        1800,
			Volume1:      0.3,
			Envelope:     AHDSR{AttackTime: 0.001, DecayTime: 0.015, SustainLvl: 0.0, ReleaseTime: 0.005},
			EnableFilter: FilterBandPass,
			FilterCutoff: 2400,
			FilterQ:      3.0,
		}.Generate(0.025, seed)
	},

	"sfx/pda_unlock_fanfare.wav": func(seed int64) *Buffer {
		f1 := SoundLayer{Wave1: WaveTriangle, Freq1: 587.33, Volume1: 0.7, Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05}}.Generate(0.15, seed) // D5
		f2 := SoundLayer{Wave1: WaveTriangle, Freq1: 739.99, Volume1: 0.8, Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.12, SustainLvl: 0.0, ReleaseTime: 0.05}}.Generate(0.15, seed+1) // F#5
		f3 := SoundLayer{Wave1: WaveTriangle, Freq1: 880.00, Volume1: 0.9, Envelope: AHDSR{AttackTime: 0.005, DecayTime: 0.45, SustainLvl: 0.0, ReleaseTime: 0.2}, ReverbRoom: 0.5, ReverbMix: 0.3}.Generate(0.55, seed+2) // A5

		buf := NewMonoBuffer(0.7)
		copy(buf.SamplesLeft[0:], f1.SamplesLeft)
		copy(buf.SamplesLeft[int(0.1*float64(SampleRate)):], f2.SamplesLeft)
		copy(buf.SamplesLeft[int(0.2*float64(SampleRate)):], f3.SamplesLeft)
		buf.Normalize(0.95)
		return buf
	},

	// -----------------------------------------------------------------
	// 4.7 Survival Voice Alerts (Formant Synthesizer)
	// -----------------------------------------------------------------
	"sfx/voice_o2_low.wav": func(seed int64) *Buffer {
		return SynthesizeOxygenLowVoice()
	},

	"sfx/voice_o2_critical.wav": func(seed int64) *Buffer {
		return SynthesizeOxygenCriticalVoice()
	},

	"sfx/voice_depth_warning.wav": func(seed int64) *Buffer {
		return SynthesizeDepthWarningVoice()
	},

	"sfx/voice_power_low.wav": func(seed int64) *Buffer {
		return SynthesizePowerLowVoice()
	},

	// -----------------------------------------------------------------
	// 4.8 Music, Biome Ambient Soundscapes & Cutscenes
	// -----------------------------------------------------------------
	"music/main_title.mp3": func(seed int64) *Buffer {
		return generateTitleTheme(seed)
	},

	"music/intro_cinematic.mp3": func(seed int64) *Buffer {
		return generateIntroTheme(seed)
	},

	"sfx/pod_atmospheric_entry.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoiseBrown,
			NoiseVolume:    0.95,
			Envelope:       AHDSR{AttackTime: 0.2, DecayTime: 1.2, SustainLvl: 0.6, ReleaseTime: 0.3},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   900,
			FilterQ:        2.0,
			FilterLFOFreq:  8.0,
			FilterLFODepth: 200.0,
			Overdrive:      0.5,
		}.Generate(1.8, seed)
	},

	"sfx/pod_water_crash.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  1.0,
			Wave1:        WaveSine,
			Freq1:        120,
			EndFreq1:     25,
			Volume1:      0.8,
			Envelope:     AHDSR{AttackTime: 0.01, DecayTime: 1.2, SustainLvl: 0.0, ReleaseTime: 0.4},
			EnableFilter: FilterLowPass,
			FilterCutoff: 800,
			FilterEndCut: 100,
			FilterQ:      2.5,
			Overdrive:    0.6,
			ReverbRoom:   0.8,
			ReverbMix:    0.4,
		}.Generate(1.8, seed)
	},

	"music/overworld_surface.mp3": func(seed int64) *Buffer {
		return generateOverworldTheme(seed)
	},

	"music/cave_shallow.mp3": func(seed int64) *Buffer {
		return generateShallowCaveAmbient(seed)
	},

	"music/cave_kelp.mp3": func(seed int64) *Buffer {
		return generateKelpCaveAmbient(seed)
	},

	"music/cave_volcanic.mp3": func(seed int64) *Buffer {
		return generateVolcanicCaveAmbient(seed)
	},

	"music/cave_abyssal.mp3": func(seed int64) *Buffer {
		return generateAbyssalDrone(seed)
	},

	"music/cave_wreckage.mp3": func(seed int64) *Buffer {
		return generateWreckageAmbient(seed)
	},

	"music/escape_outro.mp3": func(seed int64) *Buffer {
		return generateOutroFanfare(seed)
	},

	"sfx/rocket_ignition.wav": func(seed int64) *Buffer {
		return SoundLayer{
			Wave1:        WaveSawtooth,
			Freq1:        55,
			EndFreq1:     110,
			Volume1:      0.8,
			EnableNoise:  true,
			NoiseType:    WaveNoiseBrown,
			NoiseVolume:  0.9,
			Envelope:     AHDSR{AttackTime: 0.05, DecayTime: 0.8, SustainLvl: 0.8, ReleaseTime: 0.2},
			EnableFilter: FilterLowPass,
			FilterCutoff: 600,
			FilterQ:      2.5,
			Overdrive:    0.7,
		}.Generate(1.2, seed)
	},

	"sfx/rocket_liftoff_roar.wav": func(seed int64) *Buffer {
		return SoundLayer{
			EnableNoise:    true,
			NoiseType:      WaveNoiseBrown,
			NoiseVolume:    1.0,
			Wave1:          WaveSine,
			Freq1:          45,
			Volume1:        0.7,
			Envelope:       AHDSR{AttackTime: 0.1, DecayTime: 1.5, SustainLvl: 0.9, ReleaseTime: 0.3},
			EnableFilter:   FilterLowPass,
			FilterCutoff:   1200,
			FilterQ:        2.0,
			FilterLFOFreq:  14.0,
			FilterLFODepth: 300.0,
			Overdrive:      0.8,
		}.Generate(2.0, seed)
	},

	"music/game_over_theme.mp3": func(seed int64) *Buffer {
		return generateGameOverTheme(seed)
	},
}

// -----------------------------------------------------------------
// Procedural Ambient Music & Soundscape Synthesizers
// -----------------------------------------------------------------

func generateTitleTheme(seed int64) *Buffer {
	// Deep synth pads, evolving sub-bass, alien ocean arpeggios (6.0 seconds loop)
	dur := 6.0
	drone := SoundLayer{
		Wave1:          WaveSawtooth,
		Freq1:          65.41, // C2
		Volume1:        0.6,
		EnableOsc2:     true,
		Wave2:          WaveSine,
		Freq2:          130.81, // C3
		Detune2:        0.8,
		Volume2:        0.5,
		Envelope:       AHDSR{AttackTime: 0.8, DecayTime: 4.5, SustainLvl: 0.7, ReleaseTime: 0.7},
		EnableFilter:   FilterLowPass,
		FilterCutoff:   350,
		FilterQ:        2.5,
		FilterLFOFreq:  0.25,
		FilterLFODepth: 120.0,
		ReverbRoom:     0.8,
		ReverbMix:      0.35,
	}.Generate(dur, seed)

	pad := SoundLayer{
		Wave1:          WaveTriangle,
		Freq1:          261.63, // C4
		Volume1:        0.4,
		EnableOsc2:     true,
		Wave2:          WaveSine,
		Freq2:          392.00, // G4
		Detune2:        1.2,
		Volume2:        0.35,
		Envelope:       AHDSR{AttackTime: 1.2, DecayTime: 3.8, SustainLvl: 0.6, ReleaseTime: 1.0},
		EnableFilter:   FilterBandPass,
		FilterCutoff:   800,
		FilterQ:        3.0,
		FilterLFOFreq:  0.15,
		FilterLFODepth: 200.0,
		ReverbRoom:     0.85,
		ReverbMix:      0.45,
	}.Generate(dur, seed+1)

	return MixBuffers(drone, pad)
}

func generateIntroTheme(seed int64) *Buffer {
	dur := 4.0
	alarm := SoundLayer{
		Wave1:          WaveSquare,
		Freq1:          440,
		Duty1:          0.5,
		Volume1:        0.4,
		VibratoFreq:    6.0,
		VibratoDepth:   80.0,
		Envelope:       AHDSR{AttackTime: 0.5, DecayTime: 3.0, SustainLvl: 0.6, ReleaseTime: 0.5},
		EnableFilter:   FilterBandPass,
		FilterCutoff:   880,
		FilterQ:        3.0,
	}.Generate(dur, seed)

	wind := SoundLayer{
		EnableNoise:    true,
		NoiseType:      WaveNoiseBrown,
		NoiseVolume:    0.9,
		Envelope:       AHDSR{AttackTime: 0.3, DecayTime: 3.0, SustainLvl: 0.7, ReleaseTime: 0.7},
		EnableFilter:   FilterLowPass,
		FilterCutoff:   700,
		FilterEndCut:   200,
		FilterQ:        2.0,
		Overdrive:      0.4,
	}.Generate(dur, seed+1)

	return MixBuffers(alarm, wind)
}

func generateOverworldTheme(seed int64) *Buffer {
	dur := 6.0
	// Bright ocean synth with gentle wave noise
	waves := SoundLayer{
		EnableNoise:    true,
		NoiseType:      WaveNoisePink,
		NoiseVolume:    0.5,
		Envelope:       AHDSR{AttackTime: 1.0, DecayTime: 4.0, SustainLvl: 0.6, ReleaseTime: 1.0},
		EnableFilter:   FilterBandPass,
		FilterCutoff:   600,
		FilterQ:        1.8,
		FilterLFOFreq:  0.2,
		FilterLFODepth: 250.0,
	}.Generate(dur, seed)

	chords := SoundLayer{
		Wave1:          WaveSine,
		Freq1:          329.63, // E4
		Volume1:        0.5,
		EnableOsc2:     true,
		Wave2:          WaveTriangle,
		Freq2:          440.00, // A4
		Detune2:        0.5,
		Volume2:        0.4,
		Envelope:       AHDSR{AttackTime: 0.8, DecayTime: 4.2, SustainLvl: 0.6, ReleaseTime: 1.0},
		EnableFilter:   FilterLowPass,
		FilterCutoff:   900,
		FilterQ:        2.0,
		ReverbRoom:     0.75,
		ReverbMix:      0.35,
	}.Generate(dur, seed+1)

	return MixBuffers(waves, chords)
}

func generateShallowCaveAmbient(seed int64) *Buffer {
	dur := 6.0
	waterEcho := SoundLayer{
		EnableNoise:    true,
		NoiseType:      WaveNoiseBrown,
		NoiseVolume:    0.6,
		Envelope:       AHDSR{AttackTime: 0.8, DecayTime: 4.5, SustainLvl: 0.6, ReleaseTime: 0.7},
		EnableFilter:   FilterLowPass,
		FilterCutoff:   380,
		FilterQ:        2.0,
		FilterLFOFreq:  0.15,
		FilterLFODepth: 80.0,
	}.Generate(dur, seed)

	warmPad := SoundLayer{
		Wave1:        WaveSine,
		Freq1:        146.83, // D3
		Volume1:      0.5,
		EnableOsc2:   true,
		Wave2:        WaveTriangle,
		Freq2:        220.00, // A3
		Volume2:      0.4,
		Envelope:     AHDSR{AttackTime: 1.5, DecayTime: 3.5, SustainLvl: 0.6, ReleaseTime: 1.0},
		EnableFilter: FilterLowPass,
		FilterCutoff: 500,
		FilterQ:      2.5,
		ReverbRoom:   0.85,
		ReverbMix:    0.4,
	}.Generate(dur, seed+1)

	return MixBuffers(waterEcho, warmPad)
}

func generateKelpCaveAmbient(seed int64) *Buffer {
	dur := 6.0
	pulse := SoundLayer{
		Wave1:          WaveSine,
		Freq1:          110, // A2
		Volume1:        0.6,
		EnableOsc2:     true,
		Wave2:          WaveTriangle,
		Freq2:          164.81, // E3
		Volume2:        0.4,
		VibratoFreq:    0.3,
		VibratoDepth:   4.0,
		Envelope:       AHDSR{AttackTime: 1.0, DecayTime: 4.0, SustainLvl: 0.6, ReleaseTime: 1.0},
		EnableFilter:   FilterBandPass,
		FilterCutoff:   450,
		FilterQ:        3.5,
		FilterLFOFreq:  0.2,
		FilterLFODepth: 150.0,
		ReverbRoom:     0.8,
		ReverbMix:      0.35,
	}.Generate(dur, seed)

	return pulse
}

func generateVolcanicCaveAmbient(seed int64) *Buffer {
	dur := 6.0
	rumble := SoundLayer{
		Wave1:          WaveSawtooth,
		Freq1:          45,
		Volume1:        0.7,
		EnableNoise:    true,
		NoiseType:      WaveNoiseBrown,
		NoiseVolume:    0.8,
		Envelope:       AHDSR{AttackTime: 0.8, DecayTime: 4.5, SustainLvl: 0.7, ReleaseTime: 0.7},
		EnableFilter:   FilterLowPass,
		FilterCutoff:   180,
		FilterQ:        3.0,
		FilterLFOFreq:  0.4,
		FilterLFODepth: 60.0,
		Overdrive:      0.4,
	}.Generate(dur, seed)

	sizzle := SoundLayer{
		EnableNoise:    true,
		NoiseType:      WaveNoisePink,
		NoiseVolume:    0.4,
		Envelope:       AHDSR{AttackTime: 1.0, DecayTime: 4.0, SustainLvl: 0.5, ReleaseTime: 1.0},
		EnableFilter:   FilterBandPass,
		FilterCutoff:   1400,
		FilterQ:        3.0,
		FilterLFOFreq:  2.0,
		FilterLFODepth: 300.0,
	}.Generate(dur, seed+1)

	return MixBuffers(rumble, sizzle)
}

func generateAbyssalDrone(seed int64) *Buffer {
	dur := 6.0
	drone := SoundLayer{
		Wave1:          WaveSine,
		Freq1:          36.71, // D1 (Sub-bass)
		Volume1:        0.8,
		EnableOsc2:     true,
		Wave2:          WaveTriangle,
		Freq2:          55.00, // A1
		Detune2:        0.3,
		Volume2:        0.6,
		Envelope:       AHDSR{AttackTime: 1.5, DecayTime: 3.5, SustainLvl: 0.8, ReleaseTime: 1.0},
		EnableFilter:   FilterLowPass,
		FilterCutoff:   140,
		FilterQ:        3.5,
		FilterLFOFreq:  0.1,
		FilterLFODepth: 40.0,
		ReverbRoom:     0.95,
		ReverbMix:      0.5,
	}.Generate(dur, seed)

	creak := SoundLayer{
		Wave1:        WaveSawtooth,
		Freq1:        98.0,
		EndFreq1:     65.41,
		Volume1:      0.3,
		Envelope:     AHDSR{AttackTime: 2.0, DecayTime: 3.0, SustainLvl: 0.2, ReleaseTime: 1.0},
		EnableFilter: FilterBandPass,
		FilterCutoff: 300,
		FilterQ:      6.0,
		ReverbRoom:   0.9,
		ReverbMix:    0.6,
	}.Generate(dur, seed+1)

	return MixBuffers(drone, creak)
}

func generateWreckageAmbient(seed int64) *Buffer {
	dur := 6.0
	hull := SoundLayer{
		Wave1:          WaveSawtooth,
		Freq1:          82.41, // E2
		Volume1:        0.5,
		EnableNoise:    true,
		NoiseType:      WaveNoiseBrown,
		NoiseVolume:    0.4,
		Envelope:       AHDSR{AttackTime: 1.2, DecayTime: 4.0, SustainLvl: 0.6, ReleaseTime: 0.8},
		EnableFilter:   FilterBandPass,
		FilterCutoff:   250,
		FilterQ:        4.0,
		FilterLFOFreq:  0.2,
		FilterLFODepth: 80.0,
		ReverbRoom:     0.9,
		ReverbMix:      0.45,
	}.Generate(dur, seed)

	return hull
}

func generateOutroFanfare(seed int64) *Buffer {
	dur := 5.0
	horn := SoundLayer{
		Wave1:        WaveSawtooth,
		Freq1:        261.63, // C4
		Volume1:      0.7,
		EnableOsc2:   true,
		Wave2:        WaveTriangle,
		Freq2:        523.25, // C5
		Volume2:      0.5,
		Envelope:     AHDSR{AttackTime: 0.1, DecayTime: 3.5, SustainLvl: 0.6, ReleaseTime: 1.4},
		EnableFilter: FilterLowPass,
		FilterCutoff: 1200,
		FilterQ:      2.5,
		ReverbRoom:   0.8,
		ReverbMix:    0.35,
	}.Generate(dur, seed)

	return horn
}

func generateGameOverTheme(seed int64) *Buffer {
	dur := 5.0
	return SoundLayer{
		Wave1:        WaveSine,
		Freq1:        65.41, // C2
		EndFreq1:     32.70, // C1
		Volume1:      0.8,
		EnableOsc2:   true,
		Wave2:        WaveTriangle,
		Freq2:        98.0, // G2
		EndFreq2:     49.0, // G1
		Volume2:      0.5,
		Envelope:     AHDSR{AttackTime: 0.8, DecayTime: 3.5, SustainLvl: 0.0, ReleaseTime: 0.7, Curve: CurveExponential},
		EnableFilter: FilterLowPass,
		FilterCutoff: 200,
		FilterEndCut: 40,
		FilterQ:      2.0,
		ReverbRoom:   0.9,
		ReverbMix:    0.45,
	}.Generate(dur, seed)
}
