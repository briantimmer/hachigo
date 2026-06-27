package site

import (
	"fmt"
	"github.com/briantimmer/hachigo/pkg/config"
	"github.com/briantimmer/hachigo/pkg/content"
	"github.com/briantimmer/hachigo/pkg/renderer"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Build coordinates the whole site generation process
func Build(cfg *config.Config) error {
	sourceDir := cfg.Source
	destDir := cfg.Destination

	fmt.Printf("Cleaning destination directory: %s\n", destDir)
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("error cleaning destination: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination: %v", err)
	}

	fmt.Println("Scanning source directory...")
	files, err := Scan(sourceDir)
	if err != nil {
		return fmt.Errorf("error scanning directory: %v", err)
	}

	// Instantiate Liquid Renderer with _includes directory path
	includesPath := filepath.Join(sourceDir, "_includes")
	liquidRenderer := renderer.NewLiquidRenderer(includesPath)

	// Load Layouts and build inheritance maps
	layouts := make(map[string]string)
	layoutParents := make(map[string]string)

	for _, layoutPath := range files.Layouts {
		name := filepath.Base(layoutPath)
		name = strings.TrimSuffix(name, filepath.Ext(name))

		absPath := filepath.Join(sourceDir, layoutPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("error reading layout %s: %v", layoutPath, err)
		}

		fm, body, err := content.ParseFrontmatterAndBody(string(data))
		if err != nil {
			return fmt.Errorf("error parsing layout frontmatter %s: %v", layoutPath, err)
		}

		layouts[name] = body
		if parent, ok := fm["layout"].(string); ok && parent != "" && parent != "nil" {
			layoutParents[name] = parent
		}
	}

	// Parse all posts
	var posts []*content.Post
	for _, postPath := range files.Posts {
		post, err := content.ParsePost(sourceDir, postPath, cfg.Permalink)
		if err != nil {
			fmt.Printf("Warning: skipping invalid post %s: %v\n", postPath, err)
			continue
		}
		posts = append(posts, post)
	}

	// Sort posts chronologically descending (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	// Pre-render post markdown bodies to HTML
	for _, p := range posts {
		htmlBody, err := renderer.RenderMarkdown(p.Body)
		if err != nil {
			return fmt.Errorf("error rendering markdown for post %s: %v", p.FilePath, err)
		}
		p.Body = htmlBody
	}

	// Build site-wide bindings
	siteMap := cfg.ToMap()
	siteMap["time"] = time.Now()

	// Build posts list map
	allPostsMaps := make([]map[string]interface{}, len(posts))
	for i, p := range posts {
		allPostsMaps[i] = p.ToMap()
	}
	siteMap["posts"] = allPostsMaps


	// Build categories map
	categoriesMap := make(map[string][]map[string]interface{})
	for _, p := range posts {
		pMap := p.ToMap()
		for _, cat := range p.Categories {
			categoriesMap[cat] = append(categoriesMap[cat], pMap)
		}
	}
	siteMap["categories"] = categoriesMap

	// Build sorted category list for stable sidebar display
	var categoryNames []string
	for cat := range categoriesMap {
		categoryNames = append(categoryNames, cat)
	}
	sort.Strings(categoryNames)

	categoryList := make([]map[string]interface{}, len(categoryNames))
	for i, cat := range categoryNames {
		// Generate url-safe slug (lowercase, spaces to hyphens, alphanumeric only)
		slug := strings.ToLower(cat)
		slug = strings.ReplaceAll(slug, " ", "-")
		var b strings.Builder
		for _, r := range slug {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				b.WriteRune(r)
			}
		}
		cleanSlug := b.String()
		for strings.Contains(cleanSlug, "--") {
			cleanSlug = strings.ReplaceAll(cleanSlug, "--", "-")
		}
		cleanSlug = strings.Trim(cleanSlug, "-")

		categoryList[i] = map[string]interface{}{
			"name": cat,
			"slug": cleanSlug,
			"size": len(categoriesMap[cat]),
		}
	}
	siteMap["category_list"] = categoryList

	// Render Posts and Write to Files
	fmt.Printf("Rendering %d posts...\n", len(posts))
	for i, p := range posts {
		// Post page specific bindings
		pageBindings := p.ToMap()
		pageBindings["content"] = p.Body

		// Add previous / next navigation bindings
		if i < len(posts)-1 {
			pageBindings["previous"] = map[string]interface{}{
				"url":   posts[i+1].URL,
				"title": posts[i+1].Title,
			}
		} else {
			pageBindings["previous"] = nil
		}
		if i > 0 {
			pageBindings["next"] = map[string]interface{}{
				"url":   posts[i-1].URL,
				"title": posts[i-1].Title,
			}
		} else {
			pageBindings["next"] = nil
		}

		bindings := map[string]interface{}{
			"site": siteMap,
			"page": pageBindings,
		}

		// First compile page template (if it contains Liquid elements itself)
		renderedBody, err := liquidRenderer.RenderString(p.Body, bindings)
		if err != nil {
			return fmt.Errorf("error compiling post template %s: %v", p.FilePath, err)
		}

		// Now render the layout chain
		finalHTML, err := renderLayoutChain(liquidRenderer, layouts, layoutParents, p.Layout, renderedBody, bindings)
		if err != nil {
			return fmt.Errorf("error rendering layout chain for post %s: %v", p.FilePath, err)
		}

		// Write to public/blog/:year/:month/:day/:slug/index.html
		destFile := filepath.Join(destDir, p.URL, "index.html")
		if err := writeOutputFile(destFile, finalHTML); err != nil {
			return fmt.Errorf("error writing post HTML: %v", err)
		}
	}

	// Render Paginated Index pages
	fmt.Println("Rendering paginated index pages...")
	var indexTemplate string
	for _, pagePath := range files.Pages {
		if isHomepageTemplate(pagePath) {
			absPath := filepath.Join(sourceDir, pagePath)
			data, err := os.ReadFile(absPath)
			if err != nil {
				return err
			}
			_, indexTemplate, _ = content.ParseFrontmatterAndBody(string(data))
			break
		}
	}

	if indexTemplate != "" {
		paginatorList := PaginatePosts(posts, cfg.Paginate)
		for _, paginator := range paginatorList {
			bindings := map[string]interface{}{
				"site":      siteMap,
				"paginator": paginator.ToMap(),
				"page": map[string]interface{}{
					"title": cfg.Title,
				},
			}

			// Render index page template
			renderedIndex, err := liquidRenderer.RenderString(indexTemplate, bindings)
			if err != nil {
				return fmt.Errorf("error rendering index page template (page %d): %v", paginator.Page, err)
			}

			finalHTML, err := renderLayoutChain(liquidRenderer, layouts, layoutParents, "default", renderedIndex, bindings)
			if err != nil {
				return fmt.Errorf("error rendering layout chain for index (page %d): %v", paginator.Page, err)
			}

			var destFiles []string
			if paginator.Page == 1 {
				destFiles = []string{
					filepath.Join(destDir, "index.html"),
					filepath.Join(destDir, "blog/index.html"),
				}
			} else {
				destFiles = []string{
					filepath.Join(destDir, fmt.Sprintf("posts/%d/index.html", paginator.Page)),
				}
			}

			for _, destFile := range destFiles {
				if err := writeOutputFile(destFile, finalHTML); err != nil {
					return fmt.Errorf("error writing index page %d: %v", paginator.Page, err)
				}
			}
		}
	}

	// Parse and Render other Pages
	fmt.Printf("Rendering pages...\n")
	for _, pagePath := range files.Pages {
		// Skip index page since we handled it in pagination
		if isHomepageTemplate(pagePath) {
			continue
		}

		page, err := content.ParsePage(sourceDir, pagePath)
		if err != nil {
			fmt.Printf("Warning: skipping invalid page %s: %v\n", pagePath, err)
			continue
		}

		// Convert Markdown body to HTML if markdown
		var renderedBody string
		if strings.HasSuffix(pagePath, ".md") || strings.HasSuffix(pagePath, ".markdown") {
			htmlBody, err := renderer.RenderMarkdown(page.Body)
			if err != nil {
				return fmt.Errorf("error rendering markdown for page %s: %v", pagePath, err)
			}
			renderedBody = htmlBody
		} else {
			renderedBody = page.Body
		}

		bindings := map[string]interface{}{
			"site": siteMap,
			"page": page.ToMap(),
		}

		// Pre-render page content in case it has Liquid codes
		compiledBody, err := liquidRenderer.RenderString(renderedBody, bindings)
		if err != nil {
			return fmt.Errorf("error compiling page template %s: %v", pagePath, err)
		}

		finalHTML, err := renderLayoutChain(liquidRenderer, layouts, layoutParents, page.Layout, compiledBody, bindings)
		if err != nil {
			return fmt.Errorf("error rendering layout chain for page %s: %v", pagePath, err)
		}

		var destFile string
		if strings.Contains(filepath.Base(page.URL), ".") {
			destFile = filepath.Join(destDir, page.URL)
		} else {
			destFile = filepath.Join(destDir, page.URL, "index.html")
		}

		if err := writeOutputFile(destFile, finalHTML); err != nil {
			return fmt.Errorf("error writing page %s: %v", pagePath, err)
		}
	}

	// Generate Category Index Pages
	if _, ok := layouts["category_index"]; ok {
		fmt.Println("Generating category pages...")
		categoryDir := cfg.CategoryDir
		if categoryDir == "" {
			categoryDir = "blog/categories"
		}

		for cat := range categoriesMap {
			catSlug := sanitizeSlug(cat)

			bindings := map[string]interface{}{
				"site": siteMap,
				"page": map[string]interface{}{
					"title":       fmt.Sprintf("Category: %s", cat),
					"category":    cat,
					"description": fmt.Sprintf("Category: %s", cat),
				},
				"category": cat,
			}

			// Render the category index layout directly (which behaves as a template)
			// Wait, the category_index.html is a layout file. We can render it with initial content empty.
			categoryIndexTemplate := layouts["category_index"]
			renderedCat, err := liquidRenderer.RenderString(categoryIndexTemplate, bindings)
			if err != nil {
				return fmt.Errorf("error compiling category template for %s: %v", cat, err)
			}

			parentLayout := layoutParents["category_index"]
			if parentLayout == "" {
				parentLayout = "default"
			}

			finalHTML, err := renderLayoutChain(liquidRenderer, layouts, layoutParents, parentLayout, renderedCat, bindings)
			if err != nil {
				return fmt.Errorf("error rendering layout chain for category index %s: %v", cat, err)
			}

			destFile := filepath.Join(destDir, categoryDir, catSlug, "index.html")
			if err := writeOutputFile(destFile, finalHTML); err != nil {
				return fmt.Errorf("error writing category %s: %v", cat, err)
			}
		}
	}

	// Copy Static Assets
	fmt.Println("Copying static assets...")
	if err := CopyAssets(sourceDir, destDir, files.StaticFiles); err != nil {
		return fmt.Errorf("error copying static assets: %v", err)
	}

	fmt.Println("Site generation complete!")
	return nil
}

func renderLayoutChain(renderer *renderer.LiquidRenderer, layouts map[string]string, layoutParents map[string]string, layoutName string, initialContent string, bindings map[string]interface{}) (string, error) {
	if layoutName == "" || layoutName == "nil" {
		return initialContent, nil
	}

	layoutContent, exists := layouts[layoutName]
	if !exists {
		return "", fmt.Errorf("layout %s not found", layoutName)
	}

	// Create a copy of bindings to avoid mutation side-effects
	runBindings := make(map[string]interface{}, len(bindings))
	for k, v := range bindings {
		runBindings[k] = v
	}
	runBindings["content"] = initialContent

	rendered, err := renderer.RenderString(layoutContent, runBindings)
	if err != nil {
		return "", err
	}

	parentLayout := layoutParents[layoutName]
	if parentLayout != "" {
		return renderLayoutChain(renderer, layouts, layoutParents, parentLayout, rendered, runBindings)
	}

	return rendered, nil
}

func writeOutputFile(filePath string, content string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return os.WriteFile(filePath, []byte(content), 0644)
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

func isHomepageTemplate(pagePath string) bool {
	dir := filepath.Dir(pagePath)
	return (dir == "." || dir == "") && strings.HasPrefix(filepath.Base(pagePath), "index.")
}
