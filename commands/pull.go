package commands

type Pull struct {
	Paths []string `json:"paths"`
}

func (c *Pull) CmdType() string {
	return "PULL"
}
