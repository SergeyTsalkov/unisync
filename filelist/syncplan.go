package filelist

import (
	"fmt"
	"io/fs"
	"sort"
	"unisync/config"
)

type SyncPlanItem struct {
	Path string
	Mode fs.FileMode
}

type SyncPlan struct {
	Pull        []*SyncPlanItem
	Push        []*SyncPlanItem
	LocalMkdir  []*SyncPlanItem
	RemoteMkdir []*SyncPlanItem
	LocalChmod  []*SyncPlanItem
	RemoteChmod []*SyncPlanItem
	LocalDel    []*SyncPlanItem
	RemoteDel   []*SyncPlanItem
}

func NewSyncPlan() *SyncPlan {
	plan := &SyncPlan{
		Pull:        []*SyncPlanItem{},
		Push:        []*SyncPlanItem{},
		LocalMkdir:  []*SyncPlanItem{},
		RemoteMkdir: []*SyncPlanItem{},
		LocalChmod:  []*SyncPlanItem{},
		RemoteChmod: []*SyncPlanItem{},

		// LocalDel:    []*SyncPlanItem{},
		// RemoteDel:   []*SyncPlanItem{},
	}

	return plan
}

func (plan *SyncPlan) Clean() {
	sort.Slice(plan.LocalMkdir, func(i, j int) bool { return len(plan.LocalMkdir[i].Path) < len(plan.LocalMkdir[j].Path) })
	sort.Slice(plan.RemoteMkdir, func(i, j int) bool { return len(plan.RemoteMkdir[i].Path) < len(plan.RemoteMkdir[j].Path) })
}

func (plan *SyncPlan) Mkdir(isLocal bool, item *FileListItem) {
	mask := config.C.Chmod.DirMask.Perm()
	var baseMode fs.FileMode
	if isLocal {
		baseMode = config.C.Chmod.LocalDir.Perm()
	} else {
		baseMode = config.C.Chmod.RemoteDir.Perm()
	}
	syncItem := &SyncPlanItem{Path: item.Path, Mode: ModeMask(baseMode, item.Mode, mask)}

	if isLocal {
		plan.LocalMkdir = append(plan.LocalMkdir, syncItem)
	} else {
		plan.RemoteMkdir = append(plan.RemoteMkdir, syncItem)
	}
}

// srcItem's mode is used as source to change item's mode
func (plan *SyncPlan) Chmod(isLocal bool, srcItem, item *FileListItem) {
	var mask fs.FileMode
	if item.IsDir {
		mask = config.C.Chmod.DirMask.Perm()
	} else {
		mask = config.C.Chmod.Mask.Perm()
	}
	syncItem := &SyncPlanItem{Path: item.Path, Mode: ModeMask(item.Mode, srcItem.Mode, mask)}

	if isLocal {
		plan.LocalChmod = append(plan.LocalChmod, syncItem)
	} else {
		plan.RemoteChmod = append(plan.RemoteChmod, syncItem)
	}
}

func (plan *SyncPlan) Sync(toLocal bool, srcItem, item *FileListItem) {
	var baseMode fs.FileMode
	if item != nil && item.Mode != 0 {
		baseMode = item.Mode
	} else if toLocal {
		baseMode = config.C.Chmod.Local.Perm()
	} else {
		baseMode = config.C.Chmod.Remote.Perm()
	}

	mode := baseMode
	if srcItem.Mode != 0 {
		mask := config.C.Chmod.Mask.Perm()
		mode = ModeMask(baseMode, srcItem.Mode, mask)
	}

	syncItem := &SyncPlanItem{Path: srcItem.Path, Mode: mode}
	if toLocal {
		plan.Pull = append(plan.Pull, syncItem)
	} else {
		plan.Push = append(plan.Push, syncItem)
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
