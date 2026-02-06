// Package userconfig manages the user's desired Brave settings in ~/.config/cowardly/cowardly.yaml.
// Used for --reapply and to detect when settings have been reverted (e.g. by MDM after restart).
package userconfig

import (
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

// fileShape is the on-disk shape of cowardly.yaml.
type fileShape struct {
	Preset    string       `yaml:"preset,omitempty"`
	ApplyFile string       `yaml:"apply_file,omitempty"`
	Settings  []settingRow `yaml:"settings,omitempty"`
}

// DesiredState is the in-memory representation of the user's last-applied / desired state.
type DesiredState struct {
	Preset    string          // preset id, if last apply was a preset
	ApplyFile string          // path to file, if last apply was from file
	Settings  []brave.Setting // snapshot of settings (used by reapply)
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
	var f fileShape
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if len(f.Settings) == 0 {
		return nil, nil
	}
	rows := make([]presets.SettingRow, len(f.Settings))
	for i, r := range f.Settings {
		rows[i] = presets.SettingRow{Key: r.Key, Value: r.Value, Type: r.Type}
	}
	settings, err := presets.ConvertSettingRows(rows)
	if err != nil {
		return nil, fmt.Errorf("convert settings: %w", err)
	}
	return &DesiredState{
		Preset:    f.Preset,
		ApplyFile: f.ApplyFile,
		Settings:  settings,
	}, nil
}

// WritePreset writes the given preset id and settings snapshot to the config file.
func WritePreset(presetID string, settings []brave.Setting) error {
	return write(&fileShape{
		Preset:   presetID,
		Settings: settingsToRows(settings),
	})
}

// WriteApplyFile writes the given apply-file path and settings snapshot to the config file.
func WriteApplyFile(applyFilePath string, settings []brave.Setting) error {
	return write(&fileShape{
		ApplyFile: applyFilePath,
		Settings:  settingsToRows(settings),
	})
}

// WriteSettings writes only the settings snapshot (e.g. after Custom apply in TUI).
func WriteSettings(settings []brave.Setting) error {
	return write(&fileShape{Settings: settingsToRows(settings)})
}

func write(f *fileShape) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

func settingsToRows(settings []brave.Setting) []settingRow {
	rows := make([]settingRow, len(settings))
	for i, s := range settings {
		rows[i] = settingRow{Key: s.Key, Value: s.Value, Type: string(s.Type)}
	}
	return rows
}
