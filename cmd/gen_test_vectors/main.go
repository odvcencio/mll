package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/odvcencio/mll"
)

func main() {
	outDir := "testdata/v1"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatal(err)
	}
	if err := genMinimal(outDir); err != nil {
		log.Fatalf("minimal: %v", err)
	}
	if err := genTinyEmbed(outDir); err != nil {
		log.Fatalf("tiny_embed: %v", err)
	}
	fmt.Println("wrote test vectors")
}

// genMinimal writes a minimal sealed artifact containing only HEAD + STRG.
// Uses WithSkipRequirementCheck to isolate canonicalization testing from
// section-specific machinery.
func genMinimal(dir string) error {
	strg := mll.NewStringTableBuilder()
	strg.Intern("minimal")
	strg.CanonicalizeLexicographic()
	nameIdx, _ := strg.Lookup("minimal")

	head := mll.HeadSection{Name: nameIdx}
	var headBuf bytes.Buffer
	head.Write(&headBuf)

	var strgBuf bytes.Buffer
	strg.Write(&strgBuf)

	sections := []mll.SectionInput{
		{
			Tag:           mll.TagHEAD,
			Body:          headBuf.Bytes(),
			DigestBody:    head.DigestBody(mll.ProfileSealed),
			Flags:         mll.SectionFlagRequired,
			SchemaVersion: 1,
		},
		{Tag: mll.TagSTRG, Body: strgBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
	}

	var out bytes.Buffer
	wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0, mll.WithSkipRequirementCheck())
	for _, s := range sections {
		wr.AddSection(s)
	}
	if err := wr.Finish(); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "minimal.mllb"), out.Bytes(), 0644); err != nil {
		return err
	}
	hash := wr.ContentHash()
	return os.WriteFile(filepath.Join(dir, "minimal.hash"), []byte(hex.EncodeToString(hash[:])+"\n"), 0644)
}

// genTinyEmbed writes a representative sealed inference artifact with all
// required sections for a sealed file: HEAD, STRG, DIMS, TYPE, PARM, ENTR, TNSR.
func genTinyEmbed(dir string) error {
	strg := mll.NewStringTableBuilder()
	strg.Intern("tiny_embed")
	strg.Intern("token_embedding")
	strg.Intern("t_embed")
	strg.Intern("forward")
	strg.Intern("D")
	strg.CanonicalizeLexicographic()

	nameIdx, _ := strg.Lookup("tiny_embed")
	paramNameIdx, _ := strg.Lookup("token_embedding")
	typeNameIdx, _ := strg.Lookup("t_embed")
	entryNameIdx, _ := strg.Lookup("forward")
	dimNameIdx, _ := strg.Lookup("D")

	// DIMS: D = 384 (static).
	dims := mll.NewDimsBuilder()
	dims.Add(mll.DimDecl{NameIdx: dimNameIdx, Bound: mll.DimBoundStatic, Value: 384})

	// TYPE: one tensor type shape=[D].
	dDim := mll.DimSymbol("D")
	dDim.SymbolIdx = dimNameIdx
	tys := mll.NewTypeBuilder()
	tys.AddTensorType(typeNameIdx, mll.DTypeF32, []mll.Dimension{dDim})

	// PARM: one parameter token_embedding : t_embed.
	parm := mll.NewParmBuilder()
	parm.Add(mll.ParmDecl{
		NameIdx:   paramNameIdx,
		TypeRef:   mll.Ref{Tag: mll.TagTYPE, Index: 0},
		Trainable: false,
	})

	// ENTR: one pipeline entry point.
	entr := mll.NewEntrBuilder()
	entr.Add(mll.EntryPoint{NameIdx: entryNameIdx, Kind: mll.EntryKindPipeline})

	// TNSR: one f32 tensor of 384 elements (= 1536 bytes).
	tnsr := mll.NewTnsrBuilder()
	tnsr.Add(mll.TensorEntry{
		NameIdx: paramNameIdx,
		DType:   mll.DTypeF32,
		Shape:   []uint64{384},
		Data:    make([]byte, 384*4),
	})

	head := mll.HeadSection{Name: nameIdx}

	var headBuf, strgBuf, dimsBuf, typeBuf, parmBuf, entrBuf, tnsrBuf bytes.Buffer
	head.Write(&headBuf)
	strg.Write(&strgBuf)
	dims.Write(&dimsBuf)
	tys.Write(&typeBuf)
	parm.Write(&parmBuf)
	entr.Write(&entrBuf)
	tnsr.Write(&tnsrBuf)

	sections := []mll.SectionInput{
		{Tag: mll.TagHEAD, Body: headBuf.Bytes(), DigestBody: head.DigestBody(mll.ProfileSealed), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagSTRG, Body: strgBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagDIMS, Body: dimsBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTYPE, Body: typeBuf.Bytes(), Flags: 0, SchemaVersion: 1},
		{Tag: mll.TagPARM, Body: parmBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagENTR, Body: entrBuf.Bytes(), Flags: mll.SectionFlagRequired, SchemaVersion: 1},
		{Tag: mll.TagTNSR, Body: tnsrBuf.Bytes(), Flags: mll.SectionFlagRequired | mll.SectionFlagAligned, SchemaVersion: 1},
	}

	var out bytes.Buffer
	wr := mll.NewWriter(&out, mll.ProfileSealed, mll.V1_0)
	for _, s := range sections {
		wr.AddSection(s)
	}
	if err := wr.Finish(); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "tiny_embed.mllb"), out.Bytes(), 0644); err != nil {
		return err
	}
	hash := wr.ContentHash()
	return os.WriteFile(filepath.Join(dir, "tiny_embed.hash"), []byte(hex.EncodeToString(hash[:])+"\n"), 0644)
}
