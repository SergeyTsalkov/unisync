package commands

type ReqList struct {
	Path string `json:"path"`
}

func (c *ReqList) CmdType() string {
	return "REQLIST"
}
