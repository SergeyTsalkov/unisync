package filelist

import (
	"io/fs"
	"path/filepath"
	"runtime"
)

type FileListItem struct {
	Path       string      `json:"path"`
	Size       int64       `json:"size"`
	IsDir      bool        `json:"is_dir"`
	ModifiedAt int64       `json:"modified_at"`
	Mode       fs.FileMode `json:"mode,omitempty"`
}

type FileList []*FileListItem

func Make(basepath string) (FileList, error) {
	list := FileList{}
	basepath = filepath.Clean(basepath)

	err := filepath.Walk(basepath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relpath, err := filepath.Rel(basepath, path)
		if err != nil {
			return err
		}
		if relpath == "." {
			return nil
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return nil
		}

		item := &FileListItem{
			Path:  filepath.ToSlash(relpath),
			IsDir: info.IsDir(),
		}
		if !item.IsDir {
			item.Size = info.Size()
			item.ModifiedAt = info.ModTime().Unix()
		}
		if runtime.GOOS != "windows" {
			item.Mode = info.Mode().Perm()
		}

		list = append(list, item)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}

// func (list FileList) Encode() []byte {
// 	output, _ := json.MarshalIndent(list, "", "  ")
// 	return output
// }

// func Parse(bytes []byte) (FileList, error) {
// 	list := FileList{}

// 	err := json.Unmarshal(bytes, &list)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid json: %w", err)
// 	}
// 	return list, nil
// }
