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

	jsonC "encoding/json"
)

type Server struct {
	version    int
	in         io.Reader
	out        io.Writer
	basepath   string
	buffersize int
}

func New(in io.Reader, out io.Writer) *Server {
	return &Server{
		in:         in,
		out:        out,
		buffersize: 1000000,
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
		if len(line) == 0 {
			continue
		}

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
	json := strings.TrimSpace(strings.TrimPrefix(line, cmd))

	if cmd != "HELLO" && server.version == 0 {
		return fmt.Errorf("must log in with HELLO first")
	}

	switch cmd {
	case "HELLO":
		return server.handleHELLO(json)
	case "REQLIST":
		return server.handleREQLIST(json)
	case "PULL":
		return server.handlePULL(json)
	default:
		return fmt.Errorf("invalid command")
	}

	return nil
}

func (server *Server) handleHELLO(json string) error {
	args := &commands.Hello{}
	err := commands.ParseCommand(json, args)
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
	fmt.Fprintf(server.out, "OK\n")
	return nil
}

func (server *Server) handleREQLIST(json string) error {
	args := &commands.ReqList{}
	err := commands.ParseCommand(json, args)
	if err != nil {
		return err
	}

	list, err := filelist.Make(server.mkPath(args.Path))

	if err != nil {
		return err
	}

	output := list.Encode()
	reply := &commands.ResList{int64(len(output))}

	_, err = io.WriteString(server.out, reply.Encode())
	if err != nil {
		return &DeepError{err}
	}

	_, err = server.out.Write(output)
	if err != nil {
		return &DeepError{err}
	}

	return nil
}

func (server *Server) handlePULL(json string) error {
	args := &commands.Pull{}
	err := commands.ParseCommand(json, args)
	if err != nil {
		return err
	}

	if len(args.Paths) == 0 {
		return fmt.Errorf("PULL command must specify at least 1 path")
	}

	for _, path := range args.Paths {
		server.pushFile(path)
	}

	return nil
}

func (server *Server) pushFile(path string) error {
	filename := server.mkPath(path)
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, server.buffersize)
	more := true

	for more {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			more = false
		}

		push := &commands.Push{
			Path:       path,
			Length:     int64(n),
			IsDir:      info.IsDir(),
			ModifiedAt: info.ModTime().Unix(),
			More:       more,
		}

		json, err := jsonC.Marshal(push)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(server.out, "PUSH %v\n", string(json))
		if err != nil {
			return &DeepError{err}
		}

		_, err = server.out.Write(buffer[0:n])
		if err != nil {
			return &DeepError{err}
		}
	}

	return nil
}
