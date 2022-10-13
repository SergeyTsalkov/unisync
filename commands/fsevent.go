package commands

type FsEvent struct{}

func (c *FsEvent) CmdType() string {
	return "FSEVENT"
}

func (c *FsEvent) BodyLen() int {
	return 0
}
