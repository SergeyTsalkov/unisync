package commands

import (
	"encoding/json"
	"fmt"
	"strings"
)

// type CommandType interface {
// 	Hello | ReqList | ResList | Pull | Push | Mkdir
// }

type Command interface {
	CmdType() string
	BodyLen() int
}

func Encode(c Command) string {
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	return c.CmdType() + " " + string(bytes)
}

// func Parse(str string, ptr Command) error {
// 	if str == "" {
// 		return nil
// 	}

// 	err := json.Unmarshal([]byte(str), ptr)
// 	if err != nil {
// 		return fmt.Errorf("invalid json: %w", err)
// 	}
// 	return nil
// }

func Parse(str string) (cmd Command, err error) {
	str = strings.TrimSpace(str)
	word, jsonString, _ := strings.Cut(str, " ")
	word = strings.TrimSpace(strings.ToUpper(word))
	jsonString = strings.TrimSpace(jsonString)

	switch word {
	case "CHMOD":
		cmd = &Chmod{}
	case "DEL":
		cmd = &Del{}
	case "ERR":
		cmd = &Error{}
	case "FSEVENT":
		cmd = &FsEvent{}
	case "HELLO":
		cmd = &Hello{}
	case "MKDIR":
		cmd = &Mkdir{}
	case "OK":
		cmd = &Ok{}
	case "PROGRESS":
		cmd = &Progress{}
	case "PULL":
		cmd = &Pull{}
	case "PUSH":
		cmd = &Push{}
	case "REQLIST":
		cmd = &ReqList{}
	case "RESLIST":
		cmd = &ResList{}
	case "SYMLINK":
		cmd = &Symlink{}
	case "WHATSUP":
		cmd = &Whatsup{}
	default:
		err = fmt.Errorf("invalid command %v", word)
	}

	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(jsonString), cmd)
	return
}
