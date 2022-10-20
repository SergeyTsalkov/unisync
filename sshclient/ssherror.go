package sshclient

import (
	"fmt"
	"strings"
)

type ssherror struct {
	err    error
	output []byte
}

func (e *ssherror) Error() string {
	output := strings.TrimSpace(string(e.output))

	if len(output) > 0 {
		return fmt.Sprintf("%v (%v)", output, e.err)
	}
	return e.err.Error()
}
