package commands

type ReqList struct {
	Path string `json:"path"`
}

func (c *ReqList) CmdType() string {
	return "REQLIST"
}

func (c *ReqList) BodyLen() int {
	return 0
}
