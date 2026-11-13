package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type stylesTab struct {
	doc         lipgloss.Style
	highlight   lipgloss.Style
	inactiveTab lipgloss.Style
	activeTab   lipgloss.Style
	window      lipgloss.Style
}
type modelTab struct {
	Tabs      []string
	styles    *stylesTab
	activeTab int
}

func (m modelTab) InitTab() tea.Cmd {
	return nil
}

func (m model) updateFilesTab(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) updateEditorTab(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		// case "ctrl+c", "q":
		// return m, tea.Quit
		case "right", "l", "n", "tab":
			m.tab.activeTab = min(m.tab.activeTab+1, len(m.tab.Tabs))
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.tab.activeTab = max(m.tab.activeTab-1, 0)
			return m, nil
		}
	}
	return m, nil
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m model) renderTabs() string {
	s := m.tab.styles
	var renderedTabs []string

	for i, t := range m.tab.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tab.Tabs)-1, i == m.tab.activeTab
		if isActive {
			style = s.activeTab
		} else {
			style = s.inactiveTab
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "|"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "|"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	return row
}

func (m model) viewTab() tea.View {
	tabs := m.renderTabs()
	var content string

	switch m.tab.activeTab {
	case 0:
		content, _ = m.editor.ViewFilesTab()
	case 1:
		_, content = m.editor.ViewFilesTab()
	}
	view := tea.NewView(tabs + "\n" + content)
	view.AltScreen = true
	return view
}

func newStylesTab(bgIsDark bool) *stylesTab {
	lightDark := lipgloss.LightDark(bgIsDark)
	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")
	highlightColor := lightDark(lipgloss.Color("#FAD962"), lipgloss.Color("#7D56F4"))
	s := new(stylesTab)

	s.doc = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	s.inactiveTab = lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(highlightColor).
		Padding(0, 1)
	s.activeTab = s.inactiveTab.
		Border(activeTabBorder, true)
	s.window = lipgloss.NewStyle().
		BorderForeground(highlightColor).
		Padding(2, 0).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop()
	return s
}

func initialTabModel() modelTab {
	return modelTab{
		Tabs:      []string{"Files", "untitled"},
		styles:    newStylesTab(false),
		activeTab: 0,
	}
}
