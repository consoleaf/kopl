package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
)

var (
	deployPath string
	localPath  string
)

func init() {
	rootCmd.AddCommand(deployCmd)
	AddInspectorArgs(deployCmd)
	AddSSHFlags(deployCmd)

	deployCmd.Flags().StringVarP(
		&deployPath,
		"deploy-path",
		"d",
		"/mnt/us/koreader/plugins",
		"Path to the koreader directory on device. Defaults to /mnt/us/koreader",
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
	InitializeInspector()

	defer restartKOReader()
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

	return UploadDirectory(localPath, deployPath, sftp)
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

	if !strings.HasSuffix(fileInfo.Name(), ".koplugin") {
		return fmt.Errorf(
			"'%s' doesn't look like a plugin and will not be recognized by KOReader as such",
			fileInfo.Name(),
		)
	}

	return nil
}
