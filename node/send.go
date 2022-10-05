package node

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"unisync/commands"
)

func (n *Node) SendFile(path string) error {
	file, err := os.Open(n.Path(path))
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatalln("err closing file", path, ":", err)
		}
	}()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("can not SEND %v: is a directory", path)
	}

	mode := info.Mode().Perm()
	if runtime.GOOS == "windows" {
		mode = 0
	}

	more := true
	offset := int64(0)
	for more {
		len, err := file.ReadAt(Buffer, offset)
		if err == io.EOF {
			more = false
			err = nil
		} else if err != nil {
			return &DeepError{err}
		}

		push := &commands.Push{
			Path:       path,
			Length:     int64(len),
			ModifiedAt: info.ModTime().Unix(),
			Mode:       mode,
			More:       more,
		}

		err = n.SendCmdBuf(push, Buffer[0:len])
		if err != nil {
			return err
		}

		offset += int64(len)
	}

	return nil
}
