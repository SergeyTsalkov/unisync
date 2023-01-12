//go:build linux
// +build linux

package watcher

import "github.com/rjeczalik/notify"

var events notify.Event = notify.All | notify.InAttrib
var Strategy = "inotify (linux)"
