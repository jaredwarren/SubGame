package synth

import "math"

// CurveType specifies the interpolation curve for envelopes.
type CurveType int

const (
	CurveLinear CurveType = iota
	CurveExponential
	CurveLogarithmic
)

// AHDSR represents an amplitude envelope with Attack, Hold, Decay, Sustain, and Release stages.
type AHDSR struct {
	AttackTime  float64 // Duration in seconds to ramp from 0 to PeakLevel
	HoldTime    float64 // Duration in seconds to stay at PeakLevel
	DecayTime   float64 // Duration in seconds to ramp from PeakLevel to SustainLevel
	SustainLvl  float64 // Sustain amplitude (0.0 to 1.0)
	ReleaseTime float64 // Duration in seconds to ramp from SustainLevel to 0
	PeakLevel   float64 // Maximum amplitude level (default 1.0)
	Curve       CurveType
}

// DefaultAHDSR creates a standard percussive envelope.
func DefaultAHDSR() AHDSR {
	return AHDSR{
		AttackTime:  0.01,
		HoldTime:    0.0,
		DecayTime:   0.15,
		SustainLvl:  0.0,
		ReleaseTime: 0.05,
		PeakLevel:   1.0,
		Curve:       CurveExponential,
	}
}

// TotalDuration returns the total length of the envelope in seconds.
func (e AHDSR) TotalDuration() float64 {
	return e.AttackTime + e.HoldTime + e.DecayTime + e.ReleaseTime
}

// ValueAt returns the envelope multiplier [0.0, PeakLevel] at time t (seconds).
func (e AHDSR) ValueAt(t float64) float64 {
	if t < 0 {
		return 0.0
	}
	peak := e.PeakLevel
	if peak <= 0 {
		peak = 1.0
	}

	// 1. Attack
	if t < e.AttackTime {
		if e.AttackTime <= 0.0001 {
			return peak
		}
		progress := t / e.AttackTime
		return peak * applyCurve(progress, e.Curve)
	}
	t -= e.AttackTime

	// 2. Hold
	if t < e.HoldTime {
		return peak
	}
	t -= e.HoldTime

	// 3. Decay
	if t < e.DecayTime {
		if e.DecayTime <= 0.0001 {
			return e.SustainLvl * peak
		}
		progress := t / e.DecayTime
		factor := applyCurve(1.0-progress, e.Curve)
		return (e.SustainLvl + (1.0-e.SustainLvl)*factor) * peak
	}
	t -= e.DecayTime

	// 4. Release
	if t < e.ReleaseTime {
		if e.ReleaseTime <= 0.0001 {
			return 0.0
		}
		progress := t / e.ReleaseTime
		factor := applyCurve(1.0-progress, e.Curve)
		return e.SustainLvl * peak * factor
	}

	return 0.0
}

// PitchEnvelope computes a dynamic frequency value over time.
type PitchEnvelope struct {
	StartFreq float64
	EndFreq   float64
	Duration  float64
	Curve     CurveType
}

// FreqAt returns the frequency at time t (seconds).
func (p PitchEnvelope) FreqAt(t float64) float64 {
	if t <= 0 {
		return p.StartFreq
	}
	if t >= p.Duration || p.Duration <= 0.0001 {
		return p.EndFreq
	}
	progress := t / p.Duration
	if p.Curve == CurveExponential {
		// Log-frequency / exponential glide
		logStart := math.Log(math.Max(0.1, p.StartFreq))
		logEnd := math.Log(math.Max(0.1, p.EndFreq))
		return math.Exp(logStart + (logEnd-logStart)*progress)
	}
	factor := applyCurve(progress, p.Curve)
	return p.StartFreq + (p.EndFreq-p.StartFreq)*factor
}

func applyCurve(progress float64, curve CurveType) float64 {
	if progress <= 0 {
		return 0
	}
	if progress >= 1.0 {
		return 1.0
	}
	switch curve {
	case CurveExponential:
		// Slower initial rise/fall, accelerating
		return progress * progress
	case CurveLogarithmic:
		// Fast initial rise/fall, tapering off
		return math.Sqrt(progress)
	case CurveLinear:
		fallthrough
	default:
		return progress
	}
}
