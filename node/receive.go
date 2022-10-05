package node

import (
	"log"
	"os"
	"time"
	"unisync/commands"
)

var receiving = map[string]*os.File{}

func (n *Node) ReceiveFile(cmd *commands.Push, buf []byte) (done bool, err error) {
	filename := n.Path(cmd.Path)
	file := receiving[filename]
	done = !cmd.More

	// starting to receive a file
	if file == nil {
		mode := n.fileMode(filename)

		// if we got a mode of 0 (the sending side is Windows), just keep the mode we have
		if cmd.Mode != 0 {
			mask := n.Config.Chmod.Mask.Perm()
			mode = modeMask(mode, cmd.Mode, mask)
		}

		file, err = os.Create(filename)
		if err != nil {
			return
		}
		receiving[filename] = file

		err = file.Chmod(mode)
		if err != nil {
			return
		}
	}

	_, err = file.Write(buf)

	if !cmd.More {
		delete(receiving, filename)
		if err := file.Close(); err != nil {
			log.Fatalln("err closing file", filename, ":", err)
		}
		if err := os.Chtimes(filename, time.Now(), time.Unix(cmd.ModifiedAt, 0)); err != nil {
			log.Fatalln("err setting mtime", filename, ":", err)
		}
	}

	return
}
