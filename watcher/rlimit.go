//go:build (darwin && kqueue) || (darwin && !cgo) || dragonfly || freebsd || netbsd || openbsd
// +build darwin,kqueue darwin,!cgo dragonfly freebsd netbsd openbsd

package watcher

import "golang.org/x/sys/unix"

// where we use kqueue, reset our open files limit to the maximum
// reaching this limit will cause the watcher to fail and force us to
// watch for changes by polling instead
func FixOpenFilesLimit() error {
	rlim := &unix.Rlimit{}

	err := unix.Getrlimit(unix.RLIMIT_NOFILE, rlim)
	if err != nil {
		return err
	}

	rlim.Cur = rlim.Max
	return unix.Setrlimit(unix.RLIMIT_NOFILE, rlim)
}
