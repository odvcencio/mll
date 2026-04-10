package mll

import (
	"bytes"
	"testing"
)

func TestCanonicalSectionOrderSealed(t *testing.T) {
	// Unsorted input
	entries := []DirectoryEntry{
		{Tag: TagKRNL},
		{Tag: TagHEAD},
		{Tag: TagSTRG},
		{Tag: TagTNSR},
		{Tag: TagPARM},
	}
	sorted := CanonicalSectionOrder(entries, ProfileSealed)
	// Expected order per spec: HEAD, STRG, ..., PARM, ..., KRNL, ..., TNSR
	want := [][4]byte{TagHEAD, TagSTRG, TagPARM, TagKRNL, TagTNSR}
	if len(sorted) != len(want) {
		t.Fatalf("count: got %d", len(sorted))
	}
	for i, tag := range want {
		if sorted[i].Tag != tag {
			t.Errorf("slot %d: got %v, want %v", i, sorted[i].Tag, tag)
		}
	}
}

func TestCanonicalSectionOrderCustomChunksSorted(t *testing.T) {
	entries := []DirectoryEntry{
		{Tag: [4]byte{'X', 'L', 'O', 'R'}},
		{Tag: TagHEAD},
		{Tag: [4]byte{'X', 'A', 'B', 'C'}},
	}
	sorted := CanonicalSectionOrder(entries, ProfileSealed)
	// Custom chunks sort lexicographically among themselves
	if !bytes.Equal(sorted[1].Tag[:], []byte{'X', 'A', 'B', 'C'}) {
		t.Errorf("custom chunks not sorted lexicographically: %v", sorted)
	}
}
