//go:build (darwin && !kqueue && cgo) || (!darwin && !dragonfly && !freebsd && !netbsd && !openbsd)
// +build darwin,!kqueue,cgo !darwin,!dragonfly,!freebsd,!netbsd,!openbsd

package watcher

// where we use kqueue, reset our open files limit to the maximum
// reaching this limit will cause the watcher to fail and force us to
// watch for changes by polling instead
func FixOpenFilesLimit() error {
	return nil
}
