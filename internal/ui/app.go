package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cowardly/cowardly/internal/brave"
	"github.com/cowardly/cowardly/internal/config"
	"github.com/cowardly/cowardly/internal/presets"
)

type applyPresetMsg struct{ idx int }
type applyCustomMsg struct{}
type resetDoneMsg struct {
	err        error
	backupPath string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateMain:
			m.err = ""
			m.msg = ""
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				sel := m.mainList.Index()
				switch sel {
				case 0:
					m.state = statePreset
					m.presetList.ResetSelected()
					return m, nil
				case 1:
					m.state = stateCustom
					m.customIdx = 0
					return m, nil
				case 2:
					m.state = stateViewSettings
					m.viewScroll = 0
					return m, nil
				case 3:
					m.state = stateResetConfirm
					return m, nil
				case 4:
					return m, tea.Quit
				}
			}
			var cmd tea.Cmd
			m.mainList, cmd = m.mainList.Update(msg)
			return m, cmd

		case statePreset:
			switch msg.String() {
			case "q", "esc":
				m.state = stateMain
				return m, nil
			case "enter":
				idx := m.presetList.Index()
				if idx == 0 {
					m.state = stateMain
					return m, nil
				}
				return m, func() tea.Msg { return applyPresetMsg{idx: idx - 1} }
			}
			var cmd tea.Cmd
			m.presetList, cmd = m.presetList.Update(msg)
			return m, cmd

		case stateCustom:
			n := len(m.customOrder)
			switch msg.String() {
			case "q", "esc":
				m.state = stateMain
				return m, nil
			case "enter":
				return m, func() tea.Msg { return applyCustomMsg{} }
			case " ":
				if m.customIdx >= 0 && m.customIdx < n {
					idx := m.customOrder[m.customIdx]
					m.customToggles[idx] = !m.customToggles[idx]
				}
				return m, nil
			case "up", "k":
				if m.customIdx > 0 {
					m.customIdx--
				}
				return m, nil
			case "down", "j":
				if m.customIdx < n-1 {
					m.customIdx++
				}
				return m, nil
			case "a":
				for i := range m.customSettings {
					m.customToggles[i] = true
				}
				return m, nil
			case "n":
				for i := range m.customSettings {
					m.customToggles[i] = false
				}
				return m, nil
			}
			return m, nil

		case stateViewSettings:
			switch msg.String() {
			case "q", "esc", "enter":
				m.state = stateMain
				return m, nil
			case "up", "k":
				if m.viewScroll > 0 {
					m.viewScroll--
				}
				return m, nil
			case "down", "j":
				maxScroll := len(m.viewKeys) - 1
				if m.viewScroll < maxScroll {
					m.viewScroll++
				}
				return m, nil
			}
			return m, nil

		case stateResetConfirm:
			switch msg.String() {
			case "y", "Y":
				return m, func() tea.Msg {
					backupPath, _ := brave.BackupUserPlist()
					err := brave.Reset()
					return resetDoneMsg{err: err, backupPath: backupPath}
				}
			case "n", "N", "q", "esc":
				m.state = stateMain
				return m, nil
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mainList.SetSize(msg.Width/2, msg.Height/2)
		m.presetList.SetSize(msg.Width/2, msg.Height/2)
		return m, nil

	case applyPresetMsg:
		plist := presets.All()
		if msg.idx < 0 || msg.idx >= len(plist) {
			m.state = stateMain
			return m, nil
		}
		p := plist[msg.idx]
		if brave.BraveRunning() {
			m.msg = "Brave is running — quit for a clean apply. "
		} else {
			m.msg = ""
		}
		if path, err := brave.BackupUserPlist(); err == nil {
			m.msg += fmt.Sprintf("Backed up to %s. ", path)
		}
		managed, err := brave.ApplySettings(p.Settings)
		if err != nil {
			m.err = err.Error()
			m.msg = ""
		} else if managed {
			m.msg += fmt.Sprintf("Applied preset: %s (enforced). Restart Brave for changes.", p.Name)
		} else {
			m.msg += fmt.Sprintf("Applied preset: %s. Restart Brave. For enforced policies, approve the macOS authentication dialog when you apply.", p.Name)
		}
		m.state = stateMain
		return m, nil

	case applyCustomMsg:
		var toApply []brave.Setting
		for i, cs := range m.customSettings {
			if m.customToggles[i] {
				toApply = append(toApply, brave.Setting{Key: cs.Key, Value: cs.Value, Type: cs.Type})
			}
		}
		if len(toApply) == 0 {
			m.msg = "No settings selected. Toggle with Space, Apply with Enter."
			m.state = stateMain
			return m, nil
		}
		if brave.BraveRunning() {
			m.msg = "Brave is running — quit for a clean apply. "
		} else {
			m.msg = ""
		}
		if path, err := brave.BackupUserPlist(); err == nil {
			m.msg += fmt.Sprintf("Backed up to %s. ", path)
		}
		managed, err := brave.ApplySettings(toApply)
		if err != nil {
			m.err = err.Error()
			m.msg = ""
		} else if managed {
			m.msg += fmt.Sprintf("Applied %d setting(s) (enforced). Restart Brave for changes.", len(toApply))
		} else {
			m.msg += fmt.Sprintf("Applied %d setting(s). Restart Brave. For enforced policies, approve the macOS authentication dialog when you apply.", len(toApply))
		}
		m.state = stateMain
		return m, nil

	case resetDoneMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.msg = "All Brave policy settings reset. Restart Brave."
			if msg.backupPath != "" {
				m.msg = "Backed up to " + msg.backupPath + ". " + m.msg
			}
			m.msg += " If you cancelled the authentication dialog, the managed plist may still exist; run Reset again and approve to remove it."
		}
		m.state = stateMain
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.err != "" {
		return errorStyle.Render("Error: "+m.err) + "\n\nPress any key..."
	}
	if m.msg != "" {
		return successStyle.Render(m.msg) + "\n\nPress any key..."
	}

	switch m.state {
	case stateMain:
		return titleStyle.Render("Cowardly — Brave Browser Debloater") + "\n" + m.mainList.View() + dimStyle.Render("\n↑/k up  ↓/j down  enter select  q quit")
	case statePreset:
		return titleStyle.Render("Choose a preset") + "\n" + m.presetList.View() + dimStyle.Render("\nenter apply  esc back")
	case stateCustom:
		return m.customView()
	case stateViewSettings:
		return m.viewSettingsView()
	case stateResetConfirm:
		return titleStyle.Render("Reset all settings?") + "\n\n" +
			"This will remove ALL Brave policy settings and restore defaults.\n\n" +
			"Type " + activeStyle.Render("y") + " to confirm, " + activeStyle.Render("n") + " to cancel."
	default:
		return ""
	}
}

func (m model) customView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Custom — Toggle settings (Space), Apply (Enter)"))
	b.WriteString(dimStyle.Render("\n[a] select all  [n] select none  [esc] back\n\n"))

	byCat := config.CustomSettingsByCategory()
	for _, cat := range config.CategoryOrder {
		idxs, ok := byCat[cat]
		if !ok || len(idxs) == 0 {
			continue
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62")).Render(cat))
		b.WriteString("\n")
		for _, i := range idxs {
			cs := m.customSettings[i]
			on := m.customToggles[i]
			mark := " "
			if on {
				mark = checkStyle.Render("✓")
			}
			cursor := " "
			if m.customIdx < len(m.customOrder) && m.customOrder[m.customIdx] == i {
				cursor = activeStyle.Render(">")
			}
			b.WriteString(fmt.Sprintf("  %s %s %s %s\n", cursor, mark, cs.DisableWord, cs.Label))
		}
		b.WriteString("\n")
	}
	b.WriteString(dimStyle.Render("↑/k up  ↓/j down  space toggle  enter apply  esc back"))
	return b.String()
}

func (m model) viewSettingsView() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	var b strings.Builder
	b.WriteString(headerStyle.Render("Current Brave settings"))
	b.WriteString(dimStyle.Render("\n(managed overrides user; enforced = what Brave uses)\n\n"))
	for _, key := range m.viewKeys {
		managedVal, managedOK := brave.ReadManaged(key)
		userVal, userOK := brave.Read(key)
		if managedOK {
			b.WriteString("  " + checkStyle.Render("✓ ") + key + " = " + managedVal + " (enforced)\n")
		} else if userOK {
			b.WriteString("  " + checkStyle.Render("✓ ") + key + " = " + userVal + " (user)\n")
		} else {
			b.WriteString("  " + dimStyle.Render("○ ") + key + " = (not set)\n")
		}
	}
	b.WriteString(dimStyle.Render("\nq/esc back"))
	return b.String()
}
