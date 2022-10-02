package filelist

import (
	"io/fs"
	"log"
	"unisync/config"
)

type SyncPlanBuilder struct {
	Config *config.Config
	Plan   *SyncPlan
}

func NewSyncPlanBuilder(config *config.Config) *SyncPlanBuilder {
	return &SyncPlanBuilder{Config: config}
}

func (b *SyncPlanBuilder) BuildSyncPlan(localList, remoteList FileList) *SyncPlan {
	b.Plan = NewSyncPlan()
	index := indexFileList(localList, remoteList)

	for _, lists := range index {
		b.compare2Way(lists.local, lists.remote)
	}

	b.Plan.Clean()
	return b.Plan
}

func (b *SyncPlanBuilder) compare2Way(local, remote *FileListItem) {
	if local == nil && remote == nil {
		log.Fatalln("local and remote cannot both be nil")
	}

	if remote == nil {
		if local.IsDir {
			b.Plan.Mkdir(false, local)
		} else {
			b.Plan.Sync(false, local)
		}

		return
	}

	if local == nil {
		if remote.IsDir {
			b.Plan.Mkdir(true, remote)
		} else {
			b.Plan.Sync(true, remote)
		}

		return
	}

	isSame := local.ModifiedAt == remote.ModifiedAt && local.Size == remote.Size
	isDirMismatch := local.IsDir != remote.IsDir

	if isDirMismatch {
		// TODO: figure this out

	} else if !local.IsDir && !isSame {
		if b.preferLocal(local, remote) {
			b.Plan.Sync(false, local)
		} else {
			b.Plan.Sync(true, remote)
		}

	} else if !b.compareModes(local, remote) {
		if b.preferLocal(local, remote) {
			b.Plan.Chmod(false, local)
		} else {
			b.Plan.Chmod(true, remote)
		}
	}

}

func (b *SyncPlanBuilder) preferLocal(local, remote *FileListItem) bool {
	switch b.Config.Prefer {
	case "newest":
		return local.ModifiedAt >= remote.ModifiedAt
	case "oldest":
		return local.ModifiedAt <= remote.ModifiedAt
	case "local":
		return true
	case "remote":
		return false
	default:
		log.Fatalln("config.prefer must be one of: newest, oldest, local, remote")
	}

	return true
}

func (b *SyncPlanBuilder) compareModes(local, remote *FileListItem) bool {
	localMode := local.Mode.Perm()
	remoteMode := remote.Mode.Perm()

	if localMode == 0 || remoteMode == 0 {
		return true
	}
	if local.IsDir != remote.IsDir {
		return true
	}

	var mask fs.FileMode
	if local.IsDir {
		mask = b.Config.Chmod.DirMask.Perm()
	} else {
		mask = b.Config.Chmod.Mask.Perm()
	}

	localMode = localMode & mask
	remoteMode = remoteMode & mask
	return localMode == remoteMode
}
