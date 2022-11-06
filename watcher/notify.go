package watcher

import (
	"path/filepath"

	"github.com/rjeczalik/notify"
)

func (w *Watcher) StartNotify(basepath string) error {
	// where we use kqueue, reset our open files limit to the maximum
	// reaching this limit will cause the watcher to fail and force us to
	// watch for changes by polling instead
	FixOpenFilesLimit()

	eventsC := make(chan notify.EventInfo, 100)
	w.stop = func() {
		if eventsC == nil {
			return
		}
		notify.Stop(eventsC)
		close(eventsC)
		eventsC = nil
	}

	go w.notifyMonitor(basepath, eventsC)
	err := notify.Watch(filepath.Join(basepath, "..."), eventsC, events)
	if err != nil {
		w.stop()
	}
	return err
}

// separate goroutine
func (w *Watcher) notifyMonitor(basepath string, eventsC chan notify.EventInfo) {
	for event := range eventsC {
		path, _ := filepath.Rel(basepath, event.Path())
		path = filepath.ToSlash(path)
		w.Send(path)
	}
}
