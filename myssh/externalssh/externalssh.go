package externalssh

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"unisync/config"
	"unisync/log"
	"unisync/myssh"
)

type externalSshClient struct {
	sshcmd  []string
	execCmd *exec.Cmd
}

func (c *externalSshClient) cmd(format string, a ...any) *exec.Cmd {
	len := len(c.sshcmd)
	out := make([]string, len+1)
	copy(out, c.sshcmd)
	out[len] = fmt.Sprintf(format, a...)
	return exec.Command(out[0], out[1:]...)
}

func New(conf *config.Config) *externalSshClient {
	c := &externalSshClient{}

	if conf.Port != 22 {
		conf.SshOpts += " " + fmt.Sprintf("-p %v", conf.Port)
	}
	if conf.SshKey != "" {
		conf.SshOpts += " " + fmt.Sprintf("-i %v", conf.SshKey)
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

func (c *externalSshClient) Search(locations []string) (string, error) {
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
			return "", &myssh.SshError{err, output}
		}

		return location, nil
	}

	return "", fmt.Errorf("Unable to find unisync binary: %v", &myssh.SshError{err, output})
}

func (c *externalSshClient) Run(location string) (stdin io.Writer, stdout io.Reader, err error) {
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

	go c.wait(stderr)
	return
}

func (c *externalSshClient) wait(stderr io.Reader) {
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

	err := c.execCmd.Wait()
	if err != nil {
		log.Fatalln("ssh exited:", err)
	} else {
		log.Fatalln("ssh exited")
	}

}
