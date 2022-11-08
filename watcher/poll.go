package watcher

import (
	"errors"
	"time"
	"unisync/done"
	"unisync/filelist"
	"unisync/log"
)

func (w *Watcher) StartPoll(basepath string) error {
	SetDone, IsDone, _ := done.New()

	w.stop = func() {
		SetDone(errors.New(""))
	}

	go w.pollMonitor(basepath, IsDone)
	return nil
}

// separate goroutine
func (w *Watcher) pollMonitor(basepath string, IsDone done.IsDoneFn) {
	builder := filelist.NewSyncPlanBuilder("newest", 0777, 0777)
	var list filelist.FileList
	var err error

	for {
		if IsDone() != nil {
			return
		}

		var newlist filelist.FileList
		newlist, err = filelist.Make(basepath, w.ignore, true)
		if err != nil {
			break
		}

		if list != nil {
			plan := builder.BuildSyncPlan(newlist, list, nil)

			for _, file := range plan.FilesChanged() {
				w.Send(file.Path)
			}
		}

		list = newlist
		time.Sleep(w.PollFreq)
	}

	if err != nil {
		log.Warnln(err)
	}
}
