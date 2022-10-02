package commands

import "os"

type Mkdir struct {
	Dirs []*MkdirAction `json:"dirs"`
}

type MkdirAction struct {
	Path string      `json:"path"`
	Mode os.FileMode `json:"mode"`
}

func (c *Mkdir) CmdType() string {
	return "MKDIR"
}
