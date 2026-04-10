package mll

import (
	"encoding/binary"
	"fmt"
	"io"
)

// WriteUvarint writes v to w as an LEB128 unsigned varint.
func WriteUvarint(w io.Writer, v uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	_, err := w.Write(buf[:n])
	return err
}

// WriteVarint writes v to w as a zigzag-LEB128 signed varint.
func WriteVarint(w io.Writer, v int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], v)
	_, err := w.Write(buf[:n])
	return err
}

// ReadUvarint reads an LEB128 unsigned varint from b.
// Returns the value, the number of bytes consumed, and any error.
func ReadUvarint(b []byte) (uint64, int, error) {
	v, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, 0, fmt.Errorf("mll: malformed uvarint")
	}
	return v, n, nil
}

// ReadVarint reads a zigzag-LEB128 signed varint from b.
func ReadVarint(b []byte) (int64, int, error) {
	v, n := binary.Varint(b)
	if n <= 0 {
		return 0, 0, fmt.Errorf("mll: malformed varint")
	}
	return v, n, nil
}
