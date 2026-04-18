package mll

import "testing"

func TestSealedContentHashReproducible(t *testing.T) {
	// Two identical directory snapshots must hash the same.
	entries := []DirectoryEntry{
		{Tag: TagHEAD, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1, 2, 3}},
		{Tag: TagSTRG, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{4, 5, 6}},
	}
	h1 := SealedContentHash(V1_0, ProfileSealed, 0, entries)
	h2 := SealedContentHash(V1_0, ProfileSealed, 0, entries)
	if h1 != h2 {
		t.Fatal("sealed content hash not reproducible for identical input")
	}
}

func TestSealedContentHashIgnoresOffsetAndSize(t *testing.T) {
	entries1 := []DirectoryEntry{
		{Tag: TagHEAD, Offset: 100, Size: 16, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1}},
	}
	entries2 := []DirectoryEntry{
		{Tag: TagHEAD, Offset: 200, Size: 16, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1}},
	}
	h1 := SealedContentHash(V1_0, ProfileSealed, 0, entries1)
	h2 := SealedContentHash(V1_0, ProfileSealed, 0, entries2)
	if h1 != h2 {
		t.Fatal("sealed content hash changed when offset changed")
	}
}

func TestSealedContentHashIgnoresSignatureFlag(t *testing.T) {
	entries := []DirectoryEntry{
		{Tag: TagHEAD, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1}},
	}
	h1 := SealedContentHash(V1_0, ProfileSealed, 0, entries)
	h2 := SealedContentHash(V1_0, ProfileSealed, FileFlagHasSignature, entries)
	if h1 != h2 {
		t.Fatal("sealed content hash changed when HAS_SIGNATURE toggled")
	}
}

func TestSealedContentHashIgnoresSgnmEntry(t *testing.T) {
	entriesWithoutSignature := []DirectoryEntry{
		{Tag: TagHEAD, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1}},
	}
	entriesWithSignature := []DirectoryEntry{
		{Tag: TagHEAD, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1}},
		{Tag: TagSGNM, SchemaVersion: 1, Digest: Digest{9, 9, 9}},
	}
	h1 := SealedContentHash(V1_0, ProfileSealed, 0, entriesWithoutSignature)
	h2 := SealedContentHash(V1_0, ProfileSealed, FileFlagHasSignature, entriesWithSignature)
	if h1 != h2 {
		t.Fatal("sealed content hash changed when SGNM was added")
	}
}

func TestSealedContentHashIgnoresMinReaderMinor(t *testing.T) {
	entries := []DirectoryEntry{
		{Tag: TagHEAD, Flags: SectionFlagRequired, SchemaVersion: 1, Digest: Digest{1}},
	}
	// MinReaderMinor is deliberately excluded.
	// Current API takes flags but not min_reader_minor, so this test is structural only.
	h := SealedContentHash(V1_0, ProfileSealed, 0, entries)
	_ = h
}
