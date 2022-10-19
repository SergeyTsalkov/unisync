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

	sshcmd  []string
	execCmd *exec.Cmd
	stderr  io.Reader
}

func (c *SSHClient) cmd(format string, a ...any) *exec.Cmd {
	len := len(c.sshcmd)
	out := make([]string, len+1)
	copy(out, c.sshcmd)
	out[len] = fmt.Sprintf(format, a...)
	return exec.Command(out[0], out[1:]...)
}

func New(username, host, ssh_path, ssh_opts string) *SSHClient {
	c := &SSHClient{}
	sshcmd := fmt.Sprintf("%v %v %v@%v", ssh_path, ssh_opts, username, host)
	c.sshcmd = strings.Split(sshcmd, " ")
	return c
}

func (c *SSHClient) Search(locations []string) (string, error) {
	var err error
	var output []byte

	for _, location := range locations {
		execCmd := c.cmd("command -v %v", location)
		output, err = execCmd.CombinedOutput()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					continue
				}
			}
			return "", &ssherror{err, output}
		}

		return location, nil
	}

	return "", fmt.Errorf("Unable to find unisync binary: %v", &ssherror{err, output})
}

func (c *SSHClient) Run(location string) error {
	c.execCmd = c.cmd("%v -stdserver", location)

	var err error
	if c.In, err = c.execCmd.StdinPipe(); err != nil {
		panic(err)
	}
	if c.Out, err = c.execCmd.StdoutPipe(); err != nil {
		panic(err)
	}
	if c.stderr, err = c.execCmd.StderrPipe(); err != nil {
		panic(err)
	}

	err = c.execCmd.Start()
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

	err := c.execCmd.Wait()
	if err != nil {
		log.Fatalln("ssh exited:", err)
	} else {
		log.Fatalln("ssh exited")
	}

}
