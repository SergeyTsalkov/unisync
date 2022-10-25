package progresswriter

import (
	"io"
	"math"
	"sync"
	"time"
)

type progressWriter struct {
	writer  io.WriteCloser
	closed  bool
	c       chan Progress
	start   time.Time
	written int64
	total   int64
	mutex   sync.Mutex
}

type Progress struct {
	Percent int
	Eta     int
}

func New(w io.WriteCloser, total int64, c chan Progress) *progressWriter {
	pw := &progressWriter{
		writer: w,
		start:  time.Now(),
		c:      c,
		total:  total,
	}

	pw.startWatch()
	return pw
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)

	pw.mutex.Lock()
	pw.written += int64(n)
	pw.mutex.Unlock()

	return n, err
}

func (pw *progressWriter) Close() error {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()

	pw.closed = true
	return pw.writer.Close()
}

func (pw *progressWriter) startWatch() {
	go func() {
		for isOpen := true; isOpen; {
			time.Sleep(time.Second)
			isOpen = pw.watch()
		}
	}()
}

func (pw *progressWriter) watch() bool {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()

	if pw.closed {
		return false
	}

	percent := (float64(pw.written) / float64(pw.total)) * 100.0
	eta := time.Duration(0)
	if percent > 0 {
		runtime := time.Since(pw.start)
		eta = time.Duration(float64(runtime) * ((100.0 - percent) / percent))
	}
	progress := Progress{int(math.Round(percent)), int(eta.Seconds())}

	select {
	case pw.c <- progress:
	default:
	}

	return true
}
