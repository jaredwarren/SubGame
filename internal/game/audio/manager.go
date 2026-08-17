package audio

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/jaredwarren/SubGame/assets"
)

const SampleRate = 44100

// Manager manages audio playback, volume controls, SFX pooling, and music streams.
type Manager struct {
	mu           sync.Mutex
	context      *audio.Context
	sfxRawCache  map[string][]byte
	activeLoops  map[string]*audio.Player
	currentMusic *audio.Player
	musicName    string

	MasterVolume float64
	SFXVolume    float64
	MusicVolume  float64
	Muted        bool
	Submerged    bool

	rng *rand.Rand
}

var (
	globalManager *Manager
	once          sync.Once
)

// Get returns the global audio Manager singleton.
func Get() *Manager {
	once.Do(func() {
		globalManager = &Manager{
			context:      audio.NewContext(SampleRate),
			sfxRawCache:  make(map[string][]byte),
			activeLoops:  make(map[string]*audio.Player),
			MasterVolume: 1.0,
			SFXVolume:    0.85,
			MusicVolume:  0.65,
			Muted:        false,
			Submerged:    false,
			rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		}
		globalManager.preloadCommonSFX()
	})
	return globalManager
}

// preloadCommonSFX pre-caches raw bytes for audio files from assets.AudioFS.
func (m *Manager) preloadCommonSFX() {
	entries, err := assets.AudioFS.ReadDir("audio/sfx")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".wav") {
				path := filepath.Join("audio/sfx", entry.Name())
				data, readErr := assets.AudioFS.ReadFile(path)
				if readErr == nil {
					// Store with both "sfx/name.wav" and "name.wav" keys
					relKey := filepath.Join("sfx", entry.Name())
					m.sfxRawCache[relKey] = data
					m.sfxRawCache[entry.Name()] = data
				}
			}
		}
	}
}

func (m *Manager) loadRawBytes(relPath string) ([]byte, error) {
	relPath = filepath.Clean(relPath)
	if data, ok := m.sfxRawCache[relPath]; ok {
		return data, nil
	}

	// Try with "audio/" prefix in embedded FS
	fsPath := relPath
	if !strings.HasPrefix(fsPath, "audio/") {
		fsPath = "audio/" + fsPath
	}

	data, err := assets.AudioFS.ReadFile(fsPath)
	if err != nil {
		return nil, fmt.Errorf("audio file not found: %s (%w)", relPath, err)
	}

	m.sfxRawCache[relPath] = data
	return data, nil
}

// PlaySFX plays a sound effect once at default SFX volume.
func (m *Manager) PlaySFX(name string) {
	m.PlaySFXWithVolume(name, 1.0)
}

// PlaySFXWithVolume plays a sound effect with a custom volume scalar (0.0 to 1.0).
func (m *Manager) PlaySFXWithVolume(name string, volume float64) {
	if m.Muted || m.MasterVolume <= 0 || m.SFXVolume <= 0 || volume <= 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rawBytes, err := m.loadRawBytes(name)
	if err != nil {
		return
	}

	stream, err := wav.DecodeWithSampleRate(SampleRate, bytes.NewReader(rawBytes))
	if err != nil {
		return
	}

	player, err := m.context.NewPlayer(stream)
	if err != nil {
		return
	}

	vol := volume * m.SFXVolume * m.MasterVolume
	if m.Submerged && !strings.Contains(name, "voice") && !strings.Contains(name, "ui") {
		vol *= 0.9
	}
	player.SetVolume(vol)
	player.Play()
}

// PlaySFXVaried plays a sound effect with pitch/volume variation (e.g. ±5-10% for repetitive hits).
func (m *Manager) PlaySFXVaried(name string, baseVolume float64, pitchVariation float64) {
	if m.Muted || m.MasterVolume <= 0 {
		return
	}
	// Slight volume fluctuation (±5%)
	volMod := 1.0 + (m.rng.Float64()*0.1 - 0.05)
	m.PlaySFXWithVolume(name, baseVolume*volMod)
}

// PlayLoop starts or updates a looping sound effect (e.g. engine hums, scanner loop).
func (m *Manager) PlayLoop(loopID string, name string, volume float64) {
	if m.Muted || m.MasterVolume <= 0 {
		m.StopLoop(loopID)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if player, exists := m.activeLoops[loopID]; exists && player != nil {
		vol := volume * m.SFXVolume * m.MasterVolume
		player.SetVolume(vol)
		if !player.IsPlaying() {
			player.Play()
		}
		return
	}

	rawBytes, err := m.loadRawBytes(name)
	if err != nil {
		return
	}

	stream, err := wav.DecodeWithSampleRate(SampleRate, bytes.NewReader(rawBytes))
	if err != nil {
		return
	}

	length := stream.Length()
	loop := audio.NewInfiniteLoop(stream, length)
	player, err := m.context.NewPlayer(loop)
	if err != nil {
		return
	}

	vol := volume * m.SFXVolume * m.MasterVolume
	player.SetVolume(vol)
	player.Play()
	m.activeLoops[loopID] = player
}

// StopLoop stops and clears an active looping sound.
func (m *Manager) StopLoop(loopID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if player, exists := m.activeLoops[loopID]; exists && player != nil {
		player.Pause()
		player.Close()
		delete(m.activeLoops, loopID)
	}
}

// StopAllLoops stops all active looping sound effects.
func (m *Manager) StopAllLoops() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, player := range m.activeLoops {
		if player != nil {
			player.Pause()
			player.Close()
		}
		delete(m.activeLoops, id)
	}
}

// PlayMusic plays or transitions to background music.
func (m *Manager) PlayMusic(name string, volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.musicName == name && m.currentMusic != nil && m.currentMusic.IsPlaying() {
		vol := volume * m.MusicVolume * m.MasterVolume
		if m.Muted {
			vol = 0
		}
		m.currentMusic.SetVolume(vol)
		return
	}

	if m.currentMusic != nil {
		m.currentMusic.Pause()
		m.currentMusic.Close()
		m.currentMusic = nil
	}

	m.musicName = name
	if m.Muted || m.MasterVolume <= 0 || m.MusicVolume <= 0 {
		return
	}

	rawBytes, err := m.loadRawBytes(name)
	if err != nil {
		return
	}

	var stream io.Reader
	stream, err = wav.DecodeWithSampleRate(SampleRate, bytes.NewReader(rawBytes))
	if err != nil {
		return
	}

	if ws, ok := stream.(*wav.Stream); ok {
		loop := audio.NewInfiniteLoop(ws, ws.Length())
		player, pErr := m.context.NewPlayer(loop)
		if pErr == nil {
			vol := volume * m.MusicVolume * m.MasterVolume
			player.SetVolume(vol)
			player.Play()
			m.currentMusic = player
		}
	}
}

// StopMusic pauses and closes the active background music player.
func (m *Manager) StopMusic() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentMusic != nil {
		m.currentMusic.Pause()
		m.currentMusic.Close()
		m.currentMusic = nil
	}
	m.musicName = ""
}

// SetSubmerged updates the acoustic dampening state (in-cave/underwater vs surface).
func (m *Manager) SetSubmerged(submerged bool) {
	m.Submerged = submerged
}

// SetMasterVolume updates master volume (0.0 to 1.0).
func (m *Manager) SetMasterVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1.0 {
		v = 1.0
	}
	m.MasterVolume = v
	m.updateVolumes()
}

// SetSFXVolume updates sound effects volume (0.0 to 1.0).
func (m *Manager) SetSFXVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1.0 {
		v = 1.0
	}
	m.SFXVolume = v
	m.updateVolumes()
}

// SetMusicVolume updates background music volume (0.0 to 1.0).
func (m *Manager) SetMusicVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1.0 {
		v = 1.0
	}
	m.MusicVolume = v
	m.updateVolumes()
}

// ToggleMute toggles audio mute state.
func (m *Manager) ToggleMute() bool {
	m.Muted = !m.Muted
	m.updateVolumes()
	return m.Muted
}

func (m *Manager) updateVolumes() {
	if m.currentMusic != nil {
		vol := m.MusicVolume * m.MasterVolume
		if m.Muted {
			vol = 0
		}
		m.currentMusic.SetVolume(vol)
	}
	for _, player := range m.activeLoops {
		if player != nil {
			vol := m.SFXVolume * m.MasterVolume
			if m.Muted {
				vol = 0
			}
			player.SetVolume(vol)
		}
	}
}
