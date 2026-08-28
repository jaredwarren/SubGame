package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jaredwarren/SubGame/internal/synth"
)

func main() {
	outDir := flag.String("output-dir", "assets/audio", "Target output directory for synthesized audio files")
	filter := flag.String("filter", "", "Filter sounds by name substring")
	seed := flag.Int64("seed", 12345, "Base random seed for procedural generation")
	overwriteMusic := flag.Bool("overwrite-music", false, "Overwrite existing music files in assets/audio/music/")
	flag.Parse()

	startTime := time.Now()
	fmt.Printf("=== SubGame Procedural Audio DSP Generator ===\n")
	fmt.Printf("Output directory: %s\n", *outDir)
	if *filter != "" {
		fmt.Printf("Filtering sounds matching: %s\n", *filter)
	}

	generatedCount := 0
	totalBytes := 0

	for relPath, generator := range synth.SoundCatalog {
		if *filter != "" && !strings.Contains(relPath, *filter) {
			continue
		}

		targetPath := filepath.Join(*outDir, relPath)
		if strings.HasPrefix(relPath, "music/") && !*overwriteMusic {
			if info, err := os.Stat(targetPath); err == nil && info.Size() > 0 {
				// Don't overwrite existing music asset (e.g. AI-generated tracks)
				continue
			}
			if strings.HasPrefix(relPath, "music/cave_") {
				if info, err := os.Stat(filepath.Join(*outDir, "music/cave.mp3")); err == nil && info.Size() > 0 {
					continue
				}
			}
		}

		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", parentDir, err)
			os.Exit(1)
		}

		buf := generator(*seed + int64(generatedCount))
		if buf == nil {
			fmt.Fprintf(os.Stderr, "Generator returned nil for %s\n", relPath)
			continue
		}

		wavBytes, err := buf.EncodeWAV()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding WAV for %s: %v\n", relPath, err)
			continue
		}

		if err := os.WriteFile(targetPath, wavBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", targetPath, err)
			continue
		}

		generatedCount++
		totalBytes += len(wavBytes)
		fmt.Printf("  [✓] %-35s (%.2fs, %d KB)\n", relPath, buf.Duration(), len(wavBytes)/1024)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\nGenerated %d audio assets (%d KB total) in %v\n", generatedCount, totalBytes/1024, elapsed)
}
