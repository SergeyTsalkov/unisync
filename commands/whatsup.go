package commands

// when a client sends HELLO, server responds with WHATSUP
type Whatsup struct {
	Basepath string `json:"basepath"`
}

func (c *Whatsup) CmdType() string {
	return "WHATSUP"
}
