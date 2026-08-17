package synth

import (
	"math"
)

// SoundLayer defines a single synthesized audio layer.
type SoundLayer struct {
	// Oscillator 1
	Wave1      Waveform
	Freq1      float64
	EndFreq1   float64 // If > 0, glides from Freq1 to EndFreq1
	FreqCurve1 CurveType
	Duty1      float64 // For square wave
	Volume1    float64

	// Oscillator 2 (Sub-oscillator or harmonic)
	EnableOsc2 bool
	Wave2      Waveform
	Freq2      float64
	EndFreq2   float64
	Duty2      float64
	Volume2    float64
	Detune2    float64 // Frequency offset in Hz

	// Noise generator
	EnableNoise bool
	NoiseType   Waveform // WaveNoiseWhite, WaveNoisePink, WaveNoiseBrown
	NoiseVolume float64

	// Amplitude Envelope
	Envelope AHDSR

	// Filter
	EnableFilter FilterType
	FilterCutoff float64
	FilterEndCut float64 // If > 0, sweeps cutoff
	FilterQ      float64
	FilterLFOFreq float64 // LFO modulation speed in Hz
	FilterLFODepth float64 // LFO modulation depth in Hz

	// Frequency LFO (Vibrato / FM)
	VibratoFreq  float64 // Hz
	VibratoDepth float64 // Hz

	// Effects
	BitDepth         int     // 0 = disabled, 2-16
	Downsample       int     // 1 = disabled
	Overdrive        float64 // 0.0 to 2.0+
	DelayTime        float64 // Seconds
	DelayFeedback    float64
	DelayMix         float64
	ReverbRoom       float64
	ReverbMix        float64
	Pan              float64 // -1.0 to +1.0
	NormalizePeak    float64 // default 0.95
}

// Generate synthesizes the SoundLayer into an audio Buffer.
func (l SoundLayer) Generate(duration float64, seed int64) *Buffer {
	if duration <= 0.001 {
		duration = l.Envelope.TotalDuration()
	}
	if duration <= 0.001 {
		duration = 0.1
	}

	buf := NewMonoBuffer(duration)
	sr := float64(buf.SampleRate)
	numSamples := len(buf.SamplesLeft)

	osc1 := NewOscillator(l.Wave1, l.Freq1, seed)
	osc1.DutyCycle = l.Duty1

	var osc2 *Oscillator
	if l.EnableOsc2 {
		osc2 = NewOscillator(l.Wave2, l.Freq2+l.Detune2, seed+1)
		osc2.DutyCycle = l.Duty2
	}

	var noiseOsc *Oscillator
	if l.EnableNoise {
		noiseOsc = NewOscillator(l.NoiseType, 0, seed+2)
	}

	var filter *BiquadFilter
	if l.FilterCutoff > 0 {
		filter = NewBiquadFilter(l.EnableFilter, l.FilterCutoff, l.FilterQ)
	}

	freqEnv1 := PitchEnvelope{
		StartFreq: l.Freq1,
		EndFreq:   l.EndFreq1,
		Duration:  duration,
		Curve:     l.FreqCurve1,
	}
	if l.EndFreq1 <= 0 {
		freqEnv1.EndFreq = l.Freq1
	}

	freqEnv2 := PitchEnvelope{
		StartFreq: l.Freq2 + l.Detune2,
		EndFreq:   l.EndFreq2 + l.Detune2,
		Duration:  duration,
		Curve:     l.FreqCurve1,
	}
	if l.EndFreq2 <= 0 {
		freqEnv2.EndFreq = l.Freq2 + l.Detune2
	}

	for i := 0; i < numSamples; i++ {
		t := float64(i) / sr

		// Vibrato / FM modulation
		var vibOffset float64
		if l.VibratoFreq > 0 && l.VibratoDepth > 0 {
			vibOffset = math.Sin(2.0*math.Pi*l.VibratoFreq*t) * l.VibratoDepth
		}

		// Update frequencies
		f1 := freqEnv1.FreqAt(t) + vibOffset
		osc1.Frequency = f1

		var sample float64
		// Osc 1
		v1 := l.Volume1
		if v1 <= 0 {
			v1 = 1.0
		}
		sample += osc1.Next(0) * v1

		// Osc 2
		if osc2 != nil {
			f2 := freqEnv2.FreqAt(t) + vibOffset
			osc2.Frequency = f2
			sample += osc2.Next(0) * l.Volume2
		}

		// Noise
		if noiseOsc != nil {
			sample += noiseOsc.Next(0) * l.NoiseVolume
		}

		// Amplitude envelope
		amp := l.Envelope.ValueAt(t)
		sample *= amp

		// Dynamic filter modulation
		if filter != nil {
			curCutoff := l.FilterCutoff
			if l.FilterEndCut > 0 {
				cutProgress := t / duration
				curCutoff = l.FilterCutoff + (l.FilterEndCut-l.FilterCutoff)*cutProgress
			}
			if l.FilterLFOFreq > 0 && l.FilterLFODepth > 0 {
				curCutoff += math.Sin(2.0*math.Pi*l.FilterLFOFreq*t) * l.FilterLFODepth
			}
			filter.SetParams(curCutoff, l.FilterQ)
			sample = filter.ProcessSample(sample)
		}

		buf.SamplesLeft[i] = sample
	}

	// Apply post-processing effects
	if l.BitDepth > 0 || l.Downsample > 1 {
		ApplyBitcrush(buf, l.BitDepth, l.Downsample)
	}
	if l.Overdrive > 0 {
		ApplyOverdrive(buf, l.Overdrive)
	}
	if l.DelayTime > 0 && l.DelayMix > 0 {
		ApplyDelay(buf, l.DelayTime, l.DelayFeedback, l.DelayMix)
	}
	if l.ReverbMix > 0 {
		ApplyReverb(buf, l.ReverbRoom, 0.3, l.ReverbMix)
	}

	peak := l.NormalizePeak
	if peak <= 0 {
		peak = 0.95
	}
	buf.Normalize(peak)

	if l.Pan != 0.0 {
		return ConvertToStereoPan(buf, l.Pan)
	}

	return buf
}

// MixBuffers sums multiple audio buffers together into a single normalized buffer.
func MixBuffers(buffers ...*Buffer) *Buffer {
	if len(buffers) == 0 {
		return NewMonoBuffer(0.1)
	}
	maxLen := 0
	isStereo := false
	for _, b := range buffers {
		if len(b.SamplesLeft) > maxLen {
			maxLen = len(b.SamplesLeft)
		}
		if b.IsStereo {
			isStereo = true
		}
	}

	var out *Buffer
	if isStereo {
		out = NewStereoBuffer(float64(maxLen) / float64(SampleRate))
	} else {
		out = NewMonoBuffer(float64(maxLen) / float64(SampleRate))
	}

	for _, b := range buffers {
		for i := range b.SamplesLeft {
			out.SamplesLeft[i] += b.SamplesLeft[i]
		}
		if isStereo {
			for i := range b.SamplesLeft {
				if b.IsStereo {
					out.SamplesRight[i] += b.SamplesRight[i]
				} else {
					out.SamplesRight[i] += b.SamplesLeft[i]
				}
			}
		}
	}

	out.Normalize(0.95)
	return out
}
