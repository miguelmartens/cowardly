// Package presets loads preset definitions from embedded YAML files.
package presets

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cowardly/cowardly/configs"
	"github.com/cowardly/cowardly/internal/brave"
	"gopkg.in/yaml.v3"
)

// policyKeyRegex matches Chromium/Brave policy key names (PascalCase, letters and digits).
var policyKeyRegex = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)

// presetFile is the on-disk shape of a preset YAML file.
type presetFile struct {
	ID          string       `yaml:"id"`
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Settings    []settingRow `yaml:"settings"`
}

// SettingRow is one key/value/type row as in preset or config YAML. Exported for use by userconfig.
type SettingRow struct {
	Key   string      `yaml:"key"`
	Value interface{} `yaml:"value"`
	Type  string      `yaml:"type"`
}

// settingRow is an alias for internal use (presetFile, settingsFile).
type settingRow = SettingRow

var cachedPresets []Preset

// All returns built-in presets loaded from configs/presets/*.yaml (embedded).
// Order is determined by filename (01-quick, 02-max-privacy, ...).
// On load error, logs and returns nil.
func All() []Preset {
	list, _ := AllWithError()
	return list
}

// AllWithError returns presets and any load/validation error. Use this at startup to surface failures.
func AllWithError() ([]Preset, error) {
	if cachedPresets != nil {
		return cachedPresets, nil
	}
	list, err := LoadFromFS(configs.PresetsFS, "presets")
	if err != nil {
		log.Printf("presets: load failed: %v", err)
		return nil, err
	}
	cachedPresets = list
	return cachedPresets, nil
}

// LoadFromFS reads preset YAML files from the given fs.FS under the given dir (e.g. "presets").
// Files are sorted by name so order is deterministic (01-quick.yaml, 02-max-privacy.yaml, ...).
func LoadFromFS(fsys fs.FS, dir string) ([]Preset, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var out []Preset
	for _, name := range names {
		path := dir + "/" + name
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}
		var pf presetFile
		if err := yaml.Unmarshal(data, &pf); err != nil {
			return nil, fmt.Errorf("parse %q: %w", path, err)
		}
		settings, err := convertSettings(pf.Settings)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", path, err)
		}
		out = append(out, Preset{
			ID:          pf.ID,
			Name:        pf.Name,
			Description: pf.Description,
			Settings:    settings,
		})
	}
	return out, nil
}

// ConvertSettingRows converts YAML setting rows to brave settings. Used by presets and userconfig.
func ConvertSettingRows(rows []SettingRow) ([]brave.Setting, error) {
	return convertSettings(rows)
}

func convertSettings(rows []settingRow) ([]brave.Setting, error) {
	out := make([]brave.Setting, 0, len(rows))
	for i, r := range rows {
		if r.Key == "" {
			return nil, fmt.Errorf("setting %d: key is empty", i)
		}
		if !policyKeyRegex.MatchString(r.Key) {
			return nil, fmt.Errorf("setting %d %q: key must match [A-Za-z][A-Za-z0-9]* (Chromium policy name)", i, r.Key)
		}
		val, vt, err := normalizeValue(r.Value, r.Type)
		if err != nil {
			return nil, fmt.Errorf("setting %d %q: %w", i, r.Key, err)
		}
		out = append(out, brave.Setting{Key: r.Key, Value: val, Type: vt})
	}
	return out, nil
}

func normalizeValue(raw interface{}, typeStr string) (interface{}, brave.ValueType, error) {
	switch strings.ToLower(typeStr) {
	case "bool", "boolean":
		b, err := toBool(raw)
		return b, brave.TypeBool, err
	case "integer", "int":
		n, err := toInt(raw)
		return n, brave.TypeInteger, err
	case "string":
		s, err := toString(raw)
		return s, brave.TypeString, err
	default:
		return nil, "", fmt.Errorf("unknown type %q", typeStr)
	}
}

func toBool(v interface{}) (bool, error) {
	if b, ok := v.(bool); ok {
		return b, nil
	}
	// YAML may unmarshal as string "true"/"false"
	if s, ok := v.(string); ok {
		switch strings.ToLower(s) {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no", "":
			return false, nil
		}
	}
	return false, fmt.Errorf("cannot convert %T to bool", v)
}

func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case float64:
		return int(n), nil
	case json.Number:
		i, err := n.Int64()
		return int(i), err
	}
	return 0, fmt.Errorf("cannot convert %T to int", v)
}

func toString(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("cannot convert %T to string", v)
}

// settingsFile is the on-disk shape for YAML that contains only a settings list (export/import).
type settingsFile struct {
	Settings []settingRow `yaml:"settings"`
}

// LoadSettingsFromFile reads a YAML file (with a "settings" list) and returns brave settings.
func LoadSettingsFromFile(path string) ([]brave.Setting, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var f settingsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	return convertSettings(f.Settings)
}

// WriteSettingsToFile writes settings to a YAML file (same format as preset settings).
func WriteSettingsToFile(path string, settings []brave.Setting) error {
	rows := make([]settingRow, len(settings))
	for i, s := range settings {
		rows[i] = settingRow{Key: s.Key, Value: s.Value, Type: string(s.Type)}
	}
	f := settingsFile{Settings: rows}
	data, err := yaml.Marshal(&f)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}
