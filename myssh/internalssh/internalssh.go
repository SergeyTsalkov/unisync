package internalssh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"unisync/config"
	"unisync/log"
	"unisync/myssh"
	"unisync/pageant"

	"golang.org/x/crypto/ssh"
)

type internalSshClient struct {
	ssh       *ssh.Client
	locations []string
}

func New(conf *config.Config) (*internalSshClient, error) {
	var signers []ssh.Signer
	var err error

	if pSigners, err := pageant.GetSigners(); err == nil && len(pSigners) > 0 {
		signers = append(signers, pSigners...)
	}

	for _, keypath := range conf.SshKeys {
		key, err := os.ReadFile(keypath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %w", err)
		}
		signers = append(signers, signer)
	}

	if len(signers) == 0 {
		return nil, fmt.Errorf("no ssh_key available")
	}

	sshConfig := &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         config.Duration(conf.ConnectTimeout),
	}

	c := &internalSshClient{
		locations: conf.RemoteUnisyncPath,
	}
	addr := fmt.Sprintf("%v:%v", conf.Host, conf.Port)
	c.ssh, err = dial(addr, sshConfig, conf.Timeout)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	return c, nil
}

// replacement for ssh.Dial() to give us control over KeepAlive
func dial(addr string, sshConfig *ssh.ClientConfig, keepAlive int) (*ssh.Client, error) {
	dialer := net.Dialer{
		Timeout:   sshConfig.Timeout,
		KeepAlive: config.Duration(keepAlive),
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, sshConfig)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func (c *internalSshClient) search() (string, error) {
	var err error
	var output []byte

	for _, location := range c.locations {
		var session *ssh.Session
		session, err = c.ssh.NewSession()
		defer session.Close()

		if err != nil {
			return "", fmt.Errorf("unable to create ssh session: %w", err)
		}

		output, err = session.CombinedOutput("command -v " + location)
		if err != nil {
			if exitError, ok := err.(*ssh.ExitError); ok {
				if exitError.ExitStatus() == 1 {
					continue
				}
			}
			return "", &myssh.SshError{err, output}
		}

		return location, nil
	}

	return "", fmt.Errorf("Unable to find unisync binary: %v", &myssh.SshError{err, output})
}

func (c *internalSshClient) Run() (stdin io.Writer, stdout io.Reader, err error) {
	location := c.locations[0]
	if len(c.locations) > 1 {
		var err error
		location, err = c.search()
		if err != nil {
			return nil, nil, err
		}
	}

	var session *ssh.Session
	session, err = c.ssh.NewSession()
	if err != nil {
		return
	}

	var stderr io.Reader
	if stdin, err = session.StdinPipe(); err != nil {
		return
	}
	if stdout, err = session.StdoutPipe(); err != nil {
		return
	}
	if stderr, err = session.StderrPipe(); err != nil {
		return
	}

	err = session.Start(fmt.Sprintf("%v -stdserver", location))
	if err != nil {
		return
	}

	go c.wait(session, stderr)
	return

}

func (c *internalSshClient) wait(session *ssh.Session, stderr io.Reader) {
	reader := bufio.NewReader(stderr)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line != "" {
			log.Warnln("Server Error:", line)
		}

		if err != nil {
			break
		}
	}

	err := session.Wait()
	if err != nil {
		log.Fatalln("ssh exited:", err)
	} else {
		log.Fatalln("ssh exited")
	}

}
