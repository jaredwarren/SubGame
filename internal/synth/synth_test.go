package synth

import (
	"bytes"
	"math"
	"testing"
)

func TestWAVEncoding(t *testing.T) {
	buf := NewMonoBuffer(0.1)
	for i := range buf.SamplesLeft {
		buf.SamplesLeft[i] = math.Sin(2.0 * math.Pi * 440.0 * float64(i) / float64(buf.SampleRate))
	}
	buf.Normalize(0.95)

	wavBytes, err := buf.EncodeWAV()
	if err != nil {
		t.Fatalf("EncodeWAV failed: %v", err)
	}
	if len(wavBytes) < 44 {
		t.Fatalf("WAV output too small: %d bytes", len(wavBytes))
	}

	// Verify RIFF header
	if !bytes.Equal(wavBytes[0:4], []byte("RIFF")) {
		t.Errorf("Expected RIFF header, got %s", string(wavBytes[0:4]))
	}
	if !bytes.Equal(wavBytes[8:12], []byte("WAVE")) {
		t.Errorf("Expected WAVE format, got %s", string(wavBytes[8:12]))
	}
	if !bytes.Equal(wavBytes[12:16], []byte("fmt ")) {
		t.Errorf("Expected fmt subchunk, got %s", string(wavBytes[12:16]))
	}
}

func TestOscillators(t *testing.T) {
	waves := []Waveform{WaveSine, WaveTriangle, WaveSquare, WaveSawtooth, WaveNoiseWhite, WaveNoisePink, WaveNoiseBrown}

	for _, w := range waves {
		osc := NewOscillator(w, 440.0, 123)
		for i := 0; i < 500; i++ {
			s := osc.Next(0)
			if math.IsNaN(s) || math.IsInf(s, 0) {
				t.Fatalf("Waveform %d generated NaN/Inf sample", w)
			}
			if s < -1.5 || s > 1.5 {
				t.Fatalf("Waveform %d generated out-of-bounds sample: %f", w, s)
			}
		}
	}
}

func TestBiquadFilter(t *testing.T) {
	filter := NewBiquadFilter(FilterLowPass, 1000.0, 0.707)
	inBuf := NewMonoBuffer(0.05)
	for i := range inBuf.SamplesLeft {
		inBuf.SamplesLeft[i] = math.Sin(2.0 * math.Pi * 2000.0 * float64(i) / float64(inBuf.SampleRate))
	}

	filter.ProcessBuffer(inBuf)

	for _, s := range inBuf.SamplesLeft {
		if math.IsNaN(s) || math.IsInf(s, 0) {
			t.Fatalf("Filter produced NaN/Inf")
		}
	}
}

func TestFormantVoiceSynthesis(t *testing.T) {
	buf := SynthesizeOxygenLowVoice()
	if buf == nil || len(buf.SamplesLeft) == 0 {
		t.Fatalf("SynthesizeOxygenLowVoice returned empty buffer")
	}
	wav, err := buf.EncodeWAV()
	if err != nil {
		t.Fatalf("Failed to encode voice WAV: %v", err)
	}
	if len(wav) < 1000 {
		t.Fatalf("Voice WAV unexpectedly small: %d bytes", len(wav))
	}
}

func TestSoundCatalogPresets(t *testing.T) {
	if len(SoundCatalog) == 0 {
		t.Fatalf("SoundCatalog is empty")
	}

	for name, gen := range SoundCatalog {
		buf := gen(42)
		if buf == nil {
			t.Fatalf("Preset %s returned nil buffer", name)
		}
		if len(buf.SamplesLeft) == 0 {
			t.Fatalf("Preset %s returned 0 samples", name)
		}
		wav, err := buf.EncodeWAV()
		if err != nil {
			t.Fatalf("Preset %s failed to encode to WAV: %v", name, err)
		}
		if len(wav) < 44 {
			t.Fatalf("Preset %s produced invalid WAV file", name)
		}
	}
}
