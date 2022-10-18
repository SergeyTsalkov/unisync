//go:build darwin && !kqueue && cgo
// +build darwin,!kqueue,cgo

package watcher

import "github.com/rjeczalik/notify"

// FSEventsChangeOwner includes CHMOD changes for some reason
var events notify.Event = notify.All | notify.FSEventsChangeOwner
