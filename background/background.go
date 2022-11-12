package background

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

var childKey = "__UNISYNC_CHILD"

func IsChild() bool {
	return os.Getenv(childKey) != ""
}

func StartChild() error {
	var stdout io.Reader
	var stderr io.Reader
	var err error

	command := os.Args[0]
	args := os.Args[1:]
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%v=%v", childKey, 1))

	if stdout, err = cmd.StdoutPipe(); err != nil {
		panic(err)
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	select {
	case state := <-wait(cmd.Process):
		return fmt.Errorf("process exited: %v", state)
	case err := <-watch(stdout, os.Stdout):
		if err != nil {
			return fmt.Errorf("process error: %v", err)
		}
		return nil
	case err := <-watch(stderr, os.Stderr):
		if err != nil {
			return fmt.Errorf("process error: %v", err)
		}
		return nil
	}

	return nil
}

// like (*os.Process).Wait(), except returns on a channel
func wait(proc *os.Process) chan *os.ProcessState {
	c := make(chan *os.ProcessState, 1)

	go func() {
		p, _ := proc.Wait()
		c <- p
	}()

	return c
}

func watch(in io.Reader, out io.Writer) chan error {
	c := make(chan error, 1)

	go func() {
		_, err := io.Copy(out, in)
		c <- err
	}()

	return c
}
