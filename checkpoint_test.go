package mll

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Tests in this file exercise checkpoint generation/save bookkeeping in
// isolation. Checkpoint profile normally requires a full set of sections
// (HEAD, STRG, DIMS, PARM, ENTR, TNSR, OPTM). We pass SkipRequirementCheck
// to focus on checkpoint behavior. Task 9.2 will add a full-sections test.
func TestCheckpointInitialWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ckpt.mllb")
	ckpt, err := NewCheckpoint(path, CheckpointOptions{
		SlackBytes:           4096,
		SkipRequirementCheck: true,
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	strg := NewStringTableBuilder()
	head := HeadSection{Name: strg.Intern("test"), Generation: 1}
	var headBuf, strgBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)
	ckpt.SetSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	if err := ckpt.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	ckpt.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ReadBytes(data)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if r.Profile() != ProfileCheckpoint {
		t.Errorf("profile: got %d", r.Profile())
	}
}

func TestCheckpointGenerationIncrements(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ckpt.mllb")
	ckpt, err := NewCheckpoint(path, CheckpointOptions{
		SlackBytes:           0,
		SkipRequirementCheck: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ckpt.Close()
	strg := NewStringTableBuilder()
	head := HeadSection{Name: strg.Intern("t")}
	var headBuf, strgBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)
	ckpt.SetSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	if err := ckpt.Save(); err != nil {
		t.Fatal(err)
	}
	if ckpt.Generation() != 1 {
		t.Errorf("first save generation: got %d", ckpt.Generation())
	}
	if err := ckpt.Save(); err != nil {
		t.Fatal(err)
	}
	if ckpt.Generation() != 2 {
		t.Errorf("second save generation: got %d", ckpt.Generation())
	}
}

// Regression: Save must bump the on-disk HEAD.Generation, not just the
// in-memory counter.
func TestCheckpointSavePersistsGenerationOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ckpt.mllb")
	ckpt, err := NewCheckpoint(path, CheckpointOptions{SkipRequirementCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	defer ckpt.Close()
	strg := NewStringTableBuilder()
	head := HeadSection{Name: strg.Intern("t"), Generation: 0}
	var headBuf, strgBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)
	ckpt.SetSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})

	readGeneration := func() uint64 {
		r, err := ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		hb, ok := r.Section(TagHEAD)
		if !ok {
			t.Fatal("missing HEAD")
		}
		decoded, err := ReadHeadSection(hb)
		if err != nil {
			t.Fatal(err)
		}
		return decoded.Generation
	}

	if err := ckpt.Save(); err != nil {
		t.Fatal(err)
	}
	if got := readGeneration(); got != 1 {
		t.Errorf("after first save: on-disk generation = %d, want 1", got)
	}
	if err := ckpt.Save(); err != nil {
		t.Fatal(err)
	}
	if got := readGeneration(); got != 2 {
		t.Errorf("after second save: on-disk generation = %d, want 2", got)
	}

	ckpt.Close()
	reopened, err := NewCheckpoint(path, CheckpointOptions{SkipRequirementCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if reopened.Generation() != 2 {
		t.Errorf("reopened generation: got %d, want 2", reopened.Generation())
	}
}

// Regression: a freshly-saved checkpoint should pass WithDigestVerification.
func TestCheckpointFileVerifiesDigests(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ckpt.mllb")
	ckpt, err := NewCheckpoint(path, CheckpointOptions{SkipRequirementCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	defer ckpt.Close()
	strg := NewStringTableBuilder()
	head := HeadSection{Name: strg.Intern("t"), CreatedUnixMs: 42}
	var headBuf, strgBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)
	ckpt.SetSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	if err := ckpt.Save(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(path, WithDigestVerification()); err != nil {
		t.Fatalf("checkpoint digest verification failed: %v", err)
	}
}
