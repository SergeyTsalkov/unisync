package client

import (
	"fmt"
	"io"
	"unisync/commands"
	"unisync/config"
	"unisync/filelist"
	"unisync/node"
)

type Client struct {
	*node.Node
	cache          filelist.FileList
	remoteBasepath string
	waitForC       chan *node.Packet
}

func New(in io.Reader, out io.Writer, config *config.Config) (*Client, error) {
	n := node.New(in, out)
	n.Debug = true
	n.Config = config
	client := &Client{
		Node:     n,
		waitForC: make(chan *node.Packet),
	}

	err := client.SetBasepath(config.Local)
	if err != nil {
		return nil, fmt.Errorf("Unable to set basepath: %w", err)
	}

	go client.PacketReader()
	return client, nil
}

// separate goroutine
func (c *Client) PacketReader() {
	for packet := range c.Packets {
		switch packet.Command.CmdType() {
		case "FSEVENT":
			c.Watcher.Send("")

		default:
			c.waitForC <- packet
		}
	}
}

func (c *Client) Run() error {
	if err := c.RunHello(); err != nil {
		return err
	}
	if err := c.Sync(); err != nil {
		return err
	}

	for {
		select {
		case <-c.Watcher.C:
			if err := c.Sync(); err != nil {
				return err
			}
		case err := <-c.Errors:
			return err
		}
	}

	return nil
}

func (c *Client) RunHello() error {
	hello := &commands.Hello{c.Config}
	err := c.SendCmd(hello)
	if err != nil {
		return err
	}

	cmd, _, err := c.WaitFor("WHATSUP")
	if err != nil {
		return err
	}

	whatsup := cmd.(*commands.Whatsup)
	c.remoteBasepath = whatsup.Basepath
	return nil
}

func (c *Client) RunReqList() (filelist.FileList, error) {
	reqlist := &commands.ReqList{}
	err := c.SendCmd(reqlist)
	if err != nil {
		return nil, err
	}

	cmd, _, err := c.WaitFor("RESLIST")
	if err != nil {
		return nil, err
	}

	reply := cmd.(*commands.ResList)
	return reply.FileList, nil
}

func (c *Client) WaitFor(expectCmd string) (commands.Command, []byte, error) {
	packet := <-c.waitForC

	if cmdType := packet.Command.CmdType(); cmdType != expectCmd {
		return nil, nil, fmt.Errorf("expected %v from server but got %v", expectCmd, cmdType)
	}

	return packet.Command, packet.Buffer, nil
}
