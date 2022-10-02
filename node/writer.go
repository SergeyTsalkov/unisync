package node

import (
	"fmt"
	"io"
	"strings"
	"unisync/commands"
)

func (n *Node) SendCmdBuf(cmd commands.Command, buf []byte) error {
	err := n.SendString(commands.Encode(cmd))
	if err != nil {
		return err
	}

	if len(buf) > 0 {
		if n.Debug {
			fmt.Printf("-> [%v bytes]\n", len(buf))
		}

		_, err = n.Out.Write(buf)
		if err != nil {
			return &DeepError{err}
		}
	}

	return nil
}

func (n *Node) SendCmd(cmd commands.Command) error {
	return n.SendCmdBuf(cmd, nil)
}

func (n *Node) SendString(str string) error {
	str = strings.TrimSpace(str)

	if n.Debug {
		fmt.Printf("-> %v\n", str)
	}

	_, err := io.WriteString(n.Out, str+"\n")
	if err != nil {
		return &DeepError{err}
	}
	return nil
}

func (n *Node) SendErr(err error) error {
	return n.SendCmd(&commands.Error{
		Err: err.Error(),
	})
}

func (n *Node) SendPathErr(path string, err error) error {
	return n.SendCmd(&commands.Error{
		Err:  err.Error(),
		Path: path,
	})
}
