package rollingbuffer

import (
	"io"
	"bytes"
)

type RollingBuffer struct {
	buf    []byte
	offset int
	length int
}

func (b RollingBuffer) Len() int {
	return b.length
}

func (b RollingBuffer) Cap() int {
	return cap(b.buf)
}

func (b *RollingBuffer) Clear() {
	b.offset = 0
	b.length = 0
}

func (b *RollingBuffer) Bytes() []byte {
	by := make([]byte, 0, b.length)
	l := b.offset + b.length
	clip := 0
	if l > cap(b.buf) {
		clip = l % cap(b.buf)
		l = cap(b.buf)
	}
	by = append(by, b.buf[b.offset: l]...)
	if clip > 0 {
		by = append(by, b.buf[:clip]...)
	}
	return by
}

func (b *RollingBuffer) ReadBytes(n int) (int, []byte) {
	by := make([]byte, n)
	l, _ := b.Read(by)
	return l, by
}

func (b *RollingBuffer) Read(p []byte) (n int, err error) {
	if b.length == 0 {
		return 0, io.EOF
	}

	var count int
	for count < cap(p) {
		l := b.length
		if l > cap(p)-count {
			l = cap(p) - count
		}
		e := b.offset + l
		if e > cap(b.buf) {
			e = cap(b.buf)
		}
		c := (e - b.offset)
		copy(p[count:], b.buf[b.offset:e])
		count += c
		b.offset = (b.offset + c) % cap(b.buf)
		b.length -= c

		if b.length < 0 {
			break
		}
	}

	return count, nil
}

func (b *RollingBuffer) Write(p []byte) (n int, err error) {
	i, er := b.ReadFrom(bytes.NewBuffer(p))
	return int(i), er
}

func (b *RollingBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	var c int
	i, l := b.freeSpaceEnd()
	if l > 0 {
		c, err = r.Read(b.buf[i:])
		if err != nil {
			return
		}
	}
	i, l = b.freeSpaceStart()
	if l > 0 {
		cc, er := r.Read(b.buf[i:l])
		if er != nil {
			return
		}
		c += cc
	}
	b.length += c
	return int64(c), nil
}

// freeIndexEnd finds the first free index at the end of the buffer.
// returns -1 if there is no free space after the last occupied position.
func (b RollingBuffer) freeSpaceEnd() (index int, length int) {
	index = b.offset + b.length
	if index >= cap(b.buf) {
		index = cap(b.buf)
	}
	length = cap(b.buf) - index
	return index, length
}

// freeIndexStart finds the first free index at the beginning of the buffer
// returns -1 if there is no free space at the beginning of the buffer
func (b RollingBuffer) freeSpaceStart() (index int, length int) {
	e := b.offset + b.length
	if e >= cap(b.buf) {
		index = e % cap(b.buf)
	}
	return index, b.offset - index 
}

func NewRollingBuffer(capacity int) *RollingBuffer {
	return &RollingBuffer{buf: make([]byte, capacity)}
}
