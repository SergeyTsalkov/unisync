//go:build solaris || illumos
// +build solaris illumos

package watcher

import "github.com/rjeczalik/notify"

var events notify.Event = notify.All | notify.FileAttrib
var Strategy = "fen (solaris / illumos)"
