package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/briantimmer/hachigo/pkg/config"
)

// GithubDeployer deploys the built site to GitHub Pages using git force-push
type GithubDeployer struct{}

// Deploy runs the Git commands to deploy the public folder to the configured repo and branch
func (d *GithubDeployer) Deploy(cfg *config.Config) error {
	repo := cfg.Deploy.Repo
	if repo == "" {
		return fmt.Errorf("GitHub Pages deployment requires 'repo' configured in deploy settings")
	}

	branch := cfg.Deploy.Branch
	if branch == "" {
		branch = "gh-pages"
	}

	destDir := cfg.Destination
	if destDir == "" {
		destDir = "public"
	}

	// Verify destination directory exists and contains files
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path of destination directory: %v", err)
	}

	if _, err := os.Stat(absDestDir); os.IsNotExist(err) {
		return fmt.Errorf("destination directory %s does not exist; run build first", destDir)
	}

	// Verify git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git binary not found in PATH; git is required for GitHub Pages deployment")
	}

	fmt.Printf("Deploying content of '%s' to %s [%s]...\n", destDir, repo, branch)

	// Clean up any existing temporary git repo inside destination to avoid conflicts
	gitDir := filepath.Join(absDestDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("failed to clean old temporary git directory: %v", err)
	}

	// Helper to execute git command in the destination directory
	runGit := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = absDestDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Sequence of Git operations
	if err := runGit("init"); err != nil {
		return fmt.Errorf("git init failed: %v", err)
	}

	// Clean up temporary .git folder on return to keep destination directory clean
	defer func() {
		os.RemoveAll(gitDir)
	}()

	if err := runGit("add", "-A"); err != nil {
		return fmt.Errorf("git add failed: %v", err)
	}

	commitMsg := fmt.Sprintf("Site updated: %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := runGit("commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit failed: %v", err)
	}

	// Push current local HEAD to target remote repository branch
	if err := runGit("push", "--force", repo, fmt.Sprintf("HEAD:%s", branch)); err != nil {
		return fmt.Errorf("git push failed: %v", err)
	}

	fmt.Println("GitHub Pages deployment successful!")
	return nil
}
