package main

import (
	"os"
	"unisync/server"
)

func main() {
	s := server.New(os.Stdin, os.Stdout)
	s.Run()
}
