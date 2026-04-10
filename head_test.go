package mll

import (
	"bytes"
	"testing"
)

func TestHeadSectionRoundTrip(t *testing.T) {
	strg := NewStringTableBuilder()
	orig := HeadSection{
		Name:          strg.Intern("tiny_embed"),
		Description:   strg.Intern("a tiny embedding model"),
		CreatedUnixMs: 1712806400000,
		Generation:    0,
		Backends:      []uint16{1, 2}, // cuda, metal
		Capabilities:  []uint32{strg.Intern("device_execution")},
		Metadata:      nil,
	}
	var buf bytes.Buffer
	if err := orig.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	decoded, err := ReadHeadSection(buf.Bytes())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if decoded.Name != orig.Name {
		t.Errorf("name: got %d, want %d", decoded.Name, orig.Name)
	}
	if decoded.Description != orig.Description {
		t.Errorf("description: got %d, want %d", decoded.Description, orig.Description)
	}
	if decoded.CreatedUnixMs != orig.CreatedUnixMs {
		t.Errorf("created: got %d, want %d", decoded.CreatedUnixMs, orig.CreatedUnixMs)
	}
	if len(decoded.Backends) != len(orig.Backends) {
		t.Errorf("backend count: got %d, want %d", len(decoded.Backends), len(orig.Backends))
	}
	if len(decoded.Capabilities) != len(orig.Capabilities) {
		t.Errorf("capability count: got %d, want %d", len(decoded.Capabilities), len(orig.Capabilities))
	}
}

func TestHeadSectionWithMetadata(t *testing.T) {
	strg := NewStringTableBuilder()
	orig := HeadSection{
		Name: strg.Intern("test"),
		Metadata: []HeadMetadataEntry{
			{Key: strg.Intern("source"), Kind: HeadValueString, StringIdx: strg.Intern("pretrained.mllb")},
			{Key: strg.Intern("step"), Kind: HeadValueI64, I64: 12345},
			{Key: strg.Intern("temp"), Kind: HeadValueF64, F64: 0.85},
			{Key: strg.Intern("active"), Kind: HeadValueBool, Bool: true},
			{Key: strg.Intern("note"), Kind: HeadValueNull},
		},
	}
	var buf bytes.Buffer
	orig.Write(&buf)
	decoded, err := ReadHeadSection(buf.Bytes())
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(decoded.Metadata) != 5 {
		t.Fatalf("metadata count: got %d", len(decoded.Metadata))
	}
}

func TestHeadDigestBodyZerosWallClockForSealed(t *testing.T) {
	strg := NewStringTableBuilder()
	h1 := HeadSection{
		Name:          strg.Intern("test"),
		CreatedUnixMs: 1712806400000,
		Generation:    0, // sealed
	}
	h2 := HeadSection{
		Name:          strg.Intern("test"),
		CreatedUnixMs: 9999999999999, // different wall clock
		Generation:    0,
	}
	d1 := h1.DigestBody(ProfileSealed)
	d2 := h2.DigestBody(ProfileSealed)
	if !bytes.Equal(d1, d2) {
		t.Fatal("sealed HEAD digest body should be identical across wall clocks")
	}
}

func TestHeadDigestBodyPreservesWallClockForCheckpoint(t *testing.T) {
	strg := NewStringTableBuilder()
	h1 := HeadSection{Name: strg.Intern("test"), CreatedUnixMs: 100}
	h2 := HeadSection{Name: strg.Intern("test"), CreatedUnixMs: 200}
	d1 := h1.DigestBody(ProfileCheckpoint)
	d2 := h2.DigestBody(ProfileCheckpoint)
	if bytes.Equal(d1, d2) {
		t.Fatal("checkpoint HEAD digest body should reflect wall clock difference")
	}
}
