package mll

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

func TestEd25519SignatureVerification(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	sections, keyIdx := buildSignableSections(t)
	var unsigned bytes.Buffer
	unsignedWriter := NewWriter(&unsigned, ProfileSealed, V1_0)
	for _, section := range sections {
		unsignedWriter.AddSection(section)
	}
	if err := unsignedWriter.Finish(); err != nil {
		t.Fatal(err)
	}

	sgnm, err := NewEd25519SgnmSection(keyIdx, priv, unsignedWriter.ContentHash())
	if err != nil {
		t.Fatal(err)
	}
	sgnmBody := mustEncodeSection(t, sgnm)

	var signed bytes.Buffer
	signedWriter := NewWriter(&signed, ProfileSealed, V1_0)
	signedWriter.SetFileFlag(FileFlagHasSignature)
	for _, section := range sections {
		signedWriter.AddSection(section)
	}
	signedWriter.AddSection(SectionInput{Tag: TagSGNM, Body: sgnmBody, SchemaVersion: 1})
	if err := signedWriter.Finish(); err != nil {
		t.Fatal(err)
	}
	if signedWriter.ContentHash() != unsignedWriter.ContentHash() {
		t.Fatal("SGNM changed sealed content hash")
	}

	r, err := ReadBytes(signed.Bytes(), WithDigestVerification())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if err := r.VerifySignature(pub); err != nil {
		t.Fatalf("verify signature: %v", err)
	}

	wrongPub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.VerifySignature(wrongPub); err == nil {
		t.Fatal("expected wrong public key to fail verification")
	}
}

func buildSignableSections(t *testing.T) ([]SectionInput, uint32) {
	t.Helper()
	strg := NewStringTableBuilder()
	nameIdx := strg.Intern("signed")
	entryIdx := strg.Intern("forward")
	keyIdx := strg.Intern("test-key")

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
	}, keyIdx
}
