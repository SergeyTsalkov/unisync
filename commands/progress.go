package commands

type Progress struct {
  Percent int `json:"percent"`
  Eta     int `json:"eta"`
}

func (c *Progress) CmdType() string {
  return "PROGRESS"
}

func (c *Progress) BodyLen() int {
  return 0
}
