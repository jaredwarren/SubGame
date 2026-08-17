package synth

import "math"

// ApplyBitcrush reduces sample bit depth and simulates retro DAC quantization.
func ApplyBitcrush(buf *Buffer, bitDepth int, downsampleFactor int) {
	if bitDepth < 2 {
		bitDepth = 2
	}
	if bitDepth > 16 {
		bitDepth = 16
	}
	if downsampleFactor < 1 {
		downsampleFactor = 1
	}

	levels := math.Pow(2, float64(bitDepth-1))

	var lastL, lastR float64
	for i := range buf.SamplesLeft {
		if i%downsampleFactor == 0 {
			lastL = math.Round(buf.SamplesLeft[i]*levels) / levels
			if buf.IsStereo {
				lastR = math.Round(buf.SamplesRight[i]*levels) / levels
			}
		}
		buf.SamplesLeft[i] = lastL
		if buf.IsStereo {
			buf.SamplesRight[i] = lastR
		}
	}
}

// ApplyOverdrive adds warm non-linear saturation / tube overdrive.
func ApplyOverdrive(buf *Buffer, drive float64) {
	if drive <= 0.001 {
		return
	}
	for i := range buf.SamplesLeft {
		x := buf.SamplesLeft[i] * (1.0 + drive)
		buf.SamplesLeft[i] = math.Tanh(x)
	}
	if buf.IsStereo {
		for i := range buf.SamplesRight {
			x := buf.SamplesRight[i] * (1.0 + drive)
			buf.SamplesRight[i] = math.Tanh(x)
		}
	}
}

// ApplyDelay adds feedback echo / delay.
func ApplyDelay(buf *Buffer, delayTimeSeconds float64, feedback float64, mix float64) {
	if delayTimeSeconds <= 0 || mix <= 0 {
		return
	}
	delaySamples := int(delayTimeSeconds * float64(buf.SampleRate))
	if delaySamples <= 0 || delaySamples >= len(buf.SamplesLeft) {
		return
	}

	delayBufL := make([]float64, delaySamples)
	delayBufR := make([]float64, delaySamples)
	idx := 0

	for i := range buf.SamplesLeft {
		inL := buf.SamplesLeft[i]
		delayedL := delayBufL[idx]
		delayBufL[idx] = inL + delayedL*feedback
		buf.SamplesLeft[i] = inL*(1.0-mix) + delayedL*mix

		if buf.IsStereo {
			inR := buf.SamplesRight[i]
			delayedR := delayBufR[idx]
			delayBufR[idx] = inR + delayedR*feedback
			buf.SamplesRight[i] = inR*(1.0-mix) + delayedR*mix
		}

		idx++
		if idx >= delaySamples {
			idx = 0
		}
	}
}

// ApplyReverb adds a multi-tap algorithmic Schroeder reverb for spatial depth.
func ApplyReverb(buf *Buffer, roomSize float64, damping float64, mix float64) {
	if mix <= 0 {
		return
	}
	if roomSize <= 0.05 {
		roomSize = 0.5
	}
	// Comb filter delays in milliseconds
	combDelaysMS := []float64{29.7, 37.1, 41.1, 43.7}
	allPassDelaysMS := []float64{5.1, 12.6}

	combBuffers := make([][]float64, len(combDelaysMS))
	combIndices := make([]int, len(combDelaysMS))
	for c, ms := range combDelaysMS {
		d := int((ms / 1000.0) * float64(buf.SampleRate) * roomSize)
		if d < 1 {
			d = 1
		}
		combBuffers[c] = make([]float64, d)
	}

	allPassBuffers := make([][]float64, len(allPassDelaysMS))
	allPassIndices := make([]int, len(allPassDelaysMS))
	for a, ms := range allPassDelaysMS {
		d := int((ms / 1000.0) * float64(buf.SampleRate))
		if d < 1 {
			d = 1
		}
		allPassBuffers[a] = make([]float64, d)
	}

	fb := 0.7 * roomSize
	if fb > 0.95 {
		fb = 0.95
	}

	for i := range buf.SamplesLeft {
		in := buf.SamplesLeft[i]

		// Parallel Comb filters
		var combSum float64
		for c := range combBuffers {
			idx := combIndices[c]
			delayed := combBuffers[c][idx]
			combBuffers[c][idx] = in + delayed*fb*(1.0-damping)
			combSum += delayed
			combIndices[c] = (idx + 1) % len(combBuffers[c])
		}
		combSum *= 0.25

		// Series All-Pass filters
		apOut := combSum
		for a := range allPassBuffers {
			idx := allPassIndices[a]
			delayed := allPassBuffers[a][idx]
			apIn := apOut
			apOut = -0.5*apIn + delayed
			allPassBuffers[a][idx] = apIn + 0.5*delayed
			allPassIndices[a] = (idx + 1) % len(allPassBuffers[a])
		}

		buf.SamplesLeft[i] = in*(1.0-mix) + apOut*mix
		if buf.IsStereo {
			buf.SamplesRight[i] = buf.SamplesRight[i]*(1.0-mix) + apOut*mix
		}
	}
}

// ApplyStereoPan pans a mono buffer into stereo with angle [-1.0 (Left), +1.0 (Right)].
func ConvertToStereoPan(buf *Buffer, pan float64) *Buffer {
	if pan < -1.0 {
		pan = -1.0
	}
	if pan > 1.0 {
		pan = 1.0
	}

	st := NewStereoBuffer(buf.Duration())
	// Constant power panning
	angle := (pan + 1.0) * (math.Pi / 4.0) // 0 to Pi/2
	gainL := math.Cos(angle)
	gainR := math.Sin(angle)

	for i := range buf.SamplesLeft {
		s := buf.SamplesLeft[i]
		st.SamplesLeft[i] = s * gainL
		st.SamplesRight[i] = s * gainR
	}
	return st
}
