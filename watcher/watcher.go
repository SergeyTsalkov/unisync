package watcher

import (
	"path/filepath"
	"sync/atomic"
	"unisync/gitignore"

	"github.com/rjeczalik/notify"
)

type Watcher struct {
	C      chan string
	events chan notify.EventInfo
	enable *atomic.Bool
	ignore []string
}

func New() *Watcher {
	return &Watcher{
		C:      make(chan string, 1),
		events: make(chan notify.EventInfo, 100),
		enable: &atomic.Bool{},
	}
}

func (w *Watcher) Start(basepath string, ignore []string) error {
	w.ignore = ignore

	go w.monitor()
	err := notify.Watch(filepath.Join(basepath, "..."), w.events, notify.All)
	if err != nil {
		close(w.events)
	}
	return err
}

func (w *Watcher) Ready() {
	w.enable.Store(true)
}

func (w *Watcher) monitor() {
	for event := range w.events {
		path := event.Path()
		enable := w.enable.Load()

		if enable && !gitignore.MatchAny(w.ignore, path, true) {
			w.send(path)
			w.enable.Store(false)
		}
	}
}

func (w *Watcher) send(path string) {
	// if there's no buffer room, just discard the event
	select {
	case w.C <- path:
	default:
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
