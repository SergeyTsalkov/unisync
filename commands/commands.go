package commands

import (
	"encoding/json"
	"fmt"
	"log"
)

// type CommandType interface {
// 	Hello | ReqList | ResList | Pull | Push | Mkdir
// }

type Command interface {
	CmdType() string
}

func Encode(c Command) string {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}

	return c.CmdType() + " " + string(bytes)
}

func Parse(str string, ptr Command) error {
	if str == "" {
		return nil
	}

	err := json.Unmarshal([]byte(str), ptr)
	if err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}
