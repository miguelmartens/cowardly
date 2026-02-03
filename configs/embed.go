// Package configs holds configuration file templates and embedded preset data.
package configs

import "embed"

// PresetsFS contains the preset YAML files in configs/presets/*.yaml.
// Preset order is determined by filename (01-quick.yaml, 02-max-privacy.yaml, ...).
// To add a preset, add a new .yaml file in configs/presets/ and rebuild.
//
//go:embed presets/*.yaml
var PresetsFS embed.FS
