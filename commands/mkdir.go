package commands

import (
	"io/fs"
	"unisync/filelist"
)

type Mkdir struct {
	Dirs []*MkdirAction `json:"dirs"`
}

type MkdirAction struct {
	Path string      `json:"path"`
	Mode fs.FileMode `json:"mode"`
}

func (c *Mkdir) CmdType() string {
	return "MKDIR"
}

func (c *Mkdir) BodyLen() int {
	return 0
}

func MakeMkdir(items []*filelist.FileListItem) *Mkdir {
	if len(items) == 0 {
		return nil
	}

	mkdir := &Mkdir{
		Dirs: make([]*MkdirAction, len(items)),
	}
	for i, item := range items {
		mkdir.Dirs[i] = &MkdirAction{Path: item.Path, Mode: item.Mode}
	}

	return mkdir
}
