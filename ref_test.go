package mll

import (
	"bytes"
	"testing"
)

func TestRefFixedSize(t *testing.T) {
	r := Ref{Tag: TagPARM, Index: 5}
	encoded := r.Encode()
	if len(encoded) != 8 {
		t.Fatalf("ref encoded size: got %d, want 8", len(encoded))
	}
}

func TestRefRoundTrip(t *testing.T) {
	orig := Ref{Tag: TagKRNL, Index: 42}
	encoded := orig.Encode()
	decoded, err := DecodeRef(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded != orig {
		t.Fatalf("round trip mismatch: got %+v, want %+v", decoded, orig)
	}
}

func TestRefEncodedLayout(t *testing.T) {
	r := Ref{Tag: [4]byte{'P', 'A', 'R', 'M'}, Index: 0x12345678}
	encoded := r.Encode()
	wantTag := []byte{'P', 'A', 'R', 'M'}
	if !bytes.Equal(encoded[:4], wantTag) {
		t.Fatalf("tag bytes: got %v", encoded[:4])
	}
	// little-endian u32
	wantIdx := []byte{0x78, 0x56, 0x34, 0x12}
	if !bytes.Equal(encoded[4:], wantIdx) {
		t.Fatalf("idx bytes: got %v", encoded[4:])
	}
}
