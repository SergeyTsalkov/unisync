//go:build !windows
// +build !windows

package pageant

import "golang.org/x/crypto/ssh"

func GetSigners() ([]ssh.Signer, error) {
	return nil, nil
}
