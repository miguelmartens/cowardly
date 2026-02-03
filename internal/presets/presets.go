// Package presets defines Brave debloat presets loaded from configs/presets/*.yaml.
package presets

import "github.com/cowardly/cowardly/internal/brave"

// Preset is a named set of Brave settings (id, name, description, and key-value list).
type Preset struct {
	ID          string
	Name        string
	Description string
	Settings    []brave.Setting
}
