package commands

type Ok struct{}

func (c *Ok) CmdType() string {
  return "OK"
}

func (c *Ok) BodyLen() int {
  return 0
}
