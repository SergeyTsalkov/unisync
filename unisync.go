package main

import (
	"flag"
	"log"
	"os"
	"unisync/node"
	"unisync/sshclient"
)

func main() {
	stdServerFlag := flag.Bool("stdserver", false, "run server that uses stdin/stdout (internal use only)")
	sshClientFlag := flag.Bool("client", true, "connect through ssh to a remote unisync server")
	flag.Parse()

	if *stdServerFlag {
		s := node.New(os.Stdin, os.Stdout)
		err := s.Run()
		if err != nil {
			log.Fatalln(err)
		}

	} else if *sshClientFlag {

		c := sshclient.New()
		err := c.Run()
		if err != nil {
			log.Fatalln(err)
		}

		select {} // wait forever

	} else {
		flag.PrintDefaults()
	}

}
