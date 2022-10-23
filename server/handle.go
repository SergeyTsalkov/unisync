package server

import (
	"fmt"
	"os"
	"sync"
	"unisync/commands"
	"unisync/filelist"
	"unisync/node"
)

func (s *Server) handle(packet *node.Packet) error {
	cmd := packet.Command

	if cmd.CmdType() != "HELLO" && !s.loggedIn {
		return fmt.Errorf("must log in with HELLO first")
	}

	switch cmd.CmdType() {
	case "HELLO":
		return s.handleHELLO(cmd)
	case "REQLIST":
		return s.handleREQLIST(cmd)
	case "MKDIR":
		return s.handleMKDIR(cmd)
	case "SYMLINK":
		return s.handleSYMLINK(cmd)
	case "CHMOD":
		return s.handleCHMOD(cmd)
	case "DEL":
		return s.handleDEL(cmd)
	case "PULL":
		return s.handlePULL(cmd)
	case "PUSH":
		return s.handlePUSH(cmd, packet.Waiter)
	default:
		return fmt.Errorf("invalid command")
	}

	return nil
}

func (s *Server) handleHELLO(cmd commands.Command) error {
	hello := cmd.(*commands.Hello)

	s.Config = hello.Config
	err := s.SetBasepath(s.Config.Remote)
	if err != nil {
		return fmt.Errorf("Unable to set basepath: %w", err)
	}

	err = s.SetTmpdir(s.Config.RemoteTmpdir)
	if err != nil {
		return fmt.Errorf("Unable to set tmpdir: %w", err)
	}

	whatsup := &commands.Whatsup{s.GetBasepath()}
	err = s.SendCmd(whatsup)
	if err != nil {
		return err
	}

	s.loggedIn = true
	return nil
}

func (s *Server) handleREQLIST(cmd commands.Command) error {
	s.Watcher.Ready()

	reqlist := cmd.(*commands.ReqList)
	list, err := filelist.Make(s.Path(reqlist.Path), s.Config.Ignore)
	if err != nil {
		return err
	}

	reply := &commands.ResList{list}
	return s.SendCmd(reply)
}

func (s *Server) handleMKDIR(cmd commands.Command) error {
	mkdir := cmd.(*commands.Mkdir)

	for _, dir := range mkdir.Dirs {
		err := s.Mkdir(dir.Path, dir.Mode)
		if err != nil {
			return err
		}
	}

	return s.SendCmd(&commands.Ok{})
}

func (s *Server) handleSYMLINK(cmd commands.Command) error {
	symlink := cmd.(*commands.Symlink)

	for _, link := range symlink.Links {
		err := s.Symlink(link.Symlink, link.Path)
		if err != nil {
			return err
		}
	}

	return s.SendCmd(&commands.Ok{})
}

func (s *Server) handleCHMOD(cmd commands.Command) error {
	chmod := cmd.(*commands.Chmod)

	for _, action := range chmod.Actions {
		err := s.Chmod(action.Path, action.Mode)
		if err != nil {
			return err
		}
	}

	return s.SendCmd(&commands.Ok{})
}

func (s *Server) handleDEL(cmd commands.Command) error {
	del := cmd.(*commands.Del)

	for _, path := range del.Paths {
		err := os.Remove(s.Path(path))
		if err != nil {
			return err
		}
	}

	return s.SendCmd(&commands.Ok{})
}

func (s *Server) handlePULL(cmd commands.Command) error {
	pull := cmd.(*commands.Pull)

	if len(pull.Paths) == 0 {
		return fmt.Errorf("PULL command must specify at least 1 path")
	}

	for _, path := range pull.Paths {
		err := s.SendFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) handlePUSH(cmd commands.Command, waiter *sync.WaitGroup) error {
	push := cmd.(*commands.Push)
	return s.ReceiveFile(push, waiter)
}
