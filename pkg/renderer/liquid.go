package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
)

// LiquidRenderer wraps Shopify's Liquid template engine
type LiquidRenderer struct {
	engine      *liquid.Engine
	includesDir string
}

// NewLiquidRenderer creates a new LiquidRenderer and registers all Octopress filters and tags
func NewLiquidRenderer(includesDir string) *LiquidRenderer {
	engine := liquid.NewEngine()
	r := &LiquidRenderer{
		engine:      engine,
		includesDir: includesDir,
	}

	r.registerFilters()
	r.registerTags()

	return r
}

var forLoopReverseRegex = regexp.MustCompile(`(?i)\{\%\s*for\s+([^%]+?)\breverse\b([^%]*?)%\}`)

// RenderString parses and renders a Liquid template string
func (r *LiquidRenderer) RenderString(templateStr string, bindings map[string]interface{}) (string, error) {
	// Normalize Jekyll's "reverse" loop modifier to Liquid's standard "reversed"
	templateStr = forLoopReverseRegex.ReplaceAllString(templateStr, `{% for ${1}reversed${2} %}`)

	// Conver bindings to map[string]any for osteele/liquid compatibility
	b := make(map[string]any, len(bindings))
	for k, v := range bindings {
		b[k] = v
	}

	return r.engine.ParseAndRenderString(templateStr, b)
}

func (r *LiquidRenderer) registerFilters() {
	// 1. excerpt: Extract text before <!--more--> or the first paragraph </p>
	r.engine.RegisterFilter("excerpt", func(input string) string {
		// Check for <!--more-->
		reMore := regexp.MustCompile(`(?i)<!--\s*more\s*-->`)
		locMore := reMore.FindStringIndex(input)
		if locMore != nil {
			return input[:locMore[0]]
		}

		// Fall back to first paragraph (up to first </p>)
		rePara := regexp.MustCompile(`(?i)</p>`)
		locPara := rePara.FindStringIndex(input)
		if locPara != nil {
			return input[:locPara[1]]
		}

		return input
	})

	// 2. has_excerpt: Return 'true' if the input has content after the excerpt
	r.engine.RegisterFilter("has_excerpt", func(input string) string {
		// Check for <!--more-->
		reMore := regexp.MustCompile(`(?i)<!--\s*more\s*-->`)
		locMore := reMore.FindStringIndex(input)
		if locMore != nil {
			remaining := strings.TrimSpace(input[locMore[1]:])
			if remaining != "" {
				return "true"
			}
			return "false"
		}

		// Fall back to first paragraph check
		rePara := regexp.MustCompile(`(?i)</p>`)
		locPara := rePara.FindStringIndex(input)
		if locPara != nil {
			remaining := strings.TrimSpace(input[locPara[1]:])
			// Strip common enclosing tags like </div> that might remain
			remaining = strings.ReplaceAll(remaining, "</div>", "")
			remaining = strings.TrimSpace(remaining)
			if remaining != "" {
				return "true"
			}
		}

		return "false"
	})

	// 3. summary: Return first paragraph of text
	r.engine.RegisterFilter("summary", func(input string) string {
		input = strings.ReplaceAll(input, "\r\n", "\n")
		parts := strings.Split(input, "\n\n")
		if len(parts) > 0 {
			return parts[0]
		}
		return input
	})

	// 4. raw_content: Extracts inner body from wrapped HTML entry-content
	r.engine.RegisterFilter("raw_content", func(input string) string {
		re := regexp.MustCompile(`(?s)<div class="entry-content">(.*?)<\/div>\s*<(footer|\/article)>`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			return matches[1]
		}
		return input
	})

	// 5. expand_urls: Replaces relative urls with full urls
	r.engine.RegisterFilter("expand_urls", func(input string, baseURL interface{}) string {
		urlStr := "/"
		if baseURL != nil {
			urlStr = fmt.Sprintf("%v", baseURL)
		}
		urlStr = strings.TrimSuffix(urlStr, "/")

		re := regexp.MustCompile(`(\s+(href|src|poster)\s*=\s*["'])(/[^/>][^"'>]*)`)
		return re.ReplaceAllStringFunc(input, func(m string) string {
			matches := re.FindStringSubmatch(m)
			if len(matches) > 3 {
				return matches[1] + urlStr + matches[3]
			}
			return m
		})
	})

	// 6. condense_spaces: Collapses multiple spaces and tabs into a single space
	r.engine.RegisterFilter("condense_spaces", func(input string) string {
		re := regexp.MustCompile(`\s{2,}`)
		return re.ReplaceAllString(input, " ")
	})

	// 7. strip_slash: Removes trailing slash from a string
	r.engine.RegisterFilter("strip_slash", func(input string) string {
		return strings.TrimSuffix(input, "/")
	})

	// 8. shorthand_url: Returns a url without the protocol
	r.engine.RegisterFilter("shorthand_url", func(input string) string {
		re := regexp.MustCompile(`(?i)^https?://`)
		return re.ReplaceAllString(input, "")
	})

	// 9. titlecase: Capitalize words according to Chicago Manual of Style
	r.engine.RegisterFilter("titlecase", func(input string) string {
		words := strings.Fields(input)
		if len(words) == 0 {
			return input
		}

		smallWords := map[string]bool{
			"a": true, "an": true, "and": true, "as": true, "at": true,
			"but": true, "by": true, "en": true, "for": true, "if": true,
			"in": true, "of": true, "on": true, "or": true, "the": true,
			"to": true, "v": true, "v.": true, "via": true, "vs": true, "vs.": true,
		}

		for i, w := range words {
			// Strip punctuation to check if it's a small word
			cleanWord := strings.ToLower(strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					return r
				}
				return -1
			}, w))

			if i == 0 || i == len(words)-1 || !smallWords[cleanWord] {
				words[i] = capitalizeWord(w)
			} else {
				words[i] = strings.ToLower(w)
			}
		}

		return strings.Join(words, " ")
	})

	// 10. category_link: single category <a> link
	r.engine.RegisterFilter("category_link", func(category interface{}) string {
		catStr := fmt.Sprintf("%v", category)
		slug := sanitizeSlug(catStr)
		return fmt.Sprintf(`<a class='category' href='/blog/categories/%s/'>%s</a>`, slug, catStr)
	})

	// 11. category_links: comma-separated list of category <a> links
	r.engine.RegisterFilter("category_links", func(categories interface{}) string {
		var list []string
		switch v := categories.(type) {
		case []string:
			list = v
		case []interface{}:
			for _, item := range v {
				list = append(list, fmt.Sprintf("%v", item))
			}
		default:
			list = []string{fmt.Sprintf("%v", categories)}
		}
		sort.Strings(list)
		var links []string
		for _, cat := range list {
			catStr := fmt.Sprintf("%v", cat)
			slug := sanitizeSlug(catStr)
			links = append(links, fmt.Sprintf(`<a class='category' href='/blog/categories/%s/'>%s</a>`, slug, catStr))
		}
		return strings.Join(links, ", ")
	})

	// 12. date_to_html_string: format date as HTML with span tags
	r.engine.RegisterFilter("date_to_html_string", func(val interface{}) string {
		var t time.Time
		switch v := val.(type) {
		case time.Time:
			t = v
		case string:
			for _, fmtStr := range []string{
				"2006-01-02 15:04:05 -0700",
				"2006-01-02 15:04:05",
				"2006-01-02 15:04",
				"2006-01-02",
			} {
				if parsed, err := time.Parse(fmtStr, v); err == nil {
					t = parsed
					break
				}
			}
		}

		if t.IsZero() {
			return fmt.Sprintf("%v", val)
		}

		month := strings.ToUpper(t.Format("Jan"))
		day := t.Format("02")
		year := t.Format("2006")

		return fmt.Sprintf(`<span class="month">%s</span> <span class="day">%s</span> <span class="year">%s</span>`, month, day, year)
	})

	// 13. date_to_xmlschema: format date as XML schema / ISO 8601
	r.engine.RegisterFilter("date_to_xmlschema", func(val interface{}) string {
		var t time.Time
		switch v := val.(type) {
		case time.Time:
			t = v
		case string:
			for _, fmtStr := range []string{
				"2006-01-02 15:04:05 -0700",
				"2006-01-02 15:04:05",
				"2006-01-02 15:04",
				"2006-01-02",
			} {
				if parsed, err := time.Parse(fmtStr, v); err == nil {
					t = parsed
					break
				}
			}
		}

		if t.IsZero() {
			return fmt.Sprintf("%v", val)
		}

		return t.Format("2006-01-02T15:04:05-07:00") // standard XML schema format
	})

	// 14. date_to_rfc822: format date as RFC 822
	r.engine.RegisterFilter("date_to_rfc822", func(val interface{}) string {
		var t time.Time
		switch v := val.(type) {
		case time.Time:
			t = v
		case string:
			for _, fmtStr := range []string{
				"2006-01-02 15:04:05 -0700",
				"2006-01-02 15:04:05",
				"2006-01-02 15:04",
				"2006-01-02",
			} {
				if parsed, err := time.Parse(fmtStr, v); err == nil {
					t = parsed
					break
				}
			}
		}

		if t.IsZero() {
			return fmt.Sprintf("%v", val)
		}

		return t.Format(time.RFC1123Z)
	})

	// 15. datetime: format date as XML/HTML5 datetime (RFC 3339)
	r.engine.RegisterFilter("datetime", func(val interface{}) string {
		var t time.Time
		switch v := val.(type) {
		case time.Time:
			t = v
		case string:
			for _, fmtStr := range []string{
				"2006-01-02 15:04:05 -0700",
				"2006-01-02 15:04:05",
				"2006-01-02 15:04",
				"2006-01-02",
			} {
				if parsed, err := time.Parse(fmtStr, v); err == nil {
					t = parsed
					break
				}
			}
		}

		if t.IsZero() {
			return fmt.Sprintf("%v", val)
		}

		return t.Format(time.RFC3339)
	})

	// 16. cdata_escape: escapes CDATA sections in XML
	r.engine.RegisterFilter("cdata_escape", func(input string) string {
		input = strings.ReplaceAll(input, "<![CDATA[", "&lt;![CDATA[")
		input = strings.ReplaceAll(input, "]]>", "]]&gt;")
		return input
	})
}

func (r *LiquidRenderer) registerTags() {
	// Register custom include tag that resolves paths relative to _includes/
	r.engine.RegisterTag("include", func(c render.Context) (string, error) {
		val, err := c.EvaluateString(c.TagArgs())
		var rel string
		if err != nil || val == nil {
			rel = strings.Trim(c.TagArgs(), `"' `)
		} else {
			rel = fmt.Sprintf("%v", val)
		}

		filename := filepath.Join(r.includesDir, rel)
		if _, err := os.Stat(filename); err != nil {
			return "", fmt.Errorf("include file not found: %s", filename)
		}

		return c.RenderFile(filename, map[string]any{})
	})

	// Register custom include_array tag that renders multiple files specified in a config array
	r.engine.RegisterTag("include_array", func(c render.Context) (string, error) {
		arrayName := strings.TrimSpace(c.TagArgs())

		val := c.Get(arrayName)
		if val == nil {
			siteVal := c.Get("site")
			if siteMap, ok := siteVal.(map[string]interface{}); ok {
				val = siteMap[arrayName]
			} else if siteMap, ok := siteVal.(map[string]any); ok {
				val = siteMap[arrayName]
			}
		}

		if val == nil {
			return "", nil
		}

		var files []string
		switch v := val.(type) {
		case []interface{}:
			for _, item := range v {
				files = append(files, fmt.Sprintf("%v", item))
			}
		case []string:
			files = v
		default:
			files = []string{fmt.Sprintf("%v", val)}
		}

		var result strings.Builder
		for _, file := range files {
			filename := filepath.Join(r.includesDir, file)
			if _, err := os.Stat(filename); err != nil {
				fmt.Printf("Warning: include_array file not found: %s\n", filename)
				continue
			}

			content, err := c.RenderFile(filename, map[string]any{})
			if err != nil {
				return "", err
			}
			result.WriteString(content)
		}

		return result.String(), nil
	})

	// Register Octopress img tag
	RegisterImgTag(r.engine)

	// Register Octopress codeblock tag
	RegisterCodeBlockTag(r.engine)
}

func capitalizeWord(w string) string {
	if len(w) == 0 {
		return w
	}
	runes := []rune(w)
	for i, r := range runes {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			runes[i] = []rune(strings.ToUpper(string(r)))[0]
			break
		}
	}
	return string(runes)
}

func sanitizeSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}

	res := b.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	return strings.Trim(res, "-")
}
