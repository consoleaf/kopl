package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/xyproto/randomstring"
)

func init() {
	rootCmd.AddCommand(installCmd)
	AddInspectorArgs(installCmd)
	AddSSHFlags(installCmd)

	installCmd.Flags().StringVarP(
		&deployPath,
		"deploy-path",
		"d",
		"/mnt/us/koreader/plugins",
		"Path to the koreader directory on device. Defaults to /mnt/us/koreader",
	)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a remotely hosted koplugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		remoteRepo := args[0]
		err := installImpl(remoteRepo)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func installImpl(remoteRepo string) error {
	InitializeInspector()

	success := false

	defer func(success *bool) {
		if *success {
			restartKOReader()
		}
	}(&success)

	revert, err := makeRevertSSHAllowNoPassword()
	if err != nil {
		return err
	}
	defer revert()

	conn, err := connectSSH()
	if err != nil {
		return err
	}
	defer conn.Close()

	sftp, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer sftp.Close()

	randomstring.Seed()
	tmp := path.Join(os.TempDir(), randomstring.HumanFriendlyEnglishString(5))

	localRepoPath := strings.TrimSuffix(path.Join(tmp, path.Base(remoteRepo)), ".git")
	if !strings.HasSuffix(localRepoPath, ".koplugin") {
		localRepoPath = localRepoPath + ".koplugin"
	}

	url := remoteRepo
	if !strings.HasPrefix(url, "http") {
		url = "https://github.com/" + url
	}

	logger.Info(fmt.Sprintf("Cloning '%s' into '%s'...", url, localRepoPath))
	_, err = git.PlainClone(localRepoPath, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return err
	}
	defer func() {
		logger.Info(fmt.Sprintf("Deleting '%s'...", localRepoPath))
		os.RemoveAll(localRepoPath)
	}()

	logger.Info(fmt.Sprintf("Uploading '%s' to the device...", localRepoPath))
	err = UploadDirectory(localRepoPath, deployPath, sftp)
	if err != nil {
		return err
	}
	success = true
	return nil
}
