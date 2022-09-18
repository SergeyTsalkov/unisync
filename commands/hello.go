package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

type HelloCommand struct {
	Basepath string `json:"basepath"`
}

func (c *HelloCommand) Encode() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("HELLO %v\n", string(bytes))
}
