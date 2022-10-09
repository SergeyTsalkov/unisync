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
	In       *bufio.Reader
	Out      io.Writer
	Debug    bool
	IsServer bool
	Config   *config.Config
	basepath string

	receiveFile     *os.File
	receiveFullpath string
}

func New(in io.Reader, out io.Writer) *Node {
	return &Node{In: bufio.NewReader(in), Out: out}
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
func (n *Node) FileMask(baseMode, newMode os.FileMode) os.FileMode {
	return modeMask(baseMode, newMode, n.Config.ChmodMask.Perm())
}
func (n *Node) DirMask(baseMode, newMode os.FileMode) os.FileMode {
	return modeMask(baseMode, newMode, n.Config.ChmodDirMask.Perm())
}

func (n *Node) Mkdir(path string, mode fs.FileMode) error {
	fullpath := n.Path(path)

	var baseMode fs.FileMode
	if n.IsServer {
		baseMode = n.Config.ChmodRemoteDir.Perm()
	} else {
		baseMode = n.Config.ChmodLocalDir.Perm()
	}

	mode = n.DirMask(baseMode, mode)
	return os.Mkdir(fullpath, mode)
}

func (n *Node) Chmod(path string, mode fs.FileMode) error {
	filename := n.Path(path)
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}

	baseMode := info.Mode().Perm()
	if info.IsDir() {
		mode = n.DirMask(baseMode, mode)
	} else {
		mode = n.FileMask(baseMode, mode)
	}

	return os.Chmod(filename, mode)
}

func (n *Node) Symlink(old, new string) error {
	new = n.Path(new)
	os.Remove(new)
	return os.Symlink(old, new)
}
