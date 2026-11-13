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
	editor     EditorModel
	editorOpen bool
	tab        modelTab
}

func initialModel() model {
	return model{
		choices:  []string{"Open in Shell", "Open in app Window"},
		selected: make(map[int]struct{}),
		list:     initialListModel(),
		editor:   newModel(),
		tab:      initialTabModel(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.list.Init(),
		m.editor.Init(),
		m.tab.InitTab(),
	)
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

func (m model) OpenSelectedPreset() (tea.Model, tea.Cmd) {
	selected := m.list.list.SelectedItem()
	preset, ok := selected.(item)

	if !ok || len(m.editor.inputs) == 0 {
		return m, nil
	}
	m.editor.inputs[0].SetValue(preset.title)
	if len(m.editor.inputs) > 1 {
		m.editor.inputs[1].SetValue(preset.description)
	} else {
		m.editor.inputs[0].SetValue(preset.title + "\n\n" + preset.description)
	}
	m.showList = false
	m.editorOpen = true
	m.tab.activeTab = 1
	return m, m.editor.Init()
}

func (m model) ResizeWindow(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	updatedList, listCmd := m.list.Update(msg)
	m.list = updatedList.(listModel)
	cmds = append(cmds, listCmd)
	updatedEditor, listCmd := m.editor.Update(msg)
	m.editor = updatedEditor.(EditorModel)
	cmds = append(cmds, listCmd)
	return m, tea.Batch(cmds...)
}

func isEscape(msg tea.KeyPressMsg) bool {
	k := msg.Key()
	return k.Code == tea.KeyEscape || (k.Code == '[' && k.Mod&tea.ModCtrl != 0)
}

func (m model) IsShowList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case isEscape(keyMsg):
			m.showList = false
			return m, nil
		case keyMsg.String() == "enter":
			return m.OpenSelectedPreset()
		}
	}
	updatedList, cmd := m.list.Update(msg)
	m.list = updatedList.(listModel)
	return m, cmd
}

func (m model) IsEditorOpen(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && isEscape(keyMsg) {
		if !m.editor.saving {
			m.editorOpen = false
			m.showList = true
			return m, nil
		}
	}
	updatedEditor, cmd := m.editor.Update(msg)
	m.editor = updatedEditor.(EditorModel)
	m.updatedEditorTabTitle()
	return m, cmd
}

func (m *model) updatedEditorTabTitle() {
	filename := "untitled"

	if m.editor.editorIndex >= 0 &&
		m.editor.editorIndex < len(m.editor.storefilenames) &&
		m.editor.storefilenames[m.editor.editorIndex] != "" {
		filename = m.editor.storefilenames[m.editor.editorIndex]
	}
	if len(m.tab.Tabs) > 1 {
		m.tab.Tabs[1] = filename
	}
}

func (m model) editorOpenEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editor.saving {
		updated, cmd := m.editor.Update(msg)
		m.editor = updated.(EditorModel)
		return m, cmd
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if isEscape(keyMsg) && !m.editor.saving {
			m.editorOpen = false
			m.showList = true
			return m, nil
		}
		switch keyMsg.String() {
		case "tab", "right", "l":
			m.tab.activeTab = min(m.tab.activeTab+1, 1)
			m.editor.focus = focusEditor
			m.editor.inputs[m.editor.editorIndex].Focus()
			return m, nil

		case "shift+tab", "left", "h":
			m.tab.activeTab = max(m.tab.activeTab-1, 0)
			m.editor.focus = focusFiles
			m.editor.inputs[m.editor.editorIndex].Blur()
			return m, nil
		}
	}
	if m.tab.activeTab == 0 {
		updated, cmd := m.editor.Update(msg)
		m.editor = updated.(EditorModel)
		return m, cmd
	}
	updated, cmd := m.editor.Update(msg)
	m.editor = updated.(EditorModel)
	return m, cmd
}

func (m model) fileOpenEvent(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fileOpenedMsg:
		if msg.index >= 0 && msg.index < len(m.editor.inputs) {
			m.editor.inputs[msg.index].SetValue(msg.content)
			if msg.index >= len(m.editor.storefilenames) {
				m.editor.storefilenames = append(m.editor.storefilenames,
					make([]string, msg.index-len(m.editor.storefilenames)+1)...)
			}
			m.editor.storefilenames[msg.index] = msg.filename
			m.editor.editorIndex = msg.index
			m.tab.activeTab = 1
			m.updatedEditorTabTitle()
		}
	}
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	commands := map[string]func() (tea.Model, tea.Cmd){
		"ctrl+c": m.Echap, "q": m.Echap,
		"ctrl+z": m.Suspend, "space": m.Screen, "up": m.MoveUp,
		"k": m.MoveUp, "down": m.MoveDown, "j": m.MoveDown, "enter": m.Select,
	}

	if _, ok := msg.(tea.WindowSizeMsg); ok {
		return m.ResizeWindow(msg)
	}
	if _, ok := msg.(fileOpenedMsg); ok {
		return m.fileOpenEvent(msg)
	}
	if m.showList {
		return m.IsShowList(msg)
	}
	if m.editorOpen {
		return m.editorOpenEvent(msg)
	}
	switch msg := msg.(type) {
	case tea.ResumeMsg:
		m.suspending = false
		return m, nil
	case tea.KeyPressMsg:
		if isEscape(msg) {
			return m.Echap()
		}
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
	if m.editorOpen {
		return m.viewTab()
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
