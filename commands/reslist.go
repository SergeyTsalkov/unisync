package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

type ResList struct {
	Length int64 `json:"length"`
}

func (c *ResList) Type() string {
	return "RESLIST"
}

func (c *ResList) Encode() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return fmt.Sprintf("%v %v\n", c.Type(), string(bytes))
}
