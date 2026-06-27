package site

import (
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a single file from src to dst, creating any parent directories
func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	info, err := srcFile.Stat()
	if err == nil {
		return os.Chmod(dst, info.Mode())
	}
	return nil
}

// CopyAssets copies all listed relative static files from sourceDir to destDir
func CopyAssets(sourceDir, destDir string, staticFiles []string) error {
	for _, relPath := range staticFiles {
		src := filepath.Join(sourceDir, relPath)
		dst := filepath.Join(destDir, relPath)
		if err := CopyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}
