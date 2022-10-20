package internalssh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
	"unisync/config"
	"unisync/log"
	"unisync/myssh"

	"golang.org/x/crypto/ssh"
)

type internalSshClient struct {
	ssh *ssh.Client
}

func New(conf *config.Config) (*internalSshClient, error) {
	key, err := os.ReadFile(conf.SshKey)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: conf.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),

		Timeout: time.Duration(conf.ConnectTimeout) * time.Second,
	}

	c := &internalSshClient{}
	addr := fmt.Sprintf("%v:%v", conf.Host, conf.Port)
	c.ssh, err = dial(addr, config, conf.Timeout)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	return c, nil
}

// replacement for ssh.Dial() to give us control over KeepAlive
func dial(addr string, config *ssh.ClientConfig, timeout int) (*ssh.Client, error) {
	dialer := net.Dialer{
		Timeout:   config.Timeout,
		KeepAlive: time.Duration(timeout) * time.Second,
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func (c *internalSshClient) Search(locations []string) (string, error) {
	var err error
	var output []byte

	for _, location := range locations {
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

func (c *internalSshClient) Run(location string) (stdin io.Writer, stdout io.Reader, err error) {
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
