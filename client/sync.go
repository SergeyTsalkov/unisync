package client

import (
	"fmt"
	"os"
	"unisync/commands"
	"unisync/filelist"
	"unisync/log"
	"unisync/progressbar"
)

func (c *Client) Sync() error {
	c.Watcher.Ready()
	log.Printf("%v %v", "<->", "Comparing..")

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

	// if one side or the other is empty, don't use the cache
	// we'll assume that we want the empty side repopulated, and never want the full side emptied
	if len(localList) == 0 || len(remoteList) == 0 {
		c.RemoveCache()
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
		log.Printf("%v %v %v", "<-", "DEL", file.Path)
		err = os.Remove(c.Path(file.Path))
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.RemoteDel {
		log.Printf("%v %v %v", "->", "DEL", file.Path)
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
		log.Printf("%v %v %v", "<-", "MKDIR", file.Path)
		err = c.Mkdir(file.Path, file.Mode)
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.RemoteMkdir {
		log.Printf("%v %v %v", "->", "MKDIR", file.Path)
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
		log.Printf("%v %v %v", "<-", "SYMLINK", file.Path)
		err = c.Symlink(file.Symlink, file.Path)
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.RemoteMklink {
		log.Printf("%v %v %v", "->", "SYMLINK", file.Path)
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
		log.Printf("%v %v %v %v", "<-", "CHMOD", file.Path, file.Mode)
		err = c.Chmod(file.Path, file.Mode)
		if err != nil {
			return err
		}
	}

	for _, file := range syncplan.RemoteChmod {
		log.Printf("%v %v %v %v", "->", "CHMOD", file.Path, file.Mode)
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
		log.Printf("%v %v", "->", file.Path)
		stop := c.startProgressBar()
		err = c.SendFile(file.Path)
		stop()
		if err != nil {
			return err
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
			cmd, waiter, err := c.WaitFor("PUSH")
			if err != nil {
				return err
			}

			push := cmd.(*commands.Push)
			log.Printf("%v %v", "<-", push.Path)
			stop := c.startProgressBar()
			err = c.ReceiveFile(push, waiter)
			stop()
			if err != nil {
				return err
			}

			delete(paths, push.Path)
		}

	}

	return nil
}

func (c *Client) startProgressBar() func() {
	if log.ScreenOutput == nil || log.ScreenLevel > log.Notice || !progressbar.CanUse() {
		return func() {}
	}

	done := make(chan struct{})
	stop := func() {
		done <- struct{}{}
		<-done
	}

	go func() {
		for {
			select {
			case progress := <-c.Progress:
				progressbar.Draw(progress.Percent, progress.Eta)
			case <-done:
				progressbar.Reset()
				close(done)
				return
			}
		}
	}()

	return stop
}
