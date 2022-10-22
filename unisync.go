package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/myssh"
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
		conf.Username = user
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

func showHelp() {
	help :=
		`
unisync -- a continuous remote sync tool for programmers

USAGE:
  unisync myserver
    reads config file from ~/.unisync/myserver.conf and syncs according to settings

  unisync ~/localdir user@host:~/remotedir
    runs continuous syncing between localdir and remotedir

  unisync -server 18744
  	runs a direct server, listening on port 18744
  	use a client with method=directtls to connect to it

`

	fmt.Fprintf(os.Stderr, help)
	// flag.PrintDefaults()
	os.Exit(0)
}
