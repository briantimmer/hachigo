package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDeployConfig(t *testing.T) {
	configYAML := `
title: "My Blog"
author: "Test Author"
deploy:
  type: "sftp"
  host: "deploy.example.com"
  port: 2222
  user: "deployer"
  key_path: "/home/user/.ssh/custom_key"
  target_dir: "/var/www/blog"
`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.Title != "My Blog" {
		t.Errorf("expected title to be 'My Blog', got '%s'", cfg.Title)
	}

	if cfg.Deploy.Type != "sftp" {
		t.Errorf("expected deploy type to be 'sftp', got '%s'", cfg.Deploy.Type)
	}

	if cfg.Deploy.Host != "deploy.example.com" {
		t.Errorf("expected deploy host to be 'deploy.example.com', got '%s'", cfg.Deploy.Host)
	}

	if cfg.Deploy.Port != 2222 {
		t.Errorf("expected deploy port to be 2222, got %d", cfg.Deploy.Port)
	}

	if cfg.Deploy.User != "deployer" {
		t.Errorf("expected deploy user to be 'deployer', got '%s'", cfg.Deploy.User)
	}

	if cfg.Deploy.KeyPath != "/home/user/.ssh/custom_key" {
		t.Errorf("expected deploy key_path to be '/home/user/.ssh/custom_key', got '%s'", cfg.Deploy.KeyPath)
	}

	if cfg.Deploy.TargetDir != "/var/www/blog" {
		t.Errorf("expected deploy target_dir to be '/var/www/blog', got '%s'", cfg.Deploy.TargetDir)
	}
}
