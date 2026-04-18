package mll

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestValidateCatchesBadReference(t *testing.T) {
	sections := buildBadReferenceSections(t)
	var out bytes.Buffer
	wr := NewWriter(&out, ProfileSealed, V1_0)
	for _, section := range sections {
		wr.AddSection(section)
	}
	if err := wr.Finish(); err != nil {
		t.Fatal(err)
	}

	r, err := ReadBytes(out.Bytes(), WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	err = r.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "PARM[0].type") {
		t.Fatalf("validation error did not mention bad parm ref: %v", err)
	}
}

func buildBadReferenceSections(t *testing.T) []SectionInput {
	t.Helper()
	strg := NewStringTableBuilder()
	nameIdx := strg.Intern("bad_ref")
	paramIdx := strg.Intern("w")
	entryIdx := strg.Intern("forward")

	head := HeadSection{Name: nameIdx}
	var headBuf, strgBuf bytes.Buffer
	if err := head.Write(&headBuf); err != nil {
		t.Fatal(err)
	}
	if err := strg.Write(&strgBuf); err != nil {
		t.Fatal(err)
	}

	dims := NewDimsBuilder()
	types := NewTypeBuilder()
	parm := NewParmBuilder()
	parm.Add(ParmDecl{NameIdx: paramIdx, TypeRef: Ref{Tag: TagTYPE, Index: 9}})
	entr := NewEntrBuilder()
	entr.Add(EntryPoint{NameIdx: entryIdx, Kind: EntryKindPipeline})
	tnsr := NewTnsrBuilder()

	return []SectionInput{
		{Tag: TagHEAD, Body: headBuf.Bytes(), DigestBody: head.DigestBody(ProfileSealed), Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagDIMS, Body: mustEncodeSection(t, dims), Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagTYPE, Body: mustEncodeSection(t, types), SchemaVersion: 1},
		{Tag: TagPARM, Body: mustEncodeSection(t, parm), Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagENTR, Body: mustEncodeSection(t, entr), Flags: SectionFlagRequired, SchemaVersion: 1},
		{Tag: TagTNSR, Body: mustEncodeSection(t, tnsr), Flags: SectionFlagRequired | SectionFlagAligned, SchemaVersion: 1},
	}
}

func mustEncodeSection(t *testing.T, section interface{ Write(io.Writer) error }) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := section.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
