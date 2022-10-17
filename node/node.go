package node

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"unisync/config"
	"unisync/progresswriter"
	"unisync/watcher"
)

var Buffer = make([]byte, 1000000)

type Node struct {
	In       *bufio.Reader
	Out      io.Writer
	IsServer bool
	Config   *config.Config
	Watcher  *watcher.Watcher
	basepath string
	Progress chan progresswriter.Progress

	// most incoming packets go into MainC
	MainC chan *Packet

	// packets can be diverted to	SideC if they match sideCmatch
	sideCmatch map[string]struct{}
	SideC      chan *Packet

	// packet reader errors will go into Errors channel
	Errors chan error
}

func New(in io.Reader, out io.Writer) *Node {
	node := &Node{
		In:         bufio.NewReader(in),
		Out:        out,
		MainC:      make(chan *Packet),
		SideC:      make(chan *Packet),
		sideCmatch: map[string]struct{}{},
		Errors:     make(chan error),
		Watcher:    watcher.New(),
		Progress:   make(chan progresswriter.Progress),
	}

	go node.InputReader()
	return node
}

func (n *Node) SetBasepath(basepath string) error {
	if n.basepath != "" {
		return fmt.Errorf("basepath is already set")
	}

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

	err = n.Watcher.Start(basepath, n.Config.Ignore)
	if err != nil {
		return err
	}

	n.basepath = basepath
	return nil
}
func (n *Node) GetBasepath() string {
	return n.basepath
}

func (n *Node) Path(path string) string {
	if n.basepath == "" {
		panic("basepath is not set")
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
