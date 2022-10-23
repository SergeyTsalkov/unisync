package watcher

import (
	"path/filepath"
	"sync"
	"unisync/gitignore"

	"github.com/rjeczalik/notify"
)

type Watcher struct {
	C        chan string
	events   chan notify.EventInfo
	enabled  bool
	ignore   []string
	basepath string
	mutex    sync.Mutex
}

func New() *Watcher {
	return &Watcher{
		C:      make(chan string, 1),
		events: make(chan notify.EventInfo, 100),
	}
}

func (w *Watcher) Start(basepath string, ignore []string) error {
	w.ignore = ignore
	w.basepath = basepath

	go w.monitor()
	err := notify.Watch(filepath.Join(basepath, "..."), w.events, events)
	if err != nil {
		close(w.events)
	}
	return err
}

func (w *Watcher) Stop() {
	notify.Stop(w.events)
	close(w.events)
}

func (w *Watcher) Ready() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.drain()
	w.enabled = true
}

// separate goroutine
func (w *Watcher) monitor() {
	for event := range w.events {
		path, _ := filepath.Rel(w.basepath, event.Path())
		path = filepath.ToSlash(path)

		if !gitignore.MatchAny(w.ignore, path, true) {
			w.Send(path)
		}
	}
}

func (w *Watcher) Send(path string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.enabled {
		return
	}

	// once we've seen an event, don't alert on any others until Ready() is called
	w.enabled = false

	// if there's no buffer room, just discard the event
	select {
	case w.C <- path:
	default:
	}
}

func (w *Watcher) drain() {
	for {
		select {
		case <-w.C:
		default:
			return
		}
	}
}
