package commands

import "unisync/filelist"

type ResList struct {
	FileList filelist.FileList `json:"filelist"`
}

func (c *ResList) CmdType() string {
	return "RESLIST"
}

func (c *ResList) BodyLen() int {
	return 0
}
