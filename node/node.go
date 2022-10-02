package node

import (
	"bufio"
	"io"
	"log"
	"path/filepath"
)

var Buffer = make([]byte, 1000000)

type Node struct {
	Basepath string
	In       *bufio.Reader
	Out      io.Writer
	Debug    bool
}

func (n *Node) Path(path string) string {
	if n.Basepath == "" {
		log.Fatalln("basepath is not set")
	}

	path = filepath.FromSlash(path)
	path = filepath.Join(n.Basepath, path)
	path = filepath.Clean(path)
	return path
}
