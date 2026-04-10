package mll

import (
	"bytes"
	"testing"
)

func TestWriteUint16LE(t *testing.T) {
	var buf bytes.Buffer
	WriteUint16LE(&buf, 0x1234)
	want := []byte{0x34, 0x12}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("got %v, want %v", buf.Bytes(), want)
	}
}

func TestWriteUint32LE(t *testing.T) {
	var buf bytes.Buffer
	WriteUint32LE(&buf, 0xDEADBEEF)
	want := []byte{0xEF, 0xBE, 0xAD, 0xDE}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("got %v, want %v", buf.Bytes(), want)
	}
}

func TestWriteUint64LE(t *testing.T) {
	var buf bytes.Buffer
	WriteUint64LE(&buf, 0x0123456789ABCDEF)
	want := []byte{0xEF, 0xCD, 0xAB, 0x89, 0x67, 0x45, 0x23, 0x01}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("got %v, want %v", buf.Bytes(), want)
	}
}

func TestReadUint16LE(t *testing.T) {
	b := []byte{0x34, 0x12}
	got, err := ReadUint16LE(b)
	if err != nil {
		t.Fatal(err)
	}
	if got != 0x1234 {
		t.Fatalf("got %x", got)
	}
}
