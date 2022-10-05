package commands

import "unisync/filelist"

type Del struct {
	Paths []string `json:"paths"`
}

func (c *Del) CmdType() string {
	return "DEL"
}

func MakeDel(items []*filelist.FileListItem) *Del {
	if len(items) == 0 {
		return nil
	}

	del := &Del{
		Paths: make([]string, len(items)),
	}
	for i, item := range items {
		del.Paths[i] = item.Path
	}

	return del
}
