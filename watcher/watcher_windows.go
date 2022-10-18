//go:build windows
// +build windows

package watcher

import "github.com/rjeczalik/notify"

var events notify.Event = notify.All
