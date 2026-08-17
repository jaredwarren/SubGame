package audio

import (
	"testing"
)

func TestAudioPreload(t *testing.T) {
	mgr := Get()
	if mgr == nil {
		t.Fatalf("Expected non-nil audio manager")
	}

	// Verify that common SFX were cached
	if len(mgr.sfxRawCache) == 0 {
		t.Fatalf("Expected cached SFX files in manager, got 0")
	}

	// Check specific files
	sampleFiles := []string{
		"sfx/mining_hit.wav",
		"sfx/splash.wav",
		"sfx/scanner_complete.wav",
		"sfx/voice_o2_low.wav",
	}

	for _, f := range sampleFiles {
		data, err := mgr.loadRawBytes(f)
		if err != nil {
			t.Errorf("Failed to load %s: %v", f, err)
		}
		if len(data) < 44 {
			t.Errorf("File %s data too short (%d bytes)", f, len(data))
		}
	}
}

func TestVolumeAndMute(t *testing.T) {
	mgr := Get()
	mgr.SetMasterVolume(0.5)
	if mgr.MasterVolume != 0.5 {
		t.Errorf("Expected master volume 0.5, got %f", mgr.MasterVolume)
	}

	mgr.SetSFXVolume(0.7)
	if mgr.SFXVolume != 0.7 {
		t.Errorf("Expected SFX volume 0.7, got %f", mgr.SFXVolume)
	}

	muted := mgr.ToggleMute()
	if !muted || !mgr.Muted {
		t.Errorf("Expected manager to be muted")
	}
	mgr.ToggleMute()
	if mgr.Muted {
		t.Errorf("Expected manager to be unmuted")
	}
}
