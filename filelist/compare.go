package filelist

import (
	"fmt"
	"path/filepath"
	"strings"
)

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

	plan.LocalMkdir = clean(plan.LocalMkdir, plan.LocalMkdir)
	plan.LocalMkdir = clean(plan.LocalMkdir, plan.Pull)

	plan.RemoteMkdir = clean(plan.RemoteMkdir, plan.RemoteMkdir)
	plan.RemoteMkdir = clean(plan.RemoteMkdir, plan.Push)

	return plan
}

func clean(list, list2 []*FileListItem) []*FileListItem {
	cleanlist := []*FileListItem{}

	for _, item := range list {
		isClean := true

		for _, item2 := range list2 {
			if HasPrefix(item2.Path, item.Path) {
				isClean = false
				break
			}
		}

		if isClean {
			cleanlist = append(cleanlist, item)
		}
	}

	return cleanlist
}

func HasPrefix(_long, _short string) bool {
	long := strings.Split(_long, string(filepath.Separator))
	short := strings.Split(_short, string(filepath.Separator))

	if len(long) <= len(short) {
		return false
	}

	for len(long) > 0 && len(short) > 0 {
		if long[0] != short[0] {
			return false
		}
		long = long[1:]
		short = short[1:]
	}

	return true
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
