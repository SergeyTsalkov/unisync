package commands

type Hello struct {
	Basepath string `json:"basepath"`
}

func (c *Hello) CmdType() string {
	return "HELLO"
}
