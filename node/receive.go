package node

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unisync/commands"
	"unisync/progresswriter"
)

func (n *Node) ReceiveFile(push *commands.Push, waiter *sync.WaitGroup) error {
	path := push.Path
	fullpath := n.Path(path)
	mtime := time.Unix(push.ModifiedAt, 0)
	file, tempfullpath, err := n.openReceiveFile(fullpath, push.Mode.Perm(), push.Size)
	if err != nil {
		return err
	}
	defer os.Remove(tempfullpath)

	for {
		if bodyLen := int64(push.BodyLen()); bodyLen > 0 {
			bytesCopied, err := io.CopyBuffer(file, io.LimitReader(n.In, bodyLen), n.Buffer)
			if err != nil {
				return err
			}
			if bytesCopied != bodyLen {
				return fmt.Errorf("size mismatch: %v (expected %v bytes in this chunk but got %v)", path, bodyLen, bytesCopied)
			}
			waiter.Done()
		}

		if push.More {
			var cmd commands.Command
			cmd, waiter, err = n.WaitFor("PUSH")
			if err != nil {
				return err
			}
			push = cmd.(*commands.Push)

			if path != push.Path {
				return fmt.Errorf("PUSH: was expecting file %v but got %v", path, push.Path)
			}

		} else {
			break
		}
	}

	err = file.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tempfullpath, fullpath)
	if err != nil {
		return err
	}
	err = os.Chtimes(fullpath, time.Now(), mtime)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) openReceiveFile(fullpath string, receivedPerm fs.FileMode, size int64) (io.WriteCloser, string, error) {
	var perm fs.FileMode
	if info, err := os.Lstat(fullpath); err == nil {
		if info.Mode().IsDir() {
			return nil, "", fmt.Errorf("can't RECEIVE %v: is a directory", fullpath)
		}
		perm = info.Mode().Perm()
	} else {
		if n.IsServer {
			perm = n.Config.ChmodRemote.Perm()
		} else {
			perm = n.Config.ChmodLocal.Perm()
		}
	}

	// if we got a mode of 0 (the sending side is Windows), just keep the mode we have
	if receivedPerm != 0 {
		perm = n.FileMask(perm, receivedPerm)
	}

	var tmpdir string
	if n.GetTmpdir() != "" {
		tmpdir = n.GetTmpdir()
	} else {
		tmpdir, _ = filepath.Split(fullpath)
	}

	file, err := os.CreateTemp(tmpdir, ".tmp-unisync-*.tmp")
	if err != nil {
		return nil, "", err
	}

	err = file.Chmod(perm)
	if err != nil {
		return nil, "", err
	}
	return progresswriter.New(file, size, n.Progress), file.Name(), nil
}
