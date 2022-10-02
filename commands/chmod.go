package commands

import "io/fs"

type Chmod struct {
  Actions []*ChmodAction `json:"actions"`
}

type ChmodAction struct {
  Path string      `json:"path"`
  Mode fs.FileMode `json:"mode"`
}

func (c *Chmod) CmdType() string {
  return "CHMOD"
}
