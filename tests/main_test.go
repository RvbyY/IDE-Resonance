package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestIsEscape(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
		want bool
	}{
		{name: "escape", msg: tea.KeyPressMsg{Code: tea.KeyEscape}, want: true},
		{name: "ctrl open bracket", msg: tea.KeyPressMsg{Code: '[', Mod: tea.ModCtrl}, want: true},
		{name: "q", msg: tea.KeyPressMsg{Code: 'q'}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEscape(tt.msg); got != tt.want {
				t.Fatalf("isEscape() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeFromListReturnsToMain(t *testing.T) {
	m := initialModel()
	m.showList = true

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)
	if got.showList {
		t.Fatal("expected list to close")
	}
	if cmd != nil {
		t.Fatal("expected no quit command while closing list")
	}
}

func TestEscapeFromEditorReturnsToList(t *testing.T) {
	m := initialModel()
	m.editorOpen = true

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	got := next.(model)
	if got.editorOpen || !got.showList {
		t.Fatal("expected editor to close and list to open")
	}
	if cmd != nil {
		t.Fatal("expected no quit command while closing editor")
	}
}

func TestOpenSelectedPresetWithOneEditor(t *testing.T) {
	m := initialModel()
	m.showList = true
	m.editor.inputs = m.editor.inputs[:1]

	next, cmd := m.OpenSelectedPreset()
	got := next.(model)

	if !got.editorOpen || got.showList {
		t.Fatal("expected preset list to close and editor to open")
	}
	if cmd == nil {
		t.Fatal("expected editor initialization command")
	}
	if got.editor.inputs[0].Value() == "" {
		t.Fatal("expected preset content in the editor")
	}
}
