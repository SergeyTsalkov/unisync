package progresswriter

import (
	"io"
	"sync/atomic"
	"time"
)

type progressWriter struct {
	writer  io.WriteCloser
	closed  atomic.Bool
	c       chan Progress
	start   time.Time
	percent atomic.Int64
	written int64
	total   int64
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

	go pw.watch()
	return pw
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.written += int64(n)
	pw.percent.Store(int64((float64(pw.written) / float64(pw.total)) * 100.0))
	return n, err
}

func (pw *progressWriter) Close() error {
	pw.closed.Store(true)
	return pw.writer.Close()
}

func (pw *progressWriter) watch() {
	for {
		time.Sleep(time.Second)

		if closed := pw.closed.Load(); closed {
			return
		}

		percent := int(pw.percent.Load())
		eta := time.Duration(0)
		if percent > 0 {
			runtime := time.Since(pw.start)
			eta = time.Duration(float64(runtime) * (float64(100-percent) / float64(percent)))
		}
		progress := Progress{percent, int(eta.Seconds())}

		select {
		case pw.c <- progress:
		default:
		}

	}
}
