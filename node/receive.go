package node

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"
	"unisync/commands"
)

var receiving = map[string]*os.File{}

func (n *Node) ReceiveFile(cmd *commands.Push, buf []byte) (bool, error) {
	filename := n.Path(cmd.Path)
	file := receiving[filename]
	done := !cmd.More

	// starting to receive a file
	if file == nil {

		if info, err := os.Lstat(filename); err == nil {
			mode := info.Mode()
			if mode.IsDir() {
				// we are trying to write to a file that's actually a dir
				// this should never happen unless the sync step really messed up
				return done, fmt.Errorf("can not RECEIVE %v: is a directory", cmd.Path)
			}
			if mode&fs.ModeSymlink != 0 {
				// this can happen if the symlink was changed to a file
				// remove the symlink before creating the file
				err = os.Remove(filename)
				if err != nil {
					return done, err
				}
			}
		}

		// if this file was a symlink before, then we've just deleted the symlink
		// therefore, we need to find out the file's mode again
		// this will be the default mode set in the config if the file doesn't exist
		mode := n.fileMode(filename)

		// if we got a mode of 0 (the sending side is Windows), just keep the mode we have
		if cmd.Mode != 0 {
			mask := n.Config.Chmod.Mask.Perm()
			mode = modeMask(mode, cmd.Mode, mask)
		}

		var err error
		file, err = os.Create(filename)
		if err != nil {
			return done, err
		}
		receiving[filename] = file

		err = file.Chmod(mode)
		if err != nil {
			return done, err
		}
	}

	_, err := file.Write(buf)

	if done {
		delete(receiving, filename)
		if err := file.Close(); err != nil {
			log.Fatalln("err closing file", filename, ":", err)
		}
		if err := os.Chtimes(filename, time.Now(), time.Unix(cmd.ModifiedAt, 0)); err != nil {
			log.Fatalln("err setting mtime", filename, ":", err)
		}
	}

	return done, err
}
