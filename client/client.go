package client

import (
	"fmt"
	"io"
	"unisync/commands"
	"unisync/config"
	"unisync/filelist"
	"unisync/log"
	"unisync/node"
)

type Client struct {
	*node.Node
	cache          filelist.FileList
	remoteBasepath string
}

func New(in io.Reader, out io.Writer, config *config.Config) (*Client, error) {
	n := node.New(in, out)
	n.Config = config
	n.SetSideC("FSEVENT")
	client := &Client{Node: n}

	err := client.SetBasepath(config.Local)
	if err != nil {
		return nil, fmt.Errorf("Unable to set basepath: %w", err)
	}

	go client.PacketReader()
	return client, nil
}

// separate goroutine
func (c *Client) PacketReader() {
	for packet := range c.SideC {
		switch cmdType := packet.Command.CmdType(); cmdType {
		case "FSEVENT":
			c.Watcher.Send("")

		default:
			panic("invalid packet in SideC: " + cmdType)
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
		log.Printf("%v %v", "[X]", "Synced. Watching for changes..")

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
