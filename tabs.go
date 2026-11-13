package main

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type stylesTab struct {
	inactive lipgloss.Style
	active   lipgloss.Style
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func newStylesTab(bgIsDark bool) stylesTab {
	lightDark := lipgloss.LightDark(bgIsDark)
	color := lightDark(lipgloss.Color("#FAD962"), lipgloss.Color("#7D56F4"))
	inactiveBorder := tabBorderWithBottom("┴", "─", "┴")
	activeBorder := tabBorderWithBottom("┘", " ", "└")

	inactive := lipgloss.NewStyle().
		Border(inactiveBorder, true).
		BorderForeground(color).
		Padding(0, 1)

	return stylesTab{
		inactive: inactive,
		active:   inactive.Border(activeBorder, true),
	}
}

func (m EditorModel) editorTabTitle(index int) string {
	if index < len(m.storefilenames) && m.storefilenames[index] != "" {
		return m.storefilenames[index]
	}
	return fmt.Sprintf("untitled-%d", index+1)
}

func (m EditorModel) editorTabs() string {
	styles := newStylesTab(false)
	rendered := make([]string, 0, len(m.inputs))

	for i := range m.inputs {
		style := styles.inactive
		if i == m.editorIndex {
			style = styles.active
		}
		rendered = append(rendered, style.Render(m.editorTabTitle(i)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
