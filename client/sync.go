package client

import (
	"log"
	"os"
	"unisync/commands"
	"unisync/filelist"
	"unisync/node"
)

func (c *Client) Sync() error {
	remoteList, err := c.RunReqList()
	if err != nil {
		return err
	}

	localList, err := filelist.Make(c.LocalPath)
	if err != nil {
		log.Fatalln(err)
	}

	syncplan := filelist.Compare(localList, remoteList)
	syncplan.Show()

	for _, file := range syncplan.LocalMkdir {
		fullpath := c.path(file.Path)
		err := os.MkdirAll(fullpath, 0755)
		if err != nil {
			return err
		}
	}

	if len(syncplan.RemoteMkdir) > 0 {
		mkdir := &commands.Mkdir{
			Dirs: []*commands.MkdirAction{},
		}

		for _, file := range syncplan.RemoteMkdir {
			mkdir.Dirs = append(mkdir.Dirs, &commands.MkdirAction{Path: file.Path})
		}

		err = c.Send(mkdir)
		if err != nil {
			return err
		}

		_, err = c.WaitFor("OK")
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.Push {
		err = node.SendFile(c.out, file.Path, c.path(file.Path))
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
