package deploy

import (
	"fmt"

	"github.com/briantimmer/hachigo/pkg/config"
)

// Deployer defines the interface that all deployment backends must implement
type Deployer interface {
	Deploy(cfg *config.Config) error
}

// Run executes the deployment based on the config settings
func Run(cfg *config.Config) error {
	if cfg.Deploy.Type == "" {
		return fmt.Errorf("no deployment type configured in config.yml (options: s3, github, sftp, ftp)")
	}

	var deployer Deployer

	switch cfg.Deploy.Type {
	case "github":
		deployer = &GithubDeployer{}
	case "s3":
		deployer = &S3Deployer{}
	case "sftp", "ftp":
		deployer = &SFTPDeployer{}
	default:
		return fmt.Errorf("unknown deployment type: %s (supported: s3, github, sftp, ftp)", cfg.Deploy.Type)
	}

	fmt.Printf("Starting deployment using engine: %s...\n", cfg.Deploy.Type)
	return deployer.Deploy(cfg)
}
