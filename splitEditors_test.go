package main

import (
	"os"
	"path/filepath"
	"testing"
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
