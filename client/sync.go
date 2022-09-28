package client

import (
	"fmt"
	"io"
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
		return err
	}

	syncplan := filelist.Compare(localList, remoteList)

	for _, file := range syncplan.LocalMkdir {
		fullpath := c.path(file.Path)
		err := os.MkdirAll(fullpath, 0755)
		if err != nil {
			return err
		}
	}

	if len(syncplan.RemoteMkdir) > 0 {
		mkdir := &commands.Mkdir{
			Dirs: make([]*commands.MkdirAction, len(syncplan.RemoteMkdir)),
		}

		for i, file := range syncplan.RemoteMkdir {
			mkdir.Dirs[i] = &commands.MkdirAction{Path: file.Path}
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

	if len(syncplan.Pull) > 0 {
		paths := map[string]bool{}
		pull := &commands.Pull{
			Paths: make([]string, len(syncplan.Pull)),
		}

		for i, file := range syncplan.Pull {
			pull.Paths[i] = file.Path
			paths[file.Path] = true
		}

		err = c.Send(pull)
		if err != nil {
			return err
		}

		for len(paths) > 0 {
			json, err := c.WaitFor("PUSH")
			if err != nil {
				return err
			}

			push := &commands.Push{}
			err = commands.Parse(json, push)
			if err != nil {
				return err
			}

			if push.Length > int64(len(node.Buffer)) {
				return fmt.Errorf("Buffer length is %v, but needs to be at least %v", len(node.Buffer), push.Length)
			}

			buf := node.Buffer[0:push.Length]
			_, err = io.ReadAtLeast(c.in, buf, len(buf))
			if err != nil {
				return err
			}

			done, err := node.ReceiveFile(c.path(push.Path), push, buf)
			if err != nil {
				return err
			}
			if done {
				delete(paths, push.Path)
			}
		}

	}

	return nil
}
