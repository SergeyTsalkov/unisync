package node

import (
	"fmt"
	"strings"
	"sync"
	"unisync/commands"
	"unisync/log"
)

type Packet struct {
	Command commands.Command
	Waiter  *sync.WaitGroup
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

		log.Debugf("<- %v\n", line)

		var cmd commands.Command
		cmd, err = commands.Parse(line)
		if err != nil {
			break
		}

		packet := &Packet{Command: cmd}

		if cmd.BodyLen() > 0 {
			packet.Waiter = &sync.WaitGroup{}
			packet.Waiter.Add(1)
		}

		if _, exists := n.sideCmatch[cmd.CmdType()]; exists {
			n.SideC <- packet
		} else {
			n.MainC <- packet
		}

		if packet.Waiter != nil {
			packet.Waiter.Wait()
		}
	}

	if err != nil {
		n.Errors <- err
	}
}

func (n *Node) SetSideC(matches ...string) {
	for _, match := range matches {
		n.sideCmatch[match] = struct{}{}
	}
}

func (c *Node) WaitFor(expectCmd string) (commands.Command, *sync.WaitGroup, error) {
	packet := <-c.MainC

	if cmdType := packet.Command.CmdType(); cmdType != expectCmd {
		return nil, nil, fmt.Errorf("expected %v from server but got %v", expectCmd, cmdType)
	}

	return packet.Command, packet.Waiter, nil
}
