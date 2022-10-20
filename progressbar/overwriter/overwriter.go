package overwriter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var buffer = &bytes.Buffer{}
var lastLineCount = -1

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func TerminalWidth() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}
	return width, nil
}

func Println(a ...any) error {
	_, err := io.WriteString(buffer, fmt.Sprintln(a...))
	return err
}

func Printf(format string, a ...any) error {
	_, err := io.WriteString(buffer, fmt.Sprintf(format, a...))
	return err
}

func Flush() error {
	err := clear()
	if err != nil {
		return err
	}

	lastLineCount, err = countLines()
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, buffer)
	if err != nil {
		return err
	}

	return nil
}

func Reset() (err error) {
	err = clear()
	if err != nil {
		return
	}

	lastLineCount = -1
	buffer.Reset()
	return
}

func clear() error {
	if lastLineCount >= 0 {
		return clearLines(lastLineCount)
	}
	return nil
}

func countLines() (int, error) {
	width, err := TerminalWidth()
	if err != nil {
		return 0, err
	}

	str := string(buffer.Bytes())
	count := strings.Count(str, "\n")

	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if len(line) > width {
			count += (len(line) - 1) / width
		}
	}

	return count, nil
}
