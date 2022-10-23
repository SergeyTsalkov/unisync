package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"unisync/config"
	"unisync/log"
	"unisync/minica"
	"unisync/server"
)

func runDirectServer(addr string) error {
	cert, capool, err := getCert(true)
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

var mca *minica.MiniCA

func getCert(canMake bool) ([]tls.Certificate, *x509.CertPool, error) {
	if mca == nil {
		var err error
		fullpath := filepath.Join(config.ConfigDir(), "secure.key")
		mca, err = minica.Load(fullpath)

		if err != nil && canMake && errors.Is(err, fs.ErrNotExist) {
			mca, err = minica.New(fullpath)
			if err != nil {
				return nil, nil, fmt.Errorf("Failed to create key at %v: %w", fullpath, err)
			}

			log.Printf("Created new key at %v, make sure to copy this to the client so it can connect!", fullpath)
		} else if err != nil {
			return nil, nil, fmt.Errorf("Failed to load key at %v: %w", fullpath, err)
		}
	}

	cert, err := mca.GetCert()
	if err != nil {
		return nil, nil, err
	}

	return cert, mca.GetCAPool(), nil
}
