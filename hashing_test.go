package mll

import (
	"encoding/hex"
	"testing"
)

func TestDigestBytesSize(t *testing.T) {
	d := Digest{}
	if len(d) != 32 {
		t.Fatalf("digest size: got %d, want 32", len(d))
	}
}

func TestHashBytes(t *testing.T) {
	h := HashBytes([]byte("hello world"))
	// Well-known BLAKE3-256 hash of "hello world"
	want := "d74981efa70a0c880b8d8c1985d075dbcbf679b99a5f9914e5aaf96b831a9e24"
	got := hex.EncodeToString(h[:])
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestHasherIncremental(t *testing.T) {
	h1 := NewHasher()
	h1.Write([]byte("hello "))
	h1.Write([]byte("world"))
	d1 := h1.Sum()
	d2 := HashBytes([]byte("hello world"))
	if d1 != d2 {
		t.Fatalf("incremental mismatch")
	}
}
