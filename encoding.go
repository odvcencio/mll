package mll

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// WriteUint16LE writes a uint16 to w in little-endian order.
func WriteUint16LE(w io.Writer, v uint16) error {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

// WriteUint32LE writes a uint32 to w in little-endian order.
func WriteUint32LE(w io.Writer, v uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

// WriteUint64LE writes a uint64 to w in little-endian order.
func WriteUint64LE(w io.Writer, v uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	_, err := w.Write(buf[:])
	return err
}

// ReadUint16LE reads a uint16 from the first 2 bytes of b.
func ReadUint16LE(b []byte) (uint16, error) {
	if len(b) < 2 {
		return 0, fmt.Errorf("mll: need 2 bytes for uint16, got %d", len(b))
	}
	return binary.LittleEndian.Uint16(b), nil
}

// ReadUint32LE reads a uint32 from the first 4 bytes of b.
func ReadUint32LE(b []byte) (uint32, error) {
	if len(b) < 4 {
		return 0, fmt.Errorf("mll: need 4 bytes for uint32, got %d", len(b))
	}
	return binary.LittleEndian.Uint32(b), nil
}

// ReadUint64LE reads a uint64 from the first 8 bytes of b.
func ReadUint64LE(b []byte) (uint64, error) {
	if len(b) < 8 {
		return 0, fmt.Errorf("mll: need 8 bytes for uint64, got %d", len(b))
	}
	return binary.LittleEndian.Uint64(b), nil
}

// Float64bits returns the IEEE 754 binary representation of f.
func Float64bits(f float64) uint64 {
	return math.Float64bits(f)
}

// Float64frombits returns the floating-point number corresponding to the
// IEEE 754 binary representation b.
func Float64frombits(b uint64) float64 {
	return math.Float64frombits(b)
}
