package commands

import "unisync/config"

type Hello struct {
	Config *config.Config `json:"config"`
}

func (c *Hello) CmdType() string {
	return "HELLO"
}

func (c *Hello) BodyLen() int {
	return 0
}
