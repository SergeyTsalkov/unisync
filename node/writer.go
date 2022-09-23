package node

import (
	"fmt"
	"io"
	"strings"
	"unisync/commands"
)

type Writer struct {
	w     io.Writer
	Debug bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) SendCmdBuf(cmd commands.Command, buf []byte) error {
	err := w.SendString(commands.Encode(cmd))
	if err != nil {
		return err
	}

	if len(buf) > 0 {
		if w.Debug {
			fmt.Printf("-> [%v bytes]\n", len(buf))
		}

		_, err = w.w.Write(buf)
		if err != nil {
			return &DeepError{err}
		}
	}

	return nil
}

func (w *Writer) SendCmd(cmd commands.Command) error {
	return w.SendCmdBuf(cmd, nil)
}

func (w *Writer) SendString(str string) error {
	str = strings.TrimSpace(str)

	if w.Debug {
		fmt.Printf("-> %v\n", str)
	}

	_, err := io.WriteString(w.w, str+"\n")
	if err != nil {
		return &DeepError{err}
	}
	return nil
}

// func (w *Writer) SendStringf(format string, a ...any) error {
// 	return w.SendString(fmt.Sprintf(format, a...))
// }

func (w *Writer) SendErr(err error) error {
	return w.SendCmd(&commands.Error{
		Err: err.Error(),
	})
}

func (w *Writer) SendPathErr(path string, err error) error {
	return w.SendCmd(&commands.Error{
		Err:  err.Error(),
		Path: path,
	})
}
