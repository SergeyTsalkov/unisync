package main

import (
	"flag"
	"log"
	"os"
	"unisync/client"
	"unisync/config"
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
	s := server.New(os.Stdin, os.Stdout)
	err := s.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func runClient() {
	err := config.Parse("test.json")
	if err != nil {
		log.Fatalln(err)
	}

	sshc := sshclient.New(config.C.Host)
	err = sshc.Run()
	if err != nil {
		log.Fatalln(err)
	}

	c := client.New(sshc.Out, sshc.In)
	err = c.RunHello()
	if err != nil {
		log.Fatalln(err)
	}

	err = c.Sync()
	if err != nil {
		log.Fatalln(err)
	}

}
