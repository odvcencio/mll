package mll

import (
	"bytes"
	"testing"
)

func TestReaderRoundTrip(t *testing.T) {
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

	rdr, err := ReadBytes(out.Bytes())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if rdr.Profile() != ProfileSealed {
		t.Errorf("profile: got %d", rdr.Profile())
	}
	if rdr.Version() != V1_0 {
		t.Errorf("version: got %+v", rdr.Version())
	}
	if rdr.SectionCount() != 2 {
		t.Errorf("section count: got %d", rdr.SectionCount())
	}
	headBytes, ok := rdr.Section(TagHEAD)
	if !ok {
		t.Fatal("missing HEAD section")
	}
	decodedHead, err := ReadHeadSection(headBytes)
	if err != nil {
		t.Fatalf("decode head: %v", err)
	}
	_ = decodedHead
}

func TestReaderRejectsBadMagic(t *testing.T) {
	bad := make([]byte, HeaderSize)
	copy(bad, []byte{'X', 'X', 'X', 'X'})
	_, err := ReadBytes(bad)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestReaderRejectsDigestMismatch(t *testing.T) {
	strg := NewStringTableBuilder()
	var strgBuf bytes.Buffer
	strg.Intern("anything")
	strg.Write(&strgBuf)
	head := HeadSection{Name: strg.Intern("test")}
	var headBuf bytes.Buffer
	head.Write(&headBuf)
	var out bytes.Buffer
	wr := NewWriter(&out, ProfileSealed, V1_0, WithSkipRequirementCheck())
	wr.AddSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	wr.AddSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	wr.Finish()
	b := out.Bytes()
	b[len(b)-1] ^= 0xFF
	_, err := ReadBytes(b, WithDigestVerification())
	if err == nil {
		t.Fatal("expected digest mismatch error")
	}
}

// Regression: sealed HEAD's digest is computed over DigestBody(profile), which
// zeroes wall-clock fields. The Reader must reproduce that transform when
// verifying, otherwise any sealed artifact with a nonzero created_unix_ms
// fails verification even when uncorrupted.
func TestReaderVerifiesSealedHeadDigestAcrossWallClocks(t *testing.T) {
	build := func(createdUnixMs int64) []byte {
		strg := NewStringTableBuilder()
		head := HeadSection{
			Name:          strg.Intern("test"),
			CreatedUnixMs: createdUnixMs,
		}
		var headBuf, strgBuf bytes.Buffer
		head.Write(&headBuf)
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
			t.Fatal(err)
		}
		return out.Bytes()
	}
	for _, ts := range []int64{0, 1, 1_700_000_000_000} {
		b := build(ts)
		if _, err := ReadBytes(b, WithDigestVerification()); err != nil {
			t.Fatalf("sealed HEAD digest verification failed for ts=%d: %v", ts, err)
		}
	}
}

// Regression: if someone tampers with HEAD fields that are NOT zeroed by
// DigestBody (e.g., Name), verification must still fail.
func TestReaderDetectsSealedHeadTampering(t *testing.T) {
	strg := NewStringTableBuilder()
	head := HeadSection{Name: strg.Intern("test"), CreatedUnixMs: 42}
	var headBuf, strgBuf bytes.Buffer
	head.Write(&headBuf)
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
		t.Fatal(err)
	}
	b := out.Bytes()
	rdr, err := ReadBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range rdr.DirectoryEntries() {
		if e.Tag == TagHEAD {
			b[e.Offset] ^= 0xFF
			break
		}
	}
	if _, err := ReadBytes(b, WithDigestVerification()); err == nil {
		t.Fatal("expected digest mismatch after tampering with HEAD.name_idx")
	}
}
