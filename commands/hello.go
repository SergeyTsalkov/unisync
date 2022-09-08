package commands

import (
	"encoding/json"
	"fmt"
)

type HelloCommand struct {
	Basepath string `json:"basepath"`
}

func ParseHelloCommand(str string) (c *HelloCommand, err error) {
	c = &HelloCommand{}
	err = json.Unmarshal([]byte(str), c)
	if err != nil {
		err = fmt.Errorf("invalid json: %w", err)
	}
	return
}
