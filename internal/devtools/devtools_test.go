package devtools

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInspectWorld(t *testing.T) {
	report := inspectWorld(12345)
	if report.Width != 500 || report.Height != 500 {
		t.Fatalf("world size %dx%d, want 500x500", report.Width, report.Height)
	}
	if report.LandTiles == 0 || report.WaterTiles == 0 {
		t.Fatalf("expected land and water tiles, got land=%d water=%d", report.LandTiles, report.WaterTiles)
	}
	if len(report.BiomeCounts) < 2 {
		t.Fatalf("expected multiple biomes, got %v", report.BiomeCounts)
	}
	if len(report.png) == 0 {
		t.Fatal("expected minimap PNG")
	}
	img, err := png.Decode(bytes.NewReader(report.png))
	if err != nil {
		t.Fatalf("decode minimap: %v", err)
	}
	if img.Bounds().Dx() != 500 || img.Bounds().Dy() != 500 {
		t.Fatalf("minimap bounds %v, want 500x500", img.Bounds())
	}
}

func TestInspectCaveThermo(t *testing.T) {
	report, err := inspectCave("thermo", "thermal_barrens", 54321)
	if err != nil {
		t.Fatal(err)
	}
	if report.Width == 0 || report.Height == 0 {
		t.Fatal("expected thermo cave grid")
	}
	if report.OpenTiles == 0 {
		t.Fatal("expected open tiles")
	}
	if len(report.png) == 0 {
		t.Fatal("expected cave PNG")
	}
	foundSiphon := false
	for _, e := range report.Entities {
		if e.Name == "BrimstoneSiphon" {
			foundSiphon = true
		}
	}
	if !foundSiphon {
		t.Fatalf("expected BrimstoneSiphon in thermo cave, got %v", report.Entities)
	}
}

func TestInspectCaveUnknown(t *testing.T) {
	_, err := inspectCave("nope", "shallow_reef", 1)
	if err == nil {
		t.Fatal("expected error for unknown cave type")
	}
}

func TestInspectCaveVoid(t *testing.T) {
	report, err := inspectCave("void", "shallow_reef", 1)
	if err != nil {
		t.Fatal(err)
	}
	if report.png != nil {
		t.Fatal("void cave should have no PNG")
	}
	if report.Note == "" {
		t.Fatal("expected void note")
	}
}

func TestCatalogReport(t *testing.T) {
	rep := catalogReport()
	if len(rep.Groups) == 0 {
		t.Fatal("expected sound groups")
	}
	n := 0
	for _, g := range rep.Groups {
		n += len(g.Sounds)
	}
	if n == 0 {
		t.Fatal("expected catalog entries")
	}
}

func TestSynthesizeWAV(t *testing.T) {
	wav, err := synthesizeWAV("sfx/ui_confirm.wav", 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(wav) < 44 {
		t.Fatalf("WAV too small: %d", len(wav))
	}
	if string(wav[0:4]) != "RIFF" {
		t.Fatal("expected RIFF header")
	}
}

func TestInspectCaveShallow(t *testing.T) {
	report, err := inspectCave("shallow", "kelp_forest", 99)
	if err != nil {
		t.Fatal(err)
	}
	if report.Width != 60 || report.Height != 120 {
		t.Fatalf("shallow size %dx%d, want 60x120", report.Width, report.Height)
	}
	if report.OpenTiles == 0 {
		t.Fatal("expected open tiles")
	}
}

func TestInspectCaveShockKelpShallowChasm(t *testing.T) {
	report, err := inspectCave("shock_kelp_shallow", "kelp_forest", 42)
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasChasm {
		t.Fatal("expected chasm on shock kelp shallow layer")
	}
}

func TestHandlerWorldAndCaves(t *testing.T) {
	h := Handler()
	for _, path := range []string{"/world?seed=7", "/caves?type=thermo&seed=7"} {
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, path, nil))
		if rw.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, rw.Code)
		}
		if !strings.Contains(rw.Body.String(), "SubGame Devtools") {
			t.Fatalf("%s missing hub chrome", path)
		}
	}

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/caves/map.png?type=thermo&seed=7", nil))
	if rw.Code != http.StatusOK {
		t.Fatalf("cave png status %d", rw.Code)
	}
	if ct := rw.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("cave png content-type %q", ct)
	}
}

func TestHandlerRoutes(t *testing.T) {
	h := Handler()

	redir := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, redir)
	if rw.Code != http.StatusFound {
		t.Fatalf("GET / status %d, want 302", rw.Code)
	}

	audio := httptest.NewRequest(http.MethodGet, "/audio", nil)
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, audio)
	if rw.Code != http.StatusOK {
		t.Fatalf("GET /audio status %d", rw.Code)
	}
	if !strings.Contains(rw.Body.String(), "SubGame Devtools") {
		t.Fatal("audio page missing hub chrome")
	}

	bad := httptest.NewRequest(http.MethodGet, "/audio/wav?name=not-a-sound", nil)
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, bad)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("unknown sound status %d, want 400", rw.Code)
	}
}
