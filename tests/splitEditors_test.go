package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSelectFileLoadsContentWhileFilesHaveFocus(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "example.txt")
	if err := os.WriteFile(filename, []byte("file content"), 0644); err != nil {
		t.Fatal(err)
	}

	m := newModel()
	m.files = []string{filename}
	m.fileIndex = 0
	m.focus = focusFiles

	next, cmd := m.selectFile()
	if cmd == nil {
		t.Fatal("expected selecting a file to return an open command")
	}

	opened := cmd()
	updated, _ := next.(EditorModel).Update(opened)
	editor := updated.(EditorModel)
	if editor.focus != focusEditor {
		t.Fatal("expected selecting a file to focus the editor")
	}
	if got := editor.inputs[editor.editorIndex].Value(); got != "file content" {
		t.Fatalf("loaded content = %q, want %q", got, "file content")
	}
}

func TestEditorUpdateReportsFileOpenErrorsWhileFilesHaveFocus(t *testing.T) {
	m := newModel()
	m.focus = focusFiles

	updated, _ := m.Update(fileOpenErrorMsg{err: os.ErrNotExist})
	editor := updated.(EditorModel)
	if editor.status != os.ErrNotExist.Error() {
		t.Fatalf("status = %q, want %q", editor.status, os.ErrNotExist.Error())
	}
}

func TestFilesKeyMovesFromEditorToFiles(t *testing.T) {
	m := newModel()
	m.focus = focusEditor
	m.inputs[m.editorIndex].Focus()

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
	editor := updated.(EditorModel)
	if editor.focus != focusFiles {
		t.Fatal("expected ctrl+o to focus the files pane")
	}
	if cmd != nil {
		t.Fatal("expected moving focus to the files pane to return no command")
	}
}

func TestEditorKeymapsHandleEditorCommands(t *testing.T) {
	tests := []struct {
		name       string
		msg        tea.KeyPressMsg
		wantInputs int
	}{
		{
			name:       "add",
			msg:        tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl},
			wantInputs: initialInputs + 1,
		},
		{
			name:       "remove",
			msg:        tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl},
			wantInputs: initialInputs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel()
			m.focus = focusEditor

			updated, _ := m.Update(tt.msg)
			editor := updated.(EditorModel)
			if len(editor.inputs) != tt.wantInputs {
				t.Fatalf("input count = %d, want %d", len(editor.inputs), tt.wantInputs)
			}
		})
	}
}
