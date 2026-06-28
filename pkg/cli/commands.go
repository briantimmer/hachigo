package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/briantimmer/hachigo/pkg/config"
	"github.com/briantimmer/hachigo/pkg/server"
	"github.com/briantimmer/hachigo/pkg/site"

	"github.com/spf13/cobra"
)

var (
	configFile string
	serverPort string
	watchMode  bool
)

// Execute runs the root command
func Execute() {
	rootCmd := &cobra.Command{
		Use:     "hachigo",
		Short:   "Hachigo is a fast static blog generator built on Go",
		Long:    `Hachigo is a reverse-engineered port of Octopress in Go, enabling high-performance compilation of legacy layouts, includes, posts, and assets.`,
		Version: getVersion(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			configFile = resolveConfigFile(configFile)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yml", "Config file path")

	// Commands
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(newCmd())
	rootCmd.AddCommand(initCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func resolveConfigFile(path string) string {
	if path == "config.yml" {
		if _, err := os.Stat("config.yml"); os.IsNotExist(err) {
			if _, err := os.Stat("_config.yml"); err == nil {
				return "_config.yml"
			}
		}
	}
	return path
}

func buildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Generate static site",
		Long:  `Compile all layouts, includes, posts, pages, and copy static assets into the destination directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(configFile)
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Starting build for site: %s\n", cfg.Title)
			if err := site.Build(cfg); err != nil {
				fmt.Printf("Build failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Build completed successfully!")
		},
	}
}

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a local preview server",
		Long:  `Run a local HTTP server and automatically rebuild pages when files change.`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(configFile)
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
			if err := server.Start(cfg, serverPort, watchMode); err != nil {
				fmt.Printf("Preview server failed: %v\n", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVarP(&serverPort, "port", "p", "4000", "Port to run the preview server on")
	cmd.Flags().BoolVarP(&watchMode, "watch", "w", true, "Watch for file changes and regenerate")
	return cmd
}

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create new content",
	}

	cmd.AddCommand(newPostCmd())
	cmd.AddCommand(newPageCmd())

	return cmd
}

func newPostCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "post [title]",
		Short: "Create a new post",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(configFile)
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}

			title := args[0]
			slug := sanitizeSlug(title)
			now := time.Now()
			dateStr := now.Format("2006-01-02")
			filename := fmt.Sprintf("%s-%s.md", dateStr, slug)

			postsDir := filepath.Join(cfg.Source, "_posts")
			if err := os.MkdirAll(postsDir, 0755); err != nil {
				fmt.Printf("Error creating posts directory: %v\n", err)
				os.Exit(1)
			}

			filePath := filepath.Join(postsDir, filename)
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("Error: file %s already exists\n", filePath)
				os.Exit(1)
			}

			// Generate post template with YAML frontmatter
			content := fmt.Sprintf(`---
layout: post
title: "%s"
date: %s
comments: true
categories: 
---

Write post content here.
`, title, now.Format("2006-01-02 15:04:05 -0700"))

			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing post file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created new post: %s\n", filePath)
		},
	}
}

func newPageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "page [path/filename]",
		Short: "Create a new page",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(configFile)
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}

			inputPath := args[0]
			// Clean up input path (e.g., "about/index" -> "about/index.md" or "about" -> "about/index.md")
			var pagePath string
			var pageTitle string

			if strings.HasSuffix(inputPath, ".html") || strings.HasSuffix(inputPath, ".md") || strings.HasSuffix(inputPath, ".markdown") {
				pagePath = filepath.Join(cfg.Source, inputPath)
				base := filepath.Base(inputPath)
				pageTitle = strings.TrimSuffix(base, filepath.Ext(base))
			} else {
				// Directories get an index.md page inside them
				pagePath = filepath.Join(cfg.Source, inputPath, "index.md")
				pageTitle = filepath.Base(inputPath)
			}

			dir := filepath.Dir(pagePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
				os.Exit(1)
			}

			if _, err := os.Stat(pagePath); err == nil {
				fmt.Printf("Error: file %s already exists\n", pagePath)
				os.Exit(1)
			}

			// Generate page template with YAML frontmatter
			content := fmt.Sprintf(`---
layout: page
title: "%s"
date: %s
comments: true
sharing: true
footer: true
---

Write page content here.
`, pageTitle, time.Now().Format("2006-01-02 15:04"))

			if err := os.WriteFile(pagePath, []byte(content), 0644); err != nil {
				fmt.Printf("Error writing page file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created new page: %s\n", pagePath)
		},
	}
}

// sanitizeSlug replaces spaces and non-alphanumeric chars with hyphens
func sanitizeSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	
	// Remove non-alphanumeric, non-hyphen characters
	var b strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	
	res := b.String()
	// Collapse multiple hyphens
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	return strings.Trim(res, "-")
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new static site",
		Long:  `Initialize a new static site structure (config.yml, source folders, default Octopress-based templates, and stylesheets) in the target directory.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			fmt.Printf("Initializing new Hachigo static site in: %s...\n", targetDir)
			if err := site.InitNewSite(targetDir); err != nil {
				fmt.Printf("Error initializing site: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("New Hachigo static site successfully initialized!")
			fmt.Println("\nTo build and preview your new site:")
			if targetDir != "." {
				fmt.Printf("  cd %s\n", targetDir)
			}
			fmt.Println("  hachigo build")
			fmt.Println("  hachigo serve")
		},
	}
}

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "development"
}
