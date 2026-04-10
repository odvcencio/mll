package mll

import (
	"bytes"
	"testing"
)

func TestHeaderRoundTrip(t *testing.T) {
	orig := FileHeader{
		Version:         V1_0,
		Profile:         ProfileSealed,
		Flags:           0,
		TotalFileSize:   0x1234567890,
		SectionCount:    3,
		MinReaderMinor:  0,
	}
	var buf bytes.Buffer
	if err := orig.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.Len() != HeaderSize {
		t.Fatalf("header size: got %d, want %d", buf.Len(), HeaderSize)
	}
	decoded, err := ReadHeader(buf.Bytes())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if decoded != orig {
		t.Fatalf("round trip: got %+v, want %+v", decoded, orig)
	}
}

func TestHeaderRejectsBadMagic(t *testing.T) {
	bad := make([]byte, HeaderSize)
	copy(bad, []byte{'X', 'X', 'X', 'X'})
	_, err := ReadHeader(bad)
	if err == nil {
		t.Fatal("expected magic error")
	}
}

func TestHeaderRejectsTooSmall(t *testing.T) {
	_, err := ReadHeader(make([]byte, HeaderSize-1))
	if err == nil {
		t.Fatal("expected size error")
	}
}
