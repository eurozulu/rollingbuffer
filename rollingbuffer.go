package rollingbuffer

import "io"

type RollingBuffer struct {
	buf    []byte
	offset int
	length int
}

func (r RollingBuffer) Len() int {
	return r.length
}

func (r RollingBuffer) Cap() int {
	return cap(r.buf)
}

func (r *RollingBuffer) Clear() {
	r.offset = 0
	r.length = 0
}

func (r *RollingBuffer) Read(p []byte) (n int, err error) {
	if r.length == 0 {
		return 0, io.EOF
	}

	var count int
	for count < cap(p) {
		l := r.length
		if l > cap(p)-count {
			l = cap(p) - count
		}
		e := r.offset + l
		if e > cap(r.buf) {
			e = cap(r.buf)
		}
		c := (e - r.offset)
		copy(p[count:], r.buf[r.offset:e])
		count += c
		r.offset = (r.offset + c) % cap(r.buf)
		r.length -= c

		if r.length < 0 {
			break
		}
	}

	return count, nil
}

func (r *RollingBuffer) Write(p []byte) (n int, err error) {
	var count int
	fep := r.offset + r.length
	// Can we fill in front of existing
	if fep < cap(r.buf) {
		count = cap(r.buf) - fep
		if len(p) < count {
			count = len(p)
		}
		copy(r.buf[fep:], p[:count])
		r.length += count
		fep +=count
	}
	// fill in back fill
	if fep >= cap(r.buf) {
		fep = fep % cap(r.buf)
		c := len(p) - count
		if fep+c > r.offset {
			c = r.offset - fep
		}
		copy(r.buf[fep: fep+c], p[count: count +c])
		count += c
		r.length += c
	}

	return count, nil
}

func NewRollingBuffer(capacity int) *RollingBuffer {
	return &RollingBuffer{buf: make([]byte, capacity)}
}
