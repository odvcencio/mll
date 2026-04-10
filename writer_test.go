package mll

import (
	"bytes"
	"testing"
)

func TestWriteMinimalSealedFile(t *testing.T) {
	strg := NewStringTableBuilder()
	head := HeadSection{
		Name: strg.Intern("test"),
	}
	var headBuf bytes.Buffer
	head.Write(&headBuf)
	var strgBuf bytes.Buffer
	strg.Write(&strgBuf)

	sections := []SectionInput{
		{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1},
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

	bytesOut := out.Bytes()
	if [4]byte(bytesOut[0:4]) != Magic {
		t.Fatal("bad magic in output")
	}
	hdr, err := ReadHeader(bytesOut[:HeaderSize])
	if err != nil {
		t.Fatalf("header: %v", err)
	}
	if hdr.Profile != ProfileSealed {
		t.Errorf("profile: got %d", hdr.Profile)
	}
	if hdr.SectionCount != 2 {
		t.Errorf("section count: got %d", hdr.SectionCount)
	}
	if hdr.TotalFileSize != uint64(len(bytesOut)) {
		t.Errorf("total file size: got %d, want %d", hdr.TotalFileSize, len(bytesOut))
	}
	dirBytes := bytesOut[HeaderSize : HeaderSize+DirectoryEntrySize*2]
	entries, err := ReadDirectory(dirBytes, hdr.SectionCount)
	if err != nil {
		t.Fatalf("directory: %v", err)
	}
	if entries[0].Tag != TagHEAD {
		t.Errorf("first section: got %v, want HEAD", entries[0].Tag)
	}
	if entries[1].Tag != TagSTRG {
		t.Errorf("second section: got %v, want STRG", entries[1].Tag)
	}
	for _, e := range entries {
		body := bytesOut[e.Offset : e.Offset+e.Size]
		want := HashBytes(body)
		if want != e.Digest {
			t.Errorf("digest mismatch for %v", e.Tag)
		}
	}
}

func TestWriteReproducibleSealedHash(t *testing.T) {
	build := func() Digest {
		strg := NewStringTableBuilder()
		head := HeadSection{Name: strg.Intern("test")}
		var headBuf, strgBuf bytes.Buffer
		head.Write(&headBuf)
		strg.Write(&strgBuf)
		var out bytes.Buffer
		wr := NewWriter(&out, ProfileSealed, V1_0, WithSkipRequirementCheck())
		wr.AddSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
		wr.AddSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
		if err := wr.Finish(); err != nil {
			t.Fatal(err)
		}
		return wr.ContentHash()
	}
	h1 := build()
	h2 := build()
	if h1 != h2 {
		t.Fatalf("sealed content hash not reproducible: h1=%x h2=%x", h1, h2)
	}
}

func TestWriterRequiresRequiredSections(t *testing.T) {
	var out bytes.Buffer
	wr := NewWriter(&out, ProfileSealed, V1_0)
	err := wr.Finish()
	if err == nil {
		t.Fatal("expected error for missing required sections")
	}
}

func TestWriterSkipRequirementCheck(t *testing.T) {
	var out bytes.Buffer
	wr := NewWriter(&out, ProfileSealed, V1_0, WithSkipRequirementCheck())
	if err := wr.Finish(); err != nil {
		t.Fatalf("opt-in should bypass required check: %v", err)
	}
}

func TestWriterRejectsForbiddenSection(t *testing.T) {
	// Sealed forbids OPTM.
	var out bytes.Buffer
	wr := NewWriter(&out, ProfileSealed, V1_0, WithSkipRequirementCheck())
	wr.AddSection(SectionInput{Tag: TagOPTM, Body: []byte{0}, Flags: 0, SchemaVersion: 1})
	if err := wr.Finish(); err == nil {
		t.Fatal("expected forbidden error")
	}
}
