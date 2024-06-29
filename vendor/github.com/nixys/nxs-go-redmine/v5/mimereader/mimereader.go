package mimereader

import (
	"io"
	"net/http"
)

// Reader store first 512 bytes of readed data to determine its mime type.
type Reader struct {
	r io.Reader
	b []byte
	n int
}

const mimeBufferCapacity = 512

// New creates new mime reader
func New(r io.Reader) *Reader {
	return &Reader{
		r: r,
		b: make([]byte, mimeBufferCapacity),
		n: 0,
	}
}

// Read reads data from reader
func (m *Reader) Read(b []byte) (n int, err error) {

	n, err = m.r.Read(b)
	if n == 0 || err != nil {
		return n, err
	}

	if m.n < mimeBufferCapacity {

		// Bytes left to fill mime buffer
		c := mimeBufferCapacity - m.n

		if n < c {
			// If read bytes less than expected
			c = n
		}

		// Append mime buffer with data
		copy(m.b[m.n:], b[:c])

		// Increase mime buffer len counter
		m.n += c
	}

	return n, err
}

// DetectContentType detects content type for reader
func (m *Reader) DetectContentType() string {
	return http.DetectContentType(m.b[:m.n])
}
