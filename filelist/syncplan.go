package filelist

import (
	"fmt"
	"sort"
)

type SyncPlan struct {
	Pull        []*FileListItem
	Push        []*FileListItem
	LocalMkdir  []*FileListItem
	RemoteMkdir []*FileListItem
	LocalChmod  []*FileListItem
	RemoteChmod []*FileListItem
	LocalDel    []*FileListItem
	RemoteDel   []*FileListItem
}

func NewSyncPlan() *SyncPlan {
	plan := &SyncPlan{
		Pull:        []*FileListItem{},
		Push:        []*FileListItem{},
		LocalMkdir:  []*FileListItem{},
		RemoteMkdir: []*FileListItem{},
		LocalChmod:  []*FileListItem{},
		RemoteChmod: []*FileListItem{},

		// LocalDel:    []*FileListItem{},
		// RemoteDel:   []*FileListItem{},
	}

	return plan
}

func (plan *SyncPlan) Clean() {
	sort.Slice(plan.LocalMkdir, func(i, j int) bool { return len(plan.LocalMkdir[i].Path) < len(plan.LocalMkdir[j].Path) })
	sort.Slice(plan.RemoteMkdir, func(i, j int) bool { return len(plan.RemoteMkdir[i].Path) < len(plan.RemoteMkdir[j].Path) })
}

func (plan *SyncPlan) Mkdir(isLocal bool, item *FileListItem) {
	if isLocal {
		plan.LocalMkdir = append(plan.LocalMkdir, item)
	} else {
		plan.RemoteMkdir = append(plan.RemoteMkdir, item)
	}
}

func (plan *SyncPlan) Chmod(isLocal bool, item *FileListItem) {
	if isLocal {
		plan.LocalChmod = append(plan.LocalChmod, item)
	} else {
		plan.RemoteChmod = append(plan.RemoteChmod, item)
	}
}

func (plan *SyncPlan) Sync(toLocal bool, item *FileListItem) {
	if toLocal {
		plan.Pull = append(plan.Pull, item)
	} else {
		plan.Push = append(plan.Push, item)
	}
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
