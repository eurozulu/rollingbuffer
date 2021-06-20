package rollingbuffer

import (
	"io"
)

type RollingBuffer struct {
	base          []byte
	buckets       [2][]byte
	readOverflow  bool
	writeOverflow bool
}

func NewRollingBuffer(size int) *RollingBuffer {
	return NewRollingBufferBytes(make([]byte, size))
}

func NewRollingBufferBytes(by []byte) *RollingBuffer {
	return &RollingBuffer{base: by}
}

func (b *RollingBuffer) Read(p []byte) (n int, err error) {
	if b.Len() == 0 {
		return 0, io.EOF
	}
	by := b.Next(len(p))
	copy(p, by)
	return len(by), nil
}

func (b *RollingBuffer) Write(p []byte) (n int, err error) {
	bfree := b.Cap() - b.Len()
	l := len(p)
	if l > bfree {
		l = bfree
	}
	if l == 0 {
		// no more room in the inn
		return 0, nil
	}

	for n < l {
		i := 0
		if b.writeOverflow {
			i = 1
		}
		bk := b.buckets[i]

		if cap(bk) == 0 {
			bk = b.base[:0]
		}
		bfree = cap(bk) - len(bk)
		if bfree == 0 {
			// bucket is already full
			b.writeOverflow = !b.writeOverflow
			continue
		}
		// how much do we copy
		bl := l - n
		if bl > bfree {
			bl = bfree
		}
		// append to end of bucket and move n counter on.
		nn := n + bl
		b.buckets[i] = append(bk, p[n:nn]...)
		n = nn
	}
	return n, nil
}

func (b RollingBuffer) Len() int {
	return len(b.buckets[0]) + len(b.buckets[1])
}
func (b RollingBuffer) Cap() int {
	return cap(b.base)
}

func (b *RollingBuffer) ReadByte() (byte, error) {
	by := b.Next(1)
	if len(by) > 0 {
		return by[0], nil
	}
	return 0, io.EOF
}

func (b *RollingBuffer) UnreadByte() error {
	if b.Len() == b.Cap() {
		return io.ErrShortBuffer
	}
	i := 0
	if b.readOverflow {
		i = 1
	}
	bk := b.buckets[i]
	bos := b.Cap() - cap(bk)
	if bos <= 0 {
		// nothing to unread
		return io.ErrUnexpectedEOF
	}
	b.buckets[i] = b.base[bos-1 : bos+len(bk)]
	return nil
}

func (b *RollingBuffer) WriteByte(c byte) error {
	_, err := b.Write([]byte{c})
	return err
}

func (b *RollingBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	bFree := b.Cap() - b.Len()
	if bFree == 0 {
		return 0, nil
	}
	by := make([]byte, bFree)
	_, err = r.Read(by)
	if err != nil {
		return 0, err
	}
	l, err := b.Write(by)
	if err != nil {
		return 0, err
	}
	return int64(l), nil
}

func (b *RollingBuffer) Next(n int) []byte {
	l := b.Len()
	if l == 0 {
		return nil
	}
	if n > l {
		n = l
	}
	// stop looking once we flip back to where we started
	firstState := b.readOverflow
	var by []byte
	for len(by) < n {
		bk := b.buckets[b.readBucketIndex()]
		if len(bk) == 0 {
			// no data in current bucket, flip to other bucket.
			b.readOverflow = !b.readOverflow
			if b.readOverflow == firstState {
				break
			}
			continue
		}
		bl := n - len(by)
		if bl > len(bk) {
			bl = len(bk)
		}

		by = append(by, bk[0:bl]...)
		b.buckets[b.readBucketIndex()] = bk[bl:]
	}
	return by
}

func (b *RollingBuffer) WriteTo(w io.Writer) (n int64, err error) {
	l := b.Len()
	if l == 0 {
		return 0, io.EOF
	}

	// stop looking once we flip back to where we started
	firstState := b.readOverflow

	for n < int64(l) {
		bk := b.buckets[b.readBucketIndex()]
		if len(bk) == 0 {
			// no data in current bucket, flip to other bucket.
			b.readOverflow = !b.readOverflow
			if b.readOverflow == firstState {
				break
			}
			continue
		}
		i, err := w.Write(bk)
		if err != nil {
			return int64(i), err
		}

		// trim off the bytes read
		b.buckets[b.readBucketIndex()] = bk[i:]
		n += int64(i)
	}
	return n, nil
}

func (b *RollingBuffer) readBucketIndex() int {
	if b.readOverflow {
		return 1
	}
	return 0
}
func (b *RollingBuffer) writeIndex() int {
	if b.writeOverflow {
		return 1
	}
	return 0
}
