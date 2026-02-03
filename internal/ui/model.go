// Package ui implements the Bubble Tea TUI for cowardly.
package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/cowardly/cowardly/internal/config"
	"github.com/cowardly/cowardly/internal/presets"
)

type state int

const (
	stateMain state = iota
	statePreset
	stateCustom
	stateViewSettings
	stateResetConfirm
)

type model struct {
	state          state
	mainList       list.Model
	presetList     list.Model
	customIdx      int
	customOrder    []int // indices in display order (by category)
	customToggles  map[int]bool
	customSettings []config.CustomSetting
	viewKeys       []string
	viewScroll     int
	width          int
	height         int
	err            string
	msg            string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			MarginBottom(1)
	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)
	checkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
)

// NewModel returns the initial Bubble Tea model for the TUI.
func NewModel() model {
	mainItems := []list.Item{
		item{title: "Apply a preset", desc: "Quick Debloat, Maximum Privacy, Balanced, etc."},
		item{title: "Custom", desc: "Choose exactly which settings to apply"},
		item{title: "View current settings", desc: "See what's currently configured"},
		item{title: "Reset all to default", desc: "Remove all Brave policy settings"},
		item{title: "Exit", desc: "Quit cowardly"},
	}
	mainList := list.New(mainItems, list.NewDefaultDelegate(), 0, 0)
	mainList.Title = "Cowardly — Brave Browser Debloater"
	mainList.SetShowStatusBar(false)

	presetItems := make([]list.Item, 0, len(presets.All())+1)
	presetItems = append(presetItems, item{title: "← Back", desc: "Return to main menu"})
	for _, p := range presets.All() {
		presetItems = append(presetItems, item{title: p.Name, desc: p.Description})
	}
	presetList := list.New(presetItems, list.NewDefaultDelegate(), 0, 0)
	presetList.Title = "Choose a preset"
	presetList.SetShowStatusBar(false)

	customSettings := config.CustomSettings()
	toggles := make(map[int]bool)
	byCat := config.CustomSettingsByCategory()
	order := make([]int, 0, len(customSettings))
	for _, cat := range config.CategoryOrder {
		for _, i := range byCat[cat] {
			order = append(order, i)
			toggles[i] = false
		}
	}
	for i := range customSettings {
		toggles[i] = false
	}
	if len(order) == 0 {
		for i := range customSettings {
			order = append(order, i)
		}
	}

	viewKeys := []string{
		"MetricsReportingEnabled", "SafeBrowsingExtendedReportingEnabled",
		"UrlKeyedAnonymizedDataCollectionEnabled", "FeedbackSurveysEnabled",
		"BraveRewardsDisabled", "BraveWalletDisabled", "BraveVPNDisabled",
		"BraveAIChatEnabled", "TorDisabled", "SyncDisabled",
		"ShoppingListEnabled", "AlwaysOpenPdfExternally", "TranslateEnabled",
		"SpellcheckEnabled", "PromotionsEnabled", "DnsOverHttpsMode",
	}

	return model{
		state:          stateMain,
		mainList:       mainList,
		presetList:     presetList,
		customIdx:      0,
		customOrder:    order,
		customToggles:  toggles,
		customSettings: customSettings,
		viewKeys:       viewKeys,
		viewScroll:     0,
	}
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
