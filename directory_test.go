package mll

import (
	"bytes"
	"testing"
)

func TestDirectoryEntryRoundTrip(t *testing.T) {
	orig := DirectoryEntry{
		Tag:           TagHEAD,
		Offset:        24 + 64*3, // after header + 3 directory entries
		Size:          128,
		Digest:        Digest{0xAA, 0xBB, 0xCC},
		Flags:         SectionFlagRequired,
		SchemaVersion: 1,
	}
	var buf bytes.Buffer
	if err := orig.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.Len() != DirectoryEntrySize {
		t.Fatalf("entry size: got %d, want %d", buf.Len(), DirectoryEntrySize)
	}
	decoded, err := ReadDirectoryEntry(buf.Bytes())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if decoded != orig {
		t.Fatalf("round trip: got %+v, want %+v", decoded, orig)
	}
}

func TestDirectoryRoundTrip(t *testing.T) {
	entries := []DirectoryEntry{
		{Tag: TagHEAD, Offset: 100, Size: 16, Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagSTRG, Offset: 120, Size: 32, Flags: SectionFlagRequired, SchemaVersion: 1},
	}
	var buf bytes.Buffer
	if err := WriteDirectory(&buf, entries); err != nil {
		t.Fatalf("write: %v", err)
	}
	if buf.Len() != DirectoryEntrySize*2 {
		t.Fatalf("directory size: got %d", buf.Len())
	}
	got, err := ReadDirectory(buf.Bytes(), uint32(len(entries)))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("count: got %d", len(got))
	}
	for i := range entries {
		if got[i] != entries[i] {
			t.Errorf("entry %d: got %+v, want %+v", i, got[i], entries[i])
		}
	}
}
