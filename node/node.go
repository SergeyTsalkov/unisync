package node

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"unisync/config"
)

var Buffer = make([]byte, 1000000)

type Node struct {
	basepath string
	In       *bufio.Reader
	Out      io.Writer
	Debug    bool
	IsServer bool
	Config   *config.Config
}

func (n *Node) SetBasepath(basepath string) error {
	var err error
	basepath, err = filepath.Abs(basepath)
	if err != nil {
		return err
	}

	info, err := os.Lstat(basepath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%v is not a directory", basepath)
	}
	n.basepath = basepath
	return nil
}
func (n *Node) GetBasepath() string {
	return n.basepath
}

func (n *Node) Path(path string) string {
	if n.basepath == "" {
		log.Fatalln("basepath is not set")
	}

	path = filepath.FromSlash(path)
	path = filepath.Join(n.basepath, path)
	path = filepath.Clean(path)
	return path
}

func modeMask(baseMode, newMode, mask os.FileMode) os.FileMode {
	mode := newMode&mask | baseMode&(^mask)
	return mode.Perm()
}

// returns filemode, or default mode if file doesn't exist
func (n *Node) fileMode(filename string) fs.FileMode {
	info, err := os.Lstat(filename)
	if err != nil {
		if n.IsServer {
			return n.Config.Chmod.Remote.Perm()
		} else {
			return n.Config.Chmod.Local.Perm()
		}
	}
	return info.Mode().Perm()
}

func (n *Node) Mkdir(path string, mode fs.FileMode) error {
	fullpath := n.Path(path)

	var baseMode fs.FileMode
	mask := n.Config.Chmod.DirMask.Perm()
	if n.IsServer {
		baseMode = n.Config.Chmod.RemoteDir.Perm()
	} else {
		baseMode = n.Config.Chmod.LocalDir.Perm()
	}

	mode = modeMask(baseMode, mode, mask)
	return os.Mkdir(fullpath, mode)
}

func (n *Node) Chmod(path string, mode fs.FileMode) error {
	filename := n.Path(path)
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}

	baseMode := info.Mode().Perm()
	var mask fs.FileMode
	if info.IsDir() {
		mask = n.Config.Chmod.DirMask.Perm()
	} else {
		mask = n.Config.Chmod.Mask.Perm()
	}

	mode = modeMask(baseMode, mode, mask)
	return os.Chmod(filename, mode)
}

func (n *Node) Symlink(old, new string) error {
	new = n.Path(new)
	os.Remove(new)
	return os.Symlink(old, new)
}
