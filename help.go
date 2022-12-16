package main

import (
	"fmt"
	"os"
)

func showHelp() {
	help :=
		`
unisync -- a continuous remote sync tool for programmers

USAGE:
  unisync myserver
    reads config file from ~/.unisync/myserver.conf and syncs according to settings

  unisync -server 18744
    runs a direct server, listening on port 18744
    use a client with method=directtls to connect to it

`

	fmt.Fprintf(os.Stderr, help)
	// flag.PrintDefaults()
	os.Exit(0)
}
