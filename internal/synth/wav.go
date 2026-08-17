package synth

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// SampleRate is the standard audio sample rate used across SubGame (44.1 kHz).
const SampleRate = 44100

// Buffer represents audio samples stored as 64-bit floating point numbers in range [-1.0, 1.0].
type Buffer struct {
	SamplesLeft  []float64
	SamplesRight []float64
	IsStereo     bool
	SampleRate   int
}

// NewMonoBuffer creates a mono buffer of a given duration in seconds.
func NewMonoBuffer(durationSeconds float64) *Buffer {
	numSamples := int(math.Ceil(durationSeconds * float64(SampleRate)))
	if numSamples < 1 {
		numSamples = 1
	}
	return &Buffer{
		SamplesLeft: make([]float64, numSamples),
		IsStereo:    false,
		SampleRate:  SampleRate,
	}
}

// NewStereoBuffer creates a stereo buffer of a given duration in seconds.
func NewStereoBuffer(durationSeconds float64) *Buffer {
	numSamples := int(math.Ceil(durationSeconds * float64(SampleRate)))
	if numSamples < 1 {
		numSamples = 1
	}
	return &Buffer{
		SamplesLeft:  make([]float64, numSamples),
		SamplesRight: make([]float64, numSamples),
		IsStereo:     true,
		SampleRate:   SampleRate,
	}
}

// Duration returns the total duration in seconds.
func (b *Buffer) Duration() float64 {
	if b.SampleRate == 0 {
		return 0
	}
	return float64(len(b.SamplesLeft)) / float64(b.SampleRate)
}

// NumSamples returns the number of sample frames.
func (b *Buffer) NumSamples() int {
	return len(b.SamplesLeft)
}

// Normalize scales the samples so the peak absolute amplitude reaches peakLevel (default 0.95).
func (b *Buffer) Normalize(peakLevel float64) {
	if peakLevel <= 0 || peakLevel > 1.0 {
		peakLevel = 0.95
	}
	maxAmp := 0.0
	for _, s := range b.SamplesLeft {
		if math.Abs(s) > maxAmp {
			maxAmp = math.Abs(s)
		}
	}
	if b.IsStereo {
		for _, s := range b.SamplesRight {
			if math.Abs(s) > maxAmp {
				maxAmp = math.Abs(s)
			}
		}
	}
	if maxAmp > 0.0001 {
		scale := peakLevel / maxAmp
		for i := range b.SamplesLeft {
			b.SamplesLeft[i] *= scale
		}
		if b.IsStereo {
			for i := range b.SamplesRight {
				b.SamplesRight[i] *= scale
			}
		}
	}
}

// EncodeWAV encodes the buffer into a standard 16-bit PCM RIFF WAV format.
func (b *Buffer) EncodeWAV() ([]byte, error) {
	var buf bytes.Buffer
	if err := b.WriteWAV(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteWAV writes the buffer to an io.Writer in standard 16-bit PCM RIFF WAV format.
func (b *Buffer) WriteWAV(w io.Writer) error {
	numChannels := uint16(1)
	if b.IsStereo {
		numChannels = 2
	}
	bitsPerSample := uint16(16)
	bytesPerSample := bitsPerSample / 8
	blockAlign := numChannels * bytesPerSample
	byteRate := uint32(b.SampleRate) * uint32(blockAlign)
	numFrames := uint32(len(b.SamplesLeft))
	subChunk2Size := numFrames * uint32(blockAlign)
	chunkSize := 36 + subChunk2Size

	// 1. RIFF header
	if _, err := w.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, chunkSize); err != nil {
		return err
	}
	if _, err := w.Write([]byte("WAVE")); err != nil {
		return err
	}

	// 2. fmt sub-chunk
	if _, err := w.Write([]byte("fmt ")); err != nil {
		return err
	}
	subChunk1Size := uint32(16) // PCM format subchunk size
	if err := binary.Write(w, binary.LittleEndian, subChunk1Size); err != nil {
		return err
	}
	audioFormat := uint16(1) // 1 = PCM (uncompressed)
	if err := binary.Write(w, binary.LittleEndian, audioFormat); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, numChannels); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(b.SampleRate)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}

	// 3. data sub-chunk
	if _, err := w.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, subChunk2Size); err != nil {
		return err
	}

	// 4. Sample data (16-bit signed integers, clamped)
	for i := 0; i < len(b.SamplesLeft); i++ {
		// Left channel
		sl := clampFloat(b.SamplesLeft[i])
		i16Left := int16(sl * 32767.0)
		if err := binary.Write(w, binary.LittleEndian, i16Left); err != nil {
			return err
		}

		// Right channel (if stereo)
		if b.IsStereo {
			sr := clampFloat(b.SamplesRight[i])
			i16Right := int16(sr * 32767.0)
			if err := binary.Write(w, binary.LittleEndian, i16Right); err != nil {
				return err
			}
		}
	}

	return nil
}

// SaveWAV saves the buffer as a WAV file at filePath.
func (b *Buffer) SaveWAV(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create wav file %s: %w", filePath, err)
	}
	defer f.Close()

	return b.WriteWAV(f)
}

func clampFloat(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < -1.0 {
		return -1.0
	}
	return v
}
