package node

import (
	"io"
	"strings"
	"unisync/commands"
	"unisync/log"
)

func (n *Node) Write(p []byte) (int, error) {
	n.writeLock.Lock()
	defer n.writeLock.Unlock()
	return n.Out.Write(p)
}

func (n *Node) SendCmdBuf(cmd commands.Command, buf []byte) error {
	err := n.SendString(commands.Encode(cmd))
	if err != nil {
		return err
	}

	if len(buf) > 0 {
		log.Debugf("-> [%v bytes]", len(buf))
		_, err = n.Write(buf)

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

	log.Debugf("-> %v", str)
	_, err := io.WriteString(n, str+"\n")
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
