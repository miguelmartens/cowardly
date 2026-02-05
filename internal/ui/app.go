package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cowardly/cowardly/internal/brave"
	"github.com/cowardly/cowardly/internal/config"
	"github.com/cowardly/cowardly/internal/presets"
	"github.com/mattn/go-runewidth"
)

type applyPresetMsg struct{ idx int }
type applyCustomMsg struct{}
type resetDoneMsg struct {
	err            error
	backupPath     string
	hadManaged     bool
	managedRemoved bool
}
type backupsListMsg struct {
	paths []string
	err   error
}
type backupDoneMsg struct {
	err error
	msg string
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
					return m, func() tea.Msg {
						paths, err := brave.ListBackups()
						return backupsListMsg{paths: paths, err: err}
					}
				case 5:
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
			case "y", "Y", "enter":
				if brave.BraveRunning() {
					m.err = "Quit Brave first (Cmd+Q), then run Reset again. If Brave is running, it can restore the plist from memory and the reset will not stick."
					m.state = stateMain
					return m, nil
				}
				return m, func() tea.Msg {
					backupPath, _ := brave.BackupUserPlist()
					hadManaged, managedRemoved, err := brave.Reset()
					return resetDoneMsg{err: err, backupPath: backupPath, hadManaged: hadManaged, managedRemoved: managedRemoved}
				}
			case "n", "N", "q", "esc":
				m.state = stateMain
				return m, nil
			}
			return m, nil

		case stateBackups:
			switch msg.String() {
			case "q", "esc":
				m.state = stateMain
				return m, nil
			case "enter":
				if len(m.backupPaths) == 0 {
					return m, nil
				}
				idx := m.backupList.Index()
				if idx < 0 || idx >= len(m.backupPaths) {
					return m, nil
				}
				m.confirmPath = m.backupPaths[idx]
				m.confirmAction = "restore"
				m.state = stateBackupConfirm
				return m, nil
			case "d":
				if len(m.backupPaths) == 0 {
					return m, nil
				}
				idx := m.backupList.Index()
				if idx < 0 || idx >= len(m.backupPaths) {
					return m, nil
				}
				m.confirmPath = m.backupPaths[idx]
				m.confirmAction = "delete"
				m.state = stateBackupConfirm
				return m, nil
			}
			var cmd tea.Cmd
			m.backupList, cmd = m.backupList.Update(msg)
			return m, cmd

		case stateBackupConfirm:
			switch msg.String() {
			case "y", "Y", "enter":
				path := m.confirmPath
				action := m.confirmAction
				return m, func() tea.Msg {
					var err error
					var doneMsg string
					if action == "restore" {
						err = brave.RestoreFromBackup(path)
						if err == nil {
							doneMsg = "Restored backup. Restart Brave for changes to take effect."
						}
					} else {
						err = brave.DeleteBackup(path)
						if err == nil {
							doneMsg = "Backup deleted."
						}
					}
					if err != nil {
						return backupDoneMsg{err: err, msg: ""}
					}
					return backupDoneMsg{err: nil, msg: doneMsg}
				}
			case "n", "N", "q", "esc":
				m.state = stateBackups
				m.confirmPath = ""
				m.confirmAction = ""
				return m, nil
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mainList.SetSize(msg.Width/2, msg.Height/2)
		m.presetList.SetSize(msg.Width/2, msg.Height/2)
		m.backupList.SetSize(msg.Width/2, msg.Height/2)
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
			m.msg += fmt.Sprintf("Backed up to:\n%s\n\n", path)
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
			m.msg += fmt.Sprintf("Backed up to:\n%s\n\n", path)
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
			if msg.backupPath != "" {
				m.msg = "Backed up to:\n" + msg.backupPath + "\n\n"
			}
			if !msg.hadManaged {
				m.msg += "User preferences cleared. No managed policy file was present, so no authentication was needed. Restart Brave."
			} else if msg.managedRemoved {
				m.msg += "All Brave policy settings reset (including managed). Restart Brave."
			} else {
				m.msg += "User preferences cleared. The managed policy file could not be removed (did you cancel the authentication?). Run Reset again and approve the dialog."
			}
		}
		m.state = stateMain
		return m, nil

	case backupsListMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			m.state = stateMain
			return m, nil
		}
		m.backupPaths = msg.paths
		items := make([]list.Item, len(msg.paths))
		for i, p := range msg.paths {
			items[i] = item{title: filepath.Base(p), desc: p}
		}
		m.backupList = list.New(items, braveListDelegate(), 0, 0)
		m.backupList.Title = "Backups (Enter restore, d delete, esc back)"
		m.backupList.Styles = braveListStyles()
		m.backupList.SetShowStatusBar(false)
		if m.width > 0 && m.height > 0 {
			m.backupList.SetSize(m.width/2, m.height/2)
		}
		m.state = stateBackups
		return m, nil

	case backupDoneMsg:
		m.confirmPath = ""
		m.confirmAction = ""
		m.state = stateMain
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.msg = msg.msg
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.err != "" {
		return errorStyle.Render("Error: "+m.err) + "\n\nPress any key..."
	}
	if m.msg != "" {
		w := m.width
		if w <= 0 {
			w = 72
		}
		return successStyle.Width(w).Render(m.msg) + "\n\nPress any key..."
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
			dimStyle.Render("Quit Brave (Cmd+Q) before resetting, or the reset may not stick.\n") +
			dimStyle.Render("You will only see an authentication dialog if a managed policy file exists.\n") + "\n" +
			"Press " + activeStyle.Render("y") + " or " + activeStyle.Render("Enter") + " to confirm, " + activeStyle.Render("n") + " or " + activeStyle.Render("Esc") + " to cancel."
	case stateBackups:
		if len(m.backupPaths) == 0 {
			return titleStyle.Render("Backups") + "\n\n" + dimStyle.Render("No backups yet. Apply a preset or reset to create one.") + "\n\n" + dimStyle.Render("esc back")
		}
		return titleStyle.Render("Backups") + "\n" + m.backupList.View() + dimStyle.Render("\nEnter restore  d delete  esc back")
	case stateBackupConfirm:
		name := filepath.Base(m.confirmPath)
		if m.confirmAction == "restore" {
			return titleStyle.Render("Restore backup?") + "\n\n" +
				"Restore " + activeStyle.Render(name) + " over current user preferences.\n\n" +
				"Press " + activeStyle.Render("y") + " or " + activeStyle.Render("Enter") + " to restore, " + activeStyle.Render("n") + " or " + activeStyle.Render("Esc") + " to cancel."
		}
		return titleStyle.Render("Delete backup?") + "\n\n" +
			"Permanently delete " + activeStyle.Render(name) + ".\n\n" +
			"Press " + activeStyle.Render("y") + " or " + activeStyle.Render("Enter") + " to delete, " + activeStyle.Render("n") + " or " + activeStyle.Render("Esc") + " to cancel."
	default:
		return ""
	}
}

func (m model) customView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Custom — Toggle settings (Space), Apply (Enter)"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("[a] select all  [n] select none  [esc] back"))
	b.WriteString("\n\n") // raw newlines so the next line is not inside any style

	// Inline(true) prevents block-level reflow so the category stays at column 0
	catStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff631c")).Inline(true)
	byCat := config.CustomSettingsByCategory()
	for _, cat := range config.CategoryOrder {
		idxs, ok := byCat[cat]
		if !ok || len(idxs) == 0 {
			continue
		}
		b.WriteString(catStyle.Render(cat))
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
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff631c")).Inline(true)
	var b strings.Builder
	b.WriteString(headerStyle.Render("Current Brave settings"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("(managed overrides user; enforced = what Brave uses)"))
	b.WriteString("\n\n") // raw newlines so the list is not inside any style
	// Inline styles so each line stays at column 0 (no block reflow)
	checkIconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4caf50")).Inline(true)
	unsetIconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Inline(true)
	maxKeyWidth := 0
	for _, key := range m.viewKeys {
		if w := runewidth.StringWidth(key); w > maxKeyWidth {
			maxKeyWidth = w
		}
	}
	prefixRaw := "  ✓ " // unstyled for width calculation
	for _, key := range m.viewKeys {
		managedVal, managedOK := brave.ReadManaged(key)
		userVal, userOK := brave.Read(key)
		paddedKey := key + strings.Repeat(" ", maxKeyWidth-runewidth.StringWidth(key))
		var valuePart, suffix string
		var line string
		if managedOK {
			valuePart = managedVal
			suffix = " (enforced)"
			line = "  " + checkIconStyle.Render("✓ ") + paddedKey + " = " + managedVal + suffix + "\n"
		} else if userOK {
			valuePart = userVal
			suffix = " (user)"
			line = "  " + checkIconStyle.Render("✓ ") + paddedKey + " = " + userVal + suffix + "\n"
		} else {
			valuePart = "(not set)"
			suffix = ""
			line = "  " + unsetIconStyle.Render("○ ") + paddedKey + " = (not set)\n"
		}
		// When line would wrap, break so "= value (enforced)" aligns in a column
		lineWidth := runewidth.StringWidth(prefixRaw+paddedKey) + runewidth.StringWidth(" = "+valuePart+suffix)
		if m.width > 0 && lineWidth > m.width {
			indent := runewidth.StringWidth(prefixRaw + paddedKey)
			if managedOK {
				b.WriteString("  " + checkIconStyle.Render("✓ ") + paddedKey + "\n")
				b.WriteString(strings.Repeat(" ", indent) + "= " + managedVal + " (enforced)\n")
			} else if userOK {
				b.WriteString("  " + checkIconStyle.Render("✓ ") + paddedKey + "\n")
				b.WriteString(strings.Repeat(" ", indent) + "= " + userVal + " (user)\n")
			} else {
				b.WriteString("  " + unsetIconStyle.Render("○ ") + paddedKey + "\n")
				b.WriteString(strings.Repeat(" ", indent) + "= (not set)\n")
			}
		} else {
			b.WriteString(line)
		}
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("q/esc back"))
	return b.String()
}
