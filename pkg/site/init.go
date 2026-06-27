package site

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:templates/*
var defaultTemplates embed.FS

// InitNewSite initializes a new static site structure in the target directory
func InitNewSite(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// Strip the "templates" prefix from the embedded filesystem paths
	subFS, err := fs.Sub(defaultTemplates, "templates")
	if err != nil {
		return err
	}

	return fs.WalkDir(subFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == "." {
			return nil
		}

		destPath := filepath.Join(targetDir, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := fs.ReadFile(subFS, path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0644)
	})
}
