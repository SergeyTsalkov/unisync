package filelist

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type FileListItem struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	IsDir      bool   `json:"is_dir"`
	ModifiedAt int64  `json:"modified_at"`
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
				item.ModifiedAt = info.ModTime().Unix()
			}

			list = append(list, item)
		}
	}

	return list, nil
}

func (list FileList) Encode() []byte {
	output, _ := json.MarshalIndent(list, "", "  ")
	return output
}
