package filelist

import (
	"log"
	"unisync/config"
)

func Sync2Way(localList, remoteList FileList) *SyncPlan {
	plan := NewSyncPlan()
	index := indexFileList(localList, remoteList)

	for _, lists := range index {
		compare2Way(lists.local, lists.remote, plan)
	}

	plan.Clean()
	return plan
}

func compare2Way(local, remote *FileListItem, plan *SyncPlan) {
	if local == nil && remote == nil {
		log.Fatalln("local and remote cannot both be nil")
	}

	if remote == nil {
		if local.IsDir {
			plan.Mkdir(false, local)
		} else {
			plan.Sync(false, local, remote)
		}

		return
	}

	if local == nil {
		if remote.IsDir {
			plan.Mkdir(true, remote)
		} else {
			plan.Sync(true, remote, local)
		}

		return
	}

	isSame := local.ModifiedAt == remote.ModifiedAt && local.Size == remote.Size
	isDirMismatch := local.IsDir != remote.IsDir

	if isDirMismatch {
		// TODO: figure this out

	} else if !local.IsDir && !isSame {
		if preferLocal(local, remote) {
			plan.Sync(false, local, remote)
		} else {
			plan.Sync(true, remote, local)
		}

	} else if !compareModes(local, remote) {
		if preferLocal(local, remote) {
			plan.Chmod(false, local, remote)
		} else {
			plan.Chmod(true, remote, local)
		}
	}

}

func preferLocal(local, remote *FileListItem) bool {
	switch config.C.Prefer {
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

func compareModes(local, remote *FileListItem) bool {
	if local.Mode == 0 || remote.Mode == 0 {
		return true
	}
	if local.IsDir != remote.IsDir {
		return true
	}

	localMode := local.Mode.Perm()
	remoteMode := remote.Mode.Perm()

	if local.IsDir {
		localMode = localMode & config.C.Chmod.DirMask.Perm()
		remoteMode = remoteMode & config.C.Chmod.DirMask.Perm()
	} else {
		localMode = localMode & config.C.Chmod.Mask.Perm()
		remoteMode = remoteMode & config.C.Chmod.Mask.Perm()
	}

	return localMode == remoteMode
}
