package commands

import (
	"unisync/filelist"
)

type Symlink struct {
	Links []*SymlinkAction `json:"links"`
}

type SymlinkAction struct {
	Path    string `json:"path"`
	Symlink string `json:"symlink"`
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
		symlink.Links[i] = &SymlinkAction{Path: item.Path, Symlink: item.Symlink}
	}

	return symlink
}
