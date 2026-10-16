package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	list       listModel
	showList   bool
}

func initialModel() model {
	return model{
		choices:  []string{"Open in Shell", "Open in app Window"},
		selected: make(map[int]struct{}),
		list:     initialListModel(),
	}
}

func (m model) Init() tea.Cmd {
	return m.list.Init()
}

func (m model) Echap() (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m model) Suspend() (tea.Model, tea.Cmd) {
	m.suspending = true
	return m, tea.Suspend
}

func (m model) Screen() (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.altscreen = !m.altscreen
	return m, cmd
}

func (m model) MoveUp() (tea.Model, tea.Cmd) {
	if m.cursor > 0 {
		m.cursor--
	}
	return m, nil
}

func (m model) MoveDown() (tea.Model, tea.Cmd) {
	if m.cursor < len(m.choices)-1 {
		m.cursor++
	}
	return m, nil
}

func (m model) Select() (tea.Model, tea.Cmd) {
	if m.cursor == 0 {
		m.showList = true
		return m, nil
	}
	_, ok := m.selected[m.cursor]
	if ok {
		delete(m.selected, m.cursor)
	} else {
		m.selected[m.cursor] = struct{}{}
	}
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	commands := map[string]func() (tea.Model, tea.Cmd){
		"ctrl+c": m.Echap,
		"q":      m.Echap,
		"esc":    m.Echap,
		"ctrl+z": m.Suspend,
		"space":  m.Screen,
		"up":     m.MoveUp,
		"k":      m.MoveUp,
		"down":   m.MoveDown,
		"j":      m.MoveDown,
		"enter":  m.Select,
	}

	if m.showList {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			m.showList = false
			return m, nil
		}
		updatedList, cmd := m.list.Update(msg)
		m.list = updatedList.(listModel)
		return m, cmd
	}
	switch msg := msg.(type) {
	case tea.ResumeMsg:
		m.suspending = false
		return m, nil
	case tea.KeyPressMsg:
		if command, ok := commands[msg.String()]; ok {
			return command()
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
	if m.showList {
		return m.list.View()
	}
	return m.ViewNewScreen()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Oh, there's been an error: %v", err)
		os.Exit(84)
	}
}
