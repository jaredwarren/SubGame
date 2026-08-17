package synth

import "math"

// FilterType defines the biquad filter mode.
type FilterType int

const (
	FilterLowPass FilterType = iota
	FilterHighPass
	FilterBandPass
	FilterNotch
)

// BiquadFilter implements a digital 2-pole resonant filter based on Robert Bristow-Johnson's Audio EQ Cookbook.
type BiquadFilter struct {
	Type       FilterType
	Cutoff     float64 // Cutoff frequency in Hz
	Q          float64 // Resonance Q factor (0.707 = Butterworth flat, >1 = resonant peak)
	SampleRate float64

	// Coefficients
	b0, b1, b2 float64
	a1, a2     float64

	// State memory
	x1, x2 float64 // Input history
	y1, y2 float64 // Output history
}

// NewBiquadFilter creates a new biquad filter.
func NewBiquadFilter(ft FilterType, cutoff float64, q float64) *BiquadFilter {
	if q <= 0.01 {
		q = 0.707
	}
	f := &BiquadFilter{
		Type:       ft,
		Cutoff:     cutoff,
		Q:          q,
		SampleRate: float64(SampleRate),
	}
	f.recalculate()
	return f
}

// SetParams updates cutoff and Q and recalculates filter coefficients.
func (f *BiquadFilter) SetParams(cutoff float64, q float64) {
	if cutoff < 10.0 {
		cutoff = 10.0
	}
	if cutoff > f.SampleRate*0.48 {
		cutoff = f.SampleRate * 0.48
	}
	if q <= 0.01 {
		q = 0.01
	}
	f.Cutoff = cutoff
	f.Q = q
	f.recalculate()
}

func (f *BiquadFilter) recalculate() {
	omega := 2.0 * math.Pi * f.Cutoff / f.SampleRate
	sn := math.Sin(omega)
	cs := math.Cos(omega)
	alpha := sn / (2.0 * f.Q)

	var b0, b1, b2, a0, a1, a2 float64

	switch f.Type {
	case FilterLowPass:
		b0 = (1.0 - cs) / 2.0
		b1 = 1.0 - cs
		b2 = (1.0 - cs) / 2.0
		a0 = 1.0 + alpha
		a1 = -2.0 * cs
		a2 = 1.0 - alpha

	case FilterHighPass:
		b0 = (1.0 + cs) / 2.0
		b1 = -(1.0 + cs)
		b2 = (1.0 + cs) / 2.0
		a0 = 1.0 + alpha
		a1 = -2.0 * cs
		a2 = 1.0 - alpha

	case FilterBandPass:
		b0 = alpha
		b1 = 0.0
		b2 = -alpha
		a0 = 1.0 + alpha
		a1 = -2.0 * cs
		a2 = 1.0 - alpha

	case FilterNotch:
		b0 = 1.0
		b1 = -2.0 * cs
		b2 = 1.0
		a0 = 1.0 + alpha
		a1 = -2.0 * cs
		a2 = 1.0 - alpha
	}

	// Normalize coefficients by a0
	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

// ProcessSample runs a single sample through the filter.
func (f *BiquadFilter) ProcessSample(in float64) float64 {
	out := f.b0*in + f.b1*f.x1 + f.b2*f.x2 - f.a1*f.y1 - f.a2*f.y2

	// Check for NaN or Inf blowup
	if math.IsNaN(out) || math.IsInf(out, 0) {
		out = 0.0
		f.Reset()
	}

	// Shift history
	f.x2 = f.x1
	f.x1 = in
	f.y2 = f.y1
	f.y1 = out

	return out
}

// Reset clears filter state.
func (f *BiquadFilter) Reset() {
	f.x1 = 0
	f.x2 = 0
	f.y1 = 0
	f.y2 = 0
}

// ProcessBuffer processes an entire audio buffer in-place.
func (f *BiquadFilter) ProcessBuffer(buf *Buffer) {
	for i := range buf.SamplesLeft {
		buf.SamplesLeft[i] = f.ProcessSample(buf.SamplesLeft[i])
	}
	if buf.IsStereo {
		f.Reset()
		for i := range buf.SamplesRight {
			buf.SamplesRight[i] = f.ProcessSample(buf.SamplesRight[i])
		}
	}
}
