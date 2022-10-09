package filelist

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"unisync/gitignore"
)

type FileListItem struct {
	Path       string      `json:"path"`
	Size       int64       `json:"size,omitempty"`
	ModifiedAt int64       `json:"modified_at,omitempty"`
	Symlink    string      `json:"symlink,omitempty"`
	IsDir      bool        `json:"is_dir,omitempty"`
	Mode       fs.FileMode `json:"mode,omitempty"`
}

type FileList []*FileListItem

func Make(basepath string, ignore []string) (FileList, error) {
	list := FileList{}
	basepath = filepath.Clean(basepath)

	err := filepath.Walk(basepath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		mode := info.Mode()
		relpath, err := filepath.Rel(basepath, path)
		if err != nil {
			return err
		}
		if relpath == "." {
			return nil
		}

		relpath = filepath.ToSlash(relpath)
		if gitignore.MatchAny(ignore, relpath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		item := &FileListItem{Path: relpath}
		if info.IsDir() {
			item.IsDir = true

		} else if mode&fs.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}

			item.Symlink = link

		} else if mode.IsRegular() {
			item.Size = info.Size()
			item.ModifiedAt = info.ModTime().Unix()

		} else {
			return nil
		}

		if runtime.GOOS != "windows" {
			item.Mode = mode.Perm()
		}

		list = append(list, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (list FileList) Encode() string {
	output, _ := json.MarshalIndent(list, "", "  ")
	return string(output)
}
