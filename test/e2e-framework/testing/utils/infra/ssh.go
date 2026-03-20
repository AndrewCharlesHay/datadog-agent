// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package infra implements utilities to interact with a Pulumi infrastructure
package infra

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func SshConnectToInstance(ip, port, user string) (*ssh.Client, error) {
	auth := []ssh.AuthMethod{}

	if sshAgentSocket, found := os.LookupEnv("SSH_AUTH_SOCK"); found {
		sshAgent, err := net.Dial("unix", sshAgentSocket)
		if err != nil {
			return nil, fmt.Errorf("failed to dial SSH agent: %w", err)
		}
		defer sshAgent.Close()

		auth = append(auth, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	if sshKeyPath, found := os.LookupEnv("E2E_AWS_PRIVATE_KEY_PATH"); found {
		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	if sshKeyPath, found := os.LookupEnv("E2E_GCP_PRIVATE_KEY_PATH"); found {
		sshKey, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key: %w", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	return ssh.Dial("tcp", ip+":"+port, &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
}

func SshRunCommand(sshClient *ssh.Client, command string) ([]byte, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	output, err := sshSession.CombinedOutput(command)
	if err != nil {
		return nil, err
	}

	return output, nil
}
