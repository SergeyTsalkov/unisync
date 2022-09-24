package node

import (
	"fmt"
	"io"
	"log"
	"os"
	"unisync/commands"
)

func SendFile(output *Writer, path, filename string) error {
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("can not SEND %v: is a directory", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatalln("err closing file", filename, ":", err)
		}
	}()

	more := true

	for more {
		n, err := file.Read(Buffer)
		if err == io.EOF {
			more = false
		} else if err != nil {
			return &DeepError{err}
		}

		push := &commands.Push{
			Path:       path,
			Length:     int64(n),
			ModifiedAt: info.ModTime().Unix(),
			More:       more,
		}

		err = output.SendCmdBuf(push, Buffer[0:n])
		if err != nil {
			return err
		}
	}

	return nil
}
