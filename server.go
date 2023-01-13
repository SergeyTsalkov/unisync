package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"
	"unisync/log"
	"unisync/server"
)

func runStdinServer() error {
	log.ScreenOutput = os.Stderr
	log.ScreenLevel = log.Warn

	s := server.New(os.Stdin, os.Stdout)
	return s.Run()
}

func runDirectServer(addr string) error {
	cert, capool, err := getCert("secure.key", true)
	if err != nil {
		return err
	}

	conf := &tls.Config{
		Certificates: cert,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    capool,
	}

	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	log.Println("listening at", addr)
	listener, err := tls.Listen("tcp", addr, conf)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		log.Println("Got connection: ", conn.RemoteAddr())
		s := server.New(conn, conn)
		go func() {
			if err := s.Run(); err != nil {
				conn.Close()
				if err == io.EOF {
					err = fmt.Errorf("client disconnected")
				}
				log.Warnln("Closed connection:", err)
			}
		}()
	}
}
