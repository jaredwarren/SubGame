package audio

import (
	"testing"
)

func TestAudioPreload(t *testing.T) {
	mgr := Get()
	if mgr == nil {
		t.Fatalf("Expected non-nil audio manager")
	}

	// Verify that common SFX were cached (raw + decoded PCM)
	if len(mgr.sfxRawCache) == 0 {
		t.Fatalf("Expected cached SFX files in manager, got 0")
	}
	if len(mgr.pcmCache) == 0 {
		t.Fatalf("Expected decoded PCM cache in manager, got 0")
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
		pcm, ok := mgr.pcmCache[f]
		if !ok || len(pcm) == 0 {
			t.Errorf("Expected decoded PCM for %s", f)
		}
	}
}

func TestGetPCMCachesDecode(t *testing.T) {
	mgr := Get()
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Music files are WAV containers with an .mp3 extension — decode once, then hit cache.
	name := "music/main_title.mp3"
	pcm1, err := mgr.getPCMLocked(name)
	if err != nil {
		t.Fatalf("getPCMLocked(%s): %v", name, err)
	}
	if len(pcm1) == 0 {
		t.Fatalf("expected non-empty PCM for %s", name)
	}
	pcm2, err := mgr.getPCMLocked(name)
	if err != nil {
		t.Fatalf("second getPCMLocked(%s): %v", name, err)
	}
	// Same backing slice from cache
	if &pcm1[0] != &pcm2[0] {
		t.Errorf("expected cached PCM to reuse the same buffer")
	}
}

func TestPruneFinishedSFX(t *testing.T) {
	mgr := Get()
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Simulate a finished (never-started) player slot with nil — prune should not panic.
	mgr.activeSFX = append(mgr.activeSFX[:0], nil)
	mgr.pruneFinishedSFXLocked()
	if len(mgr.activeSFX) != 0 {
		t.Errorf("expected nil slots pruned, got %d", len(mgr.activeSFX))
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
