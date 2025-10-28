package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	SSHPort         int
	SSHUser         string
	SSHPassword     string
	SSHIdentityPath string
)

func AddSSHFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(
		&SSHPort,
		"ssh-port",
		"s",
		0,
		"SSH port. Use this if you have a separate SSH instance running.",
	)
	cmd.Flags().StringVarP(
		&SSHUser,
		"ssh-user",
		"u",
		"root",
		"SSH username",
	)
	cmd.Flags().StringVarP(
		&SSHPassword,
		"ssh-password",
		"P",
		"",
		"SSH password",
	)
	cmd.Flags().StringVarP(
		&SSHIdentityPath,
		"ssh-identity",
		"i",
		"",
		"SSH identity file",
	)
}

func connectSSH() (*ssh.Client, error) {
	if SSHPort == 0 {
		fmt.Println("SSH port not provided. Turning on SSH over HTTP Inspector...")

		err := Inspector.SSHStop()
		if err != nil {
			return nil, err
		}

		SSHPort, err = Inspector.SSHStart()
		if err != nil {
			return nil, err
		}

		err = Inspector.SSHSetAllowNoPassword(true)
		if err != nil {
			return nil, err
		}
	}

	if SSHPort == 0 {
		return nil, fmt.Errorf("SSH port is %v", SSHPort)
	}

	var sshAuth []ssh.AuthMethod
	var signers []ssh.Signer

	sshAuth = append(sshAuth, ssh.Password(SSHPassword))
	if SSHIdentityPath != "" {
		identity, err := makeIdentityFromPath(SSHIdentityPath)
		if err != nil {
			return nil, err
		}
		signers = append(signers, identity)
	}
	signers = append(signers, makeDotSshIdentities()...)

	if len(signers) != 0 {
		sshAuth = append(sshAuth, ssh.PublicKeys(signers...))
	}

	config := &ssh.ClientConfig{
		User:            SSHUser,
		Auth:            sshAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 5,
	}

	logger.Debug("Connecting over SSH", "host", Host, "port", SSHPort)

	conn, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", Host, SSHPort),
		config,
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func makeRevertSSHAllowNoPassword() (func(), error) {
	allow, err := Inspector.SSHGetAllowNoPassword()
	if err != nil {
		return nil, err
	}

	return func() {
		if !allow {
			Inspector.SSHSetAllowNoPassword(allow)
		}
	}, nil
}

func makeIdentityFromPath(path string) (ssh.Signer, error) {
	SSHIdentity, err := os.ReadFile(SSHIdentityPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read SSH identity file: %w", err)
	}
	SSHPrivateKey, err := ssh.ParsePrivateKey(SSHIdentity)
	if err != nil {
		return nil, err
	}
	return SSHPrivateKey, nil
}

func makeDotSshIdentities() []ssh.Signer {
	var res []ssh.Signer
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	sshDirList, err := os.ReadDir(path.Join(home, ".ssh"))
	if err != nil {
		return nil
	}
	for _, file := range sshDirList {
		if file.IsDir() {
			continue
		}
		data, err := os.ReadFile(path.Join(home, ".ssh", file.Name()))
		if err != nil {
			continue
		}
		key, err := ssh.ParsePrivateKey(data)
		if err != nil {
			continue
		}
		res = append(res, key)
	}
	return res
}
