package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unisync/commands"
)

type Client struct {
	version    int
	in         *bufio.Reader
	out        io.Writer
	buffersize int

	LocalPath  string
	RemotePath string
}

func New(in io.Reader, out io.Writer) *Client {
	return &Client{
		in:         bufio.NewReader(in),
		out:        out,
		buffersize: 1000000,
	}
}

func (c *Client) RunHello() error {
	hello := &commands.HelloCommand{c.RemotePath}
	err := c.RunCommand(hello)
	if err != nil {
		return err
	}

	_, err = c.GetCommand()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RunCommand(cmd commands.Command) error {
	output := cmd.Encode()
	fmt.Printf("-> %v", output)
	_, err := io.WriteString(c.out, cmd.Encode())
	return err
}

func (c *Client) GetCommand() (string, error) {
	for {
		line, err := c.in.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		fmt.Printf("<- %v\n", line)

		// words := strings.Fields(line)
		// cmd := strings.ToUpper(words[0])
		// json := strings.TrimSpace(strings.TrimPrefix(line, cmd))
		return line, nil
	}
}
