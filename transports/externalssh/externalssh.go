package externalssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"unisync/config"
	"unisync/log"
)

type externalSshClient struct {
	sshcmd    []string
	execCmd   *exec.Cmd
	locations []string
}

func (c *externalSshClient) cmd(format string, a ...any) *exec.Cmd {
	len := len(c.sshcmd)
	out := make([]string, len+1)
	copy(out, c.sshcmd)
	out[len] = fmt.Sprintf(format, a...)
	return exec.Command(out[0], out[1:]...)
}

func New(conf *config.Config) *externalSshClient {
	c := &externalSshClient{
		locations: conf.RemoteUnisyncPath,
	}

	if conf.Port != 22 {
		conf.SshOpts += " " + fmt.Sprintf("-p %v", conf.Port)
	}
	for _, keypath := range conf.SshKeys {
		conf.SshOpts += " " + fmt.Sprintf("-i %v", keypath)
	}
	if conf.ConnectTimeout > 0 {
		conf.SshOpts += " " + fmt.Sprintf("-o ConnectTimeout=%v", conf.ConnectTimeout)
	}
	if conf.Timeout > 0 && conf.Timeout != 300 {
		// when BatchMode is on, the default ServerAliveInterval is already 300
		conf.SshOpts += " " + fmt.Sprintf("-o ServerAliveInterval=%v", conf.Timeout)
	}

	sshcmd := fmt.Sprintf("%v %v %v@%v", conf.SshPath, conf.SshOpts, conf.User, conf.Host)
	c.sshcmd = strings.Split(sshcmd, " ")
	return c
}

func (c *externalSshClient) search() (string, error) {
	var err error
	var output []byte

	for _, location := range c.locations {
		execCmd := c.cmd("command -v %v", location)
		output, err = execCmd.CombinedOutput()
		output = bytes.TrimSpace(output)
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					continue
				}
			}
			return "", fmt.Errorf("%s (%w)", output, err)
		}

		return location, nil
	}

	return "", fmt.Errorf("Unable to find unisync binary: %s (%w)", output, err)
}

func (c *externalSshClient) Run() (stdin io.Writer, stdout io.Reader, err error) {
	location := c.locations[0]
	if len(c.locations) > 1 {
		var err error
		location, err = c.search()
		if err != nil {
			return nil, nil, err
		}
	}

	c.execCmd = c.cmd("%v -stdserver", location)

	var stderr io.Reader
	if stdin, err = c.execCmd.StdinPipe(); err != nil {
		return
	}
	if stdout, err = c.execCmd.StdoutPipe(); err != nil {
		return
	}
	if stderr, err = c.execCmd.StderrPipe(); err != nil {
		return
	}

	err = c.execCmd.Start()
	if err != nil {
		return
	}

	go c.logerr(stderr)
	return
}

func (c *externalSshClient) Close() error {
	if c.execCmd == nil {
		return nil
	}
	if c.execCmd.Process != nil {
		err := c.execCmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	return c.execCmd.Wait()
}

// separate goroutine
func (c *externalSshClient) logerr(stderr io.Reader) {
	reader := bufio.NewReader(stderr)
	var err error

	for err == nil {
		var line string
		line, err = reader.ReadString('\n')
		line = strings.TrimSpace(line)

		if line != "" {
			log.Warnln("Server Says:", line)
		}
	}
}
