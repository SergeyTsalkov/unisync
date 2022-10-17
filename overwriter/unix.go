// +build !windows

package overwriter

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func clearLines(lines int) error {
	ESC := 27
	clear := fmt.Sprintf("%c[1A%c[2K", ESC, ESC)
	_, err := io.WriteString(os.Stdout, strings.Repeat(clear, lines))
	return err
}
