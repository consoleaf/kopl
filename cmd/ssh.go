package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var (
	SSHPort     int
	SSHUser     string
	SSHPassword string

	sshClient *ssh.Client
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

	config := &ssh.ClientConfig{
		User: SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(SSHPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 5,
	}
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
