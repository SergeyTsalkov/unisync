package filelist

type IndexedFileList map[string]*IndexedFileItem
type IndexedFileItem struct {
	local  *FileListItem
	remote *FileListItem
}

func indexFileList(local, remote FileList) IndexedFileList {
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

	return index
}
