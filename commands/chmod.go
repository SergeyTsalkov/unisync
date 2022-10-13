package commands

import (
  "io/fs"
  "unisync/filelist"
)

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

func (c *Chmod) BodyLen() int {
  return 0
}

func MakeChmod(items []*filelist.FileListItem) *Chmod {
  if len(items) == 0 {
    return nil
  }

  chmod := &Chmod{
    Actions: make([]*ChmodAction, len(items)),
  }
  for i, item := range items {
    chmod.Actions[i] = &ChmodAction{Path: item.Path, Mode: item.Mode}
  }

  return chmod
}
