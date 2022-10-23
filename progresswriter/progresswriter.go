package progresswriter

import (
	"io"
	"math"
	"sync/atomic"
	"time"
)

type progressWriter struct {
	writer  io.WriteCloser
	closed  atomic.Bool
	c       chan Progress
	start   time.Time
	written atomic.Int64
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
	pw.written.Add(int64(n))
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

		percent := (float64(pw.written.Load()) / float64(pw.total)) * 100.0
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

	}
}
