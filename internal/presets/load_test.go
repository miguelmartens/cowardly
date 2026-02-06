package presets

import (
	"os"
	"path"
	"testing"

	"github.com/cowardly/cowardly/internal/brave"
)

func TestConvertSettingsKeyValidation(t *testing.T) {
	tests := []struct {
		name    string
		rows    []settingRow
		wantErr bool
	}{
		{"valid key", []settingRow{{Key: "BraveRewardsDisabled", Value: true, Type: "bool"}}, false},
		{"empty key", []settingRow{{Key: "", Value: true, Type: "bool"}}, true},
		{"invalid key hyphen", []settingRow{{Key: "Brave-Rewards", Value: true, Type: "bool"}}, true},
		{"invalid key space", []settingRow{{Key: "Brave Rewards", Value: true, Type: "bool"}}, true},
		{"invalid key digit first", []settingRow{{Key: "1Key", Value: true, Type: "bool"}}, true},
		{"valid key with digits", []settingRow{{Key: "Key2", Value: 1, Type: "integer"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := convertSettings(tt.rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertSettings() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromFS_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	presetsDir := path.Join(dir, "presets")
	if err := os.Mkdir(presetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(presetsDir, "bad.yaml"), []byte("id: x\nsettings:\n  - key: K\n    value: true\n    type: invalid_type"), 0600); err != nil {
		t.Fatal(err)
	}
	fsys := os.DirFS(dir)
	_, err := LoadFromFS(fsys, "presets")
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestLoadFromFS_ValidMinimal(t *testing.T) {
	dir := t.TempDir()
	presetsDir := path.Join(dir, "presets")
	if err := os.Mkdir(presetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	yaml := `id: test
name: Test
description: Test preset
settings:
  - key: BraveRewardsDisabled
    value: true
    type: bool
`
	if err := os.WriteFile(path.Join(presetsDir, "00-test.yaml"), []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	fsys := os.DirFS(dir)
	list, err := LoadFromFS(fsys, "presets")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(list))
	}
	if list[0].ID != "test" || list[0].Name != "Test" {
		t.Errorf("unexpected preset: %+v", list[0])
	}
	if len(list[0].Settings) != 1 || list[0].Settings[0].Key != "BraveRewardsDisabled" {
		t.Errorf("unexpected settings: %+v", list[0].Settings)
	}
}

func TestLoadFromFS_NoDir(t *testing.T) {
	fsys := os.DirFS(t.TempDir())
	_, err := LoadFromFS(fsys, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent dir")
	}
}

func TestLoadPrivacyGuides(t *testing.T) {
	settings, err := LoadPrivacyGuides()
	if err != nil {
		t.Fatalf("LoadPrivacyGuides: %v", err)
	}
	if len(settings) == 0 {
		t.Error("expected non-empty supplement")
	}
	keys := make(map[string]bool)
	for _, s := range settings {
		keys[s.Key] = true
	}
	for _, k := range []string{"BraveP3AEnabled", "DefaultBraveFingerprintingV2Setting", "BraveDeAmpEnabled"} {
		if !keys[k] {
			t.Errorf("expected key %q in Privacy Guides supplement", k)
		}
	}
}

func TestPrivacyGuidesMerged(t *testing.T) {
	merged, err := PrivacyGuidesMerged("quick")
	if err != nil {
		t.Fatalf("PrivacyGuidesMerged: %v", err)
	}
	keys := make(map[string]bool)
	for _, s := range merged {
		keys[s.Key] = true
	}
	// Merged = Quick Debloat + supplement; should have both
	if !keys["BraveRewardsDisabled"] {
		t.Error("merged should include base preset key BraveRewardsDisabled")
	}
	if !keys["BraveP3AEnabled"] {
		t.Error("merged should include supplement key BraveP3AEnabled")
	}
}

func TestMergeSettingsWithSupplement(t *testing.T) {
	base := []brave.Setting{
		{Key: "A", Value: true, Type: brave.TypeBool},
		{Key: "B", Value: 1, Type: brave.TypeInteger},
	}
	supplement := []brave.Setting{
		{Key: "B", Value: 2, Type: brave.TypeInteger},  // overlay
		{Key: "C", Value: "x", Type: brave.TypeString}, // new
	}
	merged := MergeSettingsWithSupplement(base, supplement)
	byKey := make(map[string]brave.Setting)
	for _, s := range merged {
		byKey[s.Key] = s
	}
	if len(merged) != 3 {
		t.Errorf("merged should have 3 entries, got %d", len(merged))
	}
	if v, ok := byKey["A"]; !ok || v.Value != true {
		t.Errorf("A should be true from base, got %v", v.Value)
	}
	if v, ok := byKey["B"]; !ok || v.Value != 2 {
		t.Errorf("B should be 2 (overlay from supplement), got %v", v.Value)
	}
	if v, ok := byKey["C"]; !ok || v.Value != "x" {
		t.Errorf("C should be x from supplement, got %v", v.Value)
	}
}

func TestLoadFromFS_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	presetsDir := path.Join(dir, "presets")
	if err := os.Mkdir(presetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	fsys := os.DirFS(dir)
	list, err := LoadFromFS(fsys, "presets")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 presets, got %d", len(list))
	}
}
