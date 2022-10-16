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
var lastLineCount int

func Println(a ...any) error {
	_, err := io.WriteString(buffer, fmt.Sprintln(a...))
	return err
}

func Printf(format string, a ...any) error {
	_, err := io.WriteString(buffer, fmt.Sprintf(format, a...))
	return err
}

func Flush() error {
	if lastLineCount > 0 {
		err := clearLines(lastLineCount)
		if err != nil {
			return err
		}

		lastLineCount = 0
	}

	var err error
	lastLineCount, err = countLines()
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, buffer)
	if err != nil {
		return err
	}

	buffer.Reset()
	return nil
}

func countLines() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
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

func clearLines(lines int) error {
	ESC := 27
	clear := fmt.Sprintf("%c[1A%c[2K", ESC, ESC)
	_, err := io.WriteString(os.Stdout, strings.Repeat(clear, lines))
	return err
}
