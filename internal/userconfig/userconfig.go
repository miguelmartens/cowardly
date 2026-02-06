// Package userconfig manages the user's desired Brave settings in ~/.config/cowardly/cowardly.yaml.
// Used for --reapply and to detect when settings have been reverted (e.g. by MDM after restart).
package userconfig

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cowardly/cowardly/internal/brave"
	"github.com/cowardly/cowardly/internal/presets"
	"gopkg.in/yaml.v3"
)

// ConfigFileName is the name of the config file in the config directory.
const ConfigFileName = "cowardly.yaml"

// settingRow matches the on-disk YAML shape for one setting (same as presets).
type settingRow struct {
	Key   string      `yaml:"key"`
	Value interface{} `yaml:"value"`
	Type  string      `yaml:"type"`
}

// block is a settings block under preset.<id> or supplement.<id>.
type block struct {
	Settings []settingRow `yaml:"settings"`
}

// fileShapeNew is the new on-disk shape: preset.<id>.settings, supplement.<id>.settings.
type fileShapeNew struct {
	Preset     map[string]block `yaml:"preset,omitempty"`
	Supplement map[string]block `yaml:"supplement,omitempty"`
	ApplyFile  string           `yaml:"apply_file,omitempty"`
	Settings   []settingRow     `yaml:"settings,omitempty"` // for apply_file, legacy
}

// fileShapeLegacy supports the old format for backward compat.
type fileShapeLegacy struct {
	Preset     string       `yaml:"preset,omitempty"`
	BasePreset string       `yaml:"base_preset,omitempty"`
	Supplement []settingRow `yaml:"supplement,omitempty"`
	ApplyFile  string       `yaml:"apply_file,omitempty"`
	Settings   []settingRow `yaml:"settings,omitempty"`
}

// DesiredState is the in-memory representation of the user's last-applied / desired state.
type DesiredState struct {
	Preset     string          // preset id, if last apply was a preset
	BasePreset string          // for preset=privacy-guides: base preset id
	Supplement []brave.Setting // for preset=privacy-guides: supplement settings
	ApplyFile  string          // path to file, if last apply was from file
	Settings   []brave.Setting // snapshot of settings (used by reapply)
}

// ConfigDir returns ~/.config/cowardly. Creates the directory if it does not exist.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".config", "cowardly")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return dir, nil
}

// ConfigPath returns the full path to ~/.config/cowardly/cowardly.yaml.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// Read loads the desired state from ~/.config/cowardly/cowardly.yaml.
// Returns (nil, nil) if the file does not exist or is empty.
func Read() (*DesiredState, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	// Try new format first (preset.<id>.settings, supplement.<id>.settings)
	var fNew fileShapeNew
	if err := yaml.Unmarshal(data, &fNew); err == nil && (len(fNew.Preset) > 0 || fNew.ApplyFile != "" || len(fNew.Settings) > 0) {
		return readNewFormat(&fNew)
	}
	// Fall back to legacy format
	var fLegacy fileShapeLegacy
	if err := yaml.Unmarshal(data, &fLegacy); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return readLegacyFormat(&fLegacy)
}

func readNewFormat(f *fileShapeNew) (*DesiredState, error) {
	// preset + supplement (Privacy Guides)
	if f.Preset != nil {
		for presetID, pblock := range f.Preset {
			if len(pblock.Settings) == 0 {
				continue
			}
			baseSettings, err := rowsToSettings(pblock.Settings)
			if err != nil {
				return nil, fmt.Errorf("preset %q: %w", presetID, err)
			}
			var finalSettings []brave.Setting
			var preset, basePreset string
			sup, ok := f.Supplement["privacy_guides"]
			if !ok || len(sup.Settings) == 0 {
				sup, ok = f.Supplement["privacy-guides"] // backward compat
			}
			if ok && len(sup.Settings) > 0 {
				supplementSettings, err := rowsToSettings(sup.Settings)
				if err != nil {
					return nil, fmt.Errorf("supplement privacy_guides: %w", err)
				}
				finalSettings = mergeSettings(baseSettings, supplementSettings)
				preset = "privacy-guides"
				basePreset = presetID
			} else {
				finalSettings = baseSettings
				preset = presetID
			}
			return &DesiredState{
				Preset:     preset,
				BasePreset: basePreset,
				Settings:   finalSettings,
			}, nil
		}
	}
	// apply_file
	if f.ApplyFile != "" && len(f.Settings) > 0 {
		settings, err := rowsToSettings(f.Settings)
		if err != nil {
			return nil, err
		}
		return &DesiredState{ApplyFile: f.ApplyFile, Settings: settings}, nil
	}
	// legacy flat settings (including old custom format)
	if len(f.Settings) > 0 {
		settings, err := rowsToSettings(f.Settings)
		if err != nil {
			return nil, err
		}
		return &DesiredState{Settings: settings}, nil
	}
	return nil, nil
}

func readLegacyFormat(f *fileShapeLegacy) (*DesiredState, error) {
	if len(f.Settings) == 0 && (f.Preset != "privacy-guides" || len(f.Supplement) == 0) {
		return nil, nil
	}
	if f.Preset == "privacy-guides" && f.BasePreset != "" && len(f.Supplement) > 0 {
		settings, err := presets.MergePresetWithSupplement(f.BasePreset, supplementToSettings(f.Supplement))
		if err != nil {
			return nil, fmt.Errorf("privacy-guides merge: %w", err)
		}
		return &DesiredState{Preset: f.Preset, BasePreset: f.BasePreset, Settings: settings}, nil
	}
	if f.Preset == "privacy-guides" && len(f.Settings) > 0 {
		settings, err := rowsToSettings(f.Settings)
		if err != nil {
			return nil, err
		}
		return &DesiredState{Preset: f.Preset, BasePreset: f.BasePreset, Settings: settings}, nil
	}
	if len(f.Settings) > 0 {
		settings, err := rowsToSettings(f.Settings)
		if err != nil {
			return nil, err
		}
		return &DesiredState{Preset: f.Preset, ApplyFile: f.ApplyFile, Settings: settings}, nil
	}
	return nil, nil
}

func rowsToSettings(rows []settingRow) ([]brave.Setting, error) {
	sr := make([]presets.SettingRow, len(rows))
	for i, r := range rows {
		sr[i] = presets.SettingRow{Key: r.Key, Value: r.Value, Type: r.Type}
	}
	return presets.ConvertSettingRows(sr)
}

func mergeSettings(base, overlay []brave.Setting) []brave.Setting {
	byKey := make(map[string]brave.Setting)
	for _, s := range base {
		byKey[s.Key] = s
	}
	for _, s := range overlay {
		byKey[s.Key] = s
	}
	seen := make(map[string]bool)
	var out []brave.Setting
	for _, s := range base {
		out = append(out, byKey[s.Key])
		seen[s.Key] = true
	}
	for _, s := range overlay {
		if !seen[s.Key] {
			out = append(out, s)
		}
	}
	return out
}

func supplementToSettings(rows []settingRow) []brave.Setting {
	sr := make([]presets.SettingRow, len(rows))
	for i, r := range rows {
		sr[i] = presets.SettingRow{Key: r.Key, Value: r.Value, Type: r.Type}
	}
	out, _ := presets.ConvertSettingRows(sr)
	return out
}

// WritePreset writes the given preset id and settings snapshot to the config file.
func WritePreset(presetID string, settings []brave.Setting) error {
	return write(&fileShapeNew{
		Preset: map[string]block{
			presetID: {Settings: settingsToRows(settings)},
		},
	})
}

// WriteApplyFile writes the given apply-file path and settings snapshot to the config file.
func WriteApplyFile(applyFilePath string, settings []brave.Setting) error {
	return write(&fileShapeNew{
		ApplyFile: applyFilePath,
		Settings:  settingsToRows(settings),
	})
}

// WriteSettings writes preset.custom.settings (e.g. after Custom apply in TUI).
func WriteSettings(settings []brave.Setting) error {
	return write(&fileShapeNew{
		Preset: map[string]block{
			"custom": {Settings: settingsToRows(settings)},
		},
	})
}

// WritePrivacyGuides writes preset.<baseID>.settings and supplement.privacy_guides.settings.
func WritePrivacyGuides(basePresetID string) error {
	var baseSettings []brave.Setting
	if basePresetID == "custom" {
		desired, err := Read()
		if err != nil || desired == nil || len(desired.Settings) == 0 {
			return fmt.Errorf("no custom settings in config to use as base")
		}
		baseSettings = desired.Settings
	} else {
		basePreset := presets.FindPreset(basePresetID)
		if basePreset == nil {
			return fmt.Errorf("base preset %q not found", basePresetID)
		}
		baseSettings = basePreset.Settings
	}
	supplement, err := presets.LoadPrivacyGuides()
	if err != nil {
		return err
	}
	f := &fileShapeNew{
		Preset: map[string]block{
			basePresetID: {Settings: settingsToRows(baseSettings)},
		},
		Supplement: map[string]block{
			"privacy_guides": {Settings: settingsToRows(supplement)},
		},
	}
	return write(f)
}

// PrivacyGuidesBaseFromConfig returns the preset ID to use as base when applying Privacy Guides.
// Uses existing config: preset (if it's a known preset ID), base_preset (for privacy-guides), or "custom".
// Returns "" if config is empty or has no usable base (apply_file).
func PrivacyGuidesBaseFromConfig() (string, error) {
	desired, err := Read()
	if err != nil || desired == nil {
		return "", err
	}
	if desired.Preset == "privacy-guides" && desired.BasePreset != "" {
		if presets.HasPreset(desired.BasePreset) || desired.BasePreset == "custom" {
			return desired.BasePreset, nil
		}
	}
	if desired.Preset == "custom" && len(desired.Settings) > 0 {
		return "custom", nil
	}
	if desired.Preset != "" && presets.HasPreset(desired.Preset) {
		return desired.Preset, nil
	}
	return "", nil
}

func write(f *fileShapeNew) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(f); err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("close encoder: %w", err)
	}
	return os.WriteFile(path, buf.Bytes(), 0600)
}

func settingsToRows(settings []brave.Setting) []settingRow {
	rows := make([]settingRow, len(settings))
	for i, s := range settings {
		rows[i] = settingRow{Key: s.Key, Value: s.Value, Type: string(s.Type)}
	}
	return rows
}
