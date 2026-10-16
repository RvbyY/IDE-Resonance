package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

var (
	cursorColor = lipgloss.Color("212")

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.SetVirtualCursor(true)

	s := t.Styles()
	s.Cursor.Color = cursorColor
	s.Focused.Placeholder = focusedPlaceholderStyle
	s.Blurred.Placeholder = placeholderStyle
	s.Focused.CursorLine = cursorLineStyle
	s.Focused.CursorLineNumber = cursorLineStyle
	s.Focused.Base = focusedBorderStyle
	s.Blurred.Base = blurredBorderStyle
	s.Focused.EndOfBuffer = endOfBufferStyle
	s.Blurred.EndOfBuffer = endOfBufferStyle
	t.SetStyles(s)

	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}

type EditorModel struct {
	width  int
	height int
	keymap keymap
	help   help.Model
	inputs []textarea.Model
	focus  int
}

func newModel() EditorModel {
	m := EditorModel{
		inputs: make([]textarea.Model, initialInputs),
		help:   help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
	for i := range initialInputs {
		m.inputs[i] = newTextarea()
	}
	m.inputs[m.focus].Focus()
	m.updateKeybindings()
	return m
}

func (m EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m EditorModel) KeymapQuit(cmds []tea.Cmd) {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

func (m EditorModel) KeyMapNext(cmds []tea.Cmd) {
	m.inputs[m.focus].Blur()
	m.focus++
	if m.focus > len(m.inputs) - 1 {
		m.focus = 0
	}
	cmd := m.inputs[m.focus].Focus()
	cmds = append(cmds, cmd)
}

func (m EditorModel) KeyMapPrev(cmds []tea.Cmd) {
	m.inputs[m.focus].Blur()
	m.focus--
	if m.focus < 0 {
		m.focus = len(m.inputs) - 1
	}
	cmd := m.inputs[m.focus].Focus()
	cmds = append(cmds, cmd)
}

func (m EditorModel) KeymapAdd(cmds []tea.Cmd) {
	m.inputs = append(m.inputs, newTextarea())
}

func (m EditorModel) KeymapRemove(cmds []tea.Cmd) {
	m.inputs = m.inputs[:len(m.inputs) - 1]
	if m.focus > len(m.inputs) - 1 {
		m.focus = len(m.inputs) - 1
	}
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	events := map[*key.Binding]func([]tea.Cmd) {
		&m.keymap.quit: m.KeymapQuit,
		&m.keymap.next: m.KeyMapNext,
		&m.keymap.prev: m.KeyMapPrev,
		&m.keymap.add:  m.KeymapAdd,
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		for binding, handler := range events {
			if key.Matches(msg, *binding) {
				handler(cmds)
				break
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}
	m.updateKeybindings()
	m.sizeInputs()
	for i := range m.inputs {
		newModel, cmd := m.inputs[i].Update(msg)
		m.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *EditorModel) sizeInputs() {
	for i := range m.inputs {
		m.inputs[i].SetWidth(m.width / len(m.inputs))
		m.inputs[i].SetHeight(m.height - helpHeight)
	}
}

func (m *EditorModel) updateKeybindings() {
	m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
	m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
}

func (m EditorModel) inputViews() []string {
	var views []string
	for i := range m.inputs {
		views = append(views, m.inputs[i].View())
	}
	return views
}

func (m EditorModel) View() tea.View {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	v := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, m.inputViews()...) + "\n\n" + help)
	v.AltScreen = true
	return v
}

func (m EditorModel) Cursor() *tea.Cursor {
	focusedInput := m.inputs[m.focus]

	if focusedInput.VirtualCursor() {
		return nil
	}
	views := m.inputViews()
	c := focusedInput.Cursor()
	for i := range m.focus {
		c.X += lipgloss.Width(views[i])
	}
	return c
}
