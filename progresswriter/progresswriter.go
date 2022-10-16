package progresswriter

import (
	"io"
	"sync/atomic"
	"time"
)

type progressWriter struct {
	writer  io.WriteCloser
	closed  atomic.Bool
	c       chan int
	percent atomic.Int64
	written float64
	total   float64
}

func New(w io.WriteCloser, total int64, c chan int) *progressWriter {
	pw := &progressWriter{
		writer: w,
		c:      c,
		total:  float64(total),
	}

	go pw.watch()
	return pw
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.written += float64(n)
	pw.percent.Store(int64((pw.written / pw.total) * 100))
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

		percent := pw.percent.Load()
		select {
		case pw.c <- int(percent):
		default:
		}

	}
}
