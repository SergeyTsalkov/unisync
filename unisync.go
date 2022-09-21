package main

import (
	"flag"
	"log"
	"os"
	"unisync/client"
	"unisync/filelist"
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
	sshc := sshclient.New()
	err := sshc.Run()
	if err != nil {
		log.Fatalln(err)
	}

	c := client.New(sshc.Out, sshc.In)
	c.LocalPath = "/Users/sergey/test"
	c.RemotePath = "/home/sergey/test"

	err = c.RunHello()
	if err != nil {
		log.Fatalln(err)
	}

	remoteList, err := c.RunReqList()
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(string(remoteList.Encode()))

	localList, err := filelist.Make(c.LocalPath)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(string(localList.Encode()))

	syncplan := filelist.Compare(localList, remoteList)
	syncplan.Show()

}
