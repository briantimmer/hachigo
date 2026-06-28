# Hachigo (八号) 🚀

Hachigo is a high-performance static blog generator built in Go, designed to modernise and replace legacy Ruby-based Jekyll and Octopress setups.

It enables lightning-fast compilation of classic layouts, includes, date-slug posts, pages, and assets, bringing modern Go execution speeds to retro blogging setups.

## Key Features

- **Blazing Fast Compilation**: Generates 100+ posts, categories, layouts, and paginations in milliseconds.
- **Octopress & Jekyll Compatible**: Custom Liquid tag filters (e.g. `titlecase`, `excerpt`, `expand_urls`, and date formats) are built directly into the engine.
- **Chromacized Codeblocks**: High-performance, syntax-highlighted code blocks powered by [Chroma](https://github.com/alecthomas/chroma) instead of slow Ruby-based Pygments.
- **Vanilla JavaScript Modernisation**: Modernised Javascript template output removing bloat and external dependencies like jQuery.
- **Standardised on `.md`**: Standardises on the modern `.md` extension for all new content generation, while retaining full backward compatibility to parse older `.markdown` files.
- **HTTP Preview Server & File Watcher**: A local HTTP server featuring a debounced filesystem watcher (`fsnotify`) to automatically compile and serve pages instantly as you save.

---

## Installation

The easiest way to install and update Hachigo globally is using the Go toolchain:

```bash
go install github.com/briantimmer/hachigo/cmd/hachigo@latest
```

### Keeping it updated
To get the latest bugfixes and features, simply re-run:
```bash
go install github.com/briantimmer/hachigo/cmd/hachigo@latest
```

### Shell Path Setup
If the `hachigo` command is not recognized after running the installation, make sure Go's binary directory is added to your shell's `$PATH` variable:

```bash
# Add Go binary path to Zsh profile (macOS default):
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc && source ~/.zshrc
```

---

## Command Reference

Hachigo provides a simple, modern CLI based on Cobra.

### 1. Initialize a New Site
Initializes a new directory with the default Hachigo site structure, templates, stylesheets, and initial placeholder content (an about page and a welcome hello-world post).

```bash
hachigo init [directory]
```
If directory is omitted, it will initialize in the current directory.

### 2. Build the Site
Generates layouts, posts, pages, categories, and copies static assets to the target output directory (defaults to `public/`).

```bash
hachigo build [flags]
```
**Flags:**
- `-c, --config string`: Path to the configuration file (default: `config.yml` with fallback to `_config.yml`).

### 3. Local Preview Server
Launch a local development server to serve the generated site. If `-w` or `--watch` is specified, it will monitor your files and rebuild automatically on change.

```bash
hachigo serve [flags]
```
**Flags:**
- `-p, --port string`: Port to serve the site on (default: `"4000"`).
- `-w, --watch`: Enable file watcher and hot rebuilds on change (default: `true`).

### 4. Generate a New Post
Creates a new post file in the `source/_posts/` directory with standardized `.md` extension and default YAML frontmatter.

```bash
hachigo new post "My Awesome Go Journey"
```
*Outputs: `source/_posts/2026-06-27-my-awesome-go-journey.md`*

### 5. Generate a New Page
Creates a new subfolder and an index page with default YAML frontmatter.

```bash
hachigo new page "about"
```
*Outputs: `source/about/index.md`*

### 6. Deploy the Site
Automatically compiles the static site, then deploys it to your S3 bucket, GitHub Pages, or (S)FTP server based on the configuration in your `config.yml`.

```bash
hachigo deploy [flags]
```
**Flags:**
- `-c, --config string`: Path to the configuration file (default: `config.yml`).

---

## Configuration

Hachigo looks for `config.yml` in the root directory by default. If not found, it seamlessly falls back to Jekyll's default `_config.yml`.

Example config structure:
```yaml
title: "Brian's Notes"
subtitle: "Thoughts on Stuff and Junk"
author: "Brian Timmer"
url: "http://briantimmer.com"
source: source
destination: public
paginate: 10
recent_posts: 5

# Deployment Configuration (uncomment the one you need)
#
# AWS S3 (or S3-compatible, e.g. R2, Spaces):
# deploy:
#   type: s3
#   bucket: my-blog-bucket
#   region: us-east-1
#   endpoint: https://xxx.r2.cloudflarestorage.com  # optional for non-AWS
#   path: /blog                                     # optional subfolder
#
# GitHub Pages:
# deploy:
#   type: github
#   repo: git@github.com:briantimmer/briantimmer.github.io.git
#   branch: gh-pages
#
# SFTP / FTP:
# deploy:
#   type: sftp # or ftp
#   host: ftp.example.com
#   port: 22   # 21 for FTP
#   user: username
#   target_dir: /var/www/html
#   key_path: ~/.ssh/id_rsa  # optional (defaults to standard SSH key locations)
```

> [!TIP]
> To keep credentials safe, Hachigo loads SFTP/FTP passwords from the `HACHIGO_DEPLOY_PASSWORD` environment variable, and AWS/S3 credentials from standard AWS env vars (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`). Do not hardcode passwords or access keys in your `config.yml`.
