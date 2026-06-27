package renderer

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	htmlformatter "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
)

var (
	langRegex       = regexp.MustCompile(`(?i)\s*lang:(\S+)`)
	captionUrlRegex = regexp.MustCompile(`(?i)^(\S[\S\s]*)\s+(https?://\S+|/\S+)\s*(.+)?$`)
	fileExtRegex    = regexp.MustCompile(`\S[\S\s]*\w+\.(\w+)`)
)

// RegisterCodeBlockTag registers the custom {% codeblock %} block tag in Liquid
func RegisterCodeBlockTag(engine *liquid.Engine) {
	engine.RegisterBlock("codeblock", func(c render.Context) (string, error) {
		markup := c.TagArgs()
		markup = strings.TrimSpace(markup)

		// 1. Extract lang if specified (e.g. lang:ruby)
		var filetype string
		if langMatches := langRegex.FindStringSubmatch(markup); len(langMatches) > 1 {
			filetype = langMatches[1]
			markup = langRegex.ReplaceAllString(markup, "")
			markup = strings.TrimSpace(markup)
		}

		// 2. Parse Caption / URL / Link Text
		var caption string
		var file string

		if markup != "" {
			if matches := captionUrlRegex.FindStringSubmatch(markup); len(matches) > 0 {
				file = matches[1]
				link := matches[2]
				linkText := "link"
				if len(matches) > 3 && matches[3] != "" {
					linkText = strings.TrimSpace(matches[3])
				}
				caption = fmt.Sprintf("<figcaption><span>%s</span><a href='%s'>%s</a></figcaption>\n", file, link, linkText)
			} else {
				// Just a plain caption
				file = markup
				caption = fmt.Sprintf("<figcaption><span>%s</span></figcaption>\n", file)
			}
		}

		// 3. Fallback filetype from file extension if not set
		if filetype == "" && file != "" {
			if extMatches := fileExtRegex.FindStringSubmatch(file); len(extMatches) > 1 {
				filetype = extMatches[1]
			}
		}

		// Normalize filetypes
		filetype = normalizeLanguage(filetype)

		// Get the code content inside the block
		code, err := c.InnerString()
		if err != nil {
			return "", err
		}

		// Remove leading/trailing newlines
		code = strings.Trim(code, "\n\r")

		// 4. Render code block
		var codeHTML string
		if filetype != "" {
			var err error
			codeHTML, err = tableizeCode(code, filetype)
			if err != nil {
				// Fallback to unhighlighted table
				codeHTML = tableizeUnhighlighted(code, filetype)
			}
		} else {
			codeHTML = tableizeUnhighlighted(code, "")
		}

		source := "<figure class='code'>\n"
		if caption != "" {
			source += caption
		}
		source += codeHTML
		source += "</figure>"

		return source, nil
	})
}

func normalizeLanguage(lang string) string {
	lang = strings.ToLower(lang)
	switch lang {
	case "ru":
		return "ruby"
	case "m":
		return "objc"
	case "pl":
		return "perl"
	case "yml":
		return "yaml"
	case "js":
		return "javascript"
	default:
		return lang
	}
}

func tableizeCode(code, lang string) (string, error) {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("pygments")
	if style == nil {
		style = styles.Fallback
	}

	formatter := htmlformatter.New(
		htmlformatter.WithClasses(true),
		htmlformatter.PreventSurroundingPre(true),
	)

	// Split by newline
	code = strings.ReplaceAll(code, "\r\n", "\n")
	lines := strings.Split(code, "\n")

	var gutterBuilder strings.Builder
	var codeBuilder strings.Builder

	for i, line := range lines {
		gutterBuilder.WriteString(fmt.Sprintf("<span class='line-number'>%d</span>\n", i+1))

		iterator, err := lexer.Tokenise(nil, line)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := formatter.Format(&buf, style, iterator); err != nil {
			return "", err
		}

		codeBuilder.WriteString(fmt.Sprintf("<span class='line'>%s</span>\n", buf.String()))
	}

	table := fmt.Sprintf(`<div class="highlight"><table><tr><td class="gutter"><pre class="line-numbers">%s</pre></td><td class='code'><pre><code class='%s'>%s</code></pre></td></tr></table></div>`,
		gutterBuilder.String(), lang, codeBuilder.String())

	return table, nil
}

func tableizeUnhighlighted(code, lang string) string {
	code = strings.ReplaceAll(code, "\r\n", "\n")
	lines := strings.Split(code, "\n")

	var gutterBuilder strings.Builder
	var codeBuilder strings.Builder

	for i, line := range lines {
		gutterBuilder.WriteString(fmt.Sprintf("<span class='line-number'>%d</span>\n", i+1))
		escapedLine := html.EscapeString(line)
		codeBuilder.WriteString(fmt.Sprintf("<span class='line'>%s</span>\n", escapedLine))
	}

	table := fmt.Sprintf(`<div class="highlight"><table><tr><td class="gutter"><pre class="line-numbers">%s</pre></td><td class='code'><pre><code class='%s'>%s</code></pre></td></tr></table></div>`,
		gutterBuilder.String(), lang, codeBuilder.String())

	return table
}
