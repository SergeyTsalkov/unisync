package commands

import "unisync/filelist"

type Pull struct {
	Paths []string `json:"paths"`
}

func (c *Pull) CmdType() string {
	return "PULL"
}

func (c *Pull) BodyLen() int {
	return 0
}

func MakePull(items []*filelist.FileListItem) *Pull {
	if len(items) == 0 {
		return nil
	}

	pull := &Pull{
		Paths: make([]string, len(items)),
	}
	for i, item := range items {
		pull.Paths[i] = item.Path
	}

	return pull
}
