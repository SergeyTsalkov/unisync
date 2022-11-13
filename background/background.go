package background

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unisync/config"

	"github.com/shirou/gopsutil/v3/process"
)

var childEnv = "__UNISYNC_CHILD"

func pidFileName(name string) string {
	return filepath.Join(config.ConfigDir(), name+".pid")
}

func IsChild() bool {
	return os.Getenv(childEnv) != ""
}

func WritePid(name string) error {
	if name == "" {
		panic("WritePid(name) -- name can't be blank")
	}
	if !IsChild() {
		panic("WritePid(name) -- should only be used with a child process")
	}

	return os.WriteFile(pidFileName(name), []byte(strconv.Itoa(os.Getpid())), 0644)
}

func proc(name string) (*process.Process, error) {
	bytes, err := os.ReadFile(pidFileName(name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
		return nil, err
	}
	str := strings.TrimSpace(string(bytes))
	pid, err := strconv.Atoi(str)
	if err != nil || pid <= 0 {
		return nil, nil
	}

	proc, err := process.NewProcess(int32(pid))

	if err == process.ErrorProcessNotRunning {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func IsRunning(name string) (bool, error) {
	proc, err := proc(name)
	if err != nil || proc == nil {
		return false, err
	}

	isRunning, err := proc.IsRunning()
	if err != nil || !isRunning {
		return false, err
	}

	procExe, err := proc.Exe()
	if err != nil {
		return false, err
	}
	thisExe, err := os.Executable()
	if err != nil {
		return false, err
	}

	_, thisExe = filepath.Split(thisExe)
	_, procExe = filepath.Split(procExe)
	return procExe == thisExe, nil
}

func Stop(name string) error {
	isRunning, err := IsRunning(name)
	if err != nil {
		return fmt.Errorf("unable to determine if %v is running: %v", name, err)
	}
	if !isRunning {
		return fmt.Errorf("%v is not running", name)
	}

	proc, err := proc(name)
	if err != nil {
		return err
	}
	if err = proc.Kill(); err != nil {
		return err
	}
	if err = os.Remove(pidFileName(name)); err != nil {
		return fmt.Errorf("Unable to remove pid file:", err)
	}
	return nil
}

func ListRunning() {

}

func StopAll() {

}

func Start(name string) error {
	isRunning, err := IsRunning(name)
	if err != nil {
		return fmt.Errorf("unable to determine if %v is running: %v", name, err)
	}
	if isRunning {
		return fmt.Errorf("%v is already running", name)
	}

	var stdout io.Reader
	var stderr io.Reader

	command := os.Args[0]
	args := os.Args[1:]
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%v=%v", childEnv, 1))

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
