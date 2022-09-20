package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

type ReqList struct {
	Path string `json:"path"`
}

func (c *ReqList) Type() string {
	return "REQLIST"
}

func (c *ReqList) Encode() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%v %v\n", c.Type(), string(bytes))
}
