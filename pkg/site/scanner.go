package site

import (
	"os"
	"path/filepath"
	"strings"
)

// SiteFiles contains categorized lists of relative file paths in the source directory
type SiteFiles struct {
	Layouts     []string
	Includes    []string
	Posts       []string
	Pages       []string
	StaticFiles []string
}

// Scan walks the source directory and categorizes all files
func Scan(sourceDir string) (*SiteFiles, error) {
	files := &SiteFiles{
		Layouts:     []string{},
		Includes:    []string{},
		Posts:       []string{},
		Pages:       []string{},
		StaticFiles: []string{},
	}

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories from being added directly as files
		if info.IsDir() {
			return nil
		}

		// Get path relative to the source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Skip OS-specific junk files (e.g. .DS_Store)
		baseName := filepath.Base(relPath)
		if strings.HasPrefix(baseName, ".") {
			return nil
		}

		// Determine categorization based on path segments
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) > 0 {
			switch parts[0] {
			case "_layouts":
				files.Layouts = append(files.Layouts, relPath)
			case "_includes":
				files.Includes = append(files.Includes, relPath)
			case "_posts":
				if isMarkdownFile(relPath) {
					files.Posts = append(files.Posts, relPath)
				} else {
					files.StaticFiles = append(files.StaticFiles, relPath)
				}
			default:
				// Inside other directories, check if it's a renderable page
				if isRenderablePage(relPath) || hasFrontmatter(path) {
					files.Pages = append(files.Pages, relPath)
				} else {
					files.StaticFiles = append(files.StaticFiles, relPath)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func isRenderablePage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown" || ext == ".html"
}

func hasFrontmatter(absPath string) bool {
	file, err := os.Open(absPath)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, 3)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	return n >= 3 && string(buf[:3]) == "---"
}
