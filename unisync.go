package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/server"
	"unisync/sshclient"
)

func main() {
	log.Add(os.Stdout, log.Notice, "")

	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	flag.Parse()
	args := flag.Args()
	var conf *config.Config

	if len(args) == 1 {
		var err error
		conf, err = config.Parse(args[0])
		if err != nil {
			log.Fatalf("Failed parsing config file %v: %v", args[0], err)
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
		runServer()
	} else {
		if conf == nil {
			showHelp()
		}

		runClient(conf)
	}
}

func runServer() {
	log.Reset()
	log.Add(os.Stderr, log.Warn, "")

	s := server.New(os.Stdin, os.Stdout)
	err := s.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func runClient(conf *config.Config) {
	log.Printf("Connecting to %v@%v", conf.Username, conf.Host)
	sshc := sshclient.New(conf.Username, conf.Host, conf.SshPath, conf.SshOpts)

	location := conf.RemoteUnisyncPath[0]
	if len(conf.RemoteUnisyncPath) > 1 {
		var err error
		location, err = sshc.Search(conf.RemoteUnisyncPath)
		if err != nil {
			log.Fatalln("ssh error:", err)
		}
	}

	err := sshc.Run(location)
	if err != nil {
		log.Fatalln("ssh error:", err)
	}

	c, err := client.New(sshc.Out, sshc.In, conf)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.Run()
	if err != nil {
		log.Fatalln(err)
	}
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
