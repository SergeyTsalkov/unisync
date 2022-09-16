package sshclient

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type SSHClient struct {
	In  io.Writer
	Out io.Reader

	cmd    *exec.Cmd
	stderr io.Reader
}

func New() *SSHClient {
	c := &SSHClient{}
	c.cmd = exec.Command("ssh", "-e", "none", "sergey@51.79.19.179", "unisync -stdserver")
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
			fmt.Println("Server Error:", line)
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
