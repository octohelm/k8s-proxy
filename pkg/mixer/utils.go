package mixer

import (
	"bytes"
	"io"
	"math/rand"
	"time"
)

type Closers []io.Closer

func (c Closers) Close() (err error) {
	for i := range c {
		err = c[i].Close()
	}
	return
}

func ToWriteCloser(w io.Writer, c io.Closer) io.WriteCloser {
	return &struct {
		io.Writer
		io.Closer
	}{w, c}
}

func ToReadCloser(r io.Reader, c io.Closer) io.ReadCloser {
	return &struct {
		io.Reader
		io.Closer
	}{r, c}
}

func ToReadWriter(r io.Reader, w io.Writer) io.ReadWriter {
	rw := &struct {
		io.Reader
		io.Writer
		io.Closer
	}{}

	rw.Reader = r
	rw.Writer = w

	if c, ok := w.(io.Closer); ok {
		rw.Closer = c
	}

	return rw
}

func BufferReadWriter(getWriter func(w io.Writer) (io.Writer, error)) (io.ReadWriter, error) {
	buf := bytes.NewBuffer(nil)
	rw, err := getWriter(buf)
	if err != nil {
		return nil, err
	}
	return ToReadWriter(buf, rw), nil
}

func CopyReader(src io.Reader, dst io.ReadWriter) io.Reader {
	return &copyReader{src: src, dst: dst}
}

type copyReader struct {
	src  io.Reader
	dst  io.ReadWriter
	done bool
}

func (r *copyReader) Read(p []byte) (int, error) {
	if !r.done {
		buf := make([]byte, len(p))

		nr, er := r.src.Read(buf)
		if nr > 0 {
			nw, ew := r.dst.Write(buf[0:nr])
			if ew != nil {
				return 0, ew
			}
			if nr != nw {
				return 0, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return 0, er
			}
		}

		// read end
		if nr == 0 {
			if closer, ok := r.dst.(io.Closer); ok {
				_ = closer.Close()
			}
			r.done = true
		}
	}

	return r.dst.Read(p)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomKey(length int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)

	for i := range b {
		b[i] = letterRunes[rnd.Intn(len(letterRunes))]
	}

	return string(b)
}
