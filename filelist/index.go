package filelist

type IndexedFileList map[string]*IndexedFileItem
type IndexedFileItem struct {
	local  *FileListItem
	remote *FileListItem
	cache  *FileListItem
}

func indexFileList(local, remote, cache FileList) IndexedFileList {
	index := make(IndexedFileList)

	for _, file := range local {
		if index[file.Path] == nil {
			index[file.Path] = &IndexedFileItem{}
		}

		index[file.Path].local = file
	}
	for _, file := range remote {
		if index[file.Path] == nil {
			index[file.Path] = &IndexedFileItem{}
		}

		index[file.Path].remote = file
	}
	for _, file := range cache {
		if index[file.Path] == nil {
			index[file.Path] = &IndexedFileItem{}
		}

		index[file.Path].cache = file
	}

	return index
}
