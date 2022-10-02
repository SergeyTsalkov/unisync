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
	"unisync/config"
	"unisync/filelist"
	"unisync/node"
)

type Server struct {
	*node.Node
}

func New(in io.Reader, out io.Writer) *Server {
	node := &node.Node{
		In:  bufio.NewReader(in),
		Out: out,
	}

	return &Server{node}
}

func (s *Server) Run() error {
	config.IsServer = true

	for {
		line, err := s.In.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		err = s.handle(line)
		if err != nil {
			if errors.Is(err, node.ErrDeep) {
				return err
			} else {
				s.SendErr(err)
			}
		}
	}
}

func (s *Server) handle(line string) error {
	words := strings.Fields(line)
	cmd := strings.ToUpper(words[0])
	json := strings.TrimSpace(strings.TrimPrefix(line, cmd))

	if cmd != "HELLO" && s.Basepath == "" {
		return fmt.Errorf("must log in with HELLO first")
	}

	switch cmd {
	case "HELLO":
		return s.handleHELLO(json)
	case "REQLIST":
		return s.handleREQLIST(json)
	case "MKDIR":
		return s.handleMKDIR(json)
	case "PULL":
		return s.handlePULL(json)
	case "PUSH":
		return s.handlePUSH(json)
	default:
		return fmt.Errorf("invalid command")
	}

	return nil
}

func (s *Server) handleHELLO(json string) error {
	cmd := &commands.Hello{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	basepath := cmd.Basepath
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

	s.Basepath = basepath

	err = s.SendString("OK")
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) handleREQLIST(json string) error {
	cmd := &commands.ReqList{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	list, err := filelist.Make(s.Path(cmd.Path))
	if err != nil {
		return err
	}

	reply := &commands.ResList{list}
	err = s.SendCmd(reply)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) handleMKDIR(json string) error {
	cmd := &commands.Mkdir{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	for _, dir := range cmd.Dirs {
		fullpath := s.Path(dir.Path)
		err := os.Mkdir(fullpath, dir.Mode)
		if err != nil {
			return err
		}
	}

	err = s.SendString("OK")
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handlePULL(json string) error {
	cmd := &commands.Pull{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	if len(cmd.Paths) == 0 {
		return fmt.Errorf("PULL command must specify at least 1 path")
	}

	for _, path := range cmd.Paths {
		err = s.SendFile(path)
		if err != nil {
			if errors.Is(err, node.ErrDeep) {
				return err
			} else {
				s.SendPathErr(path, err)
			}
		}
	}

	return nil
}

func (s *Server) handlePUSH(json string) error {
	cmd := &commands.Push{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	buf := make([]byte, cmd.Length)
	_, err = io.ReadAtLeast(s.In, buf, len(buf))
	if err != nil {
		return err
	}

	_, err = s.ReceiveFile(cmd, buf)
	if err != nil {
		return err
	}

	return nil
}
