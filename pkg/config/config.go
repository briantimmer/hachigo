package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the Hachigo configuration loaded from _config.yml
type Config struct {
	URL              string   `yaml:"url"`
	Title            string   `yaml:"title"`
	Subtitle         string   `yaml:"subtitle"`
	Author           string   `yaml:"author"`
	SimpleSearch     string   `yaml:"simple_search"`
	Description      string   `yaml:"description"`
	DateFormat       string   `yaml:"date_format"`
	SubscribeRSS     string   `yaml:"subscribe_rss"`
	SubscribeEmail   string   `yaml:"subscribe_email"`
	Email            string   `yaml:"email"`
	Root             string   `yaml:"root"`
	Permalink        string   `yaml:"permalink"`
	Source           string   `yaml:"source"`
	Destination      string   `yaml:"destination"`
	Plugins          string   `yaml:"plugins"`
	CodeDir          string   `yaml:"code_dir"`
	CategoryDir      string   `yaml:"category_dir"`
	Markdown         string   `yaml:"markdown"`
	Paginate         int      `yaml:"paginate"`
	PaginatePath     string   `yaml:"paginate_path"`
	RecentPosts      int      `yaml:"recent_posts"`
	ExcerptLink      string   `yaml:"excerpt_link"`
	ExcerptSeparator string   `yaml:"excerpt_separator"`
	Titlecase        bool     `yaml:"titlecase"`
	DefaultAsides    []string `yaml:"default_asides"`
	CopyrightYear    int      `yaml:"copyright_year"`
	InstagramUser    string   `yaml:"instagram_user"`
	MediumUser       string   `yaml:"medium_user"`
	GoodreadsUser    string   `yaml:"goodreads_user"`
	GithubUser       string   `yaml:"github_user"`
	GithubRepoCount  int      `yaml:"github_repo_count"`
	GithubShowProfileLink bool `yaml:"github_show_profile_link"`
	GithubSkipForks  bool     `yaml:"github_skip_forks"`
}

// Load reads and parses the configuration file at the given path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Set default values matching Octopress defaults before parsing
	cfg := &Config{
		Root:             "/",
		Source:           "source",
		Destination:      "public",
		Paginate:         10,
		PaginatePath:     "posts/:num",
		RecentPosts:      5,
		ExcerptLink:      "Read on &rarr;",
		ExcerptSeparator: "<!--more-->",
		Titlecase:        true,
		GithubShowProfileLink: true,
		GithubSkipForks:  true,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ToMap converts the config to a map suitable for Liquid template context
func (c *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"url":               c.URL,
		"title":             c.Title,
		"subtitle":          c.Subtitle,
		"author":            c.Author,
		"simple_search":     c.SimpleSearch,
		"description":       c.Description,
		"date_format":       c.DateFormat,
		"subscribe_rss":     c.SubscribeRSS,
		"subscribe_email":   c.SubscribeEmail,
		"email":             c.Email,
		"root":              c.Root,
		"permalink":         c.Permalink,
		"source":            c.Source,
		"destination":       c.Destination,
		"plugins":           c.Plugins,
		"code_dir":          c.CodeDir,
		"category_dir":      c.CategoryDir,
		"markdown":          c.Markdown,
		"paginate":          c.Paginate,
		"paginate_path":     c.PaginatePath,
		"recent_posts":      c.RecentPosts,
		"excerpt_link":      c.ExcerptLink,
		"excerpt_separator": c.ExcerptSeparator,
		"titlecase":         c.Titlecase,
		"default_asides":    c.DefaultAsides,
		"copyright_year":    c.CopyrightYear,
		"instagram_user":    c.InstagramUser,
		"medium_user":       c.MediumUser,
		"goodreads_user":    c.GoodreadsUser,
		"github_user":       c.GithubUser,
		"github_repo_count":  c.GithubRepoCount,
		"github_show_profile_link": c.GithubShowProfileLink,
		"github_skip_forks":  c.GithubSkipForks,
	}
}
