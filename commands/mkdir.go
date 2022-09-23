package commands

type Mkdir struct {
	Dirs []*MkdirAction `json:"dirs"`
}

type MkdirAction struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

func (c *Mkdir) CmdType() string {
	return "MKDIR"
}
