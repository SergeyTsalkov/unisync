package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unisync/commands"
	"unisync/filelist"
	"unisync/node"
)

type Server struct {
	version  int
	in       *bufio.Reader
	out      *node.Writer
	basepath string
}

func New(in io.Reader, out io.Writer) *Server {
	return &Server{
		in:  bufio.NewReader(in),
		out: node.NewWriter(out),
	}
}

func (server *Server) path(path string) string {
	return node.Path(server.basepath, path)
}

func (server *Server) Run() error {
	for {
		line, err := server.in.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		err = server.handle(line)
		if err != nil {
			if errors.Is(err, node.ErrDeep) {
				return err
			} else {
				server.out.SendErr(err)
			}
		}
	}
}

func (server *Server) handle(line string) error {
	words := strings.Fields(line)
	cmd := strings.ToUpper(words[0])
	json := strings.TrimSpace(strings.TrimPrefix(line, cmd))

	if cmd != "HELLO" && server.version == 0 {
		return fmt.Errorf("must log in with HELLO first")
	}

	switch cmd {
	case "HELLO":
		return server.handleHELLO(json)
	case "REQLIST":
		return server.handleREQLIST(json)
	case "MKDIR":
		return server.handleMKDIR(json)
	case "PULL":
		return server.handlePULL(json)
	case "PUSH":
		return server.handlePUSH(json)
	default:
		return fmt.Errorf("invalid command")
	}

	return nil
}

func (server *Server) handleHELLO(json string) error {
	args := &commands.Hello{}
	err := commands.Parse(json, args)
	if err != nil {
		return err
	}

	basepath := args.Basepath
	basepath = filepath.Clean(basepath)

	if !filepath.IsAbs(basepath) {
		return fmt.Errorf("path must be absolute")
	}

	info, err := os.Lstat(basepath)
	if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	server.version = 1
	server.basepath = basepath

	err = server.out.SendString("OK")
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) handleREQLIST(json string) error {
	args := &commands.ReqList{}
	err := commands.Parse(json, args)
	if err != nil {
		return err
	}

	list, err := filelist.Make(server.path(args.Path))

	if err != nil {
		return err
	}

	reply := &commands.ResList{list}
	err = server.out.SendCmd(reply)
	if err != nil {
		return err
	}

	return nil
}

func (server *Server) handleMKDIR(json string) error {
	args := &commands.Mkdir{}
	err := commands.Parse(json, args)
	if err != nil {
		return err
	}

	for _, dir := range args.Dirs {
		fullpath := server.path(dir.Path)
		err := os.MkdirAll(fullpath, 0755)
		if err != nil {
			return err
		}
	}

	err = server.out.SendString("OK")
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) handlePULL(json string) error {
	args := &commands.Pull{}
	err := commands.Parse(json, args)
	if err != nil {
		return err
	}

	if len(args.Paths) == 0 {
		return fmt.Errorf("PULL command must specify at least 1 path")
	}

	for _, path := range args.Paths {
		err = node.SendFile(server.out, path, server.path(path))
		if err != nil {
			if errors.Is(err, node.ErrDeep) {
				return err
			} else {
				server.out.SendPathErr(path, err)
			}
		}
	}

	return nil
}

func (server *Server) handlePUSH(json string) error {
	args := &commands.Push{}
	err := commands.Parse(json, args)
	if err != nil {
		return err
	}

	buf := make([]byte, args.Length)
	_, err = io.ReadAtLeast(server.in, buf, len(buf))
	if err != nil {
		return err
	}

	_, err = node.ReceiveFile(server.path(args.Path), args, buf)
	if err != nil {
		return err
	}

	return nil
}
