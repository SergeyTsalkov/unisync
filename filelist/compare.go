package filelist

import "fmt"

type SyncPlan struct {
	Pull        []*FileListItem
	Push        []*FileListItem
	LocalMkdir  []*FileListItem
	RemoteMkdir []*FileListItem
	LocalDel    []*FileListItem
	RemoteDel   []*FileListItem
}

type IndexedFileList map[string]*FileListItem

func Compare(_localList, _remoteList FileList) *SyncPlan {
	plan := &SyncPlan{
		Pull:        []*FileListItem{},
		Push:        []*FileListItem{},
		LocalMkdir:  []*FileListItem{},
		RemoteMkdir: []*FileListItem{},
		LocalDel:    []*FileListItem{},
		RemoteDel:   []*FileListItem{},
	}

	localList := indexFileList(_localList)
	remoteList := indexFileList(_remoteList)

	for path, local := range localList {
		remote := remoteList[path]

		if remote == nil {
			if local.IsDir {
				plan.RemoteMkdir = append(plan.RemoteMkdir, local)
			} else {
				plan.Push = append(plan.Push, local)
			}
		} else {

			if local.IsDir && remote.IsDir {
				// both dirs, do nothing
			} else if local.ModifiedAt == remote.ModifiedAt && local.Size == remote.Size {
				// already synced, do nothing
			} else if local.ModifiedAt >= remote.ModifiedAt {
				plan.Push = append(plan.Push, local)
			} else {
				plan.Pull = append(plan.Pull, local)
			}
		}
	}

	for path, remote := range remoteList {
		local := localList[path]

		if local == nil {
			if remote.IsDir {
				plan.LocalMkdir = append(plan.LocalMkdir, remote)
			} else {
				plan.Pull = append(plan.Pull, remote)
			}
		}
	}

	return plan
}

func (p *SyncPlan) Show() {
	if len(p.Pull) > 0 {
		fmt.Println("Pull files:")

		for _, file := range p.Pull {
			fmt.Println(file.Path)
		}
	}

	if len(p.Push) > 0 {
		fmt.Println("Push files:")

		for _, file := range p.Push {
			fmt.Println(file.Path)
		}
	}

	if len(p.LocalMkdir) > 0 {
		fmt.Println("Local Mkdir:")

		for _, file := range p.LocalMkdir {
			fmt.Println(file.Path)
		}
	}

	if len(p.RemoteMkdir) > 0 {
		fmt.Println("Remote Mkdir:")

		for _, file := range p.RemoteMkdir {
			fmt.Println(file.Path)
		}
	}

}

func indexFileList(list FileList) IndexedFileList {
	indexedList := make(IndexedFileList)

	for _, file := range list {
		indexedList[file.Path] = file
	}

	return indexedList
}
