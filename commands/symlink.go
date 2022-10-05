package commands

import (
	"io/fs"
	"unisync/filelist"
)

type Symlink struct {
	Links []*SymlinkAction `json:"links"`
}

type SymlinkAction struct {
	Path    string      `json:"path"`
	Symlink string      `json:"symlink"`
	Mode    fs.FileMode `json:"mode"`
}

func (c *Symlink) CmdType() string {
	return "SYMLINK"
}

func MakeSymlink(items []*filelist.FileListItem) *Symlink {
	if len(items) == 0 {
		return nil
	}

	symlink := &Symlink{
		Links: make([]*SymlinkAction, len(items)),
	}
	for i, item := range items {
		symlink.Links[i] = &SymlinkAction{Path: item.Path, Symlink: item.Symlink, Mode: item.Mode}
	}

	return symlink
}
