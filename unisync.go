package main

import (
	"flag"
	"os"
	"unisync/client"
	"unisync/config"
	"unisync/log"
	"unisync/server"
	"unisync/sshclient"
)

func main() {
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	sshClientFlag := flag.Bool("client", true, "connect through ssh to a remote unisync server")
	flag.Parse()

	if *stdServerFlag {
		runServer()

	} else if *sshClientFlag {
		runClient()

	} else {
		flag.PrintDefaults()
	}

}

func runServer() {
	log.Add(os.Stderr, log.Warn, "")

	s := server.New(os.Stdin, os.Stdout)
	err := s.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func runClient() {
	log.Add(os.Stdout, log.Notice, "")

	conf, err := config.Parse("test")
	if err != nil {
		log.Fatalln(err)
	}

	sshc := sshclient.New(conf.Username, conf.Host)
	err = sshc.Run()
	if err != nil {
		log.Fatalln(err)
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
