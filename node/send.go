package node

import (
	"fmt"
	"io"
	"log"
	"os"
	"unisync/commands"
)

func (n *Node) SendFile(path string) error {
	filename := n.Path(path)
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
			Mode:       info.Mode().Perm(),
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
