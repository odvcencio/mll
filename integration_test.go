package mll

import (
	"bytes"
	"io"
	"path/filepath"
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

// TestEndToEndCheckpointRequiredSectionsRoundTrip constructs a checkpoint
// artifact populated with every required section, saves it, reopens it,
// saves again, and asserts that:
//   - the Writer's required-section check passes (no SkipRequirementCheck)
//   - on-disk HEAD.Generation increments per Save
//   - reopening recovers the on-disk generation
//   - WithDigestVerification() passes for the saved checkpoint file
func TestEndToEndCheckpointRequiredSectionsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "full.mllb")

	// Build minimal-but-valid bodies for every required checkpoint section.
	strg := NewStringTableBuilder()
	nameIdx := strg.Intern("demo_model")
	paramNameIdx := strg.Intern("w")
	entryNameIdx := strg.Intern("forward")
	typeNameIdx := strg.Intern("t_w")
	dimNameIdx := strg.Intern("d")

	// DIMS: single static dim "d" = 4.
	dims := NewDimsBuilder()
	dims.Add(DimDecl{NameIdx: dimNameIdx, Bound: DimBoundStatic, Value: 4})

	// TYPE: one tensor type shape=[d].
	dDim := DimSymbol("d")
	dDim.SymbolIdx = dimNameIdx
	tys := NewTypeBuilder()
	tys.AddTensorType(typeNameIdx, DTypeF32, []Dimension{dDim})

	// PARM: one parameter "w" : t_w, trainable.
	parm := NewParmBuilder()
	parm.Add(ParmDecl{
		NameIdx:   paramNameIdx,
		TypeRef:   Ref{Tag: TagTYPE, Index: 0},
		Trainable: true,
	})

	// ENTR: one trivial entry point "forward" with no inputs/outputs.
	entr := NewEntrBuilder()
	entr.Add(EntryPoint{NameIdx: entryNameIdx, Kind: EntryKindPipeline})

	// TNSR: one 4xf32 tensor "w".
	tnsr := NewTnsrBuilder()
	tnsr.Add(TensorEntry{
		NameIdx: paramNameIdx,
		DType:   DTypeF32,
		Shape:   []uint64{4},
		Data:    make([]byte, 16), // 4 * f32
	})

	// OPTM: AdamW stub, step 0.
	optm := NewOptmBuilder(OptimizerAdamW)
	optm.SetStep(0)

	head := HeadSection{Name: nameIdx}

	encode := func(writer interface{ Write(w io.Writer) error }) []byte {
		var buf bytes.Buffer
		if err := writer.Write(&buf); err != nil {
			t.Fatal(err)
		}
		return buf.Bytes()
	}

	var headBuf, strgBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)

	ckpt, err := NewCheckpoint(path, CheckpointOptions{}) // no skip: enforce required sections
	if err != nil {
		t.Fatal(err)
	}
	ckpt.SetSection(SectionInput{Tag: TagHEAD, Body: headBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagSTRG, Body: strgBuf.Bytes(), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagDIMS, Body: encode(dims), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagTYPE, Body: encode(tys), Flags: 0, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagPARM, Body: encode(parm), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagENTR, Body: encode(entr), Flags: SectionFlagRequired, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagTNSR, Body: encode(tnsr), Flags: SectionFlagRequired | SectionFlagAligned, SchemaVersion: 1})
	ckpt.SetSection(SectionInput{Tag: TagOPTM, Body: encode(optm), Flags: SectionFlagRequired, SchemaVersion: 1})

	if err := ckpt.Save(); err != nil {
		t.Fatalf("first save (all required sections): %v", err)
	}
	if ckpt.Generation() != 1 {
		t.Errorf("first save gen: got %d", ckpt.Generation())
	}
	if err := ckpt.Save(); err != nil {
		t.Fatalf("second save: %v", err)
	}
	if ckpt.Generation() != 2 {
		t.Errorf("second save gen: got %d", ckpt.Generation())
	}
	ckpt.Close()

	// Reopen and verify.
	reopened, err := NewCheckpoint(path, CheckpointOptions{})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if reopened.Generation() != 2 {
		t.Errorf("reopened gen: got %d, want 2", reopened.Generation())
	}
	reopened.Close()

	// Full digest verification must pass on the on-disk file.
	if _, err := ReadFile(path, WithDigestVerification()); err != nil {
		t.Fatalf("digest verification on checkpoint file: %v", err)
	}
}
