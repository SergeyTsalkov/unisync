package node

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
	"unisync/commands"
)

func (n *Node) ReceiveFile(cmd *commands.Push, buf []byte) error {
	fullpath := n.Path(cmd.Path)

	if n.receiveFile == nil {
		n.openReceiveFile(cmd)
	} else if n.receiveFullpath != fullpath {
		return fmt.Errorf("can't RECEIVE %v: %v is still open", fullpath, n.receiveFullpath)
	}

	_, err := n.receiveFile.Write(buf)
	if err != nil {
		return err
	}

	if !cmd.More {
		err = n.CloseReceiveFile(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) openReceiveFile(cmd *commands.Push) error {
	fullpath := n.Path(cmd.Path)

	if n.receiveFile != nil {
		return fmt.Errorf("can't RECEIVE %v: %v is still open", fullpath, n.receiveFullpath)
	}

	var perm fs.FileMode
	if info, err := os.Lstat(fullpath); err == nil {
		if info.Mode().IsDir() {
			return fmt.Errorf("can't RECEIVE %v: is a directory", fullpath)
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
	if cmd.Mode.Perm() != 0 {
		perm = n.FileMask(perm, cmd.Mode)
	}

	dir, _ := filepath.Split(fullpath)
	tempfullpath := filepath.Join(dir, ".unisync-tmp")

	var err error
	n.receiveFile, err = os.Create(tempfullpath)
	if err != nil {
		return err
	}

	n.receiveFullpath = fullpath
	err = n.receiveFile.Chmod(perm)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) CloseReceiveFile(cmd *commands.Push) error {
	file := n.receiveFile
	if file == nil {
		return nil
	}

	// always try to remove the tmpfile; this will fail if we've already moved it
	defer os.Remove(file.Name())

	err := file.Close()
	if err != nil {
		return err
	}

	if cmd != nil {
		err = os.Rename(file.Name(), n.receiveFullpath)
		if err != nil {
			return err
		}

		err = os.Chtimes(n.receiveFullpath, time.Now(), time.Unix(cmd.ModifiedAt, 0))
		if err != nil {
			return err
		}
	}

	n.receiveFile = nil
	n.receiveFullpath = ""
	return nil
}
