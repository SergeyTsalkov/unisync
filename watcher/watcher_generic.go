//go:build !darwin && !linux && !freebsd && !dragonfly && !netbsd && !openbsd && !windows && !kqueue && !solaris && !illumos
// +build !darwin,!linux,!freebsd,!dragonfly,!netbsd,!openbsd,!windows,!kqueue,!solaris,!illumos

package watcher

import "github.com/rjeczalik/notify"

var events notify.Event = notify.All
