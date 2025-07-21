package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
)

func UploadDirectory(localPath string, remoteParentDir string, client *sftp.Client) error {
	logger.Info(fmt.Sprintf(
		"Starting upload of local directory '%s' to '%s'...",
		localPath,
		remoteParentDir,
	))

	remoteParentDir = path.Join(remoteParentDir, path.Base(localPath))

	err := filepath.WalkDir(
		localPath,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if ShouldSkipUpload(d) {
				logger.Debug("Skipping", "dir", d.Name())
				return filepath.SkipDir
			}

			// Determine relative path to maintain directory structure
			relPath, err := filepath.Rel(localPath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", path, err)
			}

			// Convert to Unix-style path for SFTP
			remotePath := filepath.ToSlash(filepath.Join(remoteParentDir, relPath))

			if d.IsDir() {
				// Create directory on remote if it doesn't exist
				err = client.MkdirAll(remotePath)
				if err != nil {
					return fmt.Errorf("failed to create remote directory %s: %w", remotePath, err)
				}
				logger.Debug(fmt.Sprintf("Created remote directory: %s\n", remotePath))
			} else if d.Type().IsRegular() {
				// Upload regular file
				localFile, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open local file %s: %w", path, err)
				}
				defer localFile.Close()

				remoteFile, err := client.Create(remotePath)
				if err != nil {
					return fmt.Errorf("failed to create remote file %s: %w", remotePath, err)
				}
				defer remoteFile.Close()

				_, err = io.Copy(remoteFile, localFile)
				if err != nil {
					return fmt.Errorf("failed to copy file %s to %s: %w", path, remotePath, err)
				}

				// Set permissions
				info, err := d.Info()
				if err == nil {
					if err := remoteFile.Chmod(info.Mode()); err != nil {
						logger.Warn(fmt.Sprintf("Warning: Failed to set permissions for %s: %v", remotePath, err))
					}
				}

				logger.Info(fmt.Sprintf("Uploaded file: %s -> %s", path, remotePath))
			} else {
				logger.Info(fmt.Sprintf("Skipping unsupported file type: %s (%v)\n", path, d.Type()))
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("error during directory walk/upload: %w", err)
	}

	return nil
}

func ShouldSkipUpload(d fs.DirEntry) bool {
	if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
		return true
	}

	if d.IsDir() && d.Name() == "koreader" {
		logger.Debug("Skipping (assumed) koreader submodule")
		return true
	}

	return false
}
