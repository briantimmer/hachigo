package renderer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/render"
)

var (
	// Matches class, src, width, height, title
	imgMarkupRegex = regexp.MustCompile(`(?i)^(?:(\S.*\s+)?)(https?://\S+|/\S+|\S+/)(?:\s+(\d+))?(?:\s+(\d+))?(?:\s+(.+))?$`)
	// Matches "title" "alt"
	titleAltRegex = regexp.MustCompile(`^["']([^"']+)["']\s+["']([^"']+)["']$`)
)

// RegisterImgTag registers the custom {% img %} tag in Liquid
func RegisterImgTag(engine *liquid.Engine) {
	engine.RegisterTag("img", func(c render.Context) (string, error) {
		markup := c.TagArgs()
		markup = strings.TrimSpace(markup)
		if markup == "" {
			return "", fmt.Errorf("img tag requires arguments")
		}

		matches := imgMarkupRegex.FindStringSubmatch(markup)
		if len(matches) == 0 {
			return fmt.Sprintf("<!-- Error parsing img tag arguments: %s -->", markup), nil
		}

		classAttr := strings.TrimSpace(matches[1])
		srcAttr := strings.TrimSpace(matches[2])
		widthAttr := strings.TrimSpace(matches[3])
		heightAttr := strings.TrimSpace(matches[4])
		titleAttr := strings.TrimSpace(matches[5])

		var titleVal, altVal string

		if titleAttr != "" {
			taMatches := titleAltRegex.FindStringSubmatch(titleAttr)
			if len(taMatches) > 2 {
				titleVal = taMatches[1]
				altVal = taMatches[2]
			} else {
				titleVal = strings.Trim(titleAttr, `"'`)
				altVal = titleVal
			}
		}

		classAttr = strings.ReplaceAll(classAttr, `"`, "")

		var sb strings.Builder
		sb.WriteString("<img")
		if classAttr != "" {
			sb.WriteString(fmt.Sprintf(` class="%s"`, classAttr))
		}
		sb.WriteString(fmt.Sprintf(` src="%s"`, srcAttr))
		if widthAttr != "" {
			sb.WriteString(fmt.Sprintf(` width="%s"`, widthAttr))
		}
		if heightAttr != "" {
			sb.WriteString(fmt.Sprintf(` height="%s"`, heightAttr))
		}
		if titleVal != "" {
			sb.WriteString(fmt.Sprintf(` title="%s"`, strings.ReplaceAll(titleVal, `"`, `&quot;`)))
		}
		if altVal != "" {
			sb.WriteString(fmt.Sprintf(` alt="%s"`, strings.ReplaceAll(altVal, `"`, `&quot;`)))
		}
		sb.WriteString(">")

		return sb.String(), nil
	})
}
