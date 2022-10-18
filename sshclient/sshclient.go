package sshclient

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"unisync/log"
)

type SSHClient struct {
	In  io.Writer
	Out io.Reader

	cmd    *exec.Cmd
	stderr io.Reader
}

func New(username, host string) *SSHClient {
	c := &SSHClient{}

	cmd := strings.Split(fmt.Sprintf("-e none -o BatchMode=yes -o ConnectTimeout=30 -o StrictHostKeyChecking=no %v@%v unisync -stdserver", username, host), " ")
	c.cmd = exec.Command("ssh", cmd...)
	var err error

	c.In, err = c.cmd.StdinPipe()
	if err != nil {
		log.Fatalln(err)
	}

	c.Out, err = c.cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}

	return c
}

func (c *SSHClient) Run() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}

	go c.Wait()
	return nil
}

func (c *SSHClient) Wait() {
	reader := bufio.NewReader(c.stderr)
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

	err := c.cmd.Wait()
	if err != nil {
		log.Fatalln("ssh exited:", err)
	} else {
		log.Fatalln("ssh exited")
	}

}
