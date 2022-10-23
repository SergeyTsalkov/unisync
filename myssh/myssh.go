package myssh

import (
	"fmt"
	"strings"
)

type SshError struct {
	Err    error
	Output []byte
}

func (e *SshError) Error() string {
	output := strings.TrimSpace(string(e.Output))

	if len(output) > 0 {
		return fmt.Sprintf("%v (%v)", output, e.Err)
	}
	return e.Err.Error()
}
