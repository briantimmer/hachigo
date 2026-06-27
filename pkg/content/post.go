package content

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	postFilenameRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)\.(md|markdown)$`)
	dateFormats       = []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04 -0700",
		"2006-01-02 15:04",
		"2006-01-02",
	}
)

// Post represents a parsed blog post
type Post struct {
	FilePath     string                 // relative path
	FilenameDate time.Time              // date from YYYY-MM-DD prefix
	Slug         string                 // slug from filename
	Title        string                 // title (frontmatter, falls back to slug)
	Date         time.Time              // final date (frontmatter, falls back to FilenameDate)
	Layout       string                 // layout template name (defaults to "post")
	Comments     bool                   // if comments are enabled
	Categories   []string               // categories/tags list
	Metadata     map[string]interface{} // raw frontmatter fields
	Body         string                 // Markdown body
	URL          string                 // e.g. /blog/2019/10/24/the-man-in-the-arena/
}

// ParsePost reads and parses a post file
func ParsePost(sourceDir, relPath string, permalinkFormat string) (*Post, error) {
	absPath := filepath.Join(sourceDir, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(relPath)
	matches := postFilenameRegex.FindStringSubmatch(filename)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid post filename format: %s", filename)
	}

	filenameDateStr := matches[1]
	slug := matches[2]

	filenameDate, err := time.Parse("2006-01-02", filenameDateStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing filename date %s: %v", filenameDateStr, err)
	}

	fm, body, err := ParseFrontmatterAndBody(string(data))
	if err != nil {
		return nil, fmt.Errorf("error parsing frontmatter in %s: %v", relPath, err)
	}

	// Extract basic fields
	title := slug
	if t, ok := fm["title"].(string); ok {
		title = t
	}

	// Final date resolves: frontmatter date -> filename date
	finalDate := filenameDate
	if dStr, ok := fm["date"]; ok {
		if parsedDate, err := parseDate(dStr); err == nil {
			finalDate = parsedDate
		} else {
			fmt.Printf("Warning: could not parse frontmatter date '%v' in %s: %v. Falling back to filename date.\n", dStr, relPath, err)
		}
	}

	layout := "post"
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

	comments := true
	if c, ok := fm["comments"].(bool); ok {
		comments = c
	}

	categories := parseCategories(fm["categories"])

	post := &Post{
		FilePath:     relPath,
		FilenameDate: filenameDate,
		Slug:         slug,
		Title:        title,
		Date:         finalDate,
		Layout:       layout,
		Comments:     comments,
		Categories:   categories,
		Metadata:     fm,
		Body:         body,
	}

	post.URL = post.generateURL(permalinkFormat)
	return post, nil
}

// ParseFrontmatterAndBody splits the content by the standard --- delimiters
func ParseFrontmatterAndBody(content string) (map[string]interface{}, string, error) {
	fm := make(map[string]interface{})
	
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	
	if !strings.HasPrefix(content, "---\n") {
		return fm, content, nil
	}

	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		return fm, content, nil
	}

	yamlContent := parts[1]
	bodyContent := parts[2]

	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, "", err
	}

	return fm, bodyContent, nil
}

func parseDate(val interface{}) (time.Time, error) {
	// If yaml parser parsed it as time.Time already (yaml.v3 does this for some ISO dates)
	if t, ok := val.(time.Time); ok {
		return t, nil
	}

	strVal := strings.TrimSpace(fmt.Sprintf("%v", val))
	for _, fmtStr := range dateFormats {
		if t, err := time.Parse(fmtStr, strVal); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format: %s", strVal)
}

func parseCategories(val interface{}) []string {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case string:
		// Single category string, possibly space-separated or comma-separated
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		if strings.Contains(v, " ") {
			return strings.Fields(v)
		}
		return []string{v}
	case []interface{}:
		var result []string
		for _, item := range v {
			if str := fmt.Sprintf("%v", item); str != "" {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{fmt.Sprintf("%v", val)}
	}
}

func (p *Post) generateURL(format string) string {
	// Defaults to /blog/:year/:month/:day/:title/
	if format == "" {
		format = "/blog/:year/:month/:day/:title/"
	}

	year := p.Date.Format("2006")
	month := p.Date.Format("01")
	day := p.Date.Format("02")

	url := format
	url = strings.ReplaceAll(url, ":year", year)
	url = strings.ReplaceAll(url, ":month", month)
	url = strings.ReplaceAll(url, ":day", day)
	url = strings.ReplaceAll(url, ":title", p.Slug)

	// Ensure prefix and suffix slashes are correct
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	// Jekyll links typically end in a slash or a filename extension
	// If it doesn't end with a slash or has an extension, append slash
	if !strings.HasSuffix(url, "/") && !strings.Contains(filepath.Base(url), ".") {
		url = url + "/"
	}

	return url
}

// ToMap converts the post to a map representation for templating
func (p *Post) ToMap() map[string]interface{} {
	// Standard Liquid variables for a post
	postMap := map[string]interface{}{
		"title":      p.Title,
		"date":       p.Date,
		"comments":   p.Comments,
		"categories": p.Categories,
		"url":        p.URL,
		"content":    p.Body,
		"layout":     p.Layout,
	}

	// Merge with all original metadata
	for k, v := range p.Metadata {
		if _, exists := postMap[k]; !exists {
			postMap[k] = v
		}
	}

	return postMap
}
