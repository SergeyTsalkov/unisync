package main

import (
	"log"
	"os"
	"unisync/server"
)

func main() {
	s := server.New(os.Stdin, os.Stdout)
	err := s.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
