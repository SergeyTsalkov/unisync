package client

import (
	"fmt"
	"io"
	"os"
	"unisync/commands"
	"unisync/config"
	"unisync/filelist"
	"unisync/log"
	"unisync/node"
	"unisync/progresswriter"
)

type Client struct {
	*node.Node
	Background     bool
	cache          filelist.FileList
	remoteBasepath string
}

func New(in io.Reader, out io.Writer, config *config.Config) (*Client, error) {
	n := node.New(in, out)
	n.Config = config
	n.SetSideC("FSEVENT", "PROGRESS")
	client := &Client{Node: n}

	err := client.SetTmpdir(config.TmpdirLocal)
	if err != nil {
		return nil, fmt.Errorf("Unable to set tmpdir: %w", err)
	}

	return client, nil
}

// separate goroutine
func (c *Client) SideChannelReader() {
	for packet := range c.SideC {
		switch cmdType := packet.Command.CmdType(); cmdType {
		case "FSEVENT":
			c.Watcher.Send("")

		case "PROGRESS":
			c.handlePROGRESS(packet.Command)

		default:
			panic("invalid packet in SideC: " + cmdType)
		}
	}
}

func (c *Client) Run() (bool, error) {
	err := c.SetBasepath(c.Config.Local)
	if err != nil {
		return false, fmt.Errorf("Unable to set basepath: %w", err)
	}

	go c.SideChannelReader()
	defer c.Watcher.Stop()

	if err := c.RunHello(); err != nil {
		return false, err
	}
	if err := c.Sync(); err != nil {
		return false, err
	}

	if c.Config.WatchLocal == "0" && c.Config.WatchRemote == "0" {
		log.Printf("%v %v", "[X]", "Synced. All done..")
		return true, nil
	}

	// this will let the -start runner know to detach and let us run in the background
	// if we don't close stdout, Unix produces SIGPIPE when we next try to write to it
	// unfortunately os.Stderr can't be closed because of a potential issue in go (see go "os" docs)
	// but that's okay because we'll never use os.Stderr
	if c.Background {
		log.Printf("%v %v", "[X]", "Synced. Will run in background..")
		log.ScreenOutput = nil
		os.Stdout.Close()
	}

	for {
		log.Printf("%v %v", "[X]", "Synced. Watching for changes..")

		select {
		case <-c.Watcher.C:
			if err := c.Sync(); err != nil {
				return true, err
			}
		case err := <-c.DoneC():
			return true, err
		}
	}

	return true, fmt.Errorf("unexpected exit")
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
	log.Printf("Syncing: %v <-> %v@%v:%v", c.GetBasepath(), c.Config.User, c.Config.Host, c.remoteBasepath)
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

func (c *Client) handlePROGRESS(cmd commands.Command) {
	progress := cmd.(*commands.Progress)

	select {
	case c.Progress <- progresswriter.Progress{progress.Percent, progress.Eta}:
	default:
	}
}
