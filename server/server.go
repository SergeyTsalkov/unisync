package server

import (
	"fmt"
	"io"
	"unisync/commands"
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
	go s.monitorProgress()

	// this is safe because this main thread is the only one that pushes to s.Progress (via receive.go)
	defer close(s.Progress)

	for {
		select {
		case packet, ok := <-s.MainC:
			if !ok {
				if err := s.IsDone(); err != nil {
					return err
				} else {
					return fmt.Errorf("connection closed")
				}
			}

			err := s.handle(packet)
			if err != nil {
				s.SendErr(err)
				return err
			}

		case <-s.Watcher.C:
			err := s.handleWatch()
			if err != nil {
				s.SendErr(err)
				return err
			}

		case err := <-s.DoneC():
			return err
		}
	}
}

func (s *Server) handleWatch() error {
	return s.SendCmd(&commands.FsEvent{})
}

// separate goroutine
func (s *Server) monitorProgress() {
	var err error
	for progress := range s.Progress {
		err = s.SendCmd(&commands.Progress{progress.Percent, progress.Eta})
		if err != nil {
			break
		}
	}

	if err != nil {
		s.SetDone(err)
	}
}
