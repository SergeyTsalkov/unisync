package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unisync/commands"
	"unisync/filelist"
	"unisync/node"
)

type Client struct {
	version    int
	in         *bufio.Reader
	out        *node.Writer
	buffersize int

	LocalPath  string
	RemotePath string
}

func New(in io.Reader, out io.Writer) *Client {
	writer := node.NewWriter(out)
	writer.Debug = true

	return &Client{
		in:         bufio.NewReader(in),
		out:        writer,
		buffersize: 1000000,
	}
}

func (c *Client) path(path string) string {
	return node.Path(c.LocalPath, path)
}

func (c *Client) RunHello() error {
	cmd := &commands.Hello{c.RemotePath}
	err := c.Send(cmd)
	if err != nil {
		return err
	}

	_, err = c.WaitFor("OK")
	return err
}

func (c *Client) RunReqList() (filelist.FileList, error) {
	cmd := &commands.ReqList{}
	err := c.Send(cmd)
	if err != nil {
		return nil, err
	}

	json, err := c.WaitFor("RESLIST")
	if err != nil {
		return nil, err
	}

	reply := &commands.ResList{}
	err = commands.Parse(json, reply)
	if err != nil {
		return nil, err
	}

	return reply.FileList, nil
}

func (c *Client) Send(cmd commands.Command) error {
	return c.out.SendCmd(cmd)
}

func (c *Client) GetCommand() (cmd string, json string, err error) {
	var line string

	for {
		line, err = c.in.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		fmt.Printf("<- %v\n", line)

		words := strings.Fields(line)
		cmd = strings.ToUpper(words[0])
		json = strings.TrimSpace(strings.TrimPrefix(line, cmd))
		return
	}
}

func (c *Client) WaitFor(expectCmd string) (json string, err error) {
	var cmd string
	cmd, json, err = c.GetCommand()
	if err != nil {
		return
	}
	if cmd != expectCmd {
		return "", fmt.Errorf("expected %v from server but got %v", expectCmd, cmd)
	}

	return
}
