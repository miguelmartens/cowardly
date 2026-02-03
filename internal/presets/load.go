// Package presets loads preset definitions from embedded YAML files.
package presets

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/cowardly/cowardly/configs"
	"github.com/cowardly/cowardly/internal/brave"
	"gopkg.in/yaml.v3"
)

// presetFile is the on-disk shape of a preset YAML file.
type presetFile struct {
	ID          string       `yaml:"id"`
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Settings    []settingRow `yaml:"settings"`
}

type settingRow struct {
	Key   string      `yaml:"key"`
	Value interface{} `yaml:"value"`
	Type  string      `yaml:"type"`
}

var cachedPresets []Preset

// All returns built-in presets loaded from configs/presets/*.yaml (embedded).
// Order is determined by filename (01-quick, 02-max-privacy, ...).
// On load error, logs and returns nil.
func All() []Preset {
	if cachedPresets != nil {
		return cachedPresets
	}
	list, err := LoadFromFS(configs.PresetsFS, "presets")
	if err != nil {
		log.Printf("presets: load failed: %v", err)
		return nil
	}
	cachedPresets = list
	return cachedPresets
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

func convertSettings(rows []settingRow) ([]brave.Setting, error) {
	out := make([]brave.Setting, 0, len(rows))
	for i, r := range rows {
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
