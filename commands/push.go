package commands

import "io/fs"

type Push struct {
	Path       string      `json:"path"`
	ModifiedAt int64       `json:"modified_at"`
	Mode       fs.FileMode `json:"mode"`
	Length     int64       `json:"length"`
	More       bool        `json:"more"`
}

func (c *Push) CmdType() string {
	return "PUSH"
}
