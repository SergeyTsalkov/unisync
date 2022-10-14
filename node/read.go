package node

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unisync/commands"
)

type Packet struct {
	Command commands.Command
	Buffer  []byte
}

// separate goroutine
func (n *Node) InputReader() {
	var err error

	for {
		var line string
		line, err = n.In.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if n.Debug {
			fmt.Fprintf(os.Stderr, "<- %v\n", line)
		}

		var cmd commands.Command
		cmd, err = commands.Parse(line)
		if err != nil {
			break
		}

		packet := &Packet{Command: cmd}

		if cmd.BodyLen() > 0 {
			packet.Buffer = make([]byte, cmd.BodyLen())
			_, err = io.ReadAtLeast(n.In, packet.Buffer, len(packet.Buffer))
			if err != nil {
				break
			}
		}

		n.Packets <- packet
	}

	if err != nil {
		n.Errors <- err
	}
}
