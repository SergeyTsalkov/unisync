package node

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"unisync/commands"
)

func (n *Node) SendFile(path string) error {
	filename := n.Path(path)
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}

	mode := info.Mode()
	if mode.IsDir() {
		return fmt.Errorf("can not SEND %v: is a directory", path)
	}
	if mode&fs.ModeSymlink != 0 {
		return fmt.Errorf("can not SEND %v: is a symlink", path)
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic("error closing file: " + err.Error())
		}
	}()

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
			Length:     len,
			ModifiedAt: info.ModTime().Unix(),
			Mode:       mode.Perm(),
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
