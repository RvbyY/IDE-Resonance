package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"os"
)

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type model struct {
	altscreen  bool
	quitting   bool
	suspending bool
	choices    []string
	cursor     int
	selected   map[int]struct{}
}

func initialModel() model {
	return model{
		choices:  []string{"Open in Shell", "Open in app Window"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ResumeMsg:
		m.suspending = false
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "ctrl+z":
			m.suspending = true
			return m, tea.Suspend
		case "space":
			var cmd tea.Cmd
			m.altscreen = !m.altscreen
			return m, cmd
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m model) DisplayText(s string) string {
	s += "Resonance IDE\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}
	return s
}

func (m model) ViewNewScreen() tea.View {
	var mode string

	if m.altscreen {
		mode = " altscreen mode "
	} else {
		mode = " inline mode "
	}
	s := fmt.Sprintf("\n\nYou're in %s\n\n", keywordStyle.Render(mode))
	s += helpStyle.Render("space: switch modes • ctrl-z: suspend • q: exit\n\n")
	if m.suspending {
		v := tea.NewView("")
		v.AltScreen = m.altscreen
		return v
	}
	if m.quitting {
		v := tea.NewView("Bye!\n")
		v.AltScreen = m.altscreen
		return v
	}
	v := tea.NewView(m.DisplayText(s))
	v.AltScreen = m.altscreen
	return v
}

func (m model) View() tea.View {
	return m.ViewNewScreen()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(84)
	}
}
