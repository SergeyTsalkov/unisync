//go:build windows
// +build windows

package pageant

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func GetSigners() ([]ssh.Signer, error) {
	agentConn, err := NewConn()
	if err != nil {
		return nil, err
	}
	defer agentConn.Close()

	signers, err := agent.NewClient(agentConn).Signers()
	if err != nil {
		return nil, err
	}
	return signers, nil
}
