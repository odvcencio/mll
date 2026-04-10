package mll

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DirectoryEntry describes one section in the MLL file directory.
type DirectoryEntry struct {
	Tag           [4]byte
	Offset        uint64
	Size          uint64
	Digest        Digest
	Flags         uint16
	SchemaVersion uint16
}

// Write encodes e to w as exactly DirectoryEntrySize bytes.
func (e DirectoryEntry) Write(w io.Writer) error {
	buf := make([]byte, DirectoryEntrySize)
	copy(buf[0:4], e.Tag[:])
	binary.LittleEndian.PutUint64(buf[4:12], e.Offset)
	binary.LittleEndian.PutUint64(buf[12:20], e.Size)
	copy(buf[20:52], e.Digest[:])
	binary.LittleEndian.PutUint16(buf[52:54], e.Flags)
	binary.LittleEndian.PutUint16(buf[54:56], e.SchemaVersion)
	// pad[8]: zero
	_, err := w.Write(buf)
	return err
}

// ReadDirectoryEntry decodes one directory entry from the first DirectoryEntrySize bytes of b.
func ReadDirectoryEntry(b []byte) (DirectoryEntry, error) {
	if len(b) < DirectoryEntrySize {
		return DirectoryEntry{}, fmt.Errorf("mll: directory entry needs %d bytes, got %d", DirectoryEntrySize, len(b))
	}
	var e DirectoryEntry
	copy(e.Tag[:], b[0:4])
	e.Offset = binary.LittleEndian.Uint64(b[4:12])
	e.Size = binary.LittleEndian.Uint64(b[12:20])
	copy(e.Digest[:], b[20:52])
	e.Flags = binary.LittleEndian.Uint16(b[52:54])
	e.SchemaVersion = binary.LittleEndian.Uint16(b[54:56])
	return e, nil
}

// WriteDirectory writes a slice of directory entries to w.
func WriteDirectory(w io.Writer, entries []DirectoryEntry) error {
	for _, e := range entries {
		if err := e.Write(w); err != nil {
			return err
		}
	}
	return nil
}

// ReadDirectory reads n directory entries from b.
func ReadDirectory(b []byte, n uint32) ([]DirectoryEntry, error) {
	need := int(n) * DirectoryEntrySize
	if len(b) < need {
		return nil, fmt.Errorf("mll: directory needs %d bytes for %d entries, got %d", need, n, len(b))
	}
	entries := make([]DirectoryEntry, n)
	for i := uint32(0); i < n; i++ {
		e, err := ReadDirectoryEntry(b[int(i)*DirectoryEntrySize:])
		if err != nil {
			return nil, err
		}
		entries[i] = e
	}
	return entries, nil
}
