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

	for tries := 1; tries < 3; tries++ {
		syncplan, localList, err := c.MakeSyncPlan()
		if err != nil {
			return err
		}

		if syncplan.IsSynced() {
			return c.SaveCache(localList)
		}

		err = c.RunSyncPlan(syncplan)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("Unable to sync after several tries!")
}

func (c *Client) MakeSyncPlan() (*filelist.SyncPlan, filelist.FileList, error) {
	remoteList, err := c.RunReqList()
	if err != nil {
		return nil, nil, err
	}

	localList, err := filelist.Make(c.GetBasepath())
	if err != nil {
		return nil, nil, err
	}

	cacheList, err := c.Cache()
	if err != nil {
		return nil, nil, err
	}

	b := filelist.NewSyncPlanBuilder(c.Config)
	syncplan := b.BuildSyncPlan(localList, remoteList, cacheList)
	return syncplan, localList, nil
}

func (c *Client) RunSyncPlan(syncplan *filelist.SyncPlan) error {
	var err error
	for _, file := range syncplan.LocalDel {
		err = os.Remove(c.Path(file.Path))
		if err != nil {
			return err
		}
	}

	if len(syncplan.RemoteDel) > 0 {
		del := commands.MakeDel(syncplan.RemoteDel)
		err = c.SendCmd(del)
		if err != nil {
			return err
		}
		_, err = c.WaitFor("OK")
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.LocalMkdir {
		err = c.Mkdir(file.Path, file.Mode)
		if err != nil {
			return err
		}
	}

	if len(syncplan.RemoteMkdir) > 0 {
		mkdir := commands.MakeMkdir(syncplan.RemoteMkdir)
		err = c.SendCmd(mkdir)
		if err != nil {
			return err
		}
		_, err = c.WaitFor("OK")
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.LocalMklink {
		err = c.Symlink(file.Symlink, file.Path)
		if err != nil {
			return err
		}
	}

	if len(syncplan.RemoteMklink) > 0 {
		symlink := commands.MakeSymlink(syncplan.RemoteMklink)
		err = c.SendCmd(symlink)
		if err != nil {
			return err
		}
		_, err = c.WaitFor("OK")
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.LocalChmod {
		err = c.Chmod(file.Path, file.Mode)
		if err != nil {
			return err
		}
	}

	if len(syncplan.RemoteChmod) > 0 {
		chmod := commands.MakeChmod(syncplan.RemoteChmod)
		err = c.SendCmd(chmod)
		if err != nil {
			return err
		}
		_, err = c.WaitFor("OK")
		if err != nil {
			return err
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
			_, err = io.ReadAtLeast(c.In, buf, len(buf))
			if err != nil {
				return err
			}

			done, err := c.ReceiveFile(push, buf)
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
