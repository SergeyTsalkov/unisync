package filelist

import (
	"fmt"
	"sort"
)

type SyncPlan struct {
	PullFile    []*FileListItem
	PushFile    []*FileListItem
	LocalMkdir  []*FileListItem
	RemoteMkdir []*FileListItem
	LocalChmod  []*FileListItem
	RemoteChmod []*FileListItem
	LocalDel    []*FileListItem
	RemoteDel   []*FileListItem
}

func NewSyncPlan() *SyncPlan {
	plan := &SyncPlan{
		PullFile:    []*FileListItem{},
		PushFile:    []*FileListItem{},
		LocalMkdir:  []*FileListItem{},
		RemoteMkdir: []*FileListItem{},
		LocalChmod:  []*FileListItem{},
		RemoteChmod: []*FileListItem{},
		LocalDel:    []*FileListItem{},
		RemoteDel:   []*FileListItem{},
	}

	return plan
}

func (plan *SyncPlan) Clean() {
	sort.Slice(plan.LocalMkdir, func(i, j int) bool { return len(plan.LocalMkdir[i].Path) < len(plan.LocalMkdir[j].Path) })
	sort.Slice(plan.RemoteMkdir, func(i, j int) bool { return len(plan.RemoteMkdir[i].Path) < len(plan.RemoteMkdir[j].Path) })
	sort.Slice(plan.LocalDel, func(i, j int) bool { return len(plan.LocalDel[i].Path) > len(plan.LocalDel[j].Path) })
	sort.Slice(plan.RemoteDel, func(i, j int) bool { return len(plan.RemoteDel[i].Path) > len(plan.RemoteDel[j].Path) })
}

func (plan *SyncPlan) ChmodRemote(item *FileListItem) {
	plan.RemoteChmod = append(plan.RemoteChmod, item)
}
func (plan *SyncPlan) ChmodLocal(item *FileListItem) {
	plan.LocalChmod = append(plan.LocalChmod, item)
}
func (plan *SyncPlan) DelLocal(item *FileListItem) {
	plan.LocalDel = append(plan.LocalDel, item)
}
func (plan *SyncPlan) DelRemote(item *FileListItem) {
	plan.RemoteDel = append(plan.RemoteDel, item)
}

func (plan *SyncPlan) Push(item *FileListItem) {
	if item.IsDir {
		plan.RemoteMkdir = append(plan.RemoteMkdir, item)
	} else {
		plan.PushFile = append(plan.PushFile, item)
	}
}

func (plan *SyncPlan) Pull(item *FileListItem) {
	if item.IsDir {
		plan.LocalMkdir = append(plan.LocalMkdir, item)
	} else {
		plan.PullFile = append(plan.PullFile, item)
	}
}

func (plan *SyncPlan) IsSynced() bool {
	return len(plan.PullFile) == 0 &&
		len(plan.PushFile) == 0 &&
		len(plan.LocalMkdir) == 0 &&
		len(plan.RemoteMkdir) == 0 &&
		len(plan.LocalChmod) == 0 &&
		len(plan.RemoteChmod) == 0 &&
		len(plan.LocalDel) == 0 &&
		len(plan.RemoteDel) == 0
}

func (p *SyncPlan) Show() {
	if len(p.PullFile) > 0 {
		fmt.Println("Pull files:")

		for _, file := range p.PullFile {
			fmt.Println(file.Path)
		}
	}

	if len(p.PushFile) > 0 {
		fmt.Println("Push files:")

		for _, file := range p.PushFile {
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
