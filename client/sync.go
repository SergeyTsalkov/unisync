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

func (c *Client) Sync() (bool, error) {
	remoteList, err := c.RunReqList()
	if err != nil {
		return false, err
	}

	localList, err := filelist.Make(c.Basepath)
	if err != nil {
		return false, err
	}

	cacheList, err := c.Cache()
	if err != nil {
		return false, err
	}

	b := filelist.NewSyncPlanBuilder(c.Config)
	syncplan := b.BuildSyncPlan(localList, remoteList, cacheList)

	if syncplan.IsSynced() {
		err = c.SaveCache(localList)
		return true, err
	}

	for _, file := range syncplan.LocalDel {
		err := os.Remove(file.Path)
		if err != nil {
			return false, err
		}
	}

	for _, file := range syncplan.LocalMkdir {
		err := c.Mkdir(file.Path, file.Mode)
		if err != nil {
			return false, err
		}
	}

	if len(syncplan.RemoteMkdir) > 0 {
		mkdir := commands.MakeMkdir(syncplan.RemoteMkdir)
		err = c.SendCmd(mkdir)
		if err != nil {
			return false, err
		}
		_, err = c.WaitFor("OK")
		if err != nil {
			return false, err
		}
	}

	for _, file := range syncplan.LocalChmod {
		err := c.Chmod(file.Path, file.Mode)
		if err != nil {
			return false, err
		}
	}

	if len(syncplan.RemoteChmod) > 0 {
		chmod := commands.MakeChmod(syncplan.RemoteChmod)
		err = c.SendCmd(chmod)
		if err != nil {
			return false, err
		}
		_, err = c.WaitFor("OK")
		if err != nil {
			return false, err
		}
	}

	for _, file := range syncplan.PushFile {
		err = c.SendFile(file.Path)
		if err != nil {
			log.Println(err)
		}
	}

	if len(syncplan.PullFile) > 0 {
		paths := map[string]bool{}
		for _, file := range syncplan.PullFile {
			paths[file.Path] = true
		}

		pull := commands.MakePull(syncplan.PullFile)
		err = c.SendCmd(pull)
		if err != nil {
			return false, err
		}

		for len(paths) > 0 {
			json, err := c.WaitFor("PUSH")
			if err != nil {
				return false, err
			}

			push := &commands.Push{}
			err = commands.Parse(json, push)
			if err != nil {
				return false, err
			}

			if push.Length > int64(len(node.Buffer)) {
				return false, fmt.Errorf("Buffer length is %v, but needs to be at least %v", len(node.Buffer), push.Length)
			}

			buf := node.Buffer[0:push.Length]
			_, err = io.ReadAtLeast(c.In, buf, len(buf))
			if err != nil {
				return false, err
			}

			done, err := c.ReceiveFile(push, buf)
			if err != nil {
				return false, err
			}
			if done {
				delete(paths, push.Path)
			}
		}

	}

	return false, nil
}
