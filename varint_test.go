package mll

import (
	"bytes"
	"testing"
)

func TestUvarintRoundTrip(t *testing.T) {
	cases := []uint64{0, 1, 127, 128, 255, 256, 16383, 16384, 0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF}
	for _, v := range cases {
		var buf bytes.Buffer
		WriteUvarint(&buf, v)
		got, n, err := ReadUvarint(buf.Bytes())
		if err != nil {
			t.Errorf("v=%d: %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("v=%d: got %d", v, got)
		}
		if n != buf.Len() {
			t.Errorf("v=%d: used %d bytes, wrote %d", v, n, buf.Len())
		}
	}
}

func TestVarintRoundTrip(t *testing.T) {
	cases := []int64{0, 1, -1, 63, -64, 64, -65, 0x7FFFFFFF, -0x80000000}
	for _, v := range cases {
		var buf bytes.Buffer
		WriteVarint(&buf, v)
		got, n, err := ReadVarint(buf.Bytes())
		if err != nil {
			t.Errorf("v=%d: %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("v=%d: got %d", v, got)
		}
		_ = n
	}
}

func TestUvarintCompactness(t *testing.T) {
	// 127 fits in 1 byte
	var buf bytes.Buffer
	WriteUvarint(&buf, 127)
	if buf.Len() != 1 {
		t.Errorf("127: used %d bytes, want 1", buf.Len())
	}
	// 128 needs 2 bytes
	buf.Reset()
	WriteUvarint(&buf, 128)
	if buf.Len() != 2 {
		t.Errorf("128: used %d bytes, want 2", buf.Len())
	}
}
