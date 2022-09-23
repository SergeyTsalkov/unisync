package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

type Push struct {
	Path       string `json:"path"`
	IsDir      bool   `json:"is_dir"`
	ModifiedAt int64  `json:"modified_at"`
	Length     int64  `json:"length"`
	More       bool   `json:"more"`
}

func (c *Push) Type() string {
	return "PUSH"
}

func (c *Push) Encode() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%v %v\n", c.Type(), string(bytes))
}
