package server

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"unisync/filelist"
)

type Server struct {
	in  io.Reader
	out io.Writer
}

func New(in io.Reader, out io.Writer) *Server {
	s := &Server{}
	s.in = in
	s.out = out
	return s
}

func (server *Server) Run() {
	for {
		reader := bufio.NewReader(server.in)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		server.handle(line)
	}
}

func (server *Server) Err(err string) {
	fmt.Fprintf(server.out, "ERR %v\n", err)
}

func (server *Server) handle(line string) {
	words := strings.Fields(line)
	cmd := strings.ToUpper(words[0])
	args := words[1:]

	switch cmd {
	case "REQINFO":
		server.handleREQINFO(args)
	default:
		server.Err("invalid command")
	}
}

func (server *Server) handleREQINFO(args []string) {
	list, err := filelist.Make("/Users/sergey/test")

	if err != nil {
		server.Err(err.Error())
		return
	}

	json := list.Encode()
	fmt.Fprintf(server.out, "RESINFO %v\n", len(json))
	server.out.Write(json)
}
