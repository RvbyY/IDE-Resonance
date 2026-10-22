package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

type filesSavedMsg struct {
	count int
}

type saveErrorMsg struct {
	err error
}

type fileOpenedMsg struct {
	index    int
	filename string
	content  string
}

type fileOpenErrorMsg struct {
	err error
}

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
	next, prev, add, remove, save, quit key.Binding
}

func saveFilesCmd(filenames []string, contents []string) tea.Cmd {
	return func() tea.Msg {
		if len(filenames) != len(contents) {
			return saveErrorMsg{
				err: fmt.Errorf(
					"expected %d filenames, got %d",
					len(contents),
					len(filenames),
				),
			}
		}
		for i, filename := range filenames {
			if filename == "" {
				return saveErrorMsg{
					err: fmt.Errorf("filename for editor %d is empty", i+1),
				}
			}
			path := filepath.Join(".", filename)
			if err := os.WriteFile(path, []byte(contents[i]), 0644); err != nil {
				return saveErrorMsg{
					err: fmt.Errorf("writing %q: %w", path, err),
				}
			}
		}
		return filesSavedMsg{count: len(contents)}
	}
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

func newFilenameInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "filename.txt"
	return input
}

type EditorModel struct {
	width     int
	height    int
	keymap    keymap
	help      help.Model
	inputs    []textarea.Model
	focus     int
	filenames []textinput.Model
	saving    bool
	saveFocus int
	status    string
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
			save: key.NewBinding(
				key.WithKeys("ctrl+s"),
				key.WithHelp("ctrl+s", "save files"),
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

func (m *EditorModel) KeymapQuit(cmds []tea.Cmd) {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
}

func (m *EditorModel) KeyMapNext(cmds []tea.Cmd) {
	m.inputs[m.focus].Blur()
	m.focus++
	if m.focus > len(m.inputs)-1 {
		m.focus = 0
	}
	cmd := m.inputs[m.focus].Focus()
	cmds = append(cmds, cmd)
}

func (m *EditorModel) KeyMapPrev(cmds []tea.Cmd) {
	m.inputs[m.focus].Blur()
	m.focus--
	if m.focus < 0 {
		m.focus = len(m.inputs) - 1
	}
	cmd := m.inputs[m.focus].Focus()
	cmds = append(cmds, cmd)
}

func (m *EditorModel) KeymapAdd(cmds []tea.Cmd) {
	m.inputs = append(m.inputs, newTextarea())
}

func (m *EditorModel) KeymapRemove(cmds []tea.Cmd) {
	m.inputs = m.inputs[:len(m.inputs)-1]
	if m.focus > len(m.inputs)-1 {
		m.focus = len(m.inputs) - 1
	}
}

func (m *EditorModel) beginSave() tea.Cmd {
	m.saving = true
	m.saveFocus = 0
	m.filenames = make([]textinput.Model, len(m.inputs))
	for i := range m.filenames {
		m.filenames[i] = newFilenameInput()
	}
	return m.filenames[0].Focus()
}

func (m *EditorModel) updateSave(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.saving = false
		return nil
	case "tab", "down":
		m.filenames[m.saveFocus].Blur()
		m.saveFocus = (m.saveFocus + 1) % len(m.filenames)
		return m.filenames[m.saveFocus].Focus()
	case "shift+tab", "up":
		m.filenames[m.saveFocus].Blur()
		m.saveFocus--
		if m.saveFocus < 0 {
			m.saveFocus = len(m.filenames) - 1
		}
		return m.filenames[m.saveFocus].Focus()
	case "enter":
		names := make([]string, len(m.filenames))
		for i := range m.filenames {
			names[i] = m.filenames[i].Value()
		}
		m.saving = false
		return m.SaveFiles(names)
	}

	var cmd tea.Cmd
	m.filenames[m.saveFocus], cmd = m.filenames[m.saveFocus].Update(msg)
	return cmd
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	events := map[*key.Binding]func([]tea.Cmd){
		&m.keymap.quit: m.KeymapQuit,
		&m.keymap.next: m.KeyMapNext,
		&m.keymap.prev: m.KeyMapPrev,
		&m.keymap.add:  m.KeymapAdd,
	}

	switch msg := msg.(type) {
	case fileOpenedMsg:
		if msg.index >= 0 && msg.index < len(m.inputs) {
			m.inputs[msg.index].SetValue(msg.content)
		}
	case fileOpenErrorMsg:
		m.status = msg.err.Error()
	case filesSavedMsg:
		m.status = fmt.Sprintf("Saved %d file(s)", msg.count)
	case saveErrorMsg:
		m.status = msg.err.Error()
	case tea.KeyPressMsg:
		if m.saving {
			return m, m.updateSave(msg)
		}
		for binding, handler := range events {
			if key.Matches(msg, *binding) {
				handler(cmds)
				break
			}
		}
		if key.Matches(msg, m.keymap.save) {
			return m, m.beginSave()
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
	if m.saving {
		var views []string
		for i := range m.filenames {
			views = append(views, fmt.Sprintf("Editor %d: %s", i+1, m.filenames[i].View()))
		}
		views = append(views, "", "tab/down: next • enter: save • esc: cancel")
		return tea.NewView(strings.Join(views, "\n"))
	}

	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.save,
		m.keymap.quit,
	})

	content := lipgloss.JoinHorizontal(lipgloss.Top, m.inputViews()...) + "\n\n" + help
	if m.status != "" {
		content += "\n" + m.status
	}
	v := tea.NewView(content)
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

func (m EditorModel) SaveFiles(filenames []string) tea.Cmd {
	contents := make([]string, len(m.inputs))

	for i := range m.inputs {
		contents[i] = m.inputs[i].Value()
	}
	return saveFilesCmd(filenames, contents)
}

func openFileCmd(filename string) tea.Cmd {
	return openFileAtCmd(filename, 0)
}

func openFileAtCmd(filename string, index int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(filename)
		if err != nil {
			return fileOpenErrorMsg{
				err: fmt.Errorf("opening %q: %w", filename, err),
			}
		}
		return fileOpenedMsg{
			index:    index,
			filename: filename,
			content:  string(content),
		}
	}
}

func (m model) OpenFile(filename string) (tea.Model, tea.Cmd) {
	m.showList = false
	m.editorOpen = true

	return m, tea.Batch(
		m.editor.Init(),
		openFileCmd(filename),
	)
}
