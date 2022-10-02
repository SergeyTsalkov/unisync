package commands

import "io/fs"

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
