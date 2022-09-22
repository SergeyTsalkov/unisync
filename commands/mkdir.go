package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

type Mkdir struct {
	Dirs []*MkdirAction `json:"dirs"`
}

type MkdirAction struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

func (c *Mkdir) Type() string {
	return "MKDIR"
}

func (c *Mkdir) Encode() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%v %v\n", c.Type(), string(bytes))
}
