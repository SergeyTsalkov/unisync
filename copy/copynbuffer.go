package copy

import (
	"io"
)

func CopyNbuffer(dst io.Writer, src io.Reader, n int64, buf []byte) (written int64, err error) {
	for n > 0 {
		if n < int64(len(buf)) {
			buf = buf[:n]
		}

		_, err = io.ReadAtLeast(src, buf, len(buf))
		if err != nil {
			return
		}

		var out int
		out, err = dst.Write(buf)
		written += int64(out)
		n -= int64(out)
		if err != nil {
			return
		}
	}

	return
}
