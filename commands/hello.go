package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

type Hello struct {
	Basepath string `json:"basepath"`
}

func (c *Hello) Type() string {
	return "HELLO"
}

func (c *Hello) Encode() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%v %v\n", c.Type(), string(bytes))
}
