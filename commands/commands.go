package commands

import (
	"encoding/json"
	"fmt"
)

type Pull struct {
	Paths []string `json:"paths"`
}

type Push struct {
	Path       string `json:"path"`
	IsDir      bool   `json:"is_dir"`
	ModifiedAt int64  `json:"modified_at"`
	Length     int64  `json:"length"`
	More       bool   `json:"more"`
}

type CommandType interface {
	Hello | ReqList | ResList | Pull | Push | Mkdir
}

type Command interface {
	Encode() string
	Type() string
}

func ParseCommand[T CommandType](str string, ptr *T) error {
	if str == "" {
		return nil
	}

	//typ := fmt.Printf("%T", *ptr)

	err := json.Unmarshal([]byte(str), ptr)
	if err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}
