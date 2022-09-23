package commands

type Push struct {
	Path       string `json:"path"`
	ModifiedAt int64  `json:"modified_at"`
	Length     int64  `json:"length"`
	More       bool   `json:"more"`
}

func (c *Push) CmdType() string {
	return "PUSH"
}
