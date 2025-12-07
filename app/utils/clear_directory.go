package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func ClearMediaDirectories() error {
	dirs := []string{
		"/app/video",
		"/app/audio",
	}

	for _, d := range dirs {
		if err := ClearDirectory(d); err != nil {
			return err
		}
	}

	return nil
}

func ClearDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		// Remove files or full subdirectories recursively
		err := os.RemoveAll(fullPath)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", fullPath, err)
		}
	}
	return nil
}
