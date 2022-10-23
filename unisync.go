package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/myssh/externalssh"
	"unisync/myssh/internalssh"
	"unisync/server"
)

func main() {
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	serverFlag := flag.String("server", "", "run server")
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
		conf.User = user
		conf.Host = host
	}

	if *stdServerFlag {
		err := runStdinServer()
		if err != nil {
			log.Fatalln(err)
		}

	} else if *serverFlag != "" {
		err := runDirectServer(*serverFlag)
		if err != nil {
			log.Fatalln(err)
		}

	} else {
		if conf == nil {
			showHelp()
		}

		runClient(conf)
	}
}

func runStdinServer() error {
	log.Reset()
	log.Add(os.Stderr, log.Warn, "")

	s := server.New(os.Stdin, os.Stdout)
	return s.Run()
}

func runClient(conf *config.Config) {
	retryTime := 5 * time.Second

	for {
		err := _runClient(conf)
		if err != nil {
			log.Warnln("Client disconnected:", err)
		}

		log.Printf("Retrying in %v..", retryTime)
		time.Sleep(retryTime)
	}
}

func _runClient(conf *config.Config) error {
	var in io.Reader
	var out io.Writer
	var err error

	log.Printf("Connecting to %v@%v (%v)", conf.User, conf.Host, conf.Method)

	if conf.Method == "internalssh" {
		sshclient, err := internalssh.New(conf)
		if err != nil {
			return err
		}

		out, in, err = sshclient.Run()
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}

	} else if conf.Method == "ssh" {
		sshclient := externalssh.New(conf)
		out, in, err = sshclient.Run()
		if err != nil {
			return fmt.Errorf("ssh error: %v", err)
		}

	} else if conf.Method == "directtls" {
		cert, capool, err := getCert(false)
		if err != nil {
			return err
		}

		tlsdialer := &tls.Dialer{
			NetDialer: &net.Dialer{
				Timeout:   config.Duration(conf.ConnectTimeout),
				KeepAlive: config.Duration(conf.Timeout),
			},
			Config: &tls.Config{
				ServerName:   "unisync",
				Certificates: cert,
				RootCAs:      capool,
			},
		}

		conn, err := tlsdialer.Dial("tcp", fmt.Sprintf("%v:%v", conf.Host, conf.Port))
		if err != nil {
			return err
		}
		out = conn
		in = conn
	} else {
		panic("conf.Method=" + conf.Method)
	}

	c, err := client.New(in, out, conf)
	if err != nil {
		return err
	}
	return c.Run()
}
