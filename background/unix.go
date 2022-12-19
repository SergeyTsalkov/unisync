//go:build !windows
// +build !windows

package background

import "syscall"

var procAttr *syscall.SysProcAttr = nil
