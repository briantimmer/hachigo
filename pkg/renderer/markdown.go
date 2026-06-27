package renderer

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

// markdownEngine is a pre-configured goldmark converter
var markdownEngine = goldmark.New(
	goldmark.WithRendererOptions(
		html.WithUnsafe(), // Enable raw HTML rendering (essential for Octopress posts)
	),
)

// Matches lines starting with hashes without a following space (e.g. ##Why? -> ## Why?)
var headingSpaceRegex = regexp.MustCompile(`(?m)^(#+)([^#\s].*)$`)

// RenderMarkdown converts Markdown content to HTML
func RenderMarkdown(content string) (string, error) {
	// Normalize space-less headers for compatibility with legacy Markdown engines
	content = headingSpaceRegex.ReplaceAllString(content, "$1 $2")

	var buf bytes.Buffer
	if err := markdownEngine.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
