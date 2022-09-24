package node

import (
	"log"
	"os"
	"path/filepath"
	"time"
	"unisync/commands"
)

var receiving = map[string]*os.File{}

func ReceiveFile(filename string, cmd *commands.Push, buf []byte) (done bool, err error) {
	file := receiving[filename]
	done = !cmd.More

	// starting to receive a file
	if file == nil {
		dir, _ := filepath.Split(filename)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return
		}

		file, err = os.Create(filename)
		if err != nil {
			return
		}
		receiving[filename] = file
	}

	if !cmd.More {
		defer func() {
			delete(receiving, filename)
			err := file.Close()
			if err != nil {
				log.Fatalln("err closing file", filename, ":", err)
			}
			err = os.Chtimes(filename, time.Now(), time.Unix(cmd.ModifiedAt, 0))
			if err != nil {
				log.Fatalln("err setting mtime", filename, ":", err)
			}
		}()
	}

	_, err = file.Write(buf)
	return
}
