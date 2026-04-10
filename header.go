package mll

import (
	"encoding/binary"
	"fmt"
	"io"
)

// FileHeader is the fixed 24-byte header of every MLL binary file.
type FileHeader struct {
	Version        Version
	Profile        Profile
	Flags          uint8
	TotalFileSize  uint64
	SectionCount   uint32
	MinReaderMinor uint8
}

// Write encodes the header to w. Always writes exactly HeaderSize bytes on success.
func (h FileHeader) Write(w io.Writer) error {
	buf := make([]byte, HeaderSize)
	// Magic
	copy(buf[0:4], Magic[:])
	// Version (u16 LE, major in high byte)
	binary.LittleEndian.PutUint16(buf[4:6], h.Version.Uint16())
	// Profile
	buf[6] = byte(h.Profile)
	// Flags
	buf[7] = h.Flags
	// Total file size
	binary.LittleEndian.PutUint64(buf[8:16], h.TotalFileSize)
	// Section count
	binary.LittleEndian.PutUint32(buf[16:20], h.SectionCount)
	// Min reader minor
	buf[20] = h.MinReaderMinor
	// Reserved (3 bytes): already zero
	_, err := w.Write(buf)
	return err
}

// ReadHeader parses an MLL file header from the first HeaderSize bytes of b.
func ReadHeader(b []byte) (FileHeader, error) {
	if len(b) < HeaderSize {
		return FileHeader{}, fmt.Errorf("mll: header needs %d bytes, got %d", HeaderSize, len(b))
	}
	// Magic check
	if [4]byte(b[0:4]) != Magic {
		return FileHeader{}, fmt.Errorf("mll: bad magic %v", b[0:4])
	}
	v := binary.LittleEndian.Uint16(b[4:6])
	h := FileHeader{
		Version:        Version{Major: uint8(v >> 8), Minor: uint8(v & 0xFF)},
		Profile:        Profile(b[6]),
		Flags:          b[7],
		TotalFileSize:  binary.LittleEndian.Uint64(b[8:16]),
		SectionCount:   binary.LittleEndian.Uint32(b[16:20]),
		MinReaderMinor: b[20],
	}
	return h, nil
}
