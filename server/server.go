package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unisync/commands"
	"unisync/filelist"
	"unisync/node"
)

type Server struct {
	loggedIn bool
	*node.Node
}

func New(in io.Reader, out io.Writer) *Server {
	node := node.New(in, out)
	node.IsServer = true
	return &Server{Node: node}
}

func (s *Server) Run() error {
	// if a network error prevented a file from being fully transmitted, delete the tmpfile
	defer s.CloseReceiveFile(nil)

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

	if cmd != "HELLO" && !s.loggedIn {
		return fmt.Errorf("must log in with HELLO first")
	}

	switch cmd {
	case "HELLO":
		return s.handleHELLO(json)
	case "REQLIST":
		return s.handleREQLIST(json)
	case "MKDIR":
		return s.handleMKDIR(json)
	case "SYMLINK":
		return s.handleSYMLINK(json)
	case "CHMOD":
		return s.handleCHMOD(json)
	case "DEL":
		return s.handleDEL(json)
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

	s.Config = cmd.Config
	s.Config.Validate()

	err = s.SetBasepath(s.Config.Remote)
	if err != nil {
		return fmt.Errorf("Unable to set basepath: %w", err)
	}

	whatsup := &commands.Whatsup{s.GetBasepath()}
	err = s.SendCmd(whatsup)
	if err != nil {
		return err
	}

	s.loggedIn = true
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
		err := s.Mkdir(dir.Path, dir.Mode)
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

func (s *Server) handleSYMLINK(json string) error {
	cmd := &commands.Symlink{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	for _, link := range cmd.Links {
		err := s.Symlink(link.Symlink, link.Path)
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

func (s *Server) handleCHMOD(json string) error {
	cmd := &commands.Chmod{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	for _, action := range cmd.Actions {
		err := s.Chmod(action.Path, action.Mode)
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

func (s *Server) handleDEL(json string) error {
	cmd := &commands.Del{}
	err := commands.Parse(json, cmd)
	if err != nil {
		return err
	}

	for _, path := range cmd.Paths {
		err := os.Remove(s.Path(path))
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

	err = s.ReceiveFile(cmd, buf)
	if err != nil {
		return err
	}

	return nil
}
