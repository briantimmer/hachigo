package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briantimmer/hachigo/pkg/config"
	"github.com/jlaffaye/ftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPDeployer deploys the built site using SFTP or FTP
type SFTPDeployer struct{}

// Deploy routes the connection to either deploySFTP or deployFTP depending on the deploy type
func (d *SFTPDeployer) Deploy(cfg *config.Config) error {
	if cfg.Deploy.Type == "sftp" {
		return d.deploySFTP(cfg)
	}
	return d.deployFTP(cfg)
}

func (d *SFTPDeployer) deploySFTP(cfg *config.Config) error {
	host := cfg.Deploy.Host
	if host == "" {
		return fmt.Errorf("SFTP deployment requires 'host' configured in deploy settings")
	}

	port := cfg.Deploy.Port
	if port == 0 {
		port = 22
	}

	user := cfg.Deploy.User
	if user == "" {
		return fmt.Errorf("SFTP deployment requires 'user' configured in deploy settings")
	}

	// 1. Gather SSH Auth Methods
	var authMethods []ssh.AuthMethod

	// Try SSH Key First
	keyPath := cfg.Deploy.KeyPath
	if keyPath == "" {
		// Attempt standard SSH Key locations
		homeDir, err := os.UserHomeDir()
		if err == nil {
			for _, k := range []string{"id_rsa", "id_ed25519", "id_dsa"} {
				p := filepath.Join(homeDir, ".ssh", k)
				if _, err := os.Stat(p); err == nil {
					keyPath = p
					break
				}
			}
		}
	}

	if keyPath != "" {
		fmt.Printf("Using SSH key: %s\n", keyPath)
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key %s: %v", keyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// Try Password from Environment Variable
	password := os.Getenv("HACHIGO_DEPLOY_PASSWORD")
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("SFTP deployment requires either a local SSH key or HACHIGO_DEPLOY_PASSWORD environment variable set")
	}

	// 2. Dial SSH Server
	sshConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Connecting to %s via SSH...\n", addr)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %v", err)
	}
	defer sshClient.Close()

	// 3. Initialize SFTP Client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("failed to start SFTP session: %v", err)
	}
	defer sftpClient.Close()

	destDir := cfg.Destination
	if destDir == "" {
		destDir = "public"
	}

	targetDir := cfg.Deploy.TargetDir
	if targetDir == "" {
		targetDir = "."
	}

	// 4. Walk destination and upload files
	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(destDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		remotePath := filepath.Join(targetDir, relPath)
		remotePath = filepath.ToSlash(remotePath)

		if info.IsDir() {
			return sftpClient.MkdirAll(remotePath)
		}

		fmt.Printf("Uploading %s to sftp://%s/%s...\n", relPath, host, remotePath)

		localFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer localFile.Close()

		parent := filepath.Dir(remotePath)
		if parent != "." && parent != "/" {
			if err := sftpClient.MkdirAll(parent); err != nil {
				return err
			}
		}

		remoteFile, err := sftpClient.Create(remotePath)
		if err != nil {
			return err
		}
		defer remoteFile.Close()

		_, err = io.Copy(remoteFile, localFile)
		return err
	})

	if err != nil {
		return err
	}

	fmt.Println("SFTP deployment successful!")
	return nil
}

func (d *SFTPDeployer) deployFTP(cfg *config.Config) error {
	host := cfg.Deploy.Host
	if host == "" {
		return fmt.Errorf("FTP deployment requires 'host' configured in deploy settings")
	}

	port := cfg.Deploy.Port
	if port == 0 {
		port = 21
	}

	user := cfg.Deploy.User
	if user == "" {
		return fmt.Errorf("FTP deployment requires 'user' configured in deploy settings")
	}

	password := os.Getenv("HACHIGO_DEPLOY_PASSWORD")
	if password == "" {
		return fmt.Errorf("FTP deployment requires HACHIGO_DEPLOY_PASSWORD environment variable set")
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Connecting to %s via FTP...\n", addr)
	ftpClient, err := ftp.Dial(addr, ftp.DialWithTimeout(15*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect via FTP: %v", err)
	}
	defer ftpClient.Logout()

	if err := ftpClient.Login(user, password); err != nil {
		return fmt.Errorf("failed to login to FTP: %v", err)
	}

	destDir := cfg.Destination
	if destDir == "" {
		destDir = "public"
	}

	targetDir := cfg.Deploy.TargetDir
	if targetDir == "" {
		targetDir = "."
	}

	// Helper to make directory recursively on FTP
	mkdirAllFTP := func(path string) error {
		path = filepath.ToSlash(path)
		parts := strings.Split(path, "/")
		current := ""
		for _, part := range parts {
			if part == "" {
				continue
			}
			if current == "" {
				current = part
			} else {
				current = current + "/" + part
			}
			// ignore folder exists errors
			_ = ftpClient.MakeDir(current)
		}
		return nil
	}

	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(destDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		remotePath := filepath.Join(targetDir, relPath)
		remotePath = filepath.ToSlash(remotePath)

		if info.IsDir() {
			return mkdirAllFTP(remotePath)
		}

		fmt.Printf("Uploading %s to ftp://%s/%s...\n", relPath, host, remotePath)

		localFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer localFile.Close()

		parent := filepath.Dir(remotePath)
		if parent != "." && parent != "/" {
			if err := mkdirAllFTP(parent); err != nil {
				return err
			}
		}

		return ftpClient.Stor(remotePath, localFile)
	})

	if err != nil {
		return err
	}

	fmt.Println("FTP deployment successful!")
	return nil
}
