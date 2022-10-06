package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unisync/commands"
	"unisync/config"
	"unisync/filelist"
	"unisync/node"
)

type Client struct {
	*node.Node
	cache filelist.FileList
}

func New(in io.Reader, out io.Writer, config *config.Config) (*Client, error) {
	node := &node.Node{
		In:     bufio.NewReader(in),
		Out:    out,
		Debug:  true,
		Config: config,
	}
	client := &Client{Node: node}

	err := client.SetBasepath(config.Local)
	if err != nil {
		return nil, fmt.Errorf("Unable to set basepath: %w", err)
	}

	return client, nil
}

func (c *Client) RunHello() error {
	cmd := &commands.Hello{c.Config}
	err := c.SendCmd(cmd)
	if err != nil {
		return err
	}

	_, err = c.WaitFor("OK")
	return err
}

func (c *Client) RunReqList() (filelist.FileList, error) {
	cmd := &commands.ReqList{}
	err := c.SendCmd(cmd)
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

func (c *Client) GetCommand() (cmd string, json string, err error) {
	var line string

	for {
		line, err = c.In.ReadString('\n')
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
