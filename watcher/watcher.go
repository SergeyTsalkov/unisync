package watcher

import (
	"path/filepath"
	"sync/atomic"
	"unisync/gitignore"

	"github.com/rjeczalik/notify"
)

type Watcher struct {
	C        chan string
	events   chan notify.EventInfo
	enabled  *atomic.Bool
	ignore   []string
	basepath string
}

func New() *Watcher {
	return &Watcher{
		C:       make(chan string, 1),
		events:  make(chan notify.EventInfo, 100),
		enabled: &atomic.Bool{},
	}
}

func (w *Watcher) Start(basepath string, ignore []string) error {
	w.ignore = ignore
	w.basepath = basepath

	go w.monitor()
	err := notify.Watch(filepath.Join(basepath, "..."), w.events, notify.All)
	if err != nil {
		close(w.events)
	}
	return err
}

func (w *Watcher) Ready() {
	w.drain()
	w.enabled.Store(true)
}

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
	if enabled := w.enabled.Load(); !enabled {
		return
	}

	// once we've seen an event, don't alert on any others until Ready() is called
	w.enabled.Store(false)

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

// func (w *Watcher) watch() {
// 	for {
// 		select {
// 		case event, ok := <-w.w.Events:
// 			if !ok {
// 				return
// 			}

// 			log.Println("Watcher Event:", event)

// 			if event.Op&fsnotify.Create != 0 {
// 				info, err := os.Lstat(event.Name)

// 				if err == nil && info.IsDir() {
// 					err = w.w.Add(event.Name)
// 					if err != nil {
// 						log.Println("Watcher error:", err)
// 					}
// 				}
// 			}

// 			w.send(event.Name)

// 		case err, ok := <-w.w.Errors:
// 			if !ok {
// 				return
// 			}
// 			log.Println("Watcher error:", err)
// 		}
// 	}
// }

// func dirs(basepath string, ignore []string) ([]string, error) {
// 	dirs := []string{}
// 	err := filepath.Walk(basepath, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		relpath, err := filepath.Rel(basepath, path)
// 		if err == nil && relpath != "." {
// 			relpath = filepath.ToSlash(relpath)
// 			if gitignore.MatchAny(ignore, relpath, info.IsDir()) {
// 				if info.IsDir() {
// 					return filepath.SkipDir
// 				}
// 				return nil
// 			}
// 		}

// 		if info.IsDir() {
// 			dirs = append(dirs, path)
// 		}

// 		return nil
// 	})

// 	return dirs, err
// }
