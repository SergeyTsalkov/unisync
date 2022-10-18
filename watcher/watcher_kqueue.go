//go:build (darwin && kqueue) || (darwin && !cgo) || dragonfly || freebsd || netbsd || openbsd
// +build darwin,kqueue darwin,!cgo dragonfly freebsd netbsd openbsd

package watcher

import "github.com/rjeczalik/notify"

var events notify.Event = notify.All | notify.NoteAttrib
