package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"unisync/filelist"
)

type Server struct {
	version  int
	in       io.Reader
	out      io.Writer
	basepath string
}

func New(in io.Reader, out io.Writer) *Server {
	return &Server{
		in:  in,
		out: out,
	}
}

func (server *Server) mkPath(path string) string {
	if server.basepath == "" {
		return ""
	}

	path = filepath.Join(server.basepath, path)
	path = filepath.Clean(path)
	return path
}

func (server *Server) Run() error {
	reader := bufio.NewReader(server.in)
	var deepError *DeepError

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)

		err = server.handle(line)
		if err != nil {
			if errors.As(err, &deepError) {
				return err
			} else {
				fmt.Fprintf(server.out, "ERR %v\n", err)
			}
		}
	}
}

func (server *Server) handle(line string) error {
	words := strings.Fields(line)
	cmd := strings.ToUpper(words[0])
	args := words[1:]

	if cmd != "HELLO" && server.version == 0 {
		return fmt.Errorf("must log in with HELLO first")
	}

	switch cmd {
	case "HELLO":
		return server.handleHELLO(args)
	case "REQINFO":
		return server.handleREQINFO(args)
	default:
		return fmt.Errorf("invalid command")
	}

	return nil
}

func (server *Server) handleHELLO(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("HELLO must have 2 arguments")
	}

	basepath := strings.Join(args[1:], "")
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

	server.version, _ = strconv.Atoi(args[0])
	server.basepath = basepath
	fmt.Fprintf(server.out, "OK\n")
	return nil
}

func (server *Server) handleREQINFO(args []string) error {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	list, err := filelist.Make(server.mkPath(path))

	if err != nil {
		return err
	}

	json := list.Encode()

	_, err = fmt.Fprintf(server.out, "RESINFO %v\n", len(json))
	if err != nil {
		return &DeepError{err}
	}

	_, err = server.out.Write(json)
	if err != nil {
		return &DeepError{err}
	}

	return nil
}
