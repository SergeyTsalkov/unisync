package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"path/filepath"
	"strings"
	"time"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/minica"
	"unisync/server"
)

func runDirectServer(addr string) error {
	mca, err := getMiniCa(true)
	if err != nil {
		return err
	}
	cert, err := mca.MakeCert()
	if err != nil {
		return err
	}

	conf := &tls.Config{
		Certificates: cert,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    mca.GetCAPool(),
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

func runDirectClient(conf *config.Config) error {
	connectTimeout := time.Duration(conf.ConnectTimeout) * time.Second
	timeout := time.Duration(conf.Timeout) * time.Second

	mca, err := getMiniCa(false)
	if err != nil {
		return err
	}
	cert, err := mca.MakeCert()
	if err != nil {
		return err
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: timeout,
		},
		Config: &tls.Config{
			ServerName:   "unisync",
			Certificates: cert,
			RootCAs:      mca.GetCAPool(),
		},
	}

	conn, err := dialer.Dial("tcp", fmt.Sprintf("%v:%v", conf.Host, conf.Port))
	if err != nil {
		return err
	}

	c, err := client.New(conn, conn, conf)
	if err != nil {
		return err
	}

	return c.Run()
}

func getMiniCa(canMake bool) (*minica.MiniCA, error) {
	fullpath := filepath.Join(config.ConfigDir(), "secure.key")
	mca, err := minica.Load(fullpath)

	if err != nil && canMake && errors.Is(err, fs.ErrNotExist) {
		mca, err = minica.New(fullpath)
		if err != nil {
			return nil, fmt.Errorf("Failed to create key at %v: %w", fullpath, err)
		}

		log.Printf("Created new key at %v, make sure to copy this to the client so it can connect!", fullpath)
	} else if err != nil {
		return nil, fmt.Errorf("Failed to load key at %v: %w", fullpath, err)
	}

	return mca, nil
}
