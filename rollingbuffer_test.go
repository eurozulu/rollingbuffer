package rollingbuffer_test

import (
	"rollingbuffer"
	"testing"
)

func TestWriteFull(t *testing.T) {
	capacity := 100
	datalen := 59
	data := make([]byte, datalen)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i + 32)
	}

	b := rollingbuffer.NewRollingBuffer(capacity)
	if b.Len() != 0 {
		t.Errorf("empty buffer length not zero")
	}

	l, err := b.Write(data)
	if err != nil {
		t.Error(err)
	}
	if l != datalen {
		t.Errorf("unexpected length returned from write. Expeced %d, found %d", datalen, l)
	}
	if b.Len() != datalen {
		t.Errorf("buffer length expected %d, found %d", datalen, b.Len())
	}

	// write again to overfill
	datalen2 := capacity - datalen
	l, err = b.Write(data)
	if err != nil {
		t.Error(err)
	}
	if l != datalen2 {
		t.Errorf("unexpected length returned from write. Expeced %d, found %d", datalen2, l)
	}
	if b.Len() != capacity {
		t.Errorf("buffer length expected %d, found %d", capacity, b.Len())
	}
}

func TestWriteReadWrite(t *testing.T) {
	capacity := 100
	datalen := 59
	data := make([]byte, datalen)
	for i := 0; i < len(data); i++ {
		data[i] = byte(i + 32)
	}

	b := rollingbuffer.NewRollingBuffer(capacity)
	l, err := b.Write(data)
	if err != nil {
		t.Error(err)
	}
	if b.Len() != datalen {
		t.Errorf("buffer length expected %d, found %d", datalen, b.Len())
	}

	rdata := make([]byte, 10)
	rl, err := b.Read(rdata)
	if err != nil {
		t.Error(err)
	}
	if rl != 10 {
		t.Errorf("unexpected read length of %d, expected %d", rl, 10)
	}
	if b.Len() != datalen - 10 {
		t.Errorf("unexpected buffer length of %d, expected %d", b.Len(), datalen - len(rdata))
	}

	// write again to overfill
	l, err = b.Write(data)
	if err != nil {
		t.Error(err)
	}
	if l != 100 - (datalen - 10) {
		t.Errorf("unexpected length returned from write. Expeced %d, found %d", 41, l)
	}
	if b.Len() != capacity {
		t.Errorf("buffer length expected %d, found %d", capacity, b.Len())
	}
}

