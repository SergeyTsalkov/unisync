package watcher

import (
	"sync"
	"time"
	"unisync/gitignore"
)

type stopFn func()

type Watcher struct {
	C        chan string
	PollFreq time.Duration
	enabled  bool
	ignore   []string
	mutex    sync.Mutex
	stop     stopFn
}

func New() *Watcher {
	return &Watcher{
		C:        make(chan string, 1),
		PollFreq: 250 * time.Millisecond,
	}
}

func (w *Watcher) Start(basepath string, ignore []string, poll bool) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.ignore = ignore

	if poll {
		return w.StartPoll(basepath)
	}

	return w.StartNotify(basepath)
}

func (w *Watcher) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.stop != nil {
		w.stop()
	}
	return nil
}

func (w *Watcher) Ready() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.drain()
	w.enabled = true
}

func (w *Watcher) Send(path string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.enabled {
		return
	}

	if gitignore.MatchAny(w.ignore, path, true) {
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
