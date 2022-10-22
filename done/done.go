package done

import "sync"

func New() (func(error), func() error, func() chan error) {
	lock := &sync.Mutex{}
	var done error
	chans := []chan error{}

	SetDone := func(err error) {
		lock.Lock()
		defer lock.Unlock()

		if done != nil {
			return
		}

		done = err
		for _, c := range chans {
			c <- done
			close(c)
		}
	}

	IsDone := func() error {
		lock.Lock()
		defer lock.Unlock()
		return done
	}

	DoneC := func() chan error {
		lock.Lock()
		defer lock.Unlock()

		c := make(chan error, 1)

		if done == nil {
			chans = append(chans, c)
		} else {
			c <- done
			close(c)
		}

		return c
	}

	return SetDone, IsDone, DoneC
}
