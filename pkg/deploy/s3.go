package deploy

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/briantimmer/hachigo/pkg/config"
)

// S3Deployer deploys the built site to AWS S3 or S3-compatible object storage
type S3Deployer struct{}

// Deploy walks the public output directory and uploads all files to the configured S3 bucket
func (d *S3Deployer) Deploy(cfg *config.Config) error {
	bucket := cfg.Deploy.Bucket
	if bucket == "" {
		return fmt.Errorf("S3 deployment requires 'bucket' configured in deploy settings")
	}

	region := cfg.Deploy.Region
	if region == "" {
		region = "us-east-1"
	}

	destDir := cfg.Destination
	if destDir == "" {
		destDir = "public"
	}

	ctx := context.TODO()

	// 1. Load AWS configuration (automatically resolves env variables like AWS_ACCESS_KEY_ID)
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// 2. Setup S3 client (supports custom endpoints for Cloudflare R2, MinIO, DO Spaces)
	var s3Client *s3.Client
	if cfg.Deploy.Endpoint != "" {
		s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Deploy.Endpoint)
			o.UsePathStyle = true
		})
	} else {
		s3Client = s3.NewFromConfig(awsCfg)
	}

	// 3. Walk destination directory and upload all files
	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(destDir, path)
		if err != nil {
			return err
		}

		// Construct target S3 Key (subfolder path prefix + relative path)
		s3Key := filepath.Join(cfg.Deploy.Path, relPath)
		s3Key = filepath.ToSlash(s3Key)
		s3Key = strings.TrimPrefix(s3Key, "/")

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Get MIME Content-Type based on extension for proper browser rendering
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		fmt.Printf("Uploading %s to s3://%s/%s (%s)...\n", relPath, bucket, s3Key, contentType)

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(s3Key),
			Body:        file,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s: %v", relPath, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println("S3 deployment successful!")
	return nil
}
