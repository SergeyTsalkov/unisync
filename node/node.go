package node

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"unisync/config"
	"unisync/done"
	"unisync/progresswriter"
	"unisync/watcher"
)

type Node struct {
	In        *bufio.Reader
	Out       io.Writer
	IsServer  bool
	Config    *config.Config
	basepath  string
	Progress  chan progresswriter.Progress
	writeLock *sync.Mutex
	tmpdir    string

	// watches for filesystem changes
	// started by SetBasepath()
	// can be stopped with Watcher.Stop()
	Watcher *watcher.Watcher

	// buffer used to send and receive files
	Buffer []byte

	// most incoming packets go into MainC
	// packets can be diverted to	SideC if they match sideCmatch
	// if packet reader encounters an error, it writes it to SetDone()
	// and closes MainC and SideC
	MainC      chan *Packet
	SideC      chan *Packet
	sideCmatch map[string]struct{}

	// announces to every goroutine when it's time to exit
	// similar to Context, except it holds an error
	// when SetDone(error) is called, every DoneC() channel will
	// push the event and IsDone() will start returning that error
	SetDone func(error)
	IsDone  func() error
	DoneC   func() chan error
}

func New(in io.Reader, out io.Writer) *Node {
	if in == nil || out == nil {
		panic("Node: in and out can't be nil")
	}

	node := &Node{
		In:         bufio.NewReader(in),
		Out:        out,
		Buffer:     make([]byte, 1000000),
		MainC:      make(chan *Packet),
		SideC:      make(chan *Packet),
		sideCmatch: map[string]struct{}{},
		Watcher:    watcher.New(),
		Progress:   make(chan progresswriter.Progress),
		writeLock:  &sync.Mutex{},
	}

	node.SetDone, node.IsDone, node.DoneC = done.New()
	go node.InputReader()
	return node
}

func (n *Node) SetTmpdir(tmpdir string) error {
	if tmpdir != "" {
		var err error
		tmpdir, err = config.ResolvePath(tmpdir)
		if err != nil {
			return err
		}
		if !config.IsDir(tmpdir) {
			return fmt.Errorf("%v is not a directory", tmpdir)
		}
	}

	n.tmpdir = tmpdir
	return nil
}
func (n *Node) GetTmpdir() string {
	return n.tmpdir
}

func (n *Node) SetBasepath(basepath string) error {
	if n.basepath != "" {
		return fmt.Errorf("basepath is already set")
	}

	var err error
	basepath, err = config.ResolvePath(basepath)
	if err != nil {
		return err
	}

	if !config.IsDir(basepath) {
		return fmt.Errorf("%v is not a directory", basepath)
	}

	var watch string
	if n.IsServer {
		watch = n.Config.WatchRemote
	} else {
		watch = n.Config.WatchLocal
	}

	n.Watcher.PollFreq = n.Config.PollFreq
	if watch == "1" {
		err = n.Watcher.Start(basepath, n.Config.Ignore, false)
		if err != nil {
			return fmt.Errorf("Unable to start filesystem monitoring: %w", err)
		}
	} else if watch == "poll" {
		err = n.Watcher.Start(basepath, n.Config.Ignore, true)
		if err != nil {
			return fmt.Errorf("Unable to start filesystem monitoring: %w", err)
		}
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
