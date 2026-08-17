package synth

// FormantTriplet holds the first three formant resonance frequencies in Hz.
type FormantTriplet struct {
	F1 float64
	F2 float64
	F3 float64
	Q1 float64
	Q2 float64
	Q3 float64
}

// Standard vowel formant profiles (male/neutral robotic register)
var (
	FormantA = FormantTriplet{F1: 730, F2: 1090, F3: 2440, Q1: 6.0, Q2: 8.0, Q3: 10.0}
	FormantE = FormantTriplet{F1: 530, F2: 1840, F3: 2480, Q1: 6.0, Q2: 9.0, Q3: 10.0}
	FormantI = FormantTriplet{F1: 270, F2: 2290, F3: 3010, Q1: 5.0, Q2: 10.0, Q3: 12.0}
	FormantO = FormantTriplet{F1: 570, F2: 840, F3: 2410, Q1: 6.0, Q2: 7.0, Q3: 10.0}
	FormantU = FormantTriplet{F1: 300, F2: 870, F3: 2240, Q1: 5.0, Q2: 7.0, Q3: 10.0}
	FormantR = FormantTriplet{F1: 490, F2: 1350, F3: 1690, Q1: 5.0, Q2: 8.0, Q3: 10.0}
	FormantN = FormantTriplet{F1: 250, F2: 1800, F3: 2500, Q1: 8.0, Q2: 12.0, Q3: 14.0}
	FormantL = FormantTriplet{F1: 380, F2: 1200, F3: 2700, Q1: 6.0, Q2: 8.0, Q3: 10.0}
)

// Phoneme represents a segment of synthetic speech.
type Phoneme struct {
	Duration float64
	Formants FormantTriplet
	Pitch    float64 // Fundamental frequency (F0) in Hz (e.g. 110Hz - 160Hz for robotic voice)
	IsNoise  bool    // True for fricatives (S, SH, TH)
	Volume   float64
}

// SynthesizePhonemes synthesizes a sequence of phonemes into an audio buffer with smooth transitions.
func SynthesizePhonemes(phonemes []Phoneme) *Buffer {
	totalDur := 0.0
	for _, p := range phonemes {
		totalDur += p.Duration
	}
	if totalDur <= 0 {
		return NewMonoBuffer(0.1)
	}

	buf := NewMonoBuffer(totalDur)
	sr := float64(buf.SampleRate)

	// Filter banks
	bp1 := NewBiquadFilter(FilterBandPass, 500, 6.0)
	bp2 := NewBiquadFilter(FilterBandPass, 1200, 8.0)
	bp3 := NewBiquadFilter(FilterBandPass, 2400, 10.0)

	carrierOsc := NewOscillator(WaveSawtooth, 120, 42)
	noiseOsc := NewOscillator(WaveNoiseWhite, 0, 42)

	sampleIdx := 0
	for _, p := range phonemes {
		numFrames := int(p.Duration * sr)
		carrierOsc.Frequency = p.Pitch

		bp1.SetParams(p.Formants.F1, p.Formants.Q1)
		bp2.SetParams(p.Formants.F2, p.Formants.Q2)
		bp3.SetParams(p.Formants.F3, p.Formants.Q3)

		// Envelope: soft ramp in and out to prevent clicks
		attackSamples := int(0.015 * sr)
		releaseSamples := int(0.020 * sr)

		for i := 0; i < numFrames && sampleIdx < len(buf.SamplesLeft); i++ {
			var raw float64
			if p.IsNoise {
				raw = noiseOsc.Next(0) * 0.7
			} else {
				// Glottal pulse approximation: sawtooth + pulse
				raw = carrierOsc.Next(0)
			}

			// Run through parallel formant bandpass filters
			f1Out := bp1.ProcessSample(raw) * 1.0
			f2Out := bp2.ProcessSample(raw) * 0.7
			f3Out := bp3.ProcessSample(raw) * 0.4
			voiced := (f1Out + f2Out + f3Out) * p.Volume

			// Windowing envelope
			env := 1.0
			if i < attackSamples && attackSamples > 0 {
				env = float64(i) / float64(attackSamples)
			} else if i > numFrames-releaseSamples && releaseSamples > 0 {
				env = float64(numFrames-i) / float64(releaseSamples)
			}

			buf.SamplesLeft[sampleIdx] = voiced * env
			sampleIdx++
		}
	}

	// Warm robotic saturation and light reverb
	ApplyOverdrive(buf, 0.25)
	ApplyReverb(buf, 0.4, 0.3, 0.15)
	buf.Normalize(0.95)

	return buf
}

// SynthesizeOxygenLowVoice produces the synthetic AI voice: "Oxygen low."
func SynthesizeOxygenLowVoice() *Buffer {
	// "Ox - y - gen - low"
	phonemes := []Phoneme{
		// Initial warning chirp
		{Duration: 0.08, Formants: FormantI, Pitch: 880, Volume: 0.4},
		{Duration: 0.06, Formants: FormantI, Pitch: 110, Volume: 0.0}, // pause

		// "Ox" -> O + K/S
		{Duration: 0.14, Formants: FormantO, Pitch: 155, Volume: 0.9},
		{Duration: 0.08, Formants: FormantTriplet{F1: 3000, F2: 4500, F3: 6000, Q1: 4, Q2: 5, Q3: 5}, Pitch: 150, IsNoise: true, Volume: 0.5}, // ks

		// "y" -> I
		{Duration: 0.10, Formants: FormantI, Pitch: 150, Volume: 0.8},

		// "gen" -> E + N
		{Duration: 0.12, Formants: FormantE, Pitch: 145, Volume: 0.85},
		{Duration: 0.12, Formants: FormantN, Pitch: 140, Volume: 0.8},

		// Short pause
		{Duration: 0.10, Formants: FormantTriplet{}, Pitch: 130, Volume: 0.0},

		// "low" -> L + O + U
		{Duration: 0.10, Formants: FormantL, Pitch: 135, Volume: 0.85},
		{Duration: 0.20, Formants: FormantO, Pitch: 130, Volume: 0.95},
		{Duration: 0.15, Formants: FormantU, Pitch: 120, Volume: 0.8},
	}
	return SynthesizePhonemes(phonemes)
}

// SynthesizeOxygenCriticalVoice produces the synthetic AI voice: "Warning: Oxygen critical."
func SynthesizeOxygenCriticalVoice() *Buffer {
	phonemes := []Phoneme{
		// High emergency double beep
		{Duration: 0.07, Formants: FormantI, Pitch: 950, Volume: 0.6},
		{Duration: 0.04, Formants: FormantI, Pitch: 100, Volume: 0.0},
		{Duration: 0.07, Formants: FormantI, Pitch: 950, Volume: 0.6},
		{Duration: 0.08, Formants: FormantI, Pitch: 100, Volume: 0.0},

		// "War - ning"
		{Duration: 0.14, Formants: FormantO, Pitch: 165, Volume: 0.9},
		{Duration: 0.12, Formants: FormantN, Pitch: 155, Volume: 0.85},
		{Duration: 0.06, Formants: FormantTriplet{}, Pitch: 100, Volume: 0.0},

		// "Ox - y - gen"
		{Duration: 0.12, Formants: FormantO, Pitch: 165, Volume: 0.9},
		{Duration: 0.06, Formants: FormantTriplet{F1: 3000, F2: 4500, F3: 6000, Q1: 4, Q2: 5, Q3: 5}, Pitch: 160, IsNoise: true, Volume: 0.4},
		{Duration: 0.08, Formants: FormantI, Pitch: 160, Volume: 0.8},
		{Duration: 0.10, Formants: FormantE, Pitch: 155, Volume: 0.85},
		{Duration: 0.08, Formants: FormantN, Pitch: 150, Volume: 0.8},

		// "Cri - ti - cal"
		{Duration: 0.06, Formants: FormantTriplet{F1: 2500, F2: 3800, F3: 5000, Q1: 4, Q2: 5, Q3: 5}, Pitch: 150, IsNoise: true, Volume: 0.5},
		{Duration: 0.12, Formants: FormantI, Pitch: 160, Volume: 0.9},
		{Duration: 0.08, Formants: FormantI, Pitch: 150, Volume: 0.8},
		{Duration: 0.14, Formants: FormantA, Pitch: 140, Volume: 0.85},
		{Duration: 0.12, Formants: FormantL, Pitch: 130, Volume: 0.8},
	}
	return SynthesizePhonemes(phonemes)
}

// SynthesizeDepthWarningVoice produces the synthetic AI voice: "Warning: Depth limit exceeded."
func SynthesizeDepthWarningVoice() *Buffer {
	phonemes := []Phoneme{
		// Emergency beep
		{Duration: 0.09, Formants: FormantI, Pitch: 880, Volume: 0.5},
		{Duration: 0.06, Formants: FormantI, Pitch: 100, Volume: 0.0},

		// "War - ning"
		{Duration: 0.14, Formants: FormantO, Pitch: 150, Volume: 0.9},
		{Duration: 0.12, Formants: FormantN, Pitch: 140, Volume: 0.85},
		{Duration: 0.08, Formants: FormantTriplet{}, Pitch: 100, Volume: 0.0},

		// "Depth" -> E + F/TH
		{Duration: 0.16, Formants: FormantE, Pitch: 145, Volume: 0.95},
		{Duration: 0.07, Formants: FormantTriplet{F1: 3000, F2: 4500, F3: 6000, Q1: 4, Q2: 5, Q3: 5}, Pitch: 140, IsNoise: true, Volume: 0.45},

		// "Li - mit"
		{Duration: 0.10, Formants: FormantL, Pitch: 140, Volume: 0.85},
		{Duration: 0.12, Formants: FormantI, Pitch: 135, Volume: 0.9},

		// "Ex - cee - ded"
		{Duration: 0.08, Formants: FormantE, Pitch: 140, Volume: 0.85},
		{Duration: 0.06, Formants: FormantTriplet{F1: 3500, F2: 5000, F3: 6500, Q1: 4, Q2: 5, Q3: 5}, Pitch: 140, IsNoise: true, Volume: 0.4},
		{Duration: 0.16, Formants: FormantI, Pitch: 135, Volume: 0.95},
		{Duration: 0.14, Formants: FormantE, Pitch: 125, Volume: 0.8},
	}
	return SynthesizePhonemes(phonemes)
}

// SynthesizePowerLowVoice produces the synthetic AI voice: "Warning: Power low."
func SynthesizePowerLowVoice() *Buffer {
	phonemes := []Phoneme{
		// Low alarm ping
		{Duration: 0.09, Formants: FormantI, Pitch: 660, Volume: 0.5},
		{Duration: 0.05, Formants: FormantI, Pitch: 100, Volume: 0.0},

		// "War - ning"
		{Duration: 0.14, Formants: FormantO, Pitch: 145, Volume: 0.9},
		{Duration: 0.12, Formants: FormantN, Pitch: 135, Volume: 0.85},
		{Duration: 0.06, Formants: FormantTriplet{}, Pitch: 100, Volume: 0.0},

		// "Pow - er"
		{Duration: 0.16, Formants: FormantA, Pitch: 140, Volume: 0.95},
		{Duration: 0.12, Formants: FormantU, Pitch: 135, Volume: 0.85},
		{Duration: 0.10, Formants: FormantR, Pitch: 130, Volume: 0.8},

		// "Low"
		{Duration: 0.10, Formants: FormantL, Pitch: 130, Volume: 0.85},
		{Duration: 0.18, Formants: FormantO, Pitch: 125, Volume: 0.95},
		{Duration: 0.12, Formants: FormantU, Pitch: 115, Volume: 0.8},
	}
	return SynthesizePhonemes(phonemes)
}
