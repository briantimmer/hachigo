package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Page represents a parsed static page (e.g., about page)
type Page struct {
	FilePath string                 // relative path (e.g., about/index.markdown)
	Title    string                 // title (frontmatter, falls back to filename)
	Date     time.Time              // date (frontmatter, falls back to file modtime)
	Layout   string                 // layout template name (defaults to "page")
	Metadata map[string]interface{} // raw frontmatter fields
	Body     string                 // Markdown or HTML body
	URL      string                 // generated URL (e.g. /about/)
}

// ParsePage reads and parses a page file
func ParsePage(sourceDir, relPath string) (*Page, error) {
	absPath := filepath.Join(sourceDir, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	fm, body, err := ParseFrontmatterAndBody(string(data))
	if err != nil {
		return nil, fmt.Errorf("error parsing frontmatter in %s: %v", relPath, err)
	}

	// Determine Title
	title := filepath.Base(relPath)
	title = strings.TrimSuffix(title, filepath.Ext(title))
	if t, ok := fm["title"].(string); ok {
		title = t
	}

	// Determine Date
	finalDate := info.ModTime()
	if dStr, ok := fm["date"]; ok {
		if parsedDate, err := parseDate(dStr); err == nil {
			finalDate = parsedDate
		}
	}

	layout := "page"
	if lVal, ok := fm["layout"]; ok {
		if lVal == nil {
			layout = ""
		} else if lStr, ok := lVal.(string); ok {
			if lStr == "null" || lStr == "nil" || lStr == "none" || lStr == "false" || lStr == "" {
				layout = ""
			} else {
				layout = lStr
			}
		} else if lBool, ok := lVal.(bool); ok && !lBool {
			layout = ""
		}
	}

	page := &Page{
		FilePath: relPath,
		Title:    title,
		Date:     finalDate,
		Layout:   layout,
		Metadata: fm,
		Body:     body,
	}

	page.URL = page.generateURL()
	return page, nil
}

func (p *Page) generateURL() string {
	urlPath := filepath.ToSlash(p.FilePath)
	ext := filepath.Ext(urlPath)

	if ext == ".md" || ext == ".markdown" {
		urlPath = strings.TrimSuffix(urlPath, ext)
		base := filepath.Base(urlPath)
		if base == "index" {
			urlPath = filepath.Dir(urlPath)
			if urlPath == "." {
				urlPath = ""
			}
		} else {
			urlPath = urlPath + ".html"
		}
	} else {
		// Non-markdown templates (e.g. index.html or atom.xml)
		base := filepath.Base(urlPath)
		if base == "index.html" {
			urlPath = filepath.Dir(urlPath)
			if urlPath == "." {
				urlPath = ""
			}
		}
	}

	// Format URL prefix and suffix
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	if strings.HasSuffix(urlPath, "/.") {
		urlPath = strings.TrimSuffix(urlPath, ".")
	}
	if !strings.HasSuffix(urlPath, "/") && !strings.Contains(filepath.Base(urlPath), ".") {
		urlPath = urlPath + "/"
	}

	return urlPath
}

// ToMap converts the page to a map representation for templating
func (p *Page) ToMap() map[string]interface{} {
	pageMap := map[string]interface{}{
		"title":   p.Title,
		"date":    p.Date,
		"url":     p.URL,
		"content": "", // Populated after markdown parsing
		"layout":  p.Layout,
	}

	// Merge raw metadata
	for k, v := range p.Metadata {
		if _, exists := pageMap[k]; !exists {
			pageMap[k] = v
		}
	}

	return pageMap
}
