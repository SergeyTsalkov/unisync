package client

import (
	"fmt"
	"log"
	"os"
	"unisync/commands"
	"unisync/filelist"
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

	localList, err := filelist.Make(c.GetBasepath(), c.Config.Ignore)
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
		_, _, err = c.WaitFor("OK")
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
		_, _, err = c.WaitFor("OK")
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
		_, _, err = c.WaitFor("OK")
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
		_, _, err = c.WaitFor("OK")
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
			cmd, buf, err := c.WaitFor("PUSH")
			if err != nil {
				return err
			}

			push := cmd.(*commands.Push)
			err = c.ReceiveFile(push, buf)
			if err != nil {
				return err
			}
			if !push.More {
				delete(paths, push.Path)
			}
		}

	}

	return nil
}
