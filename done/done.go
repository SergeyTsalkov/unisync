package done

import "sync"

type SetDoneFn func(error)
type IsDoneFn func() error
type DoneCFn func() chan error

func New() (SetDoneFn, IsDoneFn, DoneCFn) {
	lock := &sync.Mutex{}
	var done error
	chans := []chan error{}

	SetDone := func(err error) {
		lock.Lock()
		defer lock.Unlock()

		// must specify an error message
		if err == nil {
			panic("SetDone() called without an error")
		}

		// if SetDone() has already been called, do nothing
		if done != nil {
			return
		}

		// push error to every waiting DoneC channel, then close them all
		// each channel has buffer of 1, so this will not block
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
