# SubGame

A 2D sci-fi underwater survival exploration game built with [Ebitengine](https://ebitengine.org/) (Go).

Dive into uncharted alien ocean trenches, harvest mineral nodes, craft advanced survival gear and vehicles (The Skiff, Scout Sub, Heavy Mech), explore diverse biomes, and uncover decrypted PDA narrative logs to repair your escape vessel.

---

## Quick Start

### Prerequisites
- [Go](https://go.dev/dl/) 1.21+ installed.
- Operating System: macOS, Linux, or Windows.

### Running the Game
```bash
# Clone and run
git clone https://github.com/jaredwarren/SubGame.git
cd SubGame

# Launch via Makefile
make run

# Or launch directly with Go
go run ./cmd/game
```

---

## Procedural Audio DSP Engine

All game sound effects, robotic AI voice lines, and biome ambient music tracks are synthesized in **pure standard library Go** (`internal/synth`) with **zero external C/CGO dependencies** and **no third-party AI services**.

The audio system operates in dual-mode:
1. **Offline Generator CLI (`cmd/gen_audio`)**: Synthesizes 16-bit 44.1 kHz PCM RIFF `.wav` files into `assets/audio/`.
2. **Runtime Audio Engine (`internal/game/audio`)**: Loads embedded audio, manages volume/mute controls, looping channels (breathing, heartbeat, engines), pitch variation for repetitive hits, and underwater acoustic low-pass filtering.

---

## How to Generate & Update Sounds

### 1. Generating Audio Files
To batch generate or regenerate all audio files:

```bash
# Generate all audio assets into assets/audio/
make audio

# Or run the generator tool directly
go run ./cmd/gen_audio

# Filter and regenerate only specific sounds (e.g. splash or voice lines)
go run ./cmd/gen_audio --filter splash
go run ./cmd/gen_audio --filter voice
go run ./cmd/gen_audio --filter mining

# Use a custom random seed or output directory
go run ./cmd/gen_audio --seed 12345 --output-dir assets/audio
```

---

### 2. Customizing Existing Sounds (Examples)

All sound definitions live declaratively in [`internal/synth/presets.go`](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/synth/presets.go).

#### Example A: Tweaking an SFX Preset (e.g., Mining Hit)
To make the pickaxe mineral clink higher-pitched or punchier, locate `sfx/mining_hit.wav` in [`internal/synth/presets.go`](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/synth/presets.go):

```go
"sfx/mining_hit.wav": func(seed int64) *Buffer {
    b := NewSoundBuilder(0.30) // Duration in seconds

    // Layer 1: High metallic ping (change frequency or envelope)
    b.AddLayer(SoundLayer{
        Wave:      WaveTriangle,
        FreqStart: 1800.0, // <-- Increase for a higher chime
        FreqEnd:   600.0,  // <-- Rapid downward frequency slide
        Env:       NewDecayEnvelope(0.001, 0.08),
        Gain:      0.7,
    })

    // Layer 2: Sub-surface transient impact
    b.AddLayer(SoundLayer{
        Wave:      WaveNoisePink,
        Env:       NewDecayEnvelope(0.001, 0.05),
        Filter:    NewBiquadFilter(FilterBandPass, 800.0, 2.0),
        Gain:      0.8,
    })

    return b.Build().Normalize(0.95)
}
```
After editing, run `go run ./cmd/gen_audio --filter mining_hit` to rebuild the `.wav` asset.

---

#### Example B: Adding a Brand New Sound
1. Open [`internal/synth/presets.go`](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/synth/presets.go).
2. Add your new sound to `SoundCatalog`:

```go
"sfx/my_custom_laser.wav": func(seed int64) *Buffer {
    b := NewSoundBuilder(0.4) // 0.4 seconds

    // Add saw wave laser sweep with bitcrush distortion
    b.AddLayer(SoundLayer{
        Wave:       WaveSawtooth,
        FreqStart:  1200.0,
        FreqEnd:    150.0,
        Env:        NewExpDecayEnvelope(0.005, 0.35),
        Gain:       0.8,
        Bitcrush:   8,    // 8-bit retro crunch
        Drive:      1.4,  // Saturation
    })

    return b.Build().Normalize(0.95)
},
```

3. Run `go run ./cmd/gen_audio --filter my_custom_laser`.
4. Play it in game code:
```go
import "github.com/jaredwarren/SubGame/internal/game/audio"

audio.Get().PlaySFX("sfx/my_custom_laser.wav")
```

---

#### Example C: Customizing Robotic AI Voice Alerts
Voice lines are synthesized with vocal tract formant resonators ($F_1, F_2, F_3$) in [`internal/synth/formant.go`](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/synth/formant.go).

To create or adjust a voice warning:
```go
// Synthesize a robotic announcement by chaining formant vowel pulses
func SynthesizeCustomWarning() *Buffer {
    vowels := []FormantTriplet{
        Vowel_OW, // "Oh"
        Vowel_UH, // "uh"
        Vowel_EE, // "ee"
    }
    
    // Synthesize vocal cord pulses modulated through the formant resonator
    return SynthesizeFormantSequence(vowels, 110.0 /* Pitch in Hz */, 0.25 /* Duration */)
}
```

---

#### Example D: Modifying Ambient Biome Music
Ambient tracks (such as `music/cave_shallow.mp3` or `music/main_title.mp3` in [`internal/synth/presets.go`](file:///Users/jaredwarren/src/github.com/jaredwarren/SubGame/internal/synth/presets.go)) combine detuned multi-oscillators, slow low-pass filter sweeps, and algorithmic Schroeder reverberation:

```go
"music/cave_shallow.mp3": func(seed int64) *Buffer {
    b := NewSoundBuilder(6.0) // 6-second seamless ambient drone loop

    // Sub-bass root drone (Sine 65 Hz)
    b.AddLayer(SoundLayer{
        Wave:      WaveSine,
        FreqStart: 65.41,
        FreqEnd:   65.41,
        Env:       NewAHDSR(0.5, 0.0, 1.0, 4.5, 1.0),
        Gain:      0.45,
    })

    // Swaying underwater pad with resonant low-pass filter and Schroeder reverb
    b.AddLayer(SoundLayer{
        Wave:      WaveSawtooth,
        FreqStart: 130.81,
        FreqEnd:   130.81,
        Env:       NewAHDSR(1.0, 0.0, 0.8, 4.0, 1.0),
        Filter:    NewBiquadFilter(FilterLowPass, 320.0, 1.5),
        Gain:      0.25,
        Reverb:    true,
    })

    return b.Build().Normalize(0.92)
}
```

---

## Testing & Quality Gates

Run the test suite and quality checks:
```bash
# Run all unit tests (DSP synth, audio engine, game scenes, world gen)
make test

# Run verbose tests
make test-v

# Run linting and code quality gates
make check
```

---

## Controls

| Key / Input | Action |
|-------------|--------|
| **W, A, S, D** / **Arrows** | Swim / Steer Active Vehicle |
| **Shift** | Sprint Swim / Thruster Boost |
| **Left Click** | Mine Node / Catch Fauna / Use Active Item / Craft |
| **Right Click** | Scan Environment / Use Secondary Tool |
| **Tab** or **I** | Toggle Inventory & Equipment Panel |
| **1 – 5** | Quick-Select Hotbar Slots |
| **E** | Board / Exit Vehicle or Enter Base Habitat |
| **T** | Toggle Diver Flashlight |
| **Q** | Scout Sub Active Sonar Ping |
| **Space** | Launch Sonic Decoy / Chemical Deterrent / Mech Jump |
| **J** or **M** | Open Personal Data Assistant (PDA) Logs & Chart Map |
| **Esc** | Pause Menu / Close Active Interface |
