//go:build windows
// +build windows

package background

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// if we don't use DETACHED_PROCESS, the child process will die when the terminal is closed
// even if the child process's stdin/stdout/stderr are detached
// see for more info:
// https://learn.microsoft.com/en-us/windows/console/console-process-groups
// https://stackoverflow.com/questions/74854616/im-trying-to-fork-off-a-background-process-in-go-works-perfectly-on-mac-an/74856220

var procAttr = &syscall.SysProcAttr{
	CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
}
