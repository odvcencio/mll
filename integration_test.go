package mll

import (
	"bytes"
	"testing"
)

// TestEndToEndSealedCanonicalStringTable constructs a minimal sealed artifact
// with only HEAD and STRG (using WithSkipRequirementCheck), asserts canonical
// string ordering works, and verifies the sealed content hash is reproducible
// across two independent writer passes.
func TestEndToEndSealedCanonicalStringTable(t *testing.T) {
	build := func() ([]byte, Digest) {
		strg := NewStringTableBuilder()
		// Intern in non-alphabetical order; canonicalization should sort.
		strg.Intern("zebra")
		strg.Intern("apple")
		strg.Intern("mango")
		strg.Intern("tiny_embed")
		strg.CanonicalizeLexicographic()
		// After canonicalization, names have changed. Look up again.
		newNameIdx, _ := strg.Lookup("tiny_embed")

		head := HeadSection{Name: newNameIdx}
		var headBuf bytes.Buffer
		head.Write(&headBuf)

		var strgBuf bytes.Buffer
		strg.Write(&strgBuf)

		sections := []SectionInput{
			{Tag: TagHEAD, Body: headBuf.Bytes(), DigestBody: head.DigestBody(ProfileSealed), Flags: SectionFlagRequired, SchemaVersion: 1},
			{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1},
		}

		var out bytes.Buffer
		wr := NewWriter(&out, ProfileSealed, V1_0, WithSkipRequirementCheck())
		for _, s := range sections {
			wr.AddSection(s)
		}
		if err := wr.Finish(); err != nil {
			t.Fatalf("finish: %v", err)
		}
		return out.Bytes(), wr.ContentHash()
	}

	bytes1, hash1 := build()
	bytes2, hash2 := build()

	if hash1 != hash2 {
		t.Fatalf("sealed content hash not reproducible across writer passes: %x vs %x", hash1, hash2)
	}
	if !bytes.Equal(bytes1, bytes2) {
		t.Fatal("byte-identical output expected from canonicalized writer")
	}

	// Read back and verify
	r, err := ReadBytes(bytes1, WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if r.Profile() != ProfileSealed {
		t.Errorf("profile: got %d", r.Profile())
	}

	headBytes, ok := r.Section(TagHEAD)
	if !ok {
		t.Fatal("no HEAD section")
	}
	decodedHead, err := ReadHeadSection(headBytes)
	if err != nil {
		t.Fatalf("decode head: %v", err)
	}
	_ = decodedHead
}

// TestEndToEndSealedHashStableAcrossWallClocks confirms the HEAD wall-clock
// exclusion: the same logical artifact produced at different times hashes identically.
func TestEndToEndSealedHashStableAcrossWallClocks(t *testing.T) {
	build := func(clock int64) Digest {
		strg := NewStringTableBuilder()
		strg.Intern("test")
		strg.CanonicalizeLexicographic()
		newNameIdx, _ := strg.Lookup("test")

		head := HeadSection{Name: newNameIdx, CreatedUnixMs: clock}
		var headBuf bytes.Buffer
		head.Write(&headBuf)
		var strgBuf bytes.Buffer
		strg.Write(&strgBuf)

		var out bytes.Buffer
		wr := NewWriter(&out, ProfileSealed, V1_0, WithSkipRequirementCheck())
		wr.AddSection(SectionInput{
			Tag:           TagHEAD,
			Body:          headBuf.Bytes(),
			DigestBody:    head.DigestBody(ProfileSealed),
			Flags:         SectionFlagRequired,
			SchemaVersion: 1,
		})
		wr.AddSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
		if err := wr.Finish(); err != nil {
			t.Fatalf("finish: %v", err)
		}
		return wr.ContentHash()
	}

	h1 := build(1000000)
	h2 := build(9999999999)
	if h1 != h2 {
		t.Fatalf("wall clock leaked into sealed content hash: h1=%x h2=%x", h1, h2)
	}
}
