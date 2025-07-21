package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	httpinspector "github.com/Consoleaf/kopl/http_inspector"
	koreaderinspector "github.com/Consoleaf/koreader-http-inspector"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	sshPort    int
	deployPath string
	localPath  string

	sshUser     string
	sshPassword string
)

func init() {
	rootCmd.AddCommand(deployCmd)
	httpinspector.AddArgs(deployCmd)
	deployCmd.Flags().IntVarP(
		&sshPort,
		"ssh-port",
		"s",
		0,
		"SSH port. Use this if you have a separate SSH instance running.",
	)
	deployCmd.Flags().StringVarP(
		&deployPath,
		"deploy-path",
		"d",
		"/mnt/us/koreader/plugins",
		"Path to the koreader directory on device. Defaults to /mnt/us/koreader",
	)
	deployCmd.Flags().StringVarP(
		&sshUser,
		"ssh-user",
		"u",
		"root",
		"SSH username",
	)
	deployCmd.Flags().StringVarP(
		&sshPassword,
		"ssh-password",
		"P",
		"",
		"SSH password",
	)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy project to a device",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			localPath = args[0]

			err := validateLocalPath()
			if err != nil {
				log.Fatalf("Invalid positional argument, needs to be a directory of the plugin: %v", err)
			}
		} else {
			var err error
			localPath, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}
		err := deployCmdImpl(cmd, args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func deployCmdImpl(_ *cobra.Command, _ []string) error {
	insp, err := koreaderinspector.New(fmt.Sprintf("http://%s:%d/", httpinspector.Host, httpinspector.Port))
	if err != nil {
		return err
	}

	// NOTE: Defer here instead of just calling it to run this *after* insp.SSHSetAllowNoPassword(allow)
	defer restartKOReader(insp)

	if sshPort == 0 {
		fmt.Println("SSH port not provided. Turning on SSH over HTTP Inspector...")

		err = insp.SSHStop()
		if err != nil {
			return err
		}
		sshPort, err = insp.SSHStart()
		if err != nil {
			return err
		}

		allow, err := insp.SSHGetAllowNoPassword()
		if err != nil {
			return err
		}

		if !allow {
			defer insp.SSHSetAllowNoPassword(allow)
		}

		err = insp.SSHSetAllowNoPassword(true)
		if err != nil {
			return err
		}
	}

	if sshPort == 0 {
		return fmt.Errorf("SSH port is %v", sshPort)
	}

	err = syncDirToRemote()
	return err
}

func restartKOReader(insp *koreaderinspector.HTTPInspectorClient) {
	fmt.Println("Restarting KOReader...")
	err := insp.SSHStop()
	if err != nil {
		log.Fatal(err)
	}
	insp.RestartKOReader()
}

func validateLocalPath() error {
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory '%s' does not exist: %v", localPath, err)
		}
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("'%s' is not a directory", localPath)
	}

	return nil
}

func syncDirToRemote() error {
	name := filepath.Base(localPath)
	remoteBasePath := deployPath
	remoteTargetDir := filepath.Join(remoteBasePath, name)

	fmt.Printf("Preparing to sync local dir '%s' to remote '%s@%s:%v' as '%s'...\n", localPath, sshUser, httpinspector.Host, sshPort, remoteTargetDir)

	config := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 5,
	}
	conn, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", httpinspector.Host, sshPort),
		config,
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.MkdirAll(remoteTargetDir)
	if err != nil {
		return err
	}

	fmt.Printf("Starting upload of local directory '%s' to '%s'...\n", localPath, remoteTargetDir)
	err = filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Determine relative path to maintain directory structure
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if d.IsDir() && (d.Name() == ".git" ||
			strings.Contains(relPath, "/.git/") ||
			d.Name() == "koreader") {
			fmt.Printf("Skipping directory: %s\n", path)
			return filepath.SkipDir // Skip this directory and its contents
		}

		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Convert to Unix-style path for SFTP
		remotePath := filepath.ToSlash(filepath.Join(remoteTargetDir, relPath))

		if d.IsDir() {
			// Create directory on remote if it doesn't exist
			err = client.MkdirAll(remotePath)
			if err != nil {
				return fmt.Errorf("failed to create remote directory %s: %w", remotePath, err)
			}
			fmt.Printf("Created remote directory: %s\n", remotePath)
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

			// Set permissions (optional but good practice)
			info, err := d.Info()
			if err == nil {
				if err := remoteFile.Chmod(info.Mode()); err != nil {
					log.Printf("Warning: Failed to set permissions for %s: %v\n", remotePath, err)
				}
			}

			fmt.Printf("Uploaded file: %s -> %s\n", path, remotePath)
		} else {
			// Handle other file types like symlinks if necessary (not covered here)
			fmt.Printf("Skipping unsupported file type: %s (%v)\n", path, d.Type())
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error during directory walk/upload: %w", err)
	}

	return nil
}
