//go:build windows
// +build windows

package overwriter

import (
	"io"
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

func clearLines(lines int) error {
	handle := windows.Handle(os.Stdout.Fd())
	info := &windows.ConsoleScreenBufferInfo{}
	err := windows.GetConsoleScreenBufferInfo(handle, info)
	if err != nil {
		return err
	}
	width := int(info.Size.X)

	for i := 0; i <= lines; i++ {
		newPos := windows.Coord{X: 0, Y: info.CursorPosition.Y - int16(i)}
		windows.SetConsoleCursorPosition(handle, newPos)
		io.WriteString(os.Stdout, strings.Repeat(" ", width-1))
		windows.SetConsoleCursorPosition(handle, newPos)
	}

	return nil
}
