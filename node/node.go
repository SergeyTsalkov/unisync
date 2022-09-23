package node

import (
	"log"
	"path/filepath"
)

func Path(basepath, path string) string {
	if basepath == "" {
		log.Fatalln("basepath is not set")
	}

	path = filepath.FromSlash(path)
	return filepath.Clean(filepath.Join(basepath, path))
}
