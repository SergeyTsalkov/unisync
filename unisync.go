package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/minica"
	"unisync/myssh"
	"unisync/myssh/externalssh"
	"unisync/myssh/internalssh"
	"unisync/server"
)

func main() {
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	serverFlag := flag.Bool("server", false, "run server")
	flag.Parse()
	args := flag.Args()
	var conf *config.Config

	if len(args) == 1 {
		var err error
		conf, err = config.Parse(args[0])
		if err != nil {
			log.Fatalln(err)
		}
	} else if len(args) == 2 {
		userhost, remotepath, valid := strings.Cut(args[1], ":")
		if !valid {
			showHelp()
		}

		user, host, valid := strings.Cut(userhost, "@")
		if !valid {
			showHelp()
		}

		conf = config.New()
		conf.Local = args[0]
		conf.Remote = remotepath
		conf.Username = user
		conf.Host = host
	}

	if *stdServerFlag {
		err := runStdinServer()
		if err != nil {
			log.Fatalln(err)
		}

	} else if *serverFlag {
		err := runServer()
		if err != nil {
			log.Fatalln(err)
		}

	} else {
		if conf == nil {
			showHelp()
		}

		log.Printf("Connecting to %v@%v (%v)", conf.Username, conf.Host, conf.Method)

		if conf.Method == "ssh" {
			sshclient := externalssh.New(conf)
			if err := runClient(sshclient, conf); err != nil {
				log.Fatalln(err)
			}

		} else if conf.Method == "internalssh" {
			sshclient, err := internalssh.New(conf)
			if err != nil {
				log.Fatalln("ssh error:", err)
			}
			if err := runClient(sshclient, conf); err != nil {
				log.Fatalln(err)
			}

		} else if conf.Method == "directtls" {
			err := runDirectClient(conf)
			if err != nil {
				log.Fatalln(err)
			}
		}

	}
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

func runServer() error {
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

	addr := ":18744"
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
				log.Warnln(err)
			}
		}()
	}

}

func runStdinServer() error {
	log.Reset()
	log.Add(os.Stderr, log.Warn, "")

	s := server.New(os.Stdin, os.Stdout)
	return s.Run()
}

func runClient(sshclient myssh.SshClient, conf *config.Config) error {
	location := conf.RemoteUnisyncPath[0]
	if len(conf.RemoteUnisyncPath) > 1 {
		var err error
		location, err = sshclient.Search(conf.RemoteUnisyncPath)
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}
	}

	stdin, stdout, err := sshclient.Run(location)
	if err != nil {
		return fmt.Errorf("ssh error: %v", err)
	}
	c, err := client.New(stdout, stdin, conf)
	if err != nil {
		return err
	}
	return c.Run()
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

func showHelp() {
	help :=
		`
unisync -- a continuous remote sync tool for programmers

USAGE:
  unisync myserver
    reads config file from ~/.unisync/myserver.conf and syncs according to settings

  unisync ~/localdir user@host:~/remotedir
    runs continuous syncing between localdir and remotedir

`

	fmt.Fprintf(os.Stderr, help)
	// flag.PrintDefaults()
	os.Exit(0)
}
