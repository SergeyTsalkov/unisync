package node

import (
	"log"
	"path/filepath"
)

const BufferSize = 1000000

func Path(basepath, path string) string {
	if basepath == "" {
		log.Fatalln("basepath is not set")
	}

	path = filepath.FromSlash(path)
	return filepath.Clean(filepath.Join(basepath, path))
}
