package main

import (
	"os"
)

func listFiles(dir string) ([]string, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    files := make([]string, 0, len(entries))
    for _, entry := range entries {
        if !entry.IsDir() {
            files = append(files, entry.Name())
        }
    }

    return files, nil
}
