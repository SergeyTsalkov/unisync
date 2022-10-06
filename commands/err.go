package commands

type Error struct {
  Err  string `json:"err"`
  Path string `json:"path,omitempty"`
}

func (c *Error) CmdType() string {
  return "ERR"
}
