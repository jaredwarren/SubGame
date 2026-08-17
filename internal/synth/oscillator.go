package synth

import (
	"math"
	"math/rand"
)

// Waveform represents the basic oscillator generator type.
type Waveform int

const (
	WaveSine Waveform = iota
	WaveTriangle
	WaveSquare
	WaveSawtooth
	WaveNoiseWhite
	WaveNoisePink
	WaveNoiseBrown
)

// Oscillator produces raw wave samples with frequency, phase accumulation, and noise generation.
type Oscillator struct {
	Type       Waveform
	Frequency  float64
	Phase      float64
	DutyCycle  float64 // For Square wave (default 0.5)
	SampleRate float64

	// Pink noise filter state (Paul Kellet 3-pole filter)
	b0, b1, b2 float64

	// Brownian noise state (integrated random walk with leak)
	brownLast float64

	rng *rand.Rand
}

// NewOscillator creates a new oscillator configured for the given waveform and frequency.
func NewOscillator(wave Waveform, freq float64, seed int64) *Oscillator {
	if seed == 0 {
		seed = 42
	}
	return &Oscillator{
		Type:       wave,
		Frequency:  freq,
		Phase:      0.0,
		DutyCycle:  0.5,
		SampleRate: float64(SampleRate),
		rng:        rand.New(rand.NewSource(seed)),
	}
}

// Next generates the next float64 sample in range [-1.0, 1.0].
func (o *Oscillator) Next(freqOffset float64) float64 {
	curFreq := o.Frequency + freqOffset
	if curFreq < 0.1 {
		curFreq = 0.1
	}

	var sample float64
	switch o.Type {
	case WaveSine:
		sample = math.Sin(2.0 * math.Pi * o.Phase)

	case WaveTriangle:
		// Triangle wave: goes from -1 to 1 to -1 over phase [0, 1)
		p := o.Phase
		if p < 0.25 {
			sample = 4.0 * p
		} else if p < 0.75 {
			sample = 2.0 - 4.0*p
		} else {
			sample = 4.0*p - 4.0
		}

	case WaveSquare:
		duty := o.DutyCycle
		if duty <= 0.01 {
			duty = 0.01
		} else if duty >= 0.99 {
			duty = 0.99
		}
		if o.Phase < duty {
			sample = 1.0
		} else {
			sample = -1.0
		}

	case WaveSawtooth:
		// Sawtooth wave: goes from -1 to +1 over phase [0, 1)
		sample = 2.0*o.Phase - 1.0

	case WaveNoiseWhite:
		sample = o.rng.Float64()*2.0 - 1.0

	case WaveNoisePink:
		// Paul Kellet's filter for pink noise (-3dB/octave)
		white := o.rng.Float64()*2.0 - 1.0
		o.b0 = 0.99765*o.b0 + white*0.0990460
		o.b1 = 0.96300*o.b1 + white*0.1384000
		o.b2 = 0.57000*o.b2 + white*0.4368000
		pink := o.b0 + o.b1 + o.b2 + white*0.5362
		sample = pink * 0.12 // scale to approximate [-1, 1]

	case WaveNoiseBrown:
		// Brownian / Red noise (-6dB/octave integrated random walk with decay)
		white := o.rng.Float64()*2.0 - 1.0
		o.brownLast = (o.brownLast + (0.04 * white)) / 1.02
		sample = o.brownLast * 3.5
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
	}

	// Advance phase
	phaseIncr := curFreq / o.SampleRate
	o.Phase += phaseIncr
	if o.Phase >= 1.0 {
		o.Phase -= math.Floor(o.Phase)
	}

	return sample
}

// Reset resets phase and noise state.
func (o *Oscillator) Reset() {
	o.Phase = 0.0
	o.b0 = 0
	o.b1 = 0
	o.b2 = 0
	o.brownLast = 0
}
