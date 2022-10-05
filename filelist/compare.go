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

func (b *SyncPlanBuilder) BuildSyncPlan(localList, remoteList, cacheList FileList) *SyncPlan {
	b.Plan = NewSyncPlan()
	index := indexFileList(localList, remoteList, cacheList)

	for _, lists := range index {
		b.compare(lists.local, lists.remote, lists.cache)
	}

	b.Plan.Clean()
	return b.Plan
}

func (b *SyncPlanBuilder) compare(local, remote, cache *FileListItem) {
	plan := b.Plan

	if local == nil && remote == nil {
		// if local and remote don't exist, the file was deleted on both sides and is only known from the cache
		// in this case, there's nothing to sync

		return
	}

	if local != nil && remote != nil && local.IsDir != remote.IsDir {
		// if one side is a directory and the other side isn't, keep the directory
		if local.IsDir {
			plan.DelRemote(remote)
		} else {
			plan.DelLocal(local)
		}

	} else if !itemsMatch(local, remote) {
		// side can be dirty if cache exists and mismatches (because side was either changed or deleted)
		// side can be dirty if cache doesn't exist, but the file does (because we have no cache, or file was recently created)
		isLocalDirty := !itemsMatch(local, cache)
		isRemoteDirty := !itemsMatch(remote, cache)

		if isLocalDirty && !isRemoteDirty {
			// if there is a cache, one side matches it and the other doesn't (the mismatched side might have been deleted, too)
			// if there is no cache, the dirty side exists and the clean side doesn't

			if local != nil {
				plan.Push(local)
			} else {
				plan.DelRemote(remote)
			}

		} else if isRemoteDirty && !isLocalDirty {
			if remote != nil {
				plan.Pull(remote)
			} else {
				plan.DelLocal(local)
			}

		} else {
			// if there is a cache, both sides might have been changed (or one could have been changed and the other deleted)
			// if there is no cache, both sides exist and we need to pick a winner

			if b.preferLocal(local, remote) {
				plan.Push(local)
			} else {
				plan.Pull(remote)
			}
		}
	} else if !b.itemModesMatch(local, remote) {
		// mode can only be dirty if cache exists, and the relevant parts of mode don't match
		// if one side is Windows and has no mode, we say that modes always "match" and are never dirty
		isLocalModeDirty := !b.itemModesMatch(local, cache)
		isRemoteModeDirty := !b.itemModesMatch(remote, cache)

		if isLocalModeDirty && !isRemoteModeDirty {
			plan.ChmodRemote(local)

		} else if isRemoteModeDirty && !isLocalModeDirty {
			plan.ChmodLocal(remote)

		} else {
			if b.preferLocal(local, remote) {
				plan.ChmodRemote(local)
			} else {
				plan.ChmodLocal(remote)
			}
		}
	}

}

func (b *SyncPlanBuilder) preferLocal(local, remote *FileListItem) bool {
	if local == nil {
		return false
	}
	if remote == nil {
		return true
	}

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

func (b *SyncPlanBuilder) itemModesMatch(local, remote *FileListItem) bool {
	if local == nil || remote == nil {
		// if one side doesn't exist, let's say they match -- can't sync modes anyway
		return true
	}
	if local.Symlink != "" || remote.Symlink != "" {
		// if this is a symlink, let's say modes match -- can't really sync symlink modes anyway
		return true
	}

	localMode := local.Mode.Perm()
	remoteMode := remote.Mode.Perm()

	if localMode == 0 || remoteMode == 0 {
		// if one side is Windows and has no modes, let's say they match -- can't sync modes anyway
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

func itemsMatch(item, item2 *FileListItem) bool {
	if item == nil && item2 == nil {
		return true
	}
	if item == nil || item2 == nil {
		return false
	}

	if item.IsDir || item2.IsDir {
		return item.IsDir == item2.IsDir
	}
	if item.Symlink != "" || item2.Symlink != "" {
		return item.Symlink == item2.Symlink
	}

	return item.Size == item2.Size && item.ModifiedAt == item2.ModifiedAt
}
