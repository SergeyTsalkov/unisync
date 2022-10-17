//go:build !windows
// +build !windows

package overwriter

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func clearLines(lines int) error {
	ESC := fmt.Sprintf("%c", 27)
	MOVELEFT := ESC + "[0G"
	MOVEUP := ESC + "[1A"
	WIPELINE := ESC + "[2K"

	_, err := io.WriteString(os.Stdout, MOVELEFT+WIPELINE+strings.Repeat(MOVEUP+WIPELINE, lines))
	return err
}
