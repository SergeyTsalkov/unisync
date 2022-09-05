package filelist

import (
	"os"
	"path/filepath"
	"time"
)

type FileListItem struct {
	Path       string
	Size       int64
	IsDir      bool
	ModifiedAt time.Time
}

type FileList []FileListItem

func Make(basepath string) (FileList, error) {
	list := make(FileList, 0)
	basepath = filepath.Clean(basepath)

	for dirs := []string{basepath}; len(dirs) > 0; {
		dir := dirs[0]
		dirs = dirs[1:]

		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			path := filepath.Join(dir, file.Name())
			relpath, _ := filepath.Rel(basepath, path)
			if relpath == "" {
				continue
			}

			item := FileListItem{}
			item.Path = relpath
			item.IsDir = file.IsDir()

			if item.IsDir {
				dirs = append(dirs, path)
			}

			if info, err := file.Info(); err == nil {
				item.Size = info.Size()
				item.ModifiedAt = info.ModTime()
			}

			list = append(list, item)
		}
	}

	return list, nil
}
