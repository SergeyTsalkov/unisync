package filelist

import (
	"encoding/json"
	"fmt"
	"io/fs"
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
	list := FileList{}
	basepath = filepath.Clean(basepath)

	err := filepath.Walk(basepath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relpath, _ := filepath.Rel(basepath, path)
		if relpath == "." {
			return nil
		}

		item := FileListItem{
			Path:       relpath,
			IsDir:      info.IsDir(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime().Unix(),
		}

		list = append(list, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (list FileList) Encode() []byte {
	output, _ := json.MarshalIndent(list, "", "  ")
	return output
}

func Parse(bytes []byte) (*FileList, error) {
	list := &FileList{}

	err := json.Unmarshal(bytes, list)
	if err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	return list, nil
}
