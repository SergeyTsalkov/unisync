package node

import (
	"log"
	"os"
	"time"
	"unisync/commands"
	"unisync/config"
	"unisync/filelist"
)

var receiving = map[string]*os.File{}

func ReceiveFile(filename string, cmd *commands.Push, buf []byte) (done bool, err error) {
	file := receiving[filename]
	done = !cmd.More

	// starting to receive a file
	if file == nil {
		mode := cmd.Mode

		// if we're the server, just trust the mode sent by client
		// if we're the client, filter given mode through ModeMask
		if !config.IsServer {
			mask := config.C.Chmod.Mask.Perm()
			mode = filelist.ModeMask(fileMode(filename), mode, mask)
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

// returns filemode, or default mode if file doesn't exist
// only works in client context
func fileMode(filename string) os.FileMode {
	if config.IsServer {
		log.Fatalln("fileMode() should not be called from server")
	}

	info, err := os.Lstat(filename)
	if err != nil {
		return config.C.Chmod.Local.Perm()
	}
	return info.Mode().Perm()
}
